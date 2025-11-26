package crypto

import (
	"context"
	"crypto/rand"
	"fmt"
)

// RandomDigitsGenerator produce codigos numericos de la longitud indicada.
type RandomDigitsGenerator struct {
	Length int
}

func (g RandomDigitsGenerator) Generate(ctx context.Context, userID string) (string, error) {
	n := g.Length
	if n <= 0 {
		n = 6
	}
	max := 1
	for i := 0; i < n; i++ {
		max *= 10
	}
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	val := 0
	for _, v := range b {
		val = (val*256 + int(v)) % max
	}
	format := fmt.Sprintf("%%0%dd", n)
	return fmt.Sprintf(format, val), nil
}
