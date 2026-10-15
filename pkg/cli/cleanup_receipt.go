package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/jamesonstone/kit/v3/internal/cleanup"
)

func decodeCleanupReceipt(reader io.Reader) (cleanup.Receipt, error) {
	var receipt cleanup.Receipt
	const maximum = 1 << 20
	data, err := io.ReadAll(io.LimitReader(reader, maximum+1))
	if err != nil {
		return receipt, err
	}
	if len(data) > maximum {
		return receipt, fmt.Errorf("ownership receipt exceeds size limit")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return receipt, fmt.Errorf("ownership receipt must be an object")
	}
	seen := map[string]bool{}
	for decoder.More() {
		token, err = decoder.Token()
		key, ok := token.(string)
		normalized := strings.ToLower(key)
		if err != nil || !ok || seen[normalized] {
			return receipt, fmt.Errorf("ambiguous ownership receipt field")
		}
		seen[normalized] = true
		var value json.RawMessage
		if err = decoder.Decode(&value); err != nil {
			return receipt, err
		}
	}
	if _, err = decoder.Token(); err != nil {
		return receipt, err
	}
	if err = decoder.Decode(new(any)); err != io.EOF {
		return receipt, fmt.Errorf("ownership receipt must contain exactly one JSON object")
	}
	decoder = json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&receipt); err != nil {
		return receipt, fmt.Errorf("invalid ownership receipt: %w", err)
	}
	return receipt, nil
}
