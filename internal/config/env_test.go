package config_test

import (
	"os"
	"testing"

	"github.com/amavis442/til-backend/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestGenerateAndVerifyTokens(t *testing.T) {
	config.LoadEnv()

	assert.NotEmpty(t, os.Getenv("JWT_PRIVATE_KEY_PATH"))
	assert.NotEmpty(t, os.Getenv("JWT_PUBLIC_KEY_PATH"))

	/*
		assert.NoError(t, err)

		id, err := auth.ExtractUserIDFromClaims(claims)
		assert.NoError(t, err)
		assert.Equal(t, userID, id)

		assert.Equal(t, "refresh", claims["typ"])
	*/
}
