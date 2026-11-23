package mintassets

import _ "embed"

// ResolveRelease selects the official latest stable Mint tag for workflow checkout.
//
//go:embed resolve-release.sh
var ResolveRelease string
