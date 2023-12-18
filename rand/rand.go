package rand

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func RandBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	nRead, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("bytes: %w", err)
	}
	if nRead < n {
		return nil, fmt.Errorf("did not read enough bytes")
	}
	return b, nil
}

// String returns a random string using crypto/rand
// n is the number of bytes being used to generate the random string.
func String(n int) (string, error) {
	b, err := RandBytes(n)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
