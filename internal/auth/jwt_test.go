package auth_test

import (
	"log"
	"os"
	"path"
	"testing"

	"github.com/amavis442/til-backend/internal/auth"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	root := "../../"
	// Load env file before anything else
	err := godotenv.Load(path.Join(root, ".env.local"))
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	if err := auth.InitJWTKeys(root); err != nil {
		log.Fatalf("failed to load keys: %v", err)
	}

	// Run the tests
	os.Exit(m.Run())
}
func TestEnvLoaded(t *testing.T) {
	t.Log("JWT_PRIVATE_KEY_PATH =", os.Getenv("JWT_PRIVATE_KEY_PATH"))
}

func TestGenerateAndVerifyTokens(t *testing.T) {
	userID := uint(42)

	access, refresh, err := auth.GenerateTokens(userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)

	claims, err := auth.VerifyToken(refresh)
	assert.NoError(t, err)

	id, err := auth.ExtractUserIDFromClaims(claims)
	assert.NoError(t, err)
	assert.Equal(t, userID, id)

	assert.Equal(t, "refresh", claims["typ"])

}
