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
	// Ganti koma/tanda hubung dengan spasi agar “machine-learning” dan “Python, Go” terpisah dengan benar
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

// ─── Lemmatizer ───────────────────────────────────────────────────────────────

// lemmatize mengurangi token (setelah stemming) ke basis kanonis menggunakan
// pencarian kamus terlebih dahulu, kemudian aturan penghilangan sufiks bahasa Inggris.
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

type lemmaRule struct {
	suffix      string
	replacement string
	minBase     int
}

// lemmaRules: longest suffix first to avoid partial matches.
var lemmaRules = []lemmaRule{
	{"izations", "ize", 3}, {"ization", "ize", 3},
	{"ations", "", 3}, {"ation", "", 3},
	{"ments", "", 3}, {"ment", "", 3},
	{"nesses", "", 3}, {"ness", "", 3},
	{"ities", "", 3}, {"ity", "", 3},
	{"ives", "", 3}, {"ive", "", 3},
	{"ings", "", 3}, {"ning", "n", 3}, {"pping", "p", 3}, {"ing", "", 3},
	{"ies", "y", 2}, {"ied", "y", 3},
	{"ous", "", 3}, {"ical", "", 3},
	{"eds", "", 3}, {"ed", "", 3},
	{"ers", "", 4}, {"er", "", 4},
	{"als", "", 4}, {"al", "", 4},
	{"s", "", 4}, // plural nouns
}

// lemmaDict maps surface forms (pre-stem English + post-stem Indonesian) to a
// canonical base token so student and mentor texts normalize to the same term.
var lemmaDict = map[string]string{
	// ─── Tech nouns ───
	"analytics": "analytic", "analysis": "analytic", "analyses": "analytic",
	"algorithm": "algorithm", "algorithms": "algorithm",
	"database": "database", "databases": "database",
	"network": "network", "networks": "network",
	"model": "model", "models": "model",
	"system": "system", "systems": "system",
	"framework": "framework", "frameworks": "framework",
	"library": "library", "libraries": "library",
	"query": "query", "queries": "query",
	"technology": "technology", "technologies": "technology",
	"data": "data",
	// ─── English tech gerunds / nominalizations ───
	"learning": "learn", "machine": "machine",
	"computing":  "compute",
	"developing": "develop", "development": "develop", "developments": "develop",
	"building": "build",
	"training": "train",
	"modeling": "model", "modelling": "model",
	"processing":  "process",
	"engineering": "engineer",
	"managing":    "manage", "management": "manage",
	"programming": "program",
	"designing":   "design", "design": "design",
	"testing":   "test",
	"deploying": "deploy", "deployment": "deploy",
	"forecasting": "forecast",
	"analyzing":   "analyze", "analysing": "analyze",
	"visualization": "visualize", "visualizations": "visualize", "visualizing": "visualize",
	"implementing": "implement", "implementation": "implement",
	"optimizing": "optimize", "optimization": "optimize",
	"monitoring": "monitor",
	"prediction": "predict", "predictions": "predict", "predictive": "predict",
	"classification": "classify",
	"clustering":     "cluster",
	"collaboration":  "collaborate", "collaborating": "collaborate",
	// ─── Tech roles ───
	"developers": "developer", "engineers": "engineer",
	"scientists": "scientist", "analysts": "analyst",
	"researchers": "researcher", "architects": "architect",
	// ─── Indonesian tech terms → canonical ───
	"pemrograman":  "program",
	"pengembangan": "develop",
	"visualisasi":  "visualize",
	"analisis":     "analytic",
	"pembelajaran": "learn",
	"pengelolaan":  "manage",
	"perancangan":  "design",
	"pengujian":    "test",
	"penerapan":    "implement",
	"keuangan":     "finance",
	"perbankan":    "bank",
	"jaringan":     "network",
	"keamanan":     "security",
	"kecerdasan":   "intelligence",
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

// ─── Stopwords ────────────────────────────────────────────────────────────────

var idStopwords = map[string]bool{
	"yang": true, "dan": true, "di": true, "ke": true, "dari": true,
	"dengan": true, "untuk": true, "adalah": true, "pada": true, "ini": true,
	"itu": true, "atau": true, "juga": true, "sudah": true, "saya": true,
	"anda": true, "kamu": true, "kami": true, "kita": true, "mereka": true,
	"dia": true, "ia": true, "ada": true, "akan": true, "bisa": true,
	"dapat": true, "bukan": true, "tidak": true, "belum": true, "lagi": true,
	"masih": true, "hanya": true, "serta": true, "tetapi": true, "namun": true,
	"karena": true, "sehingga": true, "agar": true, "supaya": true, "maka": true,
	"oleh": true, "tentang": true, "dalam": true, "antara": true, "setelah": true,
	"sebelum": true, "ketika": true, "bila": true, "jika": true, "apakah": true,
	"bagaimana": true, "mengapa": true, "kapan": true, "dimana": true, "siapa": true,
	"apa": true, "telah": true, "sangat": true, "lebih": true, "tersebut": true,
	"bahwa": true, "hampir": true, "selain": true, "atas": true, "bawah": true,
	"selama": true, "hingga": true, "sampai": true, "saat": true, "waktu": true,
	"tahun": true, "bulan": true, "hari": true, "sebuah": true, "setiap": true,
	"semua": true, "beberapa": true, "jadi": true, "kalau": true, "kemudian": true,
	"lalu": true, "sejak": true, "sesudah": true, "berbagai": true, "seperti": true,
	"pula": true, "sini": true, "sana": true, "situ": true, "mana": true,
	"pun": true, "lah": true, "kah": true, "nya": true, "ku": true, "mu": true,
	"harus": true, "perlu": true, "boleh": true, "mau": true, "ingin": true,
	"sudahlah": true, "baik": true, "buruk": true, "besar": true, "kecil": true,
	"sama": true, "baru": true, "lama": true, "jauh": true, "dekat": true,
	"awal": true, "akhir": true, "mulai": true, "semenjak": true, "yakni": true,
	"yaitu": true,
}

var enStopwords = map[string]bool{
	"a": true, "an": true, "the": true, "and": true, "or": true, "but": true,
	"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
	"with": true, "by": true, "from": true, "is": true, "are": true, "was": true,
	"were": true, "be": true, "been": true, "being": true, "have": true, "has": true,
	"had": true, "do": true, "does": true, "did": true, "will": true, "would": true,
	"could": true, "should": true, "may": true, "might": true, "shall": true,
	"can": true, "need": true, "as": true, "about": true, "above": true,
	"after": true, "again": true, "against": true, "all": true, "am": true,
	"any": true, "because": true, "before": true, "between": true, "both": true,
	"during": true, "each": true, "few": true, "further": true, "he": true,
	"her": true, "here": true, "herself": true, "him": true, "himself": true,
	"his": true, "how": true, "i": true, "if": true, "into": true, "it": true,
	"its": true, "itself": true, "just": true, "me": true, "more": true,
	"most": true, "my": true, "myself": true, "no": true, "not": true, "now": true,
	"off": true, "once": true, "only": true, "other": true, "our": true,
	"ourselves": true, "out": true, "own": true, "same": true, "she": true,
	"so": true, "some": true, "such": true, "than": true, "that": true,
	"their": true, "them": true, "themselves": true, "then": true, "there": true,
	"these": true, "they": true, "this": true, "those": true, "through": true,
	"too": true, "under": true, "until": true, "up": true, "very": true,
	"we": true, "what": true, "when": true, "where": true, "which": true,
	"while": true, "who": true, "whom": true, "why": true, "you": true,
	"your": true, "yours": true, "yourself": true, "yourselves": true,
	"get": true, "got": true, "make": true, "made": true, "use": true,
	"used": true, "also": true, "like": true, "well": true, "even": true,
	"back": true, "good": true, "much": true, "take": true, "come": true,
	"since": true, "new": true, "work": true, "working": true, "worked": true,
}
