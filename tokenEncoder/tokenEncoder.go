package tokenEncoder

import (
	"crypto/rsa"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"os"
	"time"
)

type GroupTokenClaims struct {
	jwt.RegisteredClaims
	Groups []string `json:"groups"`
}

type TokenEncoder interface {
	EncodeGroupToken(groups []string, userId string) (string, error)
	LoadPrivateKey(path string) error
}

type TokenEncoderImpl struct {
	privateKey *rsa.PrivateKey
	expiration time.Duration
}

func NewGroupTokenEncoder(keyPath string, jwtExpirationTime time.Duration) (*TokenEncoderImpl, error) {
	tknEncoder := &TokenEncoderImpl{
		privateKey: nil,
		expiration: jwtExpirationTime,
	}

	err := tknEncoder.LoadPrivateKey(keyPath)
	return tknEncoder, err
}

func NewGroupTokenEncoderWithKey(key *rsa.PrivateKey, jwtExpirationTime time.Duration) *TokenEncoderImpl {
	return &TokenEncoderImpl{
		privateKey: key,
		expiration: jwtExpirationTime,
	}
}

func (t *TokenEncoderImpl) EncodeGroupToken(groups []string, userId string) (string, error) {
	groupToken := jwt.NewWithClaims(jwt.SigningMethodRS256, GroupTokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "ermes-central-repository",
			Subject:   userId,
			Audience:  nil,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.expiration)),
			NotBefore: nil,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.NewString(),
		},
		Groups: groups,
	})

	return groupToken.SignedString(t.privateKey)
}

func (t *TokenEncoderImpl) LoadPrivateKey(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file: %s", err)
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM(data)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %s", err)
	}

	t.privateKey = key
	return nil
}
