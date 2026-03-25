package authentication_test

import (
	"crypto/ed25519"
	"encoding/base64"
	"strings"
	"testing"
	"time"

	authentication "github.com/chan-jui-huang/go-backend-package/v2/pkg/authentication"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
)

func mustB64Key(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func newAuthFromKeys(pub ed25519.PublicKey, priv ed25519.PrivateKey, accessLife time.Duration) (*authentication.Authenticator, error) {
	cfg := authentication.Config{
		PrivateKey:           mustB64Key(priv),
		PublicKey:            mustB64Key(pub),
		AccessTokenLifeTime:  accessLife,
		RefreshTokenLifeTime: time.Hour,
	}
	return authentication.NewAuthenticator(cfg)
}

func TestIssueAndVerifyJwt(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	auth, err := newAuthFromKeys(pub, priv, time.Hour)
	require.NoError(t, err)

	now := time.Now()
	claims := jwt.RegisteredClaims{
		Audience:  jwt.ClaimStrings{"access"},
		Subject:   "alice",
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	tokStr, err := auth.IssueJwt(claims)
	require.NoError(t, err)

	tok, err := auth.VerifyJwt(tokStr)
	require.NoError(t, err)
	require.True(t, tok.Valid)

	mapClaims, ok := tok.Claims.(jwt.MapClaims)
	require.True(t, ok)
	require.Equal(t, "alice", mapClaims["sub"].(string))
}

func TestIssueAccessTokenExpiration(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	// short lived access token (JWT expiration is second-granularity)
	auth, err := newAuthFromKeys(pub, priv, 1*time.Second)
	require.NoError(t, err)

	at, err := auth.IssueAccessToken("bob")
	require.NoError(t, err)

	// immediately valid
	tok, err := auth.VerifyJwt(at)
	require.NoError(t, err)
	require.True(t, tok.Valid)

	// wait until expired with short polling to keep test fast but stable
	deadline := time.Now().Add(1500 * time.Millisecond)
	for {
		tok2, verifyErr := auth.VerifyJwt(at)
		if verifyErr != nil {
			if tok2 != nil {
				require.False(t, tok2.Valid)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("expected token to expire within deadline")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestInvalidSignatureAndTamper(t *testing.T) {
	// generate two keypairs
	pubA, privA, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	pubB, _, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	authA, err := newAuthFromKeys(pubA, privA, time.Hour)
	require.NoError(t, err)
	// authB uses a mismatched pub key (verification should fail)
	authB, err := newAuthFromKeys(pubB, privA, time.Hour)
	require.NoError(t, err)

	claims := jwt.RegisteredClaims{
		Audience:  jwt.ClaimStrings{"access"},
		Subject:   "mallory",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}

	tokStr, err := authA.IssueJwt(claims)
	require.NoError(t, err)

	// verify with different public key (should fail)
	_, err = authB.VerifyJwt(tokStr)
	require.Error(t, err)

	// tamper the token by modifying payload section
	parts := strings.Split(tokStr, ".")
	require.Len(t, parts, 3)
	if len(parts[1]) > 5 {
		b := []byte(parts[1])
		b[2] = b[2] ^ 1
		parts[1] = string(b)
	}
	tampered := strings.Join(parts, ".")
	_, err = authA.VerifyJwt(tampered)
	require.Error(t, err)
}

func TestVerifyAccessToken(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	auth, err := newAuthFromKeys(pub, priv, time.Hour)
	require.NoError(t, err)

	subject := "john"
	accessToken, err := auth.IssueAccessToken(subject)
	require.NoError(t, err)

	token, err := auth.VerifyJwt(accessToken)
	require.NoError(t, err)
	require.True(t, token.Valid)

	claims := token.Claims.(jwt.MapClaims)
	require.Equal(t, subject, claims["sub"].(string))
	// audience may be string or []any depending on jwt lib, handle both
	if aud, ok := claims["aud"].([]any); ok {
		require.Equal(t, "access", aud[0].(string))
	} else if auds, ok := claims["aud"].(string); ok {
		require.Equal(t, "access", auds)
	}
}

func TestInvalidJwtToken(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	auth, err := newAuthFromKeys(pub, priv, time.Hour)
	require.NoError(t, err)

	subject := "john"
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Audience:  jwt.ClaimStrings{"invalid"},
		Subject:   subject,
		ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Second)),
		IssuedAt:  jwt.NewNumericDate(now),
	}
	invalidJwt, err := auth.IssueJwt(claims)
	require.NoError(t, err)
	// wait until expired
	time.Sleep(time.Second)

	token, err := auth.VerifyJwt(invalidJwt)
	require.Error(t, err)
	require.True(t, token == nil || !token.Valid)
}
