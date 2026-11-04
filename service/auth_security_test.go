package service

import (
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/reganputra/skripsi-backend/models"
	"github.com/reganputra/skripsi-backend/repository"
	"github.com/reganputra/skripsi-backend/utils"
	"gorm.io/gorm"
)

type rotationRepo struct {
	repository.UserRepository
	mu          sync.Mutex
	user        models.User
	bumpBarrier chan struct{}
	bumps       int
}

func (r *rotationRepo) FindUserByID(uint) (*models.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	u := r.user
	return &u, nil
}
func (r *rotationRepo) BumpTokenVersion(uint) error {
	r.mu.Lock()
	r.bumps++
	if r.bumps == 2 {
		close(r.bumpBarrier)
	}
	r.mu.Unlock()
	<-r.bumpBarrier
	r.mu.Lock()
	r.user.TokenVersion++
	r.mu.Unlock()
	return nil
}
func (r *rotationRepo) CompareAndSwapTokenVersion(id uint, expected int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.user.IsActive || r.user.TokenVersion != expected {
		return errors.New("revoked")
	}
	r.user.TokenVersion++
	return nil
}

func TestRefreshSingleWinner(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-only-key")
	repo := &rotationRepo{user: models.User{Model: gorm.Model{ID: 1}, Role: "student", IsActive: true}, bumpBarrier: make(chan struct{})}
	svc := NewAuthService(repo, nil)
	raw, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"token_type": "refresh", "user_id": 1, "role": "student", "token_version": 0, "exp": time.Now().Add(time.Hour).Unix()}).SignedString([]byte("test-only-key"))
	var wins atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := svc.Refresh(raw); err == nil {
				wins.Add(1)
			}
		}()
	}
	wg.Wait()
	if wins.Load() != 1 {
		t.Fatalf("concurrent refresh successes=%d; want exactly 1", wins.Load())
	}
	if _, err := svc.Refresh(raw); err == nil {
		t.Fatal("replay accepted")
	}
}

func TestIssuedTokensSecurity(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-only-key")
	svc := &authService{}
	access, refresh, err := svc.issueTokens(&models.User{Model: gorm.Model{ID: 1}, Role: "student"})
	if err != nil {
		t.Fatal(err)
	}
	ac, err := utils.ParseJWT(access)
	if err != nil {
		t.Fatal(err)
	}
	rc, err := utils.ParseRefreshJWT(refresh)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := utils.ParseJWT(refresh); err == nil {
		t.Fatal("refresh accepted as access")
	}
	if _, err := utils.ParseRefreshJWT(access); err == nil {
		t.Fatal("access accepted as refresh")
	}
	if ac["exp"].(float64)-ac["iat"].(float64) != 900 {
		t.Fatal("access lifetime not 15m")
	}
	if rc["exp"].(float64)-rc["iat"].(float64) != 604800 {
		t.Fatal("refresh lifetime not 7d")
	}
	b, err := json.Marshal(LoginResponse{Token: access, RefreshToken: refresh})
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]interface{}
	_ = json.Unmarshal(b, &body)
	if _, ok := body["refresh_token"]; ok {
		t.Fatal("refresh exposed in JSON")
	}
}
