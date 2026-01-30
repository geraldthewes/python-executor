package imports

import (
	"regexp"
	"strings"
)

// importPattern matches "import X" and "import X as Y" statements
// Captures the module name(s) after "import"
var importPattern = regexp.MustCompile(`(?m)^[ \t]*import\s+([^\n#]+)`)

// fromImportPattern matches "from X import Y" statements
// Captures the module name after "from"
var fromImportPattern = regexp.MustCompile(`(?m)^[ \t]*from\s+(\S+)\s+import\s+`)

// stringPattern matches string literals (to exclude imports inside strings)
var stringPattern = regexp.MustCompile(`(?s)'''.*?'''|""".*?"""|'[^'\n]*'|"[^"\n]*"`)

// commentPattern matches comments (to exclude imports in comments)
var commentPattern = regexp.MustCompile(`(?m)#.*$`)

// ParseImports extracts all imported module names from Python code.
// It handles:
// - import X
// - import X as Y
// - import X, Y, Z
// - from X import Y
// - from X.submodule import Y
//
// It ignores imports inside string literals and comments.
func ParseImports(code string) []string {
	// Remove string literals first to avoid matching imports inside strings
	cleanCode := stringPattern.ReplaceAllString(code, "")

	// Remove comments to avoid matching imports in comments
	cleanCode = commentPattern.ReplaceAllString(cleanCode, "")

	modules := make(map[string]bool)

	// Match "import X" patterns
	matches := importPattern.FindAllStringSubmatch(cleanCode, -1)
	for _, match := range matches {
		if len(match) > 1 {
			// Handle "import X, Y, Z" and "import X as alias"
			parts := strings.Split(match[1], ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				// Handle "X as Y" - extract just X
				if idx := strings.Index(part, " as "); idx > 0 {
					part = part[:idx]
				}
				part = strings.TrimSpace(part)
				if part != "" && isValidModuleName(part) {
					// Extract top-level module
					topLevel := extractTopLevel(part)
					modules[topLevel] = true
				}
			}
		}
	}

	// Match "from X import Y" patterns
	fromMatches := fromImportPattern.FindAllStringSubmatch(cleanCode, -1)
	for _, match := range fromMatches {
		if len(match) > 1 {
			module := strings.TrimSpace(match[1])
			if module != "" && isValidModuleName(module) {
				// Extract top-level module (e.g., "sklearn" from "sklearn.model_selection")
				topLevel := extractTopLevel(module)
				modules[topLevel] = true
			}
		}
	}

	// Convert map to slice
	result := make([]string, 0, len(modules))
	for module := range modules {
		result = append(result, module)
	}

	return result
}

// extractTopLevel extracts the top-level module from a dotted name.
// e.g., "sklearn.model_selection" -> "sklearn"
func extractTopLevel(module string) string {
	if idx := strings.Index(module, "."); idx > 0 {
		return module[:idx]
	}
	return module
}

// isValidModuleName checks if a string is a valid Python module name.
// Module names must start with a letter or underscore and contain only
// letters, numbers, underscores, and dots (for submodules).
func isValidModuleName(name string) bool {
	if name == "" {
		return false
	}

	// Check first character
	first := name[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || first == '_') {
		return false
	}

	// Check remaining characters
	for i := 1; i < len(name); i++ {
		c := name[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '.') {
			return false
		}
	}

	return true
}
