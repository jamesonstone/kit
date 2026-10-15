package cleanup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// controlFingerprint proves canonical machine-generated control state before its removal/update.
func controlFingerprint(ctx context.Context, path string) (string, error) {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if info.Mode() != 0644 {
		return "", keep("KEEP_LOCAL_STATE", "noncanonical Git control file permissions or link")
	}
	metadata, err := treeMetadata(path, info, ctx)
	if err != nil {
		return "", keep("KEEP_LOCAL_STATE", "Git control attributes require preservation")
	}
	digest, err := physicalContentDigest(path, info, ctx)
	if err != nil {
		return "", err
	}
	return fingerprint(metadata, digest), nil
}

func requireCanonicalOID(data string) error {
	if len(data) != 40 && len(data) != 64 {
		return fmt.Errorf("invalid object identity")
	}
	for _, c := range data {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			return fmt.Errorf("invalid object identity")
		}
	}
	return nil
}

func controlStates(ctx context.Context, root string, names ...string) (string, error) {
	var values []string
	for _, name := range names {
		state, err := controlFingerprint(ctx, filepath.Join(root, filepath.FromSlash(name)))
		if err != nil {
			return "", err
		}
		values = append(values, state)
	}
	return fingerprint(values...), nil
}
