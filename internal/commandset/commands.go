package commandset

import (
	"sort"
	"strings"
)

var protectedPaths = []string{
	"init",
	"spec",
	"context resolve",
	"usage",
	"usage report",
	"usage status",
	"usage refresh",
	"usage clear",
	"usage enable",
	"usage disable",
	"status",
	"registry status",
	"health",
	"capabilities",
	"config check",
	"aws verify",
	"check",
	"pr fix",
	"pr orchestrate",
	"improve run",
	"rules add",
	"rules list",
	"rules view",
	"rules link",
	"reconcile",
	"dispatch",
	"instructions",
	"upgrade",
	"version",
	"completion",
}

var telemetryPaths = []string{
	"init",
	"spec",
	"context resolve",
	"status",
	"registry status",
	"health",
	"capabilities",
	"config check",
	"aws verify",
	"check",
	"pr fix",
	"pr orchestrate",
	"improve run",
	"rules add",
	"rules list",
	"rules view",
	"rules link",
	"reconcile",
	"dispatch",
	"instructions",
	"upgrade",
	"version",
	"completion",
}

func ProtectedPaths() []string {
	return append([]string(nil), protectedPaths...)
}

func TelemetryPaths() []string {
	return append([]string(nil), telemetryPaths...)
}

func IsTelemetryPath(path string) bool {
	for _, candidate := range telemetryPaths {
		if candidate == path {
			return true
		}
	}
	return false
}

func IsProtectedOrParent(path string) bool {
	for _, candidate := range protectedPaths {
		if candidate == path || strings.HasPrefix(candidate, path+" ") {
			return true
		}
	}
	return false
}

func RootNames() []string {
	seen := map[string]bool{}
	for _, path := range protectedPaths {
		for index, character := range path {
			if character == ' ' {
				seen[path[:index]] = true
				break
			}
		}
		if !containsSpace(path) {
			seen[path] = true
		}
	}
	values := make([]string, 0, len(seen))
	for value := range seen {
		values = append(values, value)
	}
	sort.Strings(values)
	return values
}

func containsSpace(value string) bool {
	for _, character := range value {
		if character == ' ' {
			return true
		}
	}
	return false
}
