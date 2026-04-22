package service

import (
	"regexp"
	"sort"
	"strings"

	"github.com/reganputra/skripsi-backend/repository"
	"github.com/reganputra/skripsi-backend/utils"
)

// RecommendResult holds a mentor profile + their similarity score against the query.
type RecommendResult struct {
	UserID          uint               `json:"user_id"`
	Name            string             `json:"name"`
	ProfilePicture  string             `json:"profile_picture"`
	MentorBio       string             `json:"mentor_bio"`
	Skills          string             `json:"skills"`
	Interests       string             `json:"interests"`
	Position        string             `json:"position"`
	CompanyName     string             `json:"company_name"`
	IndustryName    string             `json:"industry_name"`
	YearsExperience int                `json:"years_experience"`
	MatchedKeywords []string           `json:"matched_keywords,omitempty"`
	ScoreBreakdown  map[string]float64 `json:"score_breakdown,omitempty"`
	MentorQuota     int                `json:"mentor_quota"`
	SimilarityScore float64            `json:"similarity_score"`
}

var (
	minYearsPattern = regexp.MustCompile(`(?i)(?:at\s*least|more\s*than|minimum|>=|lebih\s*dari|minimal)?\s*(\d+)\s*(?:\+)?\s*(?:year|years|tahun)`)
	industryPattern = regexp.MustCompile(`(?i)([a-zA-Z0-9\-\+]+)\s+industry`)
)

func extractRecommendationFilters(text string) (cleaned string, minYears int, industry string) {
	cleaned = text
	if m := minYearsPattern.FindStringSubmatch(text); len(m) > 1 {
		for _, ch := range m[1] {
			minYears = (minYears * 10) + int(ch-'0')
		}
		cleaned = minYearsPattern.ReplaceAllString(cleaned, " ")
	}
	if m := industryPattern.FindStringSubmatch(text); len(m) > 1 {
		industry = strings.ToLower(strings.TrimSpace(m[1]))
		cleaned = industryPattern.ReplaceAllString(cleaned, " ")
	}
	return strings.TrimSpace(cleaned), minYears, industry
}

func toSet(tokens []string) map[string]struct{} {
	set := make(map[string]struct{}, len(tokens))
	for _, t := range tokens {
		if t == "" {
			continue
		}
		set[t] = struct{}{}
	}
	return set
}

func overlapKeywords(studentSet, mentorSet map[string]struct{}) ([]string, float64) {
	if len(studentSet) == 0 {
		return []string{}, 0
	}
	matches := make([]string, 0)
	for token := range studentSet {
		if _, ok := mentorSet[token]; ok {
			matches = append(matches, token)
		}
	}
	sort.Strings(matches)
	if len(matches) > 8 {
		matches = matches[:8]
	}
	return matches, float64(len(matches)) / float64(len(studentSet))
}

// RecommendationService computes mentor recommendations using TF-IDF + Cosine Similarity.
type RecommendationService interface {
	// RecommendMentors ranks available mentors against a student's query text.
	// studentText is either the student's skills+interests (auto mode) or a custom query.
	// topN limits the number of results (0 = return all).
	RecommendMentors(studentText string, topN int) ([]RecommendResult, error)
}

type recommendationService struct {
	mentorRepo repository.MentorRepository
}

func NewRecommendationService(mentorRepo repository.MentorRepository) RecommendationService {
	return &recommendationService{mentorRepo: mentorRepo}
}

func (s *recommendationService) RecommendMentors(studentText string, topN int) ([]RecommendResult, error) {
	cleanedQuery, minYears, industry := extractRecommendationFilters(studentText)
	if cleanedQuery == "" {
		cleanedQuery = studentText
	}

	// ── Step 1: Load all available mentor documents ───────────────────────────
	docs, err := s.mentorRepo.FindAllMentorDocs()
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return []RecommendResult{}, nil
	}

	// ── Step 2: Build combined text per mentor ────────────────────────────────
	// text = skills + interests + mentor_bio + position + company + industry
	filtered := make([]repository.MentorDoc, 0, len(docs))
	for _, d := range docs {
		if minYears > 0 && d.YearsExperience < minYears {
			continue
		}
		if industry != "" && !strings.Contains(strings.ToLower(d.IndustryName), industry) {
			continue
		}
		filtered = append(filtered, d)
	}
	if len(filtered) == 0 {
		return []RecommendResult{}, nil
	}

	mentorTexts := make([]string, len(filtered))
	mentorTokenSets := make([]map[string]struct{}, len(filtered))
	for i, d := range filtered {
		parts := []string{d.Skills, d.Interests, d.MentorBio, d.Position, d.CompanyName, d.IndustryName, d.ExperienceText}
		mentorTexts[i] = strings.Join(parts, " ")
		mentorTokenSets[i] = toSet(utils.Tokenize(mentorTexts[i]))
	}

	// ── Step 3: Tokenize corpus (mentor docs + student query) ─────────────────
	// Last element in corpus = student query vector
	corpus := make([][]string, len(mentorTexts)+1)
	for i, txt := range mentorTexts {
		corpus[i] = utils.Tokenize(txt)
	}
	studentTokens := utils.Tokenize(cleanedQuery)
	corpus[len(corpus)-1] = studentTokens
	studentTokenSet := toSet(studentTokens)

	// ── Step 4: Build TF-IDF matrix ───────────────────────────────────────────
	vectors := utils.BuildTFIDF(corpus)
	studentVec := vectors[len(vectors)-1]

	// ── Step 5: Compute cosine similarity for each mentor ─────────────────────
	results := make([]RecommendResult, len(filtered))
	for i, doc := range filtered {
		textScore := utils.CosineSimilarity(studentVec, vectors[i])
		matched, overlapScore := overlapKeywords(studentTokenSet, mentorTokenSets[i])

		experienceScore := 0.0
		if minYears > 0 {
			experienceScore = float64(doc.YearsExperience) / float64(minYears)
			if experienceScore > 1 {
				experienceScore = 1
			}
		}

		industryScore := 0.0
		if industry != "" && strings.Contains(strings.ToLower(doc.IndustryName), industry) {
			industryScore = 1.0
		}

		finalScore := (0.55 * textScore) + (0.20 * overlapScore) + (0.15 * experienceScore) + (0.10 * industryScore)
		if minYears == 0 && industry == "" {
			finalScore = (0.75 * textScore) + (0.25 * overlapScore)
		}

		results[i] = RecommendResult{
			UserID:          doc.UserID,
			Name:            doc.Name,
			ProfilePicture:  doc.ProfilePicture,
			MentorBio:       doc.MentorBio,
			Skills:          doc.Skills,
			Interests:       doc.Interests,
			Position:        doc.Position,
			CompanyName:     doc.CompanyName,
			IndustryName:    doc.IndustryName,
			YearsExperience: doc.YearsExperience,
			MatchedKeywords: matched,
			ScoreBreakdown: map[string]float64{
				"text_similarity": textScore,
				"keyword_overlap": overlapScore,
				"experience_fit":  experienceScore,
				"industry_fit":    industryScore,
			},
			MentorQuota:     doc.MentorQuota,
			SimilarityScore: finalScore,
		}
	}

	// ── Step 6: Sort by similarity score DESC ─────────────────────────────────
	sort.Slice(results, func(i, j int) bool {
		return results[i].SimilarityScore > results[j].SimilarityScore
	})

	// ── Step 7: Trim to topN ──────────────────────────────────────────────────
	if topN > 0 && topN < len(results) {
		results = results[:topN]
	}

	return results, nil
}
