package service

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/reganputra/skripsi-backend/repository"
	"gorm.io/gorm"
)

// ─── DTOs ─────────────────────────────────────────────────────────────────────

// ScoreDistribution groups similarity scores into 5 fixed brackets.
type ScoreDistribution struct {
	Bracket0to2  int `json:"bracket_0_to_2"`  // 0.0 – 0.2
	Bracket2to4  int `json:"bracket_2_to_4"`  // 0.2 – 0.4
	Bracket4to6  int `json:"bracket_4_to_6"`  // 0.4 – 0.6
	Bracket6to8  int `json:"bracket_6_to_8"`  // 0.6 – 0.8
	Bracket8to10 int `json:"bracket_8_to_10"` // 0.8 – 1.0
}

// SimilarityStats provides descriptive statistics over all similarity scores
// collected during a single evaluation run.
type SimilarityStats struct {
	Min          float64           `json:"min"`
	Max          float64           `json:"max"`
	Avg          float64           `json:"avg"`
	Distribution ScoreDistribution `json:"distribution"`
}

// MAPResult is the top-level response from the MAP evaluation endpoint.
type MAPResult struct {
	TopN                   int               `json:"top_n"`
	TotalStudents          int               `json:"total_students"`          // students with ≥1 request in DB
	ValidTestCases         int               `json:"valid_test_cases"`        // students whose relevant mentors are in the rec pool
	MAP                    float64           `json:"map"`
	PrecisionAt1           float64           `json:"precision_at_1"`
	PrecisionAt3           float64           `json:"precision_at_3"`
	PrecisionAt5           float64           `json:"precision_at_5"`
	RecallAt5              float64           `json:"recall_at_5"`
	NdcgAt5                float64           `json:"ndcg_at_5"`
	AvgResponseTimeMS      float64           `json:"avg_response_time_ms"`
	RecommendationCoverage float64           `json:"recommendation_coverage"`
	ZeroScoreCases         int               `json:"zero_score_cases"`
	SimilarityStats        SimilarityStats   `json:"similarity_stats"`
	PerStudent             []StudentAPResult `json:"per_student"`
}

// MRRResult is the top-level response from the MRR evaluation endpoint.
type MRRResult struct {
	TopN                   int               `json:"top_n"`
	TotalStudents          int               `json:"total_students"`          // students with ≥1 request in DB
	ValidTestCases         int               `json:"valid_test_cases"`        // students evaluated
	MRR                    float64           `json:"mrr"`
	AvgResponseTimeMS      float64           `json:"avg_response_time_ms"`
	RecommendationCoverage float64           `json:"recommendation_coverage"`
	ZeroScoreCases         int               `json:"zero_score_cases"`
	SimilarityStats        SimilarityStats   `json:"similarity_stats"`
	PerStudent             []StudentRRResult `json:"per_student"`
}

// StudentRRResult holds the per-student reciprocal rank breakdown.
type StudentRRResult struct {
	StudentID     uint           `json:"student_id"`
	StudentName   string         `json:"student_name"`
	RelevantCount int            `json:"relevant_count"` // total relevant mentors in ground truth
	FirstRank     int            `json:"first_rank"`     // rank of first relevant mentor (0 if not found)
	RR            float64        `json:"rr"`             // reciprocal rank (1.0/FirstRank, 0.0 if not found)
	RankedMentors []RankedMentor `json:"ranked_mentors"` // full ranked list with similarity scores
}

// RankedMentor holds a single recommendation entry used in the evaluation breakdown.
type RankedMentor struct {
	Rank            int     `json:"rank"`
	MentorID        uint    `json:"mentor_id"`
	MentorName      string  `json:"mentor_name"`
	SimilarityScore float64 `json:"similarity_score"`
	IsRelevant      bool    `json:"is_relevant"` // true = mentor is in student's ground truth
}

// StudentAPResult holds the per-student evaluation breakdown.
type StudentAPResult struct {
	StudentID      uint           `json:"student_id"`
	StudentName    string         `json:"student_name"`
	RelevantCount  int            `json:"relevant_count"`   // total relevant mentors in ground truth
	RelevantInPool int            `json:"relevant_in_pool"` // relevant mentors that appear in top-N rec pool
	FoundInTopN    int            `json:"found_in_top_n"`   // relevant mentors retrieved
	AP             float64        `json:"ap"`
	PrecisionAt1   float64        `json:"precision_at_1"`
	PrecisionAt3   float64        `json:"precision_at_3"`
	PrecisionAt5   float64        `json:"precision_at_5"`
	RecallAt5      float64        `json:"recall_at_5"`
	NdcgAt5        float64        `json:"ndcg_at_5"`
	RankedMentors  []RankedMentor `json:"ranked_mentors"` // full ranked list with similarity scores
}

// ─── Interface ─────────────────────────────────────────────────────────────────

type EvaluationService interface {
	// EvaluateCBF runs MAP evaluation against actual mentoring-request ground truth.
	// topN controls how many recommendations to evaluate (e.g., 10 → evaluate top-10).
	// studentIDs is an optional whitelist — if non-empty, only those students are evaluated.
	EvaluateCBF(topN int, studentIDs []uint) (*MAPResult, error)
	EvaluateCBFWithoutLemmatizer(topN int, studentIDs []uint) (*MAPResult, error)
	EvaluateCBFMRR(topN int, studentIDs []uint) (*MRRResult, error)
	EvaluateCBFMRRWithoutLemmatizer(topN int, studentIDs []uint) (*MRRResult, error)
}

// ─── Implementation ────────────────────────────────────────────────────────────

type evaluationService struct {
	requestRepo repository.MentorRequestRepository
	mentorSvc   MentorService
	db          *gorm.DB
}

func NewEvaluationService(
	requestRepo repository.MentorRequestRepository,
	mentorSvc MentorService,
	db *gorm.DB,
) EvaluationService {
	return &evaluationService{
		requestRepo: requestRepo,
		mentorSvc:   mentorSvc,
		db:          db,
	}
}

func (s *evaluationService) EvaluateCBF(topN int, studentIDs []uint) (*MAPResult, error) {
	if topN <= 0 {
		topN = 10
	}

	// Build a whitelist set for fast lookup (empty = evaluate all students)
	whitelist := make(map[uint]struct{}, len(studentIDs))
	for _, id := range studentIDs {
		whitelist[id] = struct{}{}
	}

	// ── Step 1: Fetch ground truth from DB ────────────────────────────────────
	gtRows, err := s.requestRepo.FindGroundTruth()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil ground truth: %w", err)
	}

	// Group by studentID → set of relevant mentorIDs (filtered by whitelist if set)
	gtMap := make(map[uint]map[uint]struct{})
	for _, row := range gtRows {
		// If whitelist is active and this student is not in it, skip
		if len(whitelist) > 0 {
			if _, ok := whitelist[row.StudentID]; !ok {
				continue
			}
		}
		if _, ok := gtMap[row.StudentID]; !ok {
			gtMap[row.StudentID] = make(map[uint]struct{})
		}
		gtMap[row.StudentID][row.MentorID] = struct{}{}
	}

	if len(gtMap) == 0 {
		return &MAPResult{TopN: topN}, nil
	}

	// ── Step 2: Batch-fetch student names ─────────────────────────────────────
	idsForQuery := make([]uint, 0, len(gtMap))
	for sid := range gtMap {
		idsForQuery = append(idsForQuery, sid)
	}
	type nameRow struct {
		ID   uint
		Name string
	}
	var nameRows []nameRow
	s.db.Table("users").Select("id, name").Where("id IN ?", idsForQuery).Scan(&nameRows)
	nameMap := make(map[uint]string, len(nameRows))
	for _, r := range nameRows {
		nameMap[r.ID] = r.Name
	}

	// ── Step 2b: Fetch total active mentor count for Recommendation Coverage ──
	_, totalMentors, _ := s.mentorSvc.GetAvailableMentors(1, 1, "")

	// ── Step 3: Per-student evaluation ───────────────────────────────────────
	var perStudent []StudentAPResult
	var allAPs []float64
	var sumP1, sumP3, sumP5, sumR5, sumNdcg5 float64
	var totalResponseTimeMS float64
	var zeroScoreCases int
	uniqueMentorIDs := make(map[uint]struct{})
	var allSimilarityScores []float64

	for studentID, relevantMentors := range gtMap {
		// Run CBF for this student (empty query → use student's profile) — measure response time
		start := time.Now()
		recs, err := s.mentorSvc.GetRecommendations(studentID, "", topN)
		totalResponseTimeMS += float64(time.Since(start).Milliseconds())
		if err != nil {
			// Student has no profile or other error → skip silently
			continue
		}

		// Track zero-score cases (no recommendations returned)
		if len(recs) == 0 {
			zeroScoreCases++
		}

		// Collect unique mentor IDs and similarity scores for coverage & distribution
		for _, rec := range recs {
			uniqueMentorIDs[rec.UserID] = struct{}{}
			allSimilarityScores = append(allSimilarityScores, rec.SimilarityScore)
		}

		// Extract ranked mentor UserIDs from the recommendation results
		rankedIDs := make([]uint, len(recs))
		for i, r := range recs {
			rankedIDs[i] = r.UserID
		}

		// Check how many of this student's relevant mentors actually appear in
		// the recommendation pool (top-N).
		// NOTE: Students whose relevant mentor is NOT in top-N get AP=0 and are
		// still counted in the MAP denominator. This ensures a fair comparison
		// between with-lemma and without-lemma: lemmatizer helps by surfacing
		// mentors that no-lemma misses, and that difference must be reflected in MAP.
		relevantInPool := 0
		for mid := range relevantMentors {
			for _, rid := range rankedIDs {
				if mid == rid {
					relevantInPool++
					break
				}
			}
		}

		// If relevant mentor not in pool → AP=0 (student still counted in denominator)
		var ap float64
		var hits int
		var p1, p3, p5, r5, ndcg5 float64
		if relevantInPool > 0 {
			ap, hits, p1, p3, p5, r5, ndcg5 = computeMetrics(rankedIDs, relevantMentors, topN)
		}

		name := nameMap[studentID]
		if name == "" {
			name = fmt.Sprintf("Student #%d", studentID)
		}

		// Build ranked mentor list with similarity scores and relevance flags
		rankedMentors := make([]RankedMentor, len(recs))
		for i, rec := range recs {
			_, isRel := relevantMentors[rec.UserID]
			rankedMentors[i] = RankedMentor{
				Rank:            i + 1,
				MentorID:        rec.UserID,
				MentorName:      rec.Name,
				SimilarityScore: rec.SimilarityScore,
				IsRelevant:      isRel,
			}
		}

		perStudent = append(perStudent, StudentAPResult{
			StudentID:      studentID,
			StudentName:    name,
			RelevantCount:  len(relevantMentors),
			RelevantInPool: relevantInPool,
			FoundInTopN:    hits,
			AP:             ap,
			PrecisionAt1:   p1,
			PrecisionAt3:   p3,
			PrecisionAt5:   p5,
			RecallAt5:      r5,
			NdcgAt5:        ndcg5,
			RankedMentors:  rankedMentors,
		})
		allAPs = append(allAPs, ap)
		sumP1 += p1
		sumP3 += p3
		sumP5 += p5
		sumR5 += r5
		sumNdcg5 += ndcg5
	}

	// Sort per-student results by AP descending for readability
	sort.Slice(perStudent, func(i, j int) bool {
		return perStudent[i].AP > perStudent[j].AP
	})

	// ── Step 4: Aggregate ────────────────────────────────────────────────────
	n := float64(len(allAPs))
	mapScore, avgP1, avgP3, avgP5, avgR5, avgNdcg5 := 0.0, 0.0, 0.0, 0.0, 0.0, 0.0
	if n > 0 {
		for _, ap := range allAPs {
			mapScore += ap
		}
		mapScore /= n
		avgP1 = sumP1 / n
		avgP3 = sumP3 / n
		avgP5 = sumP5 / n
		avgR5 = sumR5 / n
		avgNdcg5 = sumNdcg5 / n
	}

	// ── Step 5: Compute additional metrics ───────────────────────────────────
	avgResponseTimeMS := 0.0
	if n > 0 {
		avgResponseTimeMS = totalResponseTimeMS / n
	}

	recCoverage := 0.0
	if totalMentors > 0 {
		recCoverage = float64(len(uniqueMentorIDs)) / float64(totalMentors)
	}

	return &MAPResult{
		TopN:                   topN,
		TotalStudents:          len(gtMap),
		ValidTestCases:         len(allAPs),
		MAP:                    mapScore,
		PrecisionAt1:           avgP1,
		PrecisionAt3:           avgP3,
		PrecisionAt5:           avgP5,
		RecallAt5:              avgR5,
		NdcgAt5:                avgNdcg5,
		AvgResponseTimeMS:      avgResponseTimeMS,
		RecommendationCoverage: recCoverage,
		ZeroScoreCases:         zeroScoreCases,
		SimilarityStats:        computeSimilarityStats(allSimilarityScores),
		PerStudent:             perStudent,
	}, nil
}

func (s *evaluationService) EvaluateCBFWithoutLemmatizer(topN int, studentIDs []uint) (*MAPResult, error) {
	if topN <= 0 {
		topN = 10
	}

	whitelist := make(map[uint]struct{}, len(studentIDs))
	for _, id := range studentIDs {
		whitelist[id] = struct{}{}
	}

	gtRows, err := s.requestRepo.FindGroundTruth()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil ground truth: %w", err)
	}

	gtMap := make(map[uint]map[uint]struct{})
	for _, row := range gtRows {
		if len(whitelist) > 0 {
			if _, ok := whitelist[row.StudentID]; !ok {
				continue
			}
		}
		if _, ok := gtMap[row.StudentID]; !ok {
			gtMap[row.StudentID] = make(map[uint]struct{})
		}
		gtMap[row.StudentID][row.MentorID] = struct{}{}
	}

	if len(gtMap) == 0 {
		return &MAPResult{TopN: topN}, nil
	}

	idsForQuery := make([]uint, 0, len(gtMap))
	for sid := range gtMap {
		idsForQuery = append(idsForQuery, sid)
	}
	type nameRow struct {
		ID   uint
		Name string
	}
	var nameRows []nameRow
	s.db.Table("users").Select("id, name").Where("id IN ?", idsForQuery).Scan(&nameRows)
	nameMap := make(map[uint]string, len(nameRows))
	for _, r := range nameRows {
		nameMap[r.ID] = r.Name
	}

	_, totalMentors, _ := s.mentorSvc.GetAvailableMentors(1, 1, "")

	var perStudent []StudentAPResult
	var allAPs []float64
	var sumP1, sumP3, sumP5, sumR5, sumNdcg5 float64
	var totalResponseTimeMS float64
	var zeroScoreCases int
	uniqueMentorIDs := make(map[uint]struct{})
	var allSimilarityScores []float64

	for studentID, relevantMentors := range gtMap {
		start := time.Now()
		recs, err := s.mentorSvc.GetRecommendationsWithoutLemmatizer(studentID, "", topN)
		totalResponseTimeMS += float64(time.Since(start).Milliseconds())
		if err != nil {
			continue
		}

		if len(recs) == 0 {
			zeroScoreCases++
		}

		for _, rec := range recs {
			uniqueMentorIDs[rec.UserID] = struct{}{}
			allSimilarityScores = append(allSimilarityScores, rec.SimilarityScore)
		}

		rankedIDs := make([]uint, len(recs))
		for i, r := range recs {
			rankedIDs[i] = r.UserID
		}

		relevantInPool := 0
		for mid := range relevantMentors {
			for _, rid := range rankedIDs {
				if mid == rid {
					relevantInPool++
					break
				}
			}
		}
		// If relevant mentor not in pool → AP=0 (student still counted in denominator)
		var ap float64
		var hits int
		var p1, p3, p5, r5, ndcg5 float64
		if relevantInPool > 0 {
			ap, hits, p1, p3, p5, r5, ndcg5 = computeMetrics(rankedIDs, relevantMentors, topN)
		}

		name := nameMap[studentID]
		if name == "" {
			name = fmt.Sprintf("Student #%d", studentID)
		}

		rankedMentors := make([]RankedMentor, len(recs))
		for i, rec := range recs {
			_, isRel := relevantMentors[rec.UserID]
			rankedMentors[i] = RankedMentor{
				Rank:            i + 1,
				MentorID:        rec.UserID,
				MentorName:      rec.Name,
				SimilarityScore: rec.SimilarityScore,
				IsRelevant:      isRel,
			}
		}

		perStudent = append(perStudent, StudentAPResult{
			StudentID:      studentID,
			StudentName:    name,
			RelevantCount:  len(relevantMentors),
			RelevantInPool: relevantInPool,
			FoundInTopN:    hits,
			AP:             ap,
			PrecisionAt1:   p1,
			PrecisionAt3:   p3,
			PrecisionAt5:   p5,
			RecallAt5:      r5,
			NdcgAt5:        ndcg5,
			RankedMentors:  rankedMentors,
		})
		allAPs = append(allAPs, ap)
		sumP1 += p1
		sumP3 += p3
		sumP5 += p5
		sumR5 += r5
		sumNdcg5 += ndcg5
	}

	sort.Slice(perStudent, func(i, j int) bool {
		return perStudent[i].AP > perStudent[j].AP
	})

	n := float64(len(allAPs))
	mapScore, avgP1, avgP3, avgP5, avgR5, avgNdcg5 := 0.0, 0.0, 0.0, 0.0, 0.0, 0.0
	if n > 0 {
		for _, ap := range allAPs {
			mapScore += ap
		}
		mapScore /= n
		avgP1 = sumP1 / n
		avgP3 = sumP3 / n
		avgP5 = sumP5 / n
		avgR5 = sumR5 / n
		avgNdcg5 = sumNdcg5 / n
	}

	avgResponseTimeMS := 0.0
	if n > 0 {
		avgResponseTimeMS = totalResponseTimeMS / n
	}

	recCoverage := 0.0
	if totalMentors > 0 {
		recCoverage = float64(len(uniqueMentorIDs)) / float64(totalMentors)
	}

	return &MAPResult{
		TopN:                   topN,
		TotalStudents:          len(gtMap),
		ValidTestCases:         len(allAPs),
		MAP:                    mapScore,
		PrecisionAt1:           avgP1,
		PrecisionAt3:           avgP3,
		PrecisionAt5:           avgP5,
		RecallAt5:              avgR5,
		NdcgAt5:                avgNdcg5,
		AvgResponseTimeMS:      avgResponseTimeMS,
		RecommendationCoverage: recCoverage,
		ZeroScoreCases:         zeroScoreCases,
		SimilarityStats:        computeSimilarityStats(allSimilarityScores),
		PerStudent:             perStudent,
	}, nil
}

func (s *evaluationService) EvaluateCBFMRR(topN int, studentIDs []uint) (*MRRResult, error) {
	if topN <= 0 {
		topN = 10
	}

	whitelist := make(map[uint]struct{}, len(studentIDs))
	for _, id := range studentIDs {
		whitelist[id] = struct{}{}
	}

	gtRows, err := s.requestRepo.FindGroundTruth()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil ground truth: %w", err)
	}

	gtMap := make(map[uint]map[uint]struct{})
	for _, row := range gtRows {
		if len(whitelist) > 0 {
			if _, ok := whitelist[row.StudentID]; !ok {
				continue
			}
		}
		if _, ok := gtMap[row.StudentID]; !ok {
			gtMap[row.StudentID] = make(map[uint]struct{})
		}
		gtMap[row.StudentID][row.MentorID] = struct{}{}
	}

	if len(gtMap) == 0 {
		return &MRRResult{TopN: topN}, nil
	}

	idsForQuery := make([]uint, 0, len(gtMap))
	for sid := range gtMap {
		idsForQuery = append(idsForQuery, sid)
	}
	type nameRow struct {
		ID   uint
		Name string
	}
	var nameRows []nameRow
	s.db.Table("users").Select("id, name").Where("id IN ?", idsForQuery).Scan(&nameRows)
	nameMap := make(map[uint]string, len(nameRows))
	for _, r := range nameRows {
		nameMap[r.ID] = r.Name
	}

	_, totalMentors, _ := s.mentorSvc.GetAvailableMentors(1, 1, "")

	var perStudent []StudentRRResult
	var allRRs []float64
	var totalResponseTimeMS float64
	var zeroScoreCases int
	uniqueMentorIDs := make(map[uint]struct{})
	var allSimilarityScores []float64

	for studentID, relevantMentors := range gtMap {
		start := time.Now()
		recs, err := s.mentorSvc.GetRecommendations(studentID, "", topN)
		totalResponseTimeMS += float64(time.Since(start).Milliseconds())
		if err != nil {
			continue
		}

		if len(recs) == 0 {
			zeroScoreCases++
		}

		for _, rec := range recs {
			uniqueMentorIDs[rec.UserID] = struct{}{}
			allSimilarityScores = append(allSimilarityScores, rec.SimilarityScore)
		}

		rankedIDs := make([]uint, len(recs))
		for i, r := range recs {
			rankedIDs[i] = r.UserID
		}

		firstRank := 0
		rr := 0.0
		for i, rid := range rankedIDs {
			if _, ok := relevantMentors[rid]; ok {
				firstRank = i + 1
				rr = 1.0 / float64(firstRank)
				break
			}
		}

		name := nameMap[studentID]
		if name == "" {
			name = fmt.Sprintf("Student #%d", studentID)
		}

		rankedMentors := make([]RankedMentor, len(recs))
		for i, rec := range recs {
			_, isRel := relevantMentors[rec.UserID]
			rankedMentors[i] = RankedMentor{
				Rank:            i + 1,
				MentorID:        rec.UserID,
				MentorName:      rec.Name,
				SimilarityScore: rec.SimilarityScore,
				IsRelevant:      isRel,
			}
		}

		perStudent = append(perStudent, StudentRRResult{
			StudentID:     studentID,
			StudentName:   name,
			RelevantCount: len(relevantMentors),
			FirstRank:     firstRank,
			RR:            rr,
			RankedMentors: rankedMentors,
		})
		allRRs = append(allRRs, rr)
	}

	sort.Slice(perStudent, func(i, j int) bool {
		return perStudent[i].RR > perStudent[j].RR
	})

	n := float64(len(allRRs))
	mrrScore := 0.0
	if n > 0 {
		for _, val := range allRRs {
			mrrScore += val
		}
		mrrScore /= n
	}

	avgResponseTimeMS := 0.0
	if n > 0 {
		avgResponseTimeMS = totalResponseTimeMS / n
	}

	recCoverage := 0.0
	if totalMentors > 0 {
		recCoverage = float64(len(uniqueMentorIDs)) / float64(totalMentors)
	}

	return &MRRResult{
		TopN:                   topN,
		TotalStudents:          len(gtMap),
		ValidTestCases:         len(allRRs),
		MRR:                    mrrScore,
		AvgResponseTimeMS:      avgResponseTimeMS,
		RecommendationCoverage: recCoverage,
		ZeroScoreCases:         zeroScoreCases,
		SimilarityStats:        computeSimilarityStats(allSimilarityScores),
		PerStudent:             perStudent,
	}, nil
}

func (s *evaluationService) EvaluateCBFMRRWithoutLemmatizer(topN int, studentIDs []uint) (*MRRResult, error) {
	if topN <= 0 {
		topN = 10
	}

	whitelist := make(map[uint]struct{}, len(studentIDs))
	for _, id := range studentIDs {
		whitelist[id] = struct{}{}
	}

	gtRows, err := s.requestRepo.FindGroundTruth()
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil ground truth: %w", err)
	}

	gtMap := make(map[uint]map[uint]struct{})
	for _, row := range gtRows {
		if len(whitelist) > 0 {
			if _, ok := whitelist[row.StudentID]; !ok {
				continue
			}
		}
		if _, ok := gtMap[row.StudentID]; !ok {
			gtMap[row.StudentID] = make(map[uint]struct{})
		}
		gtMap[row.StudentID][row.MentorID] = struct{}{}
	}

	if len(gtMap) == 0 {
		return &MRRResult{TopN: topN}, nil
	}

	idsForQuery := make([]uint, 0, len(gtMap))
	for sid := range gtMap {
		idsForQuery = append(idsForQuery, sid)
	}
	type nameRow struct {
		ID   uint
		Name string
	}
	var nameRows []nameRow
	s.db.Table("users").Select("id, name").Where("id IN ?", idsForQuery).Scan(&nameRows)
	nameMap := make(map[uint]string, len(nameRows))
	for _, r := range nameRows {
		nameMap[r.ID] = r.Name
	}

	_, totalMentors, _ := s.mentorSvc.GetAvailableMentors(1, 1, "")

	var perStudent []StudentRRResult
	var allRRs []float64
	var totalResponseTimeMS float64
	var zeroScoreCases int
	uniqueMentorIDs := make(map[uint]struct{})
	var allSimilarityScores []float64

	for studentID, relevantMentors := range gtMap {
		start := time.Now()
		recs, err := s.mentorSvc.GetRecommendationsWithoutLemmatizer(studentID, "", topN)
		totalResponseTimeMS += float64(time.Since(start).Milliseconds())
		if err != nil {
			continue
		}

		if len(recs) == 0 {
			zeroScoreCases++
		}

		for _, rec := range recs {
			uniqueMentorIDs[rec.UserID] = struct{}{}
			allSimilarityScores = append(allSimilarityScores, rec.SimilarityScore)
		}

		rankedIDs := make([]uint, len(recs))
		for i, r := range recs {
			rankedIDs[i] = r.UserID
		}

		firstRank := 0
		rr := 0.0
		for i, rid := range rankedIDs {
			if _, ok := relevantMentors[rid]; ok {
				firstRank = i + 1
				rr = 1.0 / float64(firstRank)
				break
			}
		}

		name := nameMap[studentID]
		if name == "" {
			name = fmt.Sprintf("Student #%d", studentID)
		}

		rankedMentors := make([]RankedMentor, len(recs))
		for i, rec := range recs {
			_, isRel := relevantMentors[rec.UserID]
			rankedMentors[i] = RankedMentor{
				Rank:            i + 1,
				MentorID:        rec.UserID,
				MentorName:      rec.Name,
				SimilarityScore: rec.SimilarityScore,
				IsRelevant:      isRel,
			}
		}

		perStudent = append(perStudent, StudentRRResult{
			StudentID:     studentID,
			StudentName:   name,
			RelevantCount: len(relevantMentors),
			FirstRank:     firstRank,
			RR:            rr,
			RankedMentors: rankedMentors,
		})
		allRRs = append(allRRs, rr)
	}

	sort.Slice(perStudent, func(i, j int) bool {
		return perStudent[i].RR > perStudent[j].RR
	})

	n := float64(len(allRRs))
	mrrScore := 0.0
	if n > 0 {
		for _, val := range allRRs {
			mrrScore += val
		}
		mrrScore /= n
	}

	avgResponseTimeMS := 0.0
	if n > 0 {
		avgResponseTimeMS = totalResponseTimeMS / n
	}

	recCoverage := 0.0
	if totalMentors > 0 {
		recCoverage = float64(len(uniqueMentorIDs)) / float64(totalMentors)
	}

	return &MRRResult{
		TopN:                   topN,
		TotalStudents:          len(gtMap),
		ValidTestCases:         len(allRRs),
		MRR:                    mrrScore,
		AvgResponseTimeMS:      avgResponseTimeMS,
		RecommendationCoverage: recCoverage,
		ZeroScoreCases:         zeroScoreCases,
		SimilarityStats:        computeSimilarityStats(allSimilarityScores),
		PerStudent:             perStudent,
	}, nil
}

// computeSimilarityStats aggregates descriptive statistics from a slice of similarity scores.
func computeSimilarityStats(scores []float64) SimilarityStats {
	if len(scores) == 0 {
		return SimilarityStats{}
	}
	minVal := scores[0]
	maxVal := scores[0]
	sum := 0.0
	var dist ScoreDistribution
	for _, s := range scores {
		if s < minVal {
			minVal = s
		}
		if s > maxVal {
			maxVal = s
		}
		sum += s
		switch {
		case s < 0.2:
			dist.Bracket0to2++
		case s < 0.4:
			dist.Bracket2to4++
		case s < 0.6:
			dist.Bracket4to6++
		case s < 0.8:
			dist.Bracket6to8++
		default:
			dist.Bracket8to10++
		}
	}
	return SimilarityStats{
		Min:          minVal,
		Max:          maxVal,
		Avg:          sum / float64(len(scores)),
		Distribution: dist,
	}
}

// computeMetrics calculates all evaluation metrics for a single student query.
//
//   - rankedIDs — ordered list of mentor UserIDs returned by CBF (rank 1 = index 0)
//   - relevant  — set of mentor UserIDs from the student's actual ground truth
//   - topN      — cap; only positions 1…topN are considered
//
// Returns ap, hits, precision at 1/3/5, recall@5, and ndcg@5.
func computeMetrics(rankedIDs []uint, relevant map[uint]struct{}, topN int) (ap float64, hits int, p1, p3, p5, r5, ndcg5 float64) {
	capVal := topN
	if len(rankedIDs) < capVal {
		capVal = len(rankedIDs)
	}
	sumPrecision := 0.0
	hitsAt5 := 0
	for i := 0; i < capVal; i++ {
		rank := i + 1
		if _, ok := relevant[rankedIDs[i]]; ok {
			hits++
			sumPrecision += float64(hits) / float64(rank)
			if rank <= 5 {
				hitsAt5++
			}
		}
		switch rank {
		case 1:
			p1 = float64(hits) / 1.0
		case 3:
			p3 = float64(hits) / 3.0
		case 5:
			p5 = float64(hits) / 5.0
		}
	}
	if len(relevant) > 0 {
		ap = sumPrecision / float64(len(relevant))
		// Recall@5 = dokumen relevan yang ditemukan di top-5 / total relevan
		r5 = float64(hitsAt5) / float64(len(relevant))
	}
	ndcg5 = calculateNDCG5(rankedIDs, relevant)
	return
}

// calculateNDCG5 menghitung Normalized Discounted Cumulative Gain pada posisi ke-5.
// Relevansi bersifat biner: 1 jika mentor ada di ground truth, 0 jika tidak.
func calculateNDCG5(rankedIDs []uint, relevant map[uint]struct{}) float64 {
	k := 5
	if len(rankedIDs) < k {
		k = len(rankedIDs)
	}

	// DCG@5
	dcg := 0.0
	for i := 0; i < k; i++ {
		if _, ok := relevant[rankedIDs[i]]; ok {
			rank := i + 1
			dcg += 1.0 / math.Log2(float64(rank+1))
		}
	}

	if dcg == 0.0 {
		return 0.0
	}

	// IDCG@5: urutan terbaik teoretis — tempatkan semua dokumen relevan di peringkat teratas
	idealHits := len(relevant)
	if idealHits > 5 {
		idealHits = 5
	}
	idcg := 0.0
	for i := 0; i < idealHits; i++ {
		rank := i + 1
		idcg += 1.0 / math.Log2(float64(rank+1))
	}

	if idcg == 0.0 {
		return 0.0
	}
	return dcg / idcg
}
