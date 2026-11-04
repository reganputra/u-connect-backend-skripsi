package utils

import (
	"github.com/golang-jwt/jwt/v5"
	"testing"
	"time"
)

func TestParseJWTSecurity(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-only-key")
	for _, tc := range []struct {
		name   string
		method jwt.SigningMethod
		mutate func(jwt.MapClaims)
		valid  bool
	}{
		{"access", jwt.SigningMethodHS256, func(c jwt.MapClaims) {}, true},
		{"refresh rejected", jwt.SigningMethodHS256, func(c jwt.MapClaims) { c["token_type"] = "refresh" }, false},
		{"missing purpose", jwt.SigningMethodHS256, func(c jwt.MapClaims) { delete(c, "token_type") }, false},
		{"missing expiry", jwt.SigningMethodHS256, func(c jwt.MapClaims) { delete(c, "exp") }, false},
		{"expired", jwt.SigningMethodHS256, func(c jwt.MapClaims) { c["exp"] = time.Now().Add(-time.Minute).Unix() }, false},
		{"HS384", jwt.SigningMethodHS384, func(c jwt.MapClaims) {}, false},
		{"missing version", jwt.SigningMethodHS256, func(c jwt.MapClaims) { delete(c, "token_version") }, false},
		{"fractional id", jwt.SigningMethodHS256, func(c jwt.MapClaims) { c["user_id"] = 1.5 }, false},
		{"negative version", jwt.SigningMethodHS256, func(c jwt.MapClaims) { c["token_version"] = -1 }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := jwt.MapClaims{"user_id": 1, "token_version": 0, "role": "student", "token_type": "access", "exp": time.Now().Add(time.Minute).Unix()}
			tc.mutate(c)
			raw, err := jwt.NewWithClaims(tc.method, c).SignedString([]byte("test-only-key"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = ParseJWT(raw)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v error=%v", tc.valid, err)
			}
		})
	}
}

func TestParseJWTRejectsEmptySecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	raw, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"exp": time.Now().Add(time.Minute).Unix(), "user_id": 1, "token_version": 0, "token_type": "access", "role": "student"}).SignedString([]byte(""))
	if _, err := ParseJWT(raw); err == nil {
		t.Fatal("accepted token signed with empty secret")
	}
}
