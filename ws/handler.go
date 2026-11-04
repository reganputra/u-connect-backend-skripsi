package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	fws "github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/reganputra/skripsi-backend/repository"
	"github.com/reganputra/skripsi-backend/service"
	"github.com/reganputra/skripsi-backend/utils"
)

// IncomingMsg is the JSON payload a client sends over WebSocket.
type IncomingMsg struct {
	ReceiverID uint   `json:"receiver_id"`
	Content    string `json:"content"`
	ReplyToID  *uint  `json:"reply_to_id,omitempty"`
}

// Keepalive timing constants.
const (
	pongWait   = 60 * time.Second // read deadline; client must pong within this window
	pingPeriod = 30 * time.Second // server sends a Ping at this interval (must be < pongWait)
	writeWait  = 10 * time.Second // per-write deadline
)

func extractWSToken(c *fiber.Ctx) string {
	if h := c.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	parts := strings.Split(c.Get("Sec-WebSocket-Protocol"), ",")
	if len(parts) == 2 && strings.TrimSpace(parts[0]) == "Bearer" {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

// WSAuthMiddleware never accepts access tokens in URLs. Tickets are opaque,
// single-use Redis handles and the retained access token is revalidated at upgrade.
func WSAuthMiddleware(users repository.UserRepository, tickets *TicketStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Query("token") != "" {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		raw := extractWSToken(c)
		if ticket := c.Query("ticket"); ticket != "" {
			ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
			defer cancel()
			var err error
			raw, err = tickets.Consume(ctx, ticket)
			if err != nil {
				return c.SendStatus(fiber.StatusUnauthorized)
			}
		}
		claims, err := utils.ParseJWT(raw)
		if err != nil || users == nil {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		uid, ok := claims["user_id"].(float64)
		version, vok := claims["token_version"].(float64)
		if !ok || !vok || uid <= 0 || uid != float64(uint(uid)) {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		user, err := users.FindUserByID(uint(uid))
		if err != nil || user == nil || !user.IsActive || version != float64(user.TokenVersion) {
			return c.SendStatus(fiber.StatusUnauthorized)
		}
		if user.Role != "student" && user.Role != "alumni" {
			return c.SendStatus(fiber.StatusForbidden)
		}
		c.Locals("ws_user_id", user.ID)
		c.Locals("ws_role", user.Role)
		c.Locals("ws_token", raw)
		return c.Next()
	}
}

// WSHandler returns the Fiber WebSocket upgrade handler.
//
// Authentication is handled upstream by WSAuthMiddleware; this handler only
// reads the pre-validated user ID from Conn.Locals and manages the connection
// lifecycle (register → keepalive write pump → blocking read pump → unregister).
func WSHandler(hub *Hub, msgSvc service.MessageService, userRepo repository.UserRepository, notifSvc service.NotificationService) fiber.Handler {
	return fws.New(func(c *fws.Conn) {
		// ── 1. Read pre-validated identity (set by WSAuthMiddleware) ──────────
		userID, ok := c.Locals("ws_user_id").(uint)
		if !ok || userID == 0 {
			// Safety net — should never trigger with WSAuthMiddleware in place.
			log.Printf("[WS/CONN] ERROR safety-net triggered — ws_user_id missing from Locals — ip: %s", c.IP())
			writeError(c, "autentikasi gagal")
			return
		}

		log.Printf("[WS/CONN] INFO  connection established — userID: %d, ip: %s", userID, c.IP())

		// ── 2. Register client in the hub ─────────────────────────────────────
		client := &Client{
			UserID: userID,
			Conn:   c,
			Send:   make(chan []byte, 64),
			Done:   make(chan struct{}),
		}
		select {
		case hub.Register <- client:
		case <-hub.quit:
			return
		}
		defer func() {
			select {
			case hub.Unregister <- client:
			case <-hub.quit:
			}
			log.Printf("[WS/CONN] INFO  unregister queued — userID: %d", userID)
		}()

		// ── 3. Keepalive: reset read deadline on every Pong ──────────────────
		if err := c.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			log.Printf("[WS/CONN] WARN  SetReadDeadline failed — userID: %d — %v", userID, err)
		}
		c.SetPongHandler(func(string) error {
			// Uncomment the line below for verbose keepalive tracing:
			// log.Printf("[WS/CONN] DEBUG pong received — userID: %d", userID)
			return c.SetReadDeadline(time.Now().Add(pongWait))
		})

		log.Printf("[WS/CONN] INFO  pumps starting — userID: %d, pingPeriod: %s, pongWait: %s",
			userID, pingPeriod, pongWait)

		// ── 4. Write pump (hub → client) ──────────────────────────────────────
		// This goroutine is the sole owner of c.WriteMessage, which eliminates
		// data races on the underlying net.Conn.
		go func() {
			defer c.Close()
			ticker := time.NewTicker(pingPeriod)
			defer ticker.Stop()

			for {
				select {
				case payload, ok := <-client.Send:
					if !ok {
						log.Printf("[WS/PUMP] INFO  send channel closed — userID: %d", userID)
						return
					}
					_ = c.SetWriteDeadline(time.Now().Add(writeWait))
					if err := c.WriteMessage(fws.TextMessage, payload); err != nil {
						log.Printf("[WS/PUMP] ERROR write failed — userID: %d — %v", userID, err)
						return
					}

				case <-ticker.C:
					raw, _ := c.Locals("ws_token").(string)
					claims, err := utils.ParseJWT(raw)
					user, userErr := userRepo.FindUserByID(userID)
					if err != nil || userErr != nil || user == nil || !user.IsActive || claims["token_version"] != float64(user.TokenVersion) {
						return
					}
					_ = c.SetWriteDeadline(time.Now().Add(writeWait))
					if err := c.WriteMessage(fws.PingMessage, nil); err != nil {
						log.Printf("[WS/PUMP] WARN  ping failed — userID: %d — %v", userID, err)
						return
					}

				case <-client.Done:
					log.Printf("[WS/PUMP] INFO  pump stopped (connection replaced by new session) — userID: %d", userID)
					return
				}
			}
		}()

		// ── 5. Read pump (client → hub → DB → recipient) ──────────────────────
		for {
			_, raw, err := c.ReadMessage()
			if err != nil {
				// Distinguish normal browser-initiated closes from unexpected errors.
				if fws.IsCloseError(err,
					fws.CloseNormalClosure,
					fws.CloseGoingAway,
					fws.CloseNoStatusReceived,
				) {
					log.Printf("[WS/CONN] INFO  disconnected (client closed) — userID: %d", userID)
				} else {
					log.Printf("[WS/CONN] WARN  disconnected (error) — userID: %d — %v", userID, err)
				}
				break
			}

			// Recheck every incoming message: disabling/logout cannot keep using
			// an already-upgraded socket to perform writes.
			rawToken, _ := c.Locals("ws_token").(string)
			claims, authErr := utils.ParseJWT(rawToken)
			currentUser, userErr := userRepo.FindUserByID(userID)
			if authErr != nil || userErr != nil || currentUser == nil || !currentUser.IsActive || claims["token_version"] != float64(currentUser.TokenVersion) {
				return
			}
			var incoming IncomingMsg
			if err := json.Unmarshal(raw, &incoming); err != nil {
				log.Printf("[WS/MSG]  WARN  invalid JSON — userID: %d — %v", userID, err)
				sendError(client, "format pesan tidak valid")
				continue
			}
			if incoming.ReceiverID == 0 || incoming.Content == "" {
				log.Printf("[WS/MSG]  WARN  missing fields — userID: %d, receiverID: %d, contentEmpty: %v",
					userID, incoming.ReceiverID, incoming.Content == "")
				sendError(client, "receiver_id dan content wajib diisi")
				continue
			}

			log.Printf("[WS/MSG]  INFO  sending — from: %d → to: %d, len: %d",
				userID, incoming.ReceiverID, len(incoming.Content))

			// Persist message (validates follow relationship in service layer)
			msg, err := msgSvc.SendMessage(userID, incoming.ReceiverID, incoming.Content, incoming.ReplyToID)
			if err != nil {
				log.Printf("[WS/MSG]  WARN  service error — from: %d → to: %d — %v",
					userID, incoming.ReceiverID, err)
				sendError(client, err.Error())
				continue
			}

			log.Printf("[WS/MSG]  INFO  persisted — msgID: %d, from: %d → to: %d",
				msg.ID, userID, incoming.ReceiverID)

			// Deliver to recipient if online
			payload, _ := json.Marshal(OutgoingMsg{
				Type: "message",
				Data: msg,
			})
			if hub.SendToUser(incoming.ReceiverID, payload) {
				log.Printf("[WS/MSG]  INFO  delivery accepted for receiverID: %d (not a receipt acknowledgement)", incoming.ReceiverID)
			} else {
				log.Printf("[WS/MSG]  INFO  receiverID: %d is offline — saved to DB, awaiting REST fetch", incoming.ReceiverID)
			}

			// Echo back to sender as delivery confirmation
			hub.SendToUser(userID, payload)

			// Throttled persistent notification (best-effort; failures are ignored)
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

// writeError writes an error frame directly to the connection.
// Only safe BEFORE the write-pump goroutine is started (pre-registration path).
func writeError(c *fws.Conn, msg string) {
	payload, _ := json.Marshal(OutgoingMsg{Type: "error", Message: msg})
	_ = c.WriteMessage(fws.TextMessage, payload)
}

// sendError enqueues an error frame on the client's Send channel.
// Safe to call concurrently with the write-pump goroutine.
func sendError(client *Client, msg string) {
	payload, _ := json.Marshal(OutgoingMsg{Type: "error", Message: msg})
	select {
	case client.Send <- payload:
	default: // drop silently if the buffer is full
	}
}

func truncateText(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "…"
}
