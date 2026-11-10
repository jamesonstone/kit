package templates

import _ "embed"

// SOPGuidePath is the project-owned entry point for operating procedures.
const SOPGuidePath = "docs/sops/README.md"

// SOPGuide seeds the procedure directory. Existing project guides are preserved.
//
//go:embed sops.md
var SOPGuide string
