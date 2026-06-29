package service

import (
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
	CareerInterests string             `json:"career_interests"`
	Position        string             `json:"position"`
	CompanyName     string             `json:"company_name"`
	IndustryName    string             `json:"industry_name"`
	IndustryType    string             `json:"industry_type"`
	YearsExperience int                `json:"years_experience"`
	ScoreBreakdown  map[string]float64 `json:"score_breakdown,omitempty"`
	MentorQuota     int                `json:"mentor_quota"`
	SimilarityScore float64            `json:"similarity_score"`
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

// deduplicateInterests returns only the items in interests that are NOT already
// present in skills (case-insensitive, comma-separated input).
// This prevents LinkedIn-imported profiles where skills == interests from
// doubling every token and biasing the TF-IDF vector.
func deduplicateInterests(skills, interests string) string {
	// Build a set of skill items (lowercased, trimmed)
	skillSet := make(map[string]struct{})
	for _, item := range strings.Split(skills, ",") {
		norm := strings.ToLower(strings.TrimSpace(item))
		if norm != "" {
			skillSet[norm] = struct{}{}
		}
	}
	// Keep only interest items not already covered by skills
	unique := make([]string, 0)
	for _, item := range strings.Split(interests, ",") {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, exists := skillSet[strings.ToLower(trimmed)]; !exists {
			unique = append(unique, trimmed)
		}
	}
	return strings.Join(unique, ", ")
}
// RecommendationService computes mentor recommendations using TF-IDF + Cosine Similarity.
type RecommendationService interface {
	RecommendMentors(studentText string, topN int) ([]RecommendResult, error)
	RecommendMentorsWithoutLemmatizer(studentText string, topN int) ([]RecommendResult, error)
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

// minScore adalah ambang batas minimum cosine similarity agar seorang mentor
// diikutsertakan dalam hasil rekomendasi. Mentor dengan skor di bawah nilai ini
// dianggap tidak memiliki relevansi semantik yang cukup dengan profil mahasiswa.
const minScore = 0.1

type recommendationService struct {
	mentorRepo    repository.MentorRepository
	cache         corpusCache
	cacheNoLemma  corpusCache
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
		// Deduplikasi: Interests yang sudah ada di Skills dibuang
		// agar tidak ada token yang dihitung dua kali secara redundan
		uniqueInterests := deduplicateInterests(d.Skills, d.Interests)
		uniqueCareerInterests := deduplicateInterests(d.Skills, d.CareerInterests)
		// Skills diulang 2x (field boost) karena merupakan sinyal terkuat.
		// Interests dan CareerInterests yang sudah unik disertakan 1x.
		parts := []string{
			d.Skills, d.Skills,
			uniqueInterests, uniqueCareerInterests,
			d.MentorBio, d.Position, d.CompanyName, d.IndustryName, d.IndustryType, d.ExperienceText,
		}
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

// loadCorpusNoLemma returns a consistent cache snapshot for non-lemmatized corpus.
func (s *recommendationService) loadCorpusNoLemma() (corpusCacheSnapshot, error) {
	s.cacheNoLemma.mu.RLock()
	if !s.cacheNoLemma.loadedAt.IsZero() && time.Since(s.cacheNoLemma.loadedAt) < cacheTTL {
		snap := corpusCacheSnapshot{
			docs:           s.cacheNoLemma.docs,
			texts:          s.cacheNoLemma.texts,
			tokenSets:      s.cacheNoLemma.tokenSets,
			idf:            s.cacheNoLemma.idf,
			normalizedVecs: s.cacheNoLemma.normalizedVecs,
		}
		s.cacheNoLemma.mu.RUnlock()
		return snap, nil
	}
	s.cacheNoLemma.mu.RUnlock()

	// Cache miss or expired — reload from DB
	docs, err := s.mentorRepo.FindAllMentorDocs()
	if err != nil {
		return corpusCacheSnapshot{}, err
	}
	texts := make([]string, len(docs))
	sets := make([]map[string]struct{}, len(docs))
	corpusTokens := make([][]string, len(docs))
	for i, d := range docs {
		uniqueInterests := deduplicateInterests(d.Skills, d.Interests)
		uniqueCareerInterests := deduplicateInterests(d.Skills, d.CareerInterests)
		parts := []string{
			d.Skills, d.Skills,
			uniqueInterests, uniqueCareerInterests,
			d.MentorBio, d.Position, d.CompanyName, d.IndustryName, d.IndustryType, d.ExperienceText,
		}
		texts[i] = strings.Join(parts, " ")
		corpusTokens[i] = utils.TokenizeWithoutLemmatizer(texts[i])
		sets[i] = toSet(corpusTokens[i])
	}

	idf := utils.BuildIDF(corpusTokens)

	normVecs := make([]map[string]float64, len(docs))
	for i, tokens := range corpusTokens {
		normVecs[i] = utils.L2Normalize(utils.TFIDFVector(tokens, idf))
	}

	s.cacheNoLemma.mu.Lock()
	s.cacheNoLemma.docs = docs
	s.cacheNoLemma.texts = texts
	s.cacheNoLemma.tokenSets = sets
	s.cacheNoLemma.idf = idf
	s.cacheNoLemma.normalizedVecs = normVecs
	s.cacheNoLemma.loadedAt = time.Now()
	s.cacheNoLemma.mu.Unlock()

	return corpusCacheSnapshot{
		docs:           docs,
		texts:          texts,
		tokenSets:      sets,
		idf:            idf,
		normalizedVecs: normVecs,
	}, nil
}

func (s *recommendationService) RecommendMentors(studentText string, topN int) ([]RecommendResult, error) {
	cleanedQuery := strings.TrimSpace(studentText)

	// ── Step 1: Load mentor corpus (from cache or DB) ─────────────────────────
	snap, err := s.loadCorpus()
	if err != nil {
		return nil, err
	}
	if len(snap.docs) == 0 {
		return []RecommendResult{}, nil
	}

	// ── Step 2: Vectorize student query ───────────────────────────────────────
	studentTokens := utils.Tokenize(cleanedQuery)
	studentVec := utils.L2Normalize(utils.TFIDFVector(studentTokens, snap.idf))

	// ── Step 3: Score each mentor ─────────────────────────────────────────────
	results := make([]RecommendResult, len(snap.docs))
	for i, doc := range snap.docs {
		textScore := dotProduct(studentVec, snap.normalizedVecs[i])

		results[i] = RecommendResult{
			UserID:          doc.UserID,
			Name:            doc.Name,
			ProfilePicture:  doc.ProfilePicture,
			MentorBio:       doc.MentorBio,
			Skills:          doc.Skills,
			Interests:       doc.Interests,
			CareerInterests: doc.CareerInterests,
			Position:        doc.Position,
			CompanyName:     doc.CompanyName,
			IndustryName:    doc.IndustryName,
			IndustryType:    doc.IndustryType,
			YearsExperience: doc.YearsExperience,
			ScoreBreakdown: map[string]float64{
				"text_similarity": textScore,
			},
			MentorQuota:     doc.MentorQuota,
			SimilarityScore: textScore,
		}
	}

	// Filter mentor dengan skor di bawah threshold minimum
	filtered := results[:0]
	for _, r := range results {
		if r.SimilarityScore >= minScore {
			filtered = append(filtered, r)
		}
	}
	results = filtered

	sort.Slice(results, func(i, j int) bool {
		return results[i].SimilarityScore > results[j].SimilarityScore
	})

	if topN > 0 && topN < len(results) {
		results = results[:topN]
	}

	return results, nil
}

func (s *recommendationService) RecommendMentorsWithoutLemmatizer(studentText string, topN int) ([]RecommendResult, error) {
	cleanedQuery := strings.TrimSpace(studentText)

	snap, err := s.loadCorpusNoLemma()
	if err != nil {
		return nil, err
	}
	if len(snap.docs) == 0 {
		return []RecommendResult{}, nil
	}

	studentTokens := utils.TokenizeWithoutLemmatizer(cleanedQuery)
	studentVec := utils.L2Normalize(utils.TFIDFVector(studentTokens, snap.idf))

	results := make([]RecommendResult, len(snap.docs))
	for i, doc := range snap.docs {
		textScore := dotProduct(studentVec, snap.normalizedVecs[i])

		results[i] = RecommendResult{
			UserID:          doc.UserID,
			Name:            doc.Name,
			ProfilePicture:  doc.ProfilePicture,
			MentorBio:       doc.MentorBio,
			Skills:          doc.Skills,
			Interests:       doc.Interests,
			CareerInterests: doc.CareerInterests,
			Position:        doc.Position,
			CompanyName:     doc.CompanyName,
			IndustryName:    doc.IndustryName,
			IndustryType:    doc.IndustryType,
			YearsExperience: doc.YearsExperience,
			ScoreBreakdown: map[string]float64{
				"text_similarity": textScore,
			},
			MentorQuota:     doc.MentorQuota,
			SimilarityScore: textScore,
		}
	}

	// Filter mentor dengan skor di bawah threshold minimum
	filtered := results[:0]
	for _, r := range results {
		if r.SimilarityScore >= minScore {
			filtered = append(filtered, r)
		}
	}
	results = filtered

	sort.Slice(results, func(i, j int) bool {
		return results[i].SimilarityScore > results[j].SimilarityScore
	})

	if topN > 0 && topN < len(results) {
		results = results[:topN]
	}

	return results, nil
}

