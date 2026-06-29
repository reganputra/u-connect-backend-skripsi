package utils

import (
	"math"
	"regexp"
	"strings"
)

// ─── Text Cleaning ────────────────────────────────────────────────────────────

var nonAlpha = regexp.MustCompile(`[^a-zA-Z\s]`)

// Tokenize menjalankan seluruh alur kerja NLP: konversi huruf besar-kecil → pembersihan → tokenisasi → penghapusan kata pengisi → stemming.
// Input dapat berupa teks yang dipisahkan koma (keterampilan/minat) atau teks bebas.
func Tokenize(text string) []string {
	// Ganti koma/tanda hubung dengan spasi agar "machine-learning" dan "Python, Go" terpisah dengan benar
	text = strings.NewReplacer(",", " ", "-", " ", "_", " ", "/", " ").Replace(text)
	text = strings.ToLower(text)
	text = nonAlpha.ReplaceAllString(text, " ")

	words := strings.Fields(text)
	result := make([]string, 0, len(words))
	for _, w := range words {
		if len(w) < 2 {
			continue
		}
		if idStopwords[w] || enStopwords[w] {
			continue
		}
		var token string
		if base, ok := lemmaDict[w]; ok {
			token = base
		} else {
			token = lemmatize(stem(w))
		}
		if len(token) < 2 {
			continue
		}
		if idStopwords[token] || enStopwords[token] {
			continue
		}
		result = append(result, token)
	}
	return result
}

// TokenizeWithoutLemmatizer menjalankan alur kerja NLP tanpa Bilingual Lemmatizer (dictionary & rules).
func TokenizeWithoutLemmatizer(text string) []string {
	// Ganti koma/tanda hubung dengan spasi agar "machine-learning" dan "Python, Go" terpisah dengan benar
	text = strings.NewReplacer(",", " ", "-", " ", "_", " ", "/", " ").Replace(text)
	text = strings.ToLower(text)
	text = nonAlpha.ReplaceAllString(text, " ")

	words := strings.Fields(text)
	result := make([]string, 0, len(words))
	for _, w := range words {
		if len(w) < 2 {
			continue
		}
		if idStopwords[w] || enStopwords[w] {
			continue
		}
		// Hanya jalankan Stemmer Indonesia biasa (tanpa lemmaDict & lemmaRules)
		token := stem(w)
		if len(token) < 2 {
			continue
		}
		if idStopwords[token] || enStopwords[token] {
			continue
		}
		result = append(result, token)
	}
	return result
}

// ─── Lemmatizer ───────────────────────────────────────────────────────────────

// lemmatize mengurangi token (setelah stemming) ke basis kanonis menggunakan
// pencarian kamus terlebih dahulu, kemudian aturan penghilangan sufiks bahasa Inggris.
// Data kamus (lemmaDict, lemmaRules) ada di utils/nlp_lemma.go.
func lemmatize(word string) string {
	if base, ok := lemmaDict[word]; ok {
		return base
	}
	for _, r := range lemmaRules {
		if strings.HasSuffix(word, r.suffix) {
			candidate := word[:len(word)-len(r.suffix)] + r.replacement
			if len(candidate) >= r.minBase {
				return candidate
			}
		}
	}
	return word
}

// ─── Stemmer (Indonesian prefix/suffix removal) ──────────────────────────────

// Suffixes dicoba mulai dari yang terpanjang terlebih dahulu untuk menghindari penghilangan sebagian.
var idSuffixes = []string{"kan", "an", "nya", "lah", "kah", "pun", "i"}

// Prefixes diurutkan dari yang terpanjang terlebih dahulu untuk menangkap awalan majemuk (menge-, memper-) sebelum awalan yang lebih pendek.
var idPrefixes = []string{
	"menge", "memper", "diper",
	"meny", "men", "mem", "meng", "me",
	"peny", "pen", "pem", "peng", "pe",
	"ber", "ter", "per",
	"di", "ke", "se",
}

func stem(word string) string {
	if len(word) < 4 {
		return word
	}

	// 1. Hapus suffix
	for _, suf := range idSuffixes {
		if strings.HasSuffix(word, suf) && len(word)-len(suf) >= 3 {
			word = word[:len(word)-len(suf)]
			break
		}
	}

	// 2. Hapus prefix
	for _, pre := range idPrefixes {
		if strings.HasPrefix(word, pre) && len(word)-len(pre) >= 3 {
			word = word[len(pre):]
			break
		}
	}

	return word
}

// ─── TF-IDF ───────────────────────────────────────────────────────────────────

// BuildTFIDF menerima kumpulan teks yang telah ditokenisasi dan menghasilkan vektor TF-IDF (satu peta per dokumen).
func BuildTFIDF(corpus [][]string) []map[string]float64 {
	N := len(corpus)
	if N == 0 {
		return nil
	}

	// 1. Hitung Document Frequency (df): jumlah dokumen yang memuat setiap term
	df := make(map[string]int, 512)
	for _, doc := range corpus {
		seen := make(map[string]struct{}, len(doc))
		for _, t := range doc {
			if _, ok := seen[t]; !ok {
				df[t]++
				seen[t] = struct{}{}
			}
		}
	}

	// 2. Bangun vektor TF-IDF per dokumen
	vectors := make([]map[string]float64, N)
	for i, doc := range corpus {
		if len(doc) == 0 {
			vectors[i] = map[string]float64{}
			continue
		}

		// Hitung TF (frekuensi relatif)
		tf := make(map[string]float64, len(doc))
		for _, t := range doc {
			tf[t]++
		}
		totalTerms := float64(len(doc))
		for term := range tf {
			tf[term] = tf[term] / totalTerms // TF(t,d) = freq / total_terms
		}

		vec := make(map[string]float64, len(tf))
		for term, tfScore := range tf {
			idfScore := math.Log(float64(N+1)/float64(df[term]+1)) + 1.0
			vec[term] = tfScore * idfScore
		}
		vectors[i] = vec
	}

	return vectors
}

// BuildIDF menghitung tabel IDF tingkat korpus dari kumpulan dokumen yang telah ditokenisasi sebelumnya.
// Dengan memisahkan IDF dari TF, vektor mentor dapat dibuat terlebih dahulu dan disimpan dalam cache secara terpisah dari kueri apa pun.
func BuildIDF(corpus [][]string) map[string]float64 {
	N := len(corpus)
	if N == 0 {
		return map[string]float64{}
	}
	df := make(map[string]int, 512)
	for _, doc := range corpus {
		seen := make(map[string]struct{}, len(doc))
		for _, t := range doc {
			if _, ok := seen[t]; !ok {
				df[t]++
				seen[t] = struct{}{}
			}
		}
	}
	idf := make(map[string]float64, len(df))
	for term, freq := range df {
		idf[term] = math.Log(float64(N+1)/float64(freq+1)) + 1.0
	}
	return idf
}

// TFIDFVector menghitung vektor TF-IDF untuk satu dokumen menggunakan tabel IDF yang sudah disiapkan sebelumnya.
func TFIDFVector(tokens []string, idf map[string]float64) map[string]float64 {
	if len(tokens) == 0 {
		return map[string]float64{}
	}
	tf := make(map[string]float64, len(tokens))
	for _, t := range tokens {
		tf[t]++
	}
	total := float64(len(tokens))
	vec := make(map[string]float64, len(tf))
	for term, count := range tf {
		if idfScore, ok := idf[term]; ok {
			vec[term] = (count / total) * idfScore
		}
	}
	return vec
}

func L2Normalize(vec map[string]float64) map[string]float64 {
	var norm float64
	for _, v := range vec {
		norm += v * v
	}
	if norm == 0 {
		return vec
	}
	norm = math.Sqrt(norm)
	normalized := make(map[string]float64, len(vec))
	for k, v := range vec {
		normalized[k] = v / norm
	}
	return normalized
}

// CosineSimilarity menghitung kesamaan kosinus antara dua vektor TF-IDF.
// Mengembalikan nilai 0 jika salah satu vektor kosong.
func CosineSimilarity(a, b map[string]float64) float64 {
	var dot, normA, normB float64
	for k, va := range a {
		dot += va * b[k]
		normA += va * va
	}
	for _, vb := range b {
		normB += vb * vb
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
