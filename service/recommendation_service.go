package service

import (
	"math"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

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
	// #6: Score on the full match count, then cap the display slice separately.
	// Previously the cap was applied before scoring, understating overlap for rich queries.
	scoreCount := len(matches)
	if len(matches) > 8 {
		matches = matches[:8]
	}
	return matches, float64(scoreCount) / float64(len(studentSet))
}

// RecommendationService computes mentor recommendations using TF-IDF + Cosine Similarity.
type RecommendationService interface {
	RecommendMentors(studentText string, topN int) ([]RecommendResult, error)
}

// corpusCacheSnapshot is a consistent point-in-time read of all corpus cache fields.
type corpusCacheSnapshot struct {
	docs           []repository.MentorDoc
	texts          []string
	tokenSets      []map[string]struct{}
	idf            map[string]float64
	normalizedVecs []map[string]float64
}

// corpusCache holds pre-built, pre-normalized mentor vectors.
// Invalidated after cacheTTL; thread-safe via RWMutex.
type corpusCache struct {
	mu             sync.RWMutex
	docs           []repository.MentorDoc
	tokenSets      []map[string]struct{}
	texts          []string
	idf            map[string]float64       // corpus-level IDF (mentor docs only)
	normalizedVecs []map[string]float64     // L2-normalized TF-IDF vectors per mentor
	loadedAt       time.Time
}

const cacheTTL = 5 * time.Minute

type recommendationService struct {
	mentorRepo repository.MentorRepository
	cache      corpusCache
}

func NewRecommendationService(mentorRepo repository.MentorRepository) RecommendationService {
	return &recommendationService{mentorRepo: mentorRepo}
}

// dotProduct computes the dot product of two sparse vectors.
// When both vectors are L2-normalized, dot(a, b) == CosineSimilarity(a, b) — no sqrt needed.
func dotProduct(a, b map[string]float64) float64 {
	var sum float64
	for k, va := range a {
		sum += va * b[k]
	}
	return sum
}

// loadCorpus returns a consistent cache snapshot (refreshed if stale).
// On a cache miss it: queries DB → tokenizes → builds IDF → pre-normalizes mentor vectors.
func (s *recommendationService) loadCorpus() (corpusCacheSnapshot, error) {
	s.cache.mu.RLock()
	if !s.cache.loadedAt.IsZero() && time.Since(s.cache.loadedAt) < cacheTTL {
		snap := corpusCacheSnapshot{
			docs:           s.cache.docs,
			texts:          s.cache.texts,
			tokenSets:      s.cache.tokenSets,
			idf:            s.cache.idf,
			normalizedVecs: s.cache.normalizedVecs,
		}
		s.cache.mu.RUnlock()
		return snap, nil
	}
	s.cache.mu.RUnlock()

	// Cache miss or expired — reload from DB
	docs, err := s.mentorRepo.FindAllMentorDocs()
	if err != nil {
		return corpusCacheSnapshot{}, err
	}
	texts := make([]string, len(docs))
	sets := make([]map[string]struct{}, len(docs))
	corpusTokens := make([][]string, len(docs))
	for i, d := range docs {
		parts := []string{d.Skills, d.Interests, d.MentorBio, d.Position, d.CompanyName, d.IndustryName, d.ExperienceText}
		texts[i] = strings.Join(parts, " ")
		corpusTokens[i] = utils.Tokenize(texts[i])
		sets[i] = toSet(corpusTokens[i])
	}

	// Build IDF from mentor corpus only (query is NOT included — it shouldn't bias IDF)
	idf := utils.BuildIDF(corpusTokens)

	// Pre-compute and L2-normalize each mentor's TF-IDF vector once
	normVecs := make([]map[string]float64, len(docs))
	for i, tokens := range corpusTokens {
		normVecs[i] = utils.L2Normalize(utils.TFIDFVector(tokens, idf))
	}

	s.cache.mu.Lock()
	s.cache.docs = docs
	s.cache.texts = texts
	s.cache.tokenSets = sets
	s.cache.idf = idf
	s.cache.normalizedVecs = normVecs
	s.cache.loadedAt = time.Now()
	s.cache.mu.Unlock()

	return corpusCacheSnapshot{
		docs:           docs,
		texts:          texts,
		tokenSets:      sets,
		idf:            idf,
		normalizedVecs: normVecs,
	}, nil
}

func (s *recommendationService) RecommendMentors(studentText string, topN int) ([]RecommendResult, error) {
	cleanedQuery, minYears, industry := extractRecommendationFilters(studentText)
	if cleanedQuery == "" {
		cleanedQuery = studentText
	}

	// ── Step 1: Load mentor corpus (from cache or DB) ─────────────────────────
	snap, err := s.loadCorpus()
	if err != nil {
		return nil, err
	}
	if len(snap.docs) == 0 {
		return []RecommendResult{}, nil
	}

	// ── Step 2: Apply hard filters (years / industry) ─────────────────────────
	filtered := make([]repository.MentorDoc, 0, len(snap.docs))
	mentorTokenSets := make([]map[string]struct{}, 0, len(snap.docs))
	filteredNormVecs := make([]map[string]float64, 0, len(snap.docs))
	for i, d := range snap.docs {
		if minYears > 0 && d.YearsExperience < minYears {
			continue
		}
		if industry != "" && !strings.Contains(strings.ToLower(d.IndustryName), industry) {
			continue
		}
		filtered = append(filtered, d)
		mentorTokenSets = append(mentorTokenSets, snap.tokenSets[i])
		filteredNormVecs = append(filteredNormVecs, snap.normalizedVecs[i])
	}
	if len(filtered) == 0 {
		return []RecommendResult{}, nil
	}

	// ── Step 3: Vectorize student query ───────────────────────────────────────
	// IDF comes from the cached mentor corpus. Terms outside mentor vocab get weight 0.
	// Both student and mentor vectors are L2-normalized → dot product == cosine similarity.
	studentTokens := utils.Tokenize(cleanedQuery)
	studentTokenSet := toSet(studentTokens)
	studentVec := utils.L2Normalize(utils.TFIDFVector(studentTokens, snap.idf))

	// ── Step 4: Score each mentor ─────────────────────────────────────────────
	results := make([]RecommendResult, len(filtered))
	for i, doc := range filtered {
		textScore := dotProduct(studentVec, filteredNormVecs[i])
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

		// #3: Dynamic weight calibration — overlap weight grows with query richness.
		// A 1-token query shouldn't be dominated by keyword overlap; a 10-token query
		// deserves more overlap credit because there are enough terms to match meaningfully.
		overlapW := math.Min(0.30, 0.10+0.025*float64(len(studentTokens)))
		textW := 1.0 - overlapW

		finalScore := (textW * textScore) + (overlapW * overlapScore)
		if minYears > 0 || industry != "" {
			// When hard filters are active, redistribute weight: text=0.55, overlap adaptive,
			// experience=0.15, industry=0.10 — overlap gets the remainder.
			overlapW = math.Min(0.20, 0.05+0.015*float64(len(studentTokens)))
			finalScore = (0.55 * textScore) + (overlapW * overlapScore) + (0.15 * experienceScore) + (0.10 * industryScore)
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
