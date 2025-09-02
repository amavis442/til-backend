package auth

import "github.com/golang-jwt/jwt/v5"

type TokenVerifier interface {
	Verify(tokenStr string) (jwt.MapClaims, error)
	ExtractUserID(claims jwt.MapClaims) (uint, error)
}

type JWTVerifier struct{}

func (v *JWTVerifier) Verify(tokenStr string) (jwt.MapClaims, error) {
	return VerifyToken(tokenStr)
}

func (v *JWTVerifier) ExtractUserID(claims jwt.MapClaims) (uint, error) {
	return ExtractUserIDFromClaims(claims)
}
