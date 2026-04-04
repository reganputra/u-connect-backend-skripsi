package service

import (
	"sort"
	"strings"

	"github.com/reganputra/skripsi-backend/repository"
	"github.com/reganputra/skripsi-backend/utils"
)

// RecommendResult holds a mentor profile + their similarity score against the query.
type RecommendResult struct {
	UserID         uint    `json:"user_id"`
	Name           string  `json:"name"`
	ProfilePicture string  `json:"profile_picture"`
	MentorBio      string  `json:"mentor_bio"`
	Skills         string  `json:"skills"`
	Interests      string  `json:"interests"`
	Position       string  `json:"position"`
	CompanyName    string  `json:"company_name"`
	IndustryName   string  `json:"industry_name"`
	MentorQuota    int     `json:"mentor_quota"`
	SimilarityScore float64 `json:"similarity_score"`
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
	mentorTexts := make([]string, len(docs))
	for i, d := range docs {
		parts := []string{d.Skills, d.Interests, d.MentorBio, d.Position, d.CompanyName, d.IndustryName}
		mentorTexts[i] = strings.Join(parts, " ")
	}

	// ── Step 3: Tokenize corpus (mentor docs + student query) ─────────────────
	// Last element in corpus = student query vector
	corpus := make([][]string, len(mentorTexts)+1)
	for i, txt := range mentorTexts {
		corpus[i] = utils.Tokenize(txt)
	}
	corpus[len(corpus)-1] = utils.Tokenize(studentText)

	// ── Step 4: Build TF-IDF matrix ───────────────────────────────────────────
	vectors := utils.BuildTFIDF(corpus)
	studentVec := vectors[len(vectors)-1]

	// ── Step 5: Compute cosine similarity for each mentor ─────────────────────
	results := make([]RecommendResult, len(docs))
	for i, doc := range docs {
		score := utils.CosineSimilarity(studentVec, vectors[i])
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
			MentorQuota:     doc.MentorQuota,
			SimilarityScore: score,
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
