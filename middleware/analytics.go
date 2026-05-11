package middleware

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/reganputra/skripsi-backend/models"
	"gorm.io/gorm"
)

// TrackView mengembalikan middleware Fiber yang menambahkan baris PageView secara asinkron
// setelah handler merespons.
//
// Aturan:
//   - Hanya mencatat tampilan untuk permintaan GET yang mengembalikan HTTP 200.
//   - Mengekstrak ID sumber daya dari parameter rute ":id".
//   - Menangkap ID pengguna terautentikasi dari JWT dalam c.Locals("user").
//     Jika tidak ada JWT yang hadir (rute publik di masa mendatang), user_id disimpan sebagai nil.
//   - Penyisipan basis data terjadi dalam goroutine fire-and-forget sehingga tidak pernah
//     menambah latensi pada waktu respons handler.
func TrackView(targetType string, db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Jalankan handler sebenarnya terlebih dahulu.
		err := c.Next()

		// Hanya mencatat tampilan untuk permintaan GET yang berhasil.
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
// Mengembalikan nil jika token tidak ada atau rusak (misalnya untuk rute tidak terautentikasi).
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
