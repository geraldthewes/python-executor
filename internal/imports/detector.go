package imports

import (
	"sort"
	"strings"
)

// DetectRequirements analyzes Python code and returns a requirements.txt string
// containing all detected third-party packages.
//
// The function:
// 1. Parses import statements from the code
// 2. Filters out standard library modules
// 3. Maps module names to pip package names (e.g., PIL -> Pillow)
// 4. Returns a newline-separated list of packages
//
// If no third-party packages are detected, an empty string is returned.
func DetectRequirements(code string) string {
	// Parse all imports from the code
	modules := ParseImports(code)

	// Filter and map to package names
	packages := make(map[string]bool)
	for _, module := range modules {
		// Skip stdlib modules
		if IsStdlib(module) {
			continue
		}

		// Map to pip package name
		pkg := GetPackageName(module)
		packages[pkg] = true
	}

	if len(packages) == 0 {
		return ""
	}

	// Convert to sorted slice for deterministic output
	result := make([]string, 0, len(packages))
	for pkg := range packages {
		result = append(result, pkg)
	}
	sort.Strings(result)

	return strings.Join(result, "\n")
}

// MergeRequirements merges auto-detected requirements with user-provided ones.
// User-provided requirements take precedence (appear first, may have version pins).
func MergeRequirements(detected, userProvided string) string {
	if userProvided == "" {
		return detected
	}
	if detected == "" {
		return userProvided
	}

	// Parse user-provided packages (may include version specifiers)
	userPackages := make(map[string]bool)
	for _, line := range strings.Split(userProvided, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Extract package name (before any version specifier)
		pkgName := extractPackageName(line)
		userPackages[strings.ToLower(pkgName)] = true
	}

	// Add detected packages that aren't already in user-provided
	var result strings.Builder
	result.WriteString(userProvided)

	for _, line := range strings.Split(detected, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		pkgName := extractPackageName(line)
		if !userPackages[strings.ToLower(pkgName)] {
			result.WriteString("\n")
			result.WriteString(line)
		}
	}

	return result.String()
}

// extractPackageName extracts the package name from a requirements line.
// e.g., "requests>=2.28.0" -> "requests"
func extractPackageName(line string) string {
	// Find first occurrence of version specifiers
	for i, c := range line {
		if c == '=' || c == '>' || c == '<' || c == '!' || c == '[' || c == ';' {
			return strings.TrimSpace(line[:i])
		}
	}
	return strings.TrimSpace(line)
}
