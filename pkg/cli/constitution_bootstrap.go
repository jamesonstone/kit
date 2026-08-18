package cli

import (
	"strings"

	"github.com/jamesonstone/kit/v3/internal/templates"
)

func isBootstrapConstitution(content string) bool {
	return strings.TrimSpace(content) == strings.TrimSpace(templates.Constitution)
}
