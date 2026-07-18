package utils

// ─── Lemma Rules ─────────────────────────────────────────────────────────────
// Aturan suffix bahasa Inggris untuk lemmatisasi fallback (longest-first).

type lemmaRule struct {
	suffix      string
	replacement string
	minBase     int
}

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

// ─── Bilingual Lemma Dictionary ───────────────────────────────────────────────
// Memetakan surface form (EN & ID) ke bentuk kanonik agar teks mahasiswa dan
// mentor yang menggunakan variasi kata yang berbeda tetap dapat di-match dengan
// benar oleh mesin TF-IDF CBF.
//
// Dikelompokkan per domain untuk kemudahan pemeliharaan.
// Total entri: ~390 (14 domain, EN + ID).

var lemmaDict = map[string]string{

	// ─── [1] Tech Nouns ───────────────────────────────────────────────────────
	// Kata benda teknis inti yang sering muncul dalam profil IT.
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

	// ─── [2] English Tech Gerunds / Nominalizations ───────────────────────────
	// Bentuk gerund dan nominalisasi bahasa Inggris yang sering muncul di profil.
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

	// ─── [3] Tech Roles ───────────────────────────────────────────────────────
	// Normalisasi jabatan/peran teknis ke bentuk tunggal.
	"developers": "developer", "engineers": "engineer",
	"scientists": "scientist", "analysts": "analyst",
	"researchers": "researcher", "architects": "architect",

	// ─── [4] Indonesian IT Terms ──────────────────────────────────────────────
	// Istilah IT dalam bahasa Indonesia → padanan EN kanonik.
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

	// ─── [5] ML / AI / Data Science ──────────────────────────────────────────
	// Terminologi machine learning, deep learning, dan data science.
	"regression": "regression", "regressions": "regression",
	"neural":    "neural",
	"embedding": "embed", "embeddings": "embed",
	"inference":  "infer",
	"generative": "generate", "generation": "generate",
	"supervised":    "supervise",
	"unsupervised":  "unsupervise",
	"reinforcement": "reinforce",
	"convolutional": "convolve",
	"transformer":   "transform", "transformers": "transform",
	"gradient": "gradient", "gradients": "gradient",
	"hyperparameter": "hyperparameter", "hyperparameters": "hyperparameter",
	"overfitting":  "overfit",
	"underfitting": "underfit",
	"tokenization": "tokenize", "tokenizing": "tokenize",
	"normalization": "normalize", "normalizing": "normalize",
	"regularization": "regularize",
	"dimensionality": "dimension",
	"feature":        "feature", "features": "feature",
	"labeling": "label", "labelling": "label",
	"dataset": "dataset", "datasets": "dataset",
	"epoch": "epoch", "epochs": "epoch",
	"batch": "batch", "batches": "batch",
	// Indonesian ML terms
	"regresi":       "regression",
	"pengelompokan": "cluster",
	"inferensi":     "infer",
	"ekstraksi":     "extract",
	"fitur":         "feature",
	"normalisasi":   "normalize",

	// ─── [6] Cloud / DevOps / Mobile ─────────────────────────────────────────
	// Platform cloud, toolchain DevOps, dan pengembangan mobile.
	"containerization": "container", "containerizing": "container",
	"orchestration": "orchestrate", "orchestrating": "orchestrate",
	"kubernetes":   "kubernetes",
	"docker":       "docker",
	"microservice": "microservice", "microservices": "microservice",
	"serverless":  "serverless",
	"continuous":  "continuous",
	"integration": "integrate", "integrations": "integrate",
	"infrastructure": "infrastructure",
	"provisioning":   "provision",
	"scalability":    "scale",
	"availability":   "available",
	"reliability":    "reliable",
	"caching":        "cache",
	"android":        "android",
	"flutter":        "flutter",
	"mobile":         "mobile",
	"responsive":     "responsive",
	"native":         "native",
	"hybrid":         "hybrid",
	// Indonesian cloud/mobile terms
	"kontainer":     "container",
	"orkestrasi":    "orchestrate",
	"skalabilitas":  "scale",
	"infrastruktur": "infrastructure",
	"integrasi":     "integrate",
	"ketersediaan":  "available",

	// ─── [7] Bisnis & Manajemen ───────────────────────────────────────────────
	// Terminologi bisnis, strategi, operasional, dan kewirausahaan.
	"marketing": "market", "marketer": "market",
	"branding": "brand",
	"strategy": "strategy", "strategies": "strategy", "strategic": "strategy",
	"operations": "operate", "operational": "operate",
	"entrepreneurship": "entrepreneur", "entrepreneurial": "entrepreneur",
	"leadership": "lead", "leader": "lead", "leaders": "lead",
	"stakeholder": "stakeholder", "stakeholders": "stakeholder",
	"consulting": "consult", "consultant": "consult", "consultants": "consult",
	"procurement": "procure",
	"logistics":   "logistic",
	"ecommerce":   "commerce",
	"innovation":  "innovate", "innovative": "innovate",
	"planning":  "plan",
	"execution": "execute",
	"revenue":   "revenue",
	"growth":    "grow",
	"pitching":  "pitch",
	"startup":   "startup", "startups": "startup",
	// Indonesian business terms
	"pemasaran":     "market",
	"strategi":      "strategy",
	"operasional":   "operate",
	"kewirausahaan": "entrepreneur",
	"kepemimpinan":  "lead",
	"konsultasi":    "consult",
	"pengadaan":     "procure",
	"logistik":      "logistic",
	"inovasi":       "innovate",
	"pertumbuhan":   "grow",
	"pendapatan":    "revenue",
	"perencanaan":   "plan",

	// ─── [8] Keuangan & Akuntansi ─────────────────────────────────────────────
	// Terminologi keuangan, akuntansi, investasi, dan perpajakan.
	"accounting": "account",
	"auditing":   "audit",
	"investment": "invest", "investments": "invest", "investing": "invest",
	"valuation": "value", "valuations": "value",
	"portfolio": "portfolio", "portfolios": "portfolio",
	"taxation":  "tax",
	"financial": "finance", "financing": "finance",
	"budgeting": "budget",
	"risk":      "risk", "risks": "risk",
	"reporting": "report",
	"treasury":  "treasury",
	"insurance": "insure",
	"asset":     "asset", "assets": "asset",
	"liability": "liable", "liabilities": "liable",
	"equity":   "equity",
	"dividend": "dividend", "dividends": "dividend",
	"capital": "capital",
	// Indonesian finance terms
	"akuntansi":  "account",
	"investasi":  "invest",
	"perpajakan": "tax",
	"pembiayaan": "finance",
	"anggaran":   "budget",
	"risiko":     "risk",
	"pelaporan":  "report",
	"asuransi":   "insure",
	"aset":       "asset",
	"ekuitas":    "equity",
	"modal":      "capital",

	// ─── [9] Komunikasi & Media ───────────────────────────────────────────────
	// Jurnalistik, periklanan, media sosial, dan produksi konten.
	"journalism": "journalist",
	"content":    "content", "contents": "content",
	"copywriting":  "copywrite",
	"broadcasting": "broadcast",
	"advertising":  "advertise", "advertisement": "advertise", "advertisements": "advertise",
	"media":         "media",
	"communication": "communicate", "communications": "communicate",
	"storytelling": "story",
	"editing":      "edit",
	"production":   "produce",
	"publishing":   "publish",
	"photography":  "photograph",
	"videography":  "video",
	"podcast":      "podcast", "podcasting": "podcast",
	// Indonesian media terms
	"jurnalistik":  "journalist",
	"penyiaran":    "broadcast",
	"periklanan":   "advertise",
	"komunikasi":   "communicate",
	"penyuntingan": "edit",
	"penerbitan":   "publish",
	"fotografi":    "photograph",
	"produksi":     "produce",
	"konten":       "content",

	// ─── [10] Pendidikan & Pelatihan ─────────────────────────────────────────
	// Kurikulum, pedagogi, pelatihan, dan fasilitasi pembelajaran.
	"curriculum": "curriculum", "curricula": "curriculum",
	"pedagogy": "pedagogy", "pedagogical": "pedagogy",
	"coaching":     "coach",
	"facilitation": "facilitate", "facilitating": "facilitate",
	"teaching": "teach", "teacher": "teach",
	"education": "educate", "educational": "educate",
	"instruction": "instruct", "instructional": "instruct",
	"assessment": "assess",
	"evaluation": "evaluate",
	"workshop":   "workshop", "workshops": "workshop",
	"seminar": "seminar", "seminars": "seminar",
	"lecture": "lecture", "lectures": "lecture",
	"tutoring": "tutor",
	// Indonesian education terms
	"kurikulum":  "curriculum",
	"pengajaran": "teach",
	"fasilitasi": "facilitate",
	"pendidikan": "educate",
	"penilaian":  "assess",
	"lokakarya":  "workshop",

	// ─── [11] Riset & Akademik ────────────────────────────────────────────────
	// Metodologi penelitian, publikasi ilmiah, dan terminologi akademik.
	"methodology": "method", "methodologies": "method",
	"publication": "publish", "publications": "publish",
	"hypothesis": "hypothesis", "hypotheses": "hypothesis",
	"literature": "literature",
	"review":     "review", "reviews": "review",
	"survey": "survey", "surveys": "survey",
	"experiment": "experiment", "experiments": "experiment", "experimental": "experiment",
	"findings": "find",
	"citation": "cite", "citations": "cite",
	"reference": "reference", "references": "reference",
	"qualitative":  "qualitative",
	"quantitative": "quantitative",
	"dissertation": "dissertation",
	"thesis":       "thesis",
	// Indonesian research terms
	"metodologi":  "method",
	"publikasi":   "publish",
	"hipotesis":   "hypothesis",
	"literatur":   "literature",
	"tinjauan":    "review",
	"penelitian":  "research",
	"survei":      "survey",
	"eksperimen":  "experiment",
	"temuan":      "find",
	"referensi":   "reference",
	"kualitatif":  "qualitative",
	"kuantitatif": "quantitative",

	// ─── [12] Hukum & Kebijakan ───────────────────────────────────────────────
	// Regulasi, kepatuhan, tata kelola, dan kebijakan publik.
	"regulation": "regulate", "regulations": "regulate", "regulatory": "regulate",
	"compliance":  "comply",
	"legislation": "legislate", "legislative": "legislate",
	"policy": "policy", "policies": "policy",
	"governance": "govern",
	"legal":      "legal",
	"law":        "law", "laws": "law",
	"contract": "contract", "contracts": "contract",
	"arbitration":  "arbitrate",
	"mediation":    "mediate",
	"intellectual": "intellectual",
	"property":     "property",
	// Indonesian law terms
	"regulasi":  "regulate",
	"kepatuhan": "comply",
	"legislasi": "legislate",
	"kebijakan": "policy",
	"kelola":    "govern",
	"hukum":     "law",
	"kontrak":   "contract",
	"kekayaan":  "property",

	// ─── [13] Teknik Non-Software ─────────────────────────────────────────────
	// Teknik sipil, mekanik, elektro, industri, dan manufaktur.
	"structural":    "structure",
	"mechanical":    "mechanic",
	"electrical":    "electric",
	"construction":  "construct",
	"manufacturing": "manufacture",
	"industrial":    "industry",
	"automotive":    "automotive",
	"aerospace":     "aerospace",
	"robotics":      "robot",
	"automation":    "automate",
	"sensor":        "sensor", "sensors": "sensor",
	"hardware": "hardware",
	"firmware": "firmware",
	"circuit":  "circuit", "circuits": "circuit",
	// Indonesian engineering terms
	"struktural": "structure",
	"mekanik":    "mechanic",
	"konstruksi": "construct",
	"manufaktur": "manufacture",
	"industri":   "industry",
	"otomotif":   "automotive",
	"robotika":   "robot",
	"otomasi":    "automate",

	// ─── [14] Kesehatan (Umum) ────────────────────────────────────────────────
	// Terminologi klinis, farmasi, keperawatan, dan kesehatan masyarakat.
	"clinical":       "clinic",
	"pharmaceutical": "pharmacy",
	"diagnosis":      "diagnose", "diagnoses": "diagnose",
	"treatment": "treat", "treatments": "treat",
	"medical":    "medicine",
	"healthcare": "health",
	"nursing":    "nurse",
	"therapy":    "therapy", "therapies": "therapy",
	"rehabilitation": "rehabilitate",
	"nutrition":      "nutrition",
	"epidemiology":   "epidemiology",
	"laboratory":     "laboratory", "laboratories": "laboratory",
	"biomedical": "biomedical",
	// Indonesian health terms
	"klinis":       "clinic",
	"farmasi":      "pharmacy",
	"pengobatan":   "treat",
	"medis":        "medicine",
	"kesehatan":    "health",
	"keperawatan":  "nurse",
	"terapi":       "therapy",
	"rehabilitasi": "rehabilitate",
	"nutrisi":      "nutrition",
	"laboratorium": "laboratory",
}
