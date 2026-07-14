package auth

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ideamans/apple-app-store-connect/internal/config"
)

// MaxTTL is the maximum token lifetime App Store Connect accepts.
const MaxTTL = 20 * time.Minute

// DefaultTTL leaves headroom for clock skew.
const DefaultTTL = 15 * time.Minute

// ParsePrivateKey parses a PKCS#8 PEM-encoded ECDSA private key (.p8 contents).
func ParsePrivateKey(pemData []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("no PEM block found in private key")
	}
	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse PKCS#8 private key: %w", err)
	}
	key, ok := parsed.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is %T, want *ecdsa.PrivateKey", parsed)
	}
	return key, nil
}

// Token signs a JWT for the App Store Connect API.
func Token(creds *config.Credentials, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		ttl = DefaultTTL
	}
	if ttl > MaxTTL {
		return "", fmt.Errorf("token TTL %s exceeds the API maximum of %s", ttl, MaxTTL)
	}
	key, err := ParsePrivateKey(creds.PrivateKeyPEM)
	if err != nil {
		return "", err
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": creds.IssuerID,
		"iat": now.Unix(),
		"exp": now.Add(ttl).Unix(),
		"aud": "appstoreconnect-v1",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = creds.KeyID
	return token.SignedString(key)
}
