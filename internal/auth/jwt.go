package auth

import (
	"crypto/rsa"
	"fmt"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
)

func InitJWTKeys(root string) error {
	privPath := os.Getenv("JWT_PRIVATE_KEY_PATH")
	if privPath == "" {
		return fmt.Errorf("JWT_PRIVATE_KEY_PATH not set")
	}

	pubPath := os.Getenv("JWT_PUBLIC_KEY_PATH")
	if pubPath == "" {
		return fmt.Errorf("JWT_PUBLIC_KEY_PATH not set")
	}
	if root == "" {
		root = "./"
	}

	privData, err := os.ReadFile(path.Join(root, privPath))
	if err != nil {
		return fmt.Errorf("could not read private key: %w", err)
	}

	privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(privData)
	if err != nil {
		return fmt.Errorf("invalid private key: %w", err)
	}

	pubData, err := os.ReadFile(path.Join(root, pubPath))
	if err != nil {
		return fmt.Errorf("could not read public key: %w", err)
	}

	publicKey, err = jwt.ParseRSAPublicKeyFromPEM(pubData)
	if err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}
	return nil
}

/*
func init() {
	privateKeyPath := os.Getenv("JWT_PRIVATE_KEY_PATH")
	publicKeyPath := os.Getenv("JWT_PUBLIC_KEY_PATH")

	if privateKeyPath == "" || publicKeyPath == "" {
		panic("JWT_PRIVATE_KEY_PATH and JWT_PUBLIC_KEY_PATH must be set")
	}

	privBytes, err := os.ReadFile(filepath.Clean(privateKeyPath))
	if err != nil {
		panic(fmt.Sprintf("could not read private key: %v", err))
	}
	privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(privBytes)
	if err != nil {
		panic(fmt.Sprintf("could not parse private key: %v", err))
	}

	pubBytes, err := os.ReadFile(filepath.Clean(publicKeyPath))
	if err != nil {
		panic(fmt.Sprintf("could not read public key: %v", err))
	}
	publicKey, err = jwt.ParseRSAPublicKeyFromPEM(pubBytes)
	if err != nil {
		panic(fmt.Sprintf("could not parse public key: %v", err))
	}
}
*/

/*
	func GenerateTokens(userID uint) (accessToken string, refreshToken string, err error) {
		now := time.Now()
		userIDStr := strconv.FormatUint(uint64(userID), 10)

		accessClaims := jwt.MapClaims{
			"sub": userIDStr,
			"exp": now.Add(time.Minute * 15).Unix(), // 15 minutes access token
			"typ": "access",
		}

		refreshClaims := jwt.MapClaims{
			"sub": userIDStr,
			"exp": now.Add(7 * 24 * time.Hour).Unix(), // 1 week refresh token
			"typ": "refresh",
		}

		access := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims)
		refresh := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims)

		accessToken, err = access.SignedString(privateKey)
		if err != nil {
			return
		}

		refreshToken, err = refresh.SignedString(privateKey)
		return

}
*/
func GenerateTokens(userID uint) (accessToken string, refreshToken string, err error) {
	accessToken, err = GenerateAccessToken(userID)
	if err != nil {
		return
	}

	refreshToken, err = GenerateRefreshToken(userID)
	return
}

func GenerateAccessToken(userID uint) (accessToken string, err error) {
	now := time.Now()
	userIDStr := strconv.FormatUint(uint64(userID), 10)

	accessClaims := jwt.MapClaims{
		"sub": userIDStr,
		"exp": now.Add(time.Minute * 15).Unix(), // 15 minutes access token
		"typ": "access",
	}

	access := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims)
	accessToken, err = access.SignedString(privateKey)
	if err != nil {
		return
	}

	return
}

func GenerateRefreshToken(userID uint) (refreshToken string, err error) {
	now := time.Now()
	userIDStr := strconv.FormatUint(uint64(userID), 10)

	refreshClaims := jwt.MapClaims{
		"sub": userIDStr,
		"exp": now.Add(7 * 24 * time.Hour).Unix(), // 1 week refresh token
		"typ": "refresh",
	}

	refresh := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims)

	refreshToken, err = refresh.SignedString(privateKey)
	return
}

func VerifyToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return publicKey, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	return claims, nil
}

func ExtractUserIDFromClaims(claims jwt.MapClaims) (uint, error) {
	sub, ok := claims["sub"].(string)
	if !ok {
		return 0, fmt.Errorf("sub claim is not a string")
	}

	id64, err := strconv.ParseUint(sub, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid user id in sub claim: %w", err)
	}

	return uint(id64), nil
}

func IsTokenExpired(tokenStr string) (bool, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return false, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false, fmt.Errorf("invalid claims")
	}
	expVal, ok := claims["exp"]
	if !ok {
		return false, fmt.Errorf("no exp claim")
	}
	expFloat, ok := expVal.(float64)
	if !ok {
		return false, fmt.Errorf("exp claim is not a number")
	}
	expTime := time.Unix(int64(expFloat), 0)
	return time.Now().After(expTime), nil
}

var TokenExpiresAt = func(tokenStr string) (*time.Time, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	expVal, ok := claims["exp"]
	if !ok {
		return nil, fmt.Errorf("no exp claim")
	}
	expFloat, ok := expVal.(float64)
	if !ok {
		return nil, fmt.Errorf("exp claim is not a number")
	}
	expTime := time.Unix(int64(expFloat), 0)
	return &expTime, nil
}
