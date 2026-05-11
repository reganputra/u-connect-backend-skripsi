package service

import (
	"fmt"
	"sort"

	"github.com/reganputra/skripsi-backend/repository"
	"gorm.io/gorm"
)

// ─── DTOs ─────────────────────────────────────────────────────────────────────

// MAPResult is the top-level response from the MAP evaluation endpoint.
type MAPResult struct {
	TopN           int               `json:"top_n"`
	TotalStudents  int               `json:"total_students"`  // students with ≥1 request in DB
	ValidTestCases int               `json:"valid_test_cases"` // students whose relevant mentors are in the rec pool
	MAP            float64           `json:"map"`
	PrecisionAt1   float64           `json:"precision_at_1"`
	PrecisionAt3   float64           `json:"precision_at_3"`
	PrecisionAt5   float64           `json:"precision_at_5"`
	PerStudent     []StudentAPResult `json:"per_student"`
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
	RankedMentors  []RankedMentor `json:"ranked_mentors"` // full ranked list with similarity scores
}

// ─── Interface ─────────────────────────────────────────────────────────────────

type EvaluationService interface {
	// EvaluateCBF runs MAP evaluation against actual mentoring-request ground truth.
	// topN controls how many recommendations to evaluate (e.g., 10 → evaluate top-10).
	// studentIDs is an optional whitelist — if non-empty, only those students are evaluated.
	EvaluateCBF(topN int, studentIDs []uint) (*MAPResult, error)
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

	// ── Step 3: Per-student evaluation ───────────────────────────────────────
	var perStudent []StudentAPResult
	var allAPs []float64
	var sumP1, sumP3, sumP5 float64

	for studentID, relevantMentors := range gtMap {
		// Run CBF for this student (empty query → use student's profile)
		recs, err := s.mentorSvc.GetRecommendations(studentID, "", topN)
		if err != nil {
			// Student has no profile or other error → skip silently
			continue
		}

		// Extract ranked mentor UserIDs from the recommendation results
		rankedIDs := make([]uint, len(recs))
		for i, r := range recs {
			rankedIDs[i] = r.UserID
		}

		// Check how many of this student's relevant mentors actually appear in
		// the recommendation pool. If none are retrievable (e.g., mentor has
		// since deregistered), the system can never rank them → skip to avoid
		// unfairly penalising the algorithm.
		relevantInPool := 0
		for mid := range relevantMentors {
			for _, rid := range rankedIDs {
				if mid == rid {
					relevantInPool++
					break
				}
			}
		}
		if relevantInPool == 0 {
			continue
		}

		ap, hits, p1, p3, p5 := computeAP(rankedIDs, relevantMentors, topN)

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
			RankedMentors:  rankedMentors,
		})
		allAPs = append(allAPs, ap)
		sumP1 += p1
		sumP3 += p3
		sumP5 += p5
	}

	// Sort per-student results by AP descending for readability
	sort.Slice(perStudent, func(i, j int) bool {
		return perStudent[i].AP > perStudent[j].AP
	})

	// ── Step 4: Aggregate ────────────────────────────────────────────────────
	mapScore, avgP1, avgP3, avgP5 := 0.0, 0.0, 0.0, 0.0
	if n := float64(len(allAPs)); n > 0 {
		for _, ap := range allAPs {
			mapScore += ap
		}
		mapScore /= n
		avgP1 = sumP1 / n
		avgP3 = sumP3 / n
		avgP5 = sumP5 / n
	}

	return &MAPResult{
		TopN:           topN,
		TotalStudents:  len(gtMap),
		ValidTestCases: len(allAPs),
		MAP:            mapScore,
		PrecisionAt1:   avgP1,
		PrecisionAt3:   avgP3,
		PrecisionAt5:   avgP5,
		PerStudent:     perStudent,
	}, nil
}

// computeAP calculates Average Precision for a single student query.
//
//   - rankedIDs  — ordered list of mentor UserIDs returned by CBF (rank 1 = index 0)
//   - relevant   — set of mentor UserIDs from the student's actual ground truth
//   - topN       — cap; only positions 1…topN are considered
//
// Returns ap, number of hits, and precision at ranks 1, 3, 5.
func computeAP(rankedIDs []uint, relevant map[uint]struct{}, topN int) (ap float64, hits int, p1, p3, p5 float64) {
	cap := topN
	if len(rankedIDs) < cap {
		cap = len(rankedIDs)
	}
	sumPrecision := 0.0
	for i := 0; i < cap; i++ {
		rank := i + 1
		if _, ok := relevant[rankedIDs[i]]; ok {
			hits++
			sumPrecision += float64(hits) / float64(rank)
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
	}
	return
}
