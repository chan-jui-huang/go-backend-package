package random

import (
	"regexp"
	"strconv"
	"testing"
)

func TestRandomStringLengthAndCharset(t *testing.T) {
	t.Parallel()

	cases := []int{0, 1, 5, 64, 100, 1024}
	re := regexp.MustCompile(`^[A-Za-z0-9]*$`)

	for _, n := range cases {
		n := n
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			t.Parallel()
			s := RandomString(n)
			if len(s) != n {
				t.Fatalf("RandomString(%d) length = %d, want %d", n, len(s), n)
			}
			if !re.MatchString(s) {
				t.Fatalf("RandomString(%d) contains invalid chars: %q", n, s)
			}
		})
	}
}

func TestRandomStringUniqueness(t *testing.T) {
	t.Parallel()

	const runs = 50
	const n = 32
	seen := make(map[string]struct{}, runs)

	for i := 0; i < runs; i++ {
		s := RandomString(n)
		if len(s) != n {
			t.Fatalf("RandomString(%d) length = %d, want %d", n, len(s), n)
		}
		seen[s] = struct{}{}
	}

	if len(seen) <= 1 {
		t.Fatalf("RandomString generated no variation across %d runs", runs)
	}
}
