// Package kit embeds the rule corpus shipped with each Kit release so a
// binary always installs the rules it was built and tested with.
package kit

import "embed"

// Rules holds docs/references/rules/*.md at build time.
//
//go:embed docs/references/rules/*.md
var Rules embed.FS

// RulesDir is the directory inside Rules that holds the rule files.
const RulesDir = "docs/references/rules"
