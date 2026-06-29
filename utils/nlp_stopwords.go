package utils

// ─── Indonesian Stopwords ─────────────────────────────────────────────────────
// 117 entri kata fungsi bahasa Indonesia yang umum dalam teks bebas.

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

// ─── English Stopwords ────────────────────────────────────────────────────────
// 140 entri kata fungsi bahasa Inggris termasuk auxiliary verbs dan kata umum.

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
