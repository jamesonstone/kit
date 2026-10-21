package config

import (
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

// SharedMemoryConfig contains references only. Kit never authenticates or fetches them.
type SharedMemoryConfig struct {
	Enabled     bool   `yaml:"enabled,omitempty" json:"enabled"`
	Provider    string `yaml:"provider,omitempty" json:"provider,omitempty"`
	Destination string `yaml:"destination,omitempty" json:"destination,omitempty"`
	Guide       string `yaml:"guide,omitempty" json:"guide,omitempty"`
}

func (c SharedMemoryConfig) Validate() error {
	if c.Provider != "" && c.Provider != "notion" {
		return fmt.Errorf("shared_memory.provider must be notion")
	}
	if c.Enabled && c.Provider != "notion" {
		return fmt.Errorf("enabled shared_memory requires provider notion")
	}
	for _, field := range []struct{ name, value string }{{"destination", c.Destination}, {"guide", c.Guide}} {
		if !c.Enabled && field.value == "" {
			continue
		}
		u, err := url.Parse(field.value)
		if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || strings.ContainsAny(field.value, "<>`\"") || strings.IndexFunc(field.value, unicode.IsSpace) >= 0 || strings.IndexFunc(field.value, unicode.IsControl) >= 0 {
			return fmt.Errorf("shared_memory.%s must be an HTTPS reference without credentials, query parameters or whitespace", field.name)
		}
	}
	return nil
}
