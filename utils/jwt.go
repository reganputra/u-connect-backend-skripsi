package utils

import (
	"errors"
	"math"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

// ParseJWT accepts access tokens only, including for WebSocket authentication.
func ParseJWT(raw string) (jwt.MapClaims, error) { return parseToken(raw, "access") }

// ParseRefreshJWT must only be used by the refresh endpoint.
func ParseRefreshJWT(raw string) (jwt.MapClaims, error) { return parseToken(raw, "refresh") }

func parseToken(raw, purpose string) (jwt.MapClaims, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET is not configured")
	}
	token, err := jwt.Parse(raw, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())
	if err != nil {
		return nil, err
	}
	if token == nil || !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["token_type"] != purpose {
		return nil, jwt.ErrTokenInvalidClaims
	}
	if _, _, err := TokenIdentity(claims); err != nil {
		return nil, err
	}
	if role, ok := claims["role"].(string); !ok || role == "" {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

// TokenIdentity rejects missing, fractional, negative and out-of-range values.
// JSON numbers are float64, so restrict to exactly representable integers.
func TokenIdentity(claims jwt.MapClaims) (uint, int, error) {
	uid, ok := claims["user_id"].(float64)
	version, vok := claims["token_version"].(float64)
	if !ok || !vok || math.IsNaN(uid) || math.IsNaN(version) || uid < 1 || uid > 9007199254740991 || version < 0 || version > 9007199254740991 || math.Trunc(uid) != uid || math.Trunc(version) != version {
		return 0, 0, jwt.ErrTokenInvalidClaims
	}
	if float64(uint(uid)) != uid || float64(int(version)) != version {
		return 0, 0, jwt.ErrTokenInvalidClaims
	}
	return uint(uid), int(version), nil
}
