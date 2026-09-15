package instructions

import _ "embed"

// ThreadNamingPolicy is the shared installed and printable naming contract.
//
//go:embed thread_naming.md
var ThreadNamingPolicy string
