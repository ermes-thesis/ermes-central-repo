package tokenEncoder

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestEncodeGroupToken_Success(t *testing.T) {
	jwtPemPath := "../keys/private.pem"
	jwtExpirationTime := 24 * time.Hour

	tknEncoder, err := NewGroupTokenEncoder(jwtPemPath, jwtExpirationTime)
	require.NoError(t, err)
	groups := []string{"groupA", "groupB", "groupC"}
	userId := "randomUserUnique"

	signedTokenStr, err := tknEncoder.EncodeGroupToken(groups, userId)
	require.NoError(t, err)

	token, err := jwt.ParseWithClaims(signedTokenStr, &GroupTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		pKeyBytes, err := os.ReadFile("../keys/public.pem")
		if err != nil {
			return nil, err
		}
		return jwt.ParseRSAPublicKeyFromPEM(pKeyBytes)
	})
	require.NoError(t, err)
	require.True(t, token.Valid)
	customClaims, ok := token.Claims.(*GroupTokenClaims)
	require.True(t, ok)

	require.Equal(t, userId, customClaims.Subject)
	require.Equal(t, groups, customClaims.Groups)
	require.Equal(t, "ermes-central-repository", customClaims.Issuer)
	require.True(t, customClaims.ExpiresAt.After(time.Now()))

}
