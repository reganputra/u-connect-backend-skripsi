# Repository Layer — Improvement Plan

## Priority Matrix

| Prio | Area | File(s) | Effort | Risk |
|---|---|---|---|---|
| P0 | Error `Count()` tidak dicek | 13 file | Rendah | Aman — tambah `if err := ...` |
| P1 | Self-assignment bug | `mentor_request_repository.go:223` | Sangat rendah | Aman — hapus baris |
| P2 | Cascade delete tanpa transaksi | `admin_repository`, `group_article_repository`, `group_comment_repository` | Rendah | Rendah — bungkus pakai `db.Transaction` |
| P3 | Hard-coded `2` jadi constant | `mentor_request_repository.go:139` | Rendah | Rendah |
| P4 | String literal role/status → `constant.*` | 6 file | Sedang | Rendah — search & replace |
| P5 | Pola count-then-fetch helper | 10+ file | Sedang | Rendah — extract function |
| P6 | Repository DTOs → `dto/` | 7 struct | Sedang | Rendah — pure move |

---

## P0 — Error `Count()` tidak dicek

**Problem:** 13+ lokasi memanggil `.Count(&total)` dan mengabaikan `error`. Jika DB error, response mengembalikan `total: 0` tanpa pesan error — data tidak akurat dan user tidak tahu.

**Files affected:**

| File | Line |
|---|---|
| `post_repository.go` | 43 |
| `event_repository.go` | 40 |
| `profile_repository.go` | 151, 189, 226 |
| `admin_repository.go` | 107 |
| `job_repository.go` | 46 |
| `notification_repository.go` | 41 |
| `report_repository.go` | 39, 61 |
| `group_member_repository.go` | 67 |
| `event_registration_repository.go` | 94 |
| `analytics_repository.go` | 62 (internal closure) |

**Action:** Ubah setiap `q.Count(&total)` menjadi:

```go
if err := q.Count(&total).Error; err != nil {
    return nil, 0, err
}
```

Untuk `analytics_repository.go`, ubah helper `count` mengembalikan `(int64, error)`.

---

## P1 — Self-assignment `req.ApprovedAt = req.ApprovedAt`

**File:** `repository/mentor_request_repository.go:223`

**Action:** Hapus baris tersebut. Tidak ada efek, kode mati.

---

## P2 — Cascade delete tanpa transaksi

**Problem:** `admin_repository.DeletePost` dan `group_article_repository.DeleteGroupArticle` menjalankan multiple `DELETE` tanpa transaksi. Jika statement tengah gagal, data parsial terhapus tanpa rollback.

**Action:** Bungkus setiap rangkaian dengan `r.db.Transaction(func(tx *gorm.DB) error { ... })`, mengikuti pola `event_repository.DeleteEvent` dan `group_repository.DeleteGroup` yang sudah benar.

**Files:**

- `admin_repository.go:126-132` — `DeletePost`
- `group_article_repository.go:101-107` — `DeleteGroupArticle`
- `group_comment_repository.go:42-44` — `DeleteGroupComment`

---

## P3 — Hard-coded `2` (max mentor per student)

**File:** `repository/mentor_request_repository.go:139`

**Problem:** `studentApprovedCount >= 2` — angka ajaib terkubur di repository. Service layer (`mentor_service.go`) punya logika serupa. Dua sumber kebenaran.

**Action:**

1. Tambahkan constant di `constant/constant.go`:
   ```go
   const MaxMentorsPerStudent = 2
   ```
2. Ganti `2` di `mentor_request_repository.go:139` dengan `constant.MaxMentorsPerStudent`
3. Cari dan ganti juga di service layer jika ada literal yang sama

---

## P4 — String literal role/status → `constant.*`

**Problem:** Beberapa repository masih pakai string literal sebagai Go argument (bukan embedded SQL), yang sudah ada constant-nya di `constant/constant.go`.

**Action:** Search & replace:

| File | Line | Cari | Ganti |
|---|---|---|---|
| `mentor_repository.go` | 67 | `"alumni"` | `constant.RoleAlumni` |
| `mentor_repository.go` | 94 | `"alumni"` | `constant.RoleAlumni` |
| `profile_repository.go` | 84 | `"partner"` | `constant.RolePartner` |
| `profile_repository.go` | 148 | `"student"`, `"alumni"`, `"partner"` | `constant.RoleStudent`, `constant.RoleAlumni`, `constant.RolePartner` |
| `profile_repository.go` | 182 | sama seperti 148 | sama |
| `profile_repository.go` | 207, 223 | `"student"`, `"alumni"`, `"partner"` | `constant.RoleStudent`, `constant.RoleAlumni`, `constant.RolePartner` |
| `group_repository.go` | 46 | `"owner"` | `constant.RoleGroupOwner` (tambah jika belum ada) |
| `admin_repository.go` | 69 | `'pending'` | skip (embedded SQL) |
| `mentor_repository.go` | 47, 55, 158 | `'approved'` | skip (embedded SQL) |

**Note:** String di raw SQL / `Where("status = 'approved'")` tetap dilewati sesuai keputusan sebelumnya (SQL syntax, bukan Go value).

---

## P5 — Pola count-then-fetch helper

**Problem:** ~10 file mengulang pola yang sama:

```go
var total int64
r.db.Model(&T{}).Count(&total)     // error sering diabaikan
err := r.db.Where(...).Offset(offset).Limit(limit).Find(&res).Error
```

**Action:** Buat helper di `repository/helpers.go` untuk mengurangi boilerplate:

```go
// Paginate applies offset/limit to a query, computing offset from page/limit.
func Paginate(page, limit int) func(db *gorm.DB) *gorm.DB {
    return func(db *gorm.DB) *gorm.DB {
        offset := (page - 1) * limit
        return db.Offset(offset).Limit(limit)
    }
}

// CountQuery returns total rows for db query or error (never silently ignored).
func CountQuery(db *gorm.DB) (int64, error) {
    var total int64
    err := db.Count(&total).Error
    return total, err
}
```

Contoh penggunaan:

```go
total, err := CountQuery(r.db.Model(&models.Event{}))
if err != nil {
    return nil, 0, err
}
var events []models.Event
err := r.db.Scopes(Paginate(page, limit)).Order("created_at DESC").Find(&events).Error
```

---

## P6 — Repository DTOs → `dto/`

**Problem:** 7 struct hasil query didefinisikan di `repository/`, mengaburkan batas antara query logic dan response shape.

**Action:** Pindahkan struct berikut ke `dto/dto.go`:

| Struct | Source File |
|---|---|
| `AdminUserWithProfile` | `admin_repository.go` |
| `TopContentRow` | `analytics_repository.go` |
| `EnhancedStats` | `analytics_repository.go` |
| `ConversationSummary` | `message_repository.go` |
| `PostListRow` | `post_repository.go` |
| `DirectorySummary` | `profile_repository.go` |
| `MentorDoc` | `mentor_repository.go` |
| `GroundTruthEntry` | `mentor_request_repository.go` |

**Note:** Pure move — tidak ada perubahan logic. Import di service/controller perlu disesuaikan.

---

## Execution Order

```
Phase 1 — P0 + P1 (bug fixes)      → 1 session, ~30 menit
Phase 2 — P2 (transactions)         → 1 session, ~20 menit
Phase 3 — P3 + P4 (constants)       → 1 session, ~20 menit
Phase 4 — P5 + P6 (refactor)        → 1-2 session, ~45 menit
```

Setiap phase diakhiri dengan `go build ./...` dan `go vet ./...` untuk memastikan tidak ada regresi.
