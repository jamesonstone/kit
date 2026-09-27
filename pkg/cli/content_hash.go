package cli

import (
	"crypto/sha256"
	"encoding/hex"
)

func contentHash(content string) string {
	sum := sha256.Sum256([]byte(content))
	return "sha256:" + hex.EncodeToString(sum[:])
}
