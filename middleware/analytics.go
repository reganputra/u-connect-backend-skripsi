package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

// ViewCooldownDuration adalah jangka waktu minimum yang harus berlalu sebelum
// pengguna yang sama (yang telah terautentikasikan) dapat menghasilkan peristiwa tampilan kedua untuk
// konten yang sama. Permintaan yang masuk dalam rentang waktu ini akan diabaikan tanpa pemberitahuan sehingga
// penyegaran halaman berulang atau pengambilan ulang data oleh frontend tidak akan meningkatkan jumlah hitungan.
//
// 1 jam mencerminkan praktik industri umum untuk metrik "unique hourly viewer"
// dan mudah dijelaskan dalam dokumentasi akademik.
const ViewCooldownDuration = 1 * time.Hour

// TrackView adalah middleware Fiber yang menambahkan baris PageView
// secara asinkron setelah handler selesai merespons.
//
// Aturan:
//   - Hanya mencatat tampilan untuk permintaan GET yang mengembalikan HTTP 200.
//   - Mengambil resource ID dari parameter route ":id".
//   - Mengambil ID pengguna dari JWT yang disimpan dalam c.Locals("user").
//     Jika tidak ada JWT (rute publik di masa mendatang), pemeriksaan cooldown dilewati
//     dan tampilan dicatat dengan user_id nil.
//   - Gerbang deduplikasi berjalan di dalam goroutine sehingga tidak pernah menambahkan latensi
//     ke waktu respons handler.
func TrackView(targetType string, db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Jalankan handler yang sebenarnya terlebih dahulu.
		err := c.Next()

		// Hanya lacak permintaan GET yang berhasil.
		if c.Method() != "GET" || c.Response().StatusCode() != 200 {
			return err
		}

		idStr := c.Params("id")
		id, parseErr := strconv.ParseUint(idStr, 10, 64)
		if parseErr != nil || id == 0 {
			return err
		}

		userID := extractViewUserID(c)

		go func() {
			// Filter deduplikasi — hanya diterapkan untuk pengguna yang telah terotentikasi karena
			// tampilan anonim tidak dapat dikaitkan dengan identitas yang tetap.
			if userID != nil {
				cutoff := time.Now().UTC().Add(-ViewCooldownDuration)

				var count int64
				db.Model(&models.PageView{}).
					Where(
						"user_id = ? AND target_type = ? AND target_id = ? AND created_at >= ?",
						*userID, targetType, uint(id), cutoff,
					).
					Limit(1).
					Count(&count)

				if count > 0 {
					// Tampilan sudah tercatat dalam jendela cooldown — lewati penyisipan.
					return
				}
			}

			_ = db.Create(&models.PageView{
				CreatedAt:  time.Now().UTC(),
				UserID:     userID,
				TargetType: targetType,
				TargetID:   uint(id),
			}).Error
		}()

		return err
	}
}

// extractViewUserID mengekstrak user_id dari JWT yang disimpan dalam c.Locals("user").
// Mengembalikan nil jika tidak ada token yang ada atau token rusak (misalnya rute publik).
func extractViewUserID(c *fiber.Ctx) *uint {
	token, ok := c.Locals("user").(*jwt.Token)
	if !ok {
		return nil
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil
	}
	idFloat, ok := claims["user_id"].(float64)
	if !ok || idFloat <= 0 {
		return nil
	}
	id := uint(idFloat)
	return &id
}
