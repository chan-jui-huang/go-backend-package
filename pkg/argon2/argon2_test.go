package argon2

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"
	"time"

	argon2lib "golang.org/x/crypto/argon2"

	"github.com/stretchr/testify/require"
)

func TestMakeAndVerifyTableDriven(t *testing.T) {
	cases := []struct {
		name     string
		password string
	}{
		{"normal", "password"},
		{"empty", ""},
		{"long", strings.Repeat("a", 10000)},
	}

	for _, tc := range cases {
		c := tc // copy for safety
		t.Run(c.name, func(t *testing.T) {
			hash := MakeArgon2IdHash(c.password)
			require.NotEmpty(t, hash)
			// correct password verifies
			require.True(t, VerifyArgon2IdHash(c.password, hash))
			// wrong password does not verify
			require.False(t, VerifyArgon2IdHash(c.password+"x", hash))
		})
	}
}

func TestMakeArgon2IdHashWithConfigPHCString(t *testing.T) {
	cfg := &Argon2IdConfig{
		Memory:  32 * 1024,
		Time:    1,
		Threads: 1,
		KeyLen:  16,
		SaltLen: 8,
	}
	password := "phc-test"
	hash := MakeArgon2IdHashWithConfig(password, cfg)
	require.NotEmpty(t, hash)

	// PHC string: $argon2id$v=19$m=...,t=...,p=...$<salt>$<hash>
	parts := strings.Split(hash, "$")
	// parts[0]=="", parts[1]=="argon2id", parts[2]=="v=...", parts[3]=="m=...,t=...,p=...", parts[4]==salt, parts[5]==hash
	require.Len(t, parts, 6)
	require.Equal(t, "argon2id", parts[1])

	var ver int
	_, err := fmt.Sscanf(parts[2], "v=%d", &ver)
	require.NoError(t, err)
	require.Equal(t, argon2lib.Version, ver)

	var mem uint32
	var tim uint32
	var p uint8
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &mem, &tim, &p)
	require.NoError(t, err)
	require.Equal(t, cfg.Memory, mem)
	require.Equal(t, cfg.Time, tim)
	require.Equal(t, cfg.Threads, p)

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	require.NoError(t, err)
	require.Equal(t, int(cfg.SaltLen), len(salt))

	outHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	require.NoError(t, err)
	require.Equal(t, int(cfg.KeyLen), len(outHash))

	// verify should succeed
	require.True(t, VerifyArgon2IdHash(password, hash))
}

func TestVerifyArgon2IdHashConstantTimeLike(t *testing.T) {
	// Use a small but realistic config to keep test fast
	cfg := &Argon2IdConfig{Memory: 32 * 1024, Time: 1, Threads: 1, KeyLen: 16, SaltLen: 8}
	password := "timing-test"
	hash := MakeArgon2IdHashWithConfig(password, cfg)
	require.True(t, VerifyArgon2IdHash(password, hash))

	iters := 50
	var durGood time.Duration
	var durBad time.Duration
	for i := 0; i < iters; i++ {
		s := time.Now()
		VerifyArgon2IdHash(password, hash)
		durGood += time.Since(s)

		s = time.Now()
		VerifyArgon2IdHash(password+"x", hash)
		durBad += time.Since(s)
	}
	avgGood := durGood / time.Duration(iters)
	avgBad := durBad / time.Duration(iters)
	// Ensure both measured and that bad isn't wildly faster/slower than good (heuristic)
	require.True(t, avgGood > 0)
	require.True(t, avgBad > 0)
	// Allow some variance; require ratio less than 5x to catch obvious non-constant-time implementations
	ratio := float64(avgBad) / float64(avgGood)
	require.Less(t, ratio, 5.0)
}
