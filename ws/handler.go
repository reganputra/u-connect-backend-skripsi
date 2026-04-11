package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	fws "github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/reganputra/skripsi-backend/repository"
	"github.com/reganputra/skripsi-backend/service"
)

// IncomingMsg is the JSON payload a client sends over WebSocket.
type IncomingMsg struct {
	ReceiverID uint   `json:"receiver_id"`
	Content    string `json:"content"`
}

// WSHandler returns the Fiber WebSocket upgrade handler.
//
// Auth: JWT must be passed as query param ?token=<jwt> because WebSocket
// upgrade is a GET request and headers cannot be set reliably from browsers.
//
// Usage (route registration):
//
//	app.Get("/api/ws", websocket.IsWebSocketUpgrade, ws.WSHandler(hub, msgSvc, userRepo, notifSvc))
func WSHandler(hub *Hub, msgSvc service.MessageService, userRepo repository.UserRepository, notifSvc service.NotificationService) fiber.Handler {
	return fws.New(func(c *fws.Conn) {
		// ── 1. Authenticate ───────────────────────────────────────────────────
		tokenStr := c.Query("token")
		if tokenStr == "" {
			writeError(c, "token diperlukan")
			_ = c.Close()
			return
		}
		claims, err := parseJWT(tokenStr)
		if err != nil {
			writeError(c, "token tidak valid")
			_ = c.Close()
			return
		}
		userIDFloat, ok := claims["user_id"].(float64)
		role, _ := claims["role"].(string)
		if !ok || userIDFloat == 0 {
			writeError(c, "user_id tidak ditemukan dalam token")
			_ = c.Close()
			return
		}
		if role != "student" && role != "alumni" {
			writeError(c, "akses ditolak: hanya student dan alumni yang dapat menggunakan pesan")
			_ = c.Close()
			return
		}
		userID := uint(userIDFloat)

		// ── 2. Register client ────────────────────────────────────────────────
		client := &Client{
			UserID: userID,
			Conn:   c,
			Send:   make(chan []byte, 64),
			Done:   make(chan struct{}),
		}
		hub.Register <- client
		defer func() { hub.Unregister <- client }()

		// ── 3. Write pump (hub → client) ──────────────────────────────────────
		go func() {
			for {
				select {
				case payload, ok := <-client.Send:
					if !ok {
						return
					}
					if err := c.WriteMessage(1, payload); err != nil {
						return
					}
				case <-client.Done:
					return
				}
			}
		}()

		// ── 4. Read pump (client → hub → DB → recipient) ──────────────────────
		for {
			_, raw, err := c.ReadMessage()
			if err != nil {
				break // client disconnected
			}

			var incoming IncomingMsg
			if err := json.Unmarshal(raw, &incoming); err != nil {
				writeError(c, "format pesan tidak valid")
				continue
			}
			if incoming.ReceiverID == 0 || incoming.Content == "" {
				writeError(c, "receiver_id dan content wajib diisi")
				continue
			}

			// Persist message via service (validates follow relationship)
			msg, err := msgSvc.SendMessage(userID, incoming.ReceiverID, incoming.Content)
			if err != nil {
				writeError(c, err.Error())
				continue
			}

			// Deliver to recipient if online
			payload, _ := json.Marshal(OutgoingMsg{
				Type: "message",
				Data: msg,
			})
			hub.SendToUser(incoming.ReceiverID, payload)

			// Echo back to sender as confirmation
			hub.SendToUser(userID, payload)

			// 5. Throttled persistent notification
			if sender, err := userRepo.FindUserByID(userID); err == nil {
				_ = notifSvc.NotifyThrottled(
					incoming.ReceiverID,
					"new_message",
					"Pesan baru",
					fmt.Sprintf("Pesan dari %s: %s", sender.Name, truncateText(incoming.Content, 40)),
					"message",
					msg.ID,
					5*time.Minute,
				)
			}
		}
	})
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func writeError(c *fws.Conn, msg string) {
	payload, _ := json.Marshal(OutgoingMsg{Type: "error", Message: msg})
	_ = c.WriteMessage(1, payload)
}

func parseJWT(tokenStr string) (jwt.MapClaims, error) {
	secret := os.Getenv("JWT_SECRET")
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}
	// Check expiry
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, jwt.ErrTokenExpired
		}
	}
	_ = log.Printf // suppress unused import warning if log unused later
	return claims, nil
}

func truncateText(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "…"
}
