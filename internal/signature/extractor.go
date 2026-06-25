package signature

import (
	"regexp"
	"strings"
)

// ExtractErrorType identifies the error type from an error message.
func ExtractErrorType(errorMsg string) string {
	patterns := map[string]string{
		"NullPointerException":    "NullReferenceError",
		"NullReference":           "NullReferenceError",
		"nil pointer":             "NullReferenceError",
		"cannot unmarshal":        "TypeError",
		"type error":              "TypeError",
		"TypeError":               "TypeError",
		"AssertionError":          "AssertionError",
		"assert":                  "AssertionError",
		"timeout":                 "TimeoutError",
		"TimeoutError":            "TimeoutError",
		"deadline exceeded":       "TimeoutError",
		"connection refused":      "ConnectionError",
		"ConnectionError":         "ConnectionError",
		"no route to host":        "ConnectionError",
		"syntax error":            "SyntaxError",
		"SyntaxError":             "SyntaxError",
		"import error":            "ImportError",
		"ImportError":             "ImportError",
		"no module named":         "ImportError",
		"cannot find package":     "ImportError",
		"permission denied":       "PermissionError",
		"PermissionError":         "PermissionError",
		"key error":               "KeyError",
		"KeyError":                "KeyError",
		"index out of range":      "IndexError",
		"IndexError":              "IndexError",
		"panic":                   "PanicError",
		"segmentation fault":      "Segfault",
		"SIGSEGV":                 "Segfault",
	}

	lower := strings.ToLower(errorMsg)
	for pattern, errType := range patterns {
		if strings.Contains(lower, strings.ToLower(pattern)) {
			return errType
		}
	}
	return "UnknownError"
}

// NormalizeErrorMessage normalizes an error message for comparison.
func NormalizeErrorMessage(msg string) string {
	// Remove numbers (line numbers, ports, IDs, etc.)
	re := regexp.MustCompile(`\d+`)
	normalized := re.ReplaceAllString(msg, "N")

	// Remove quoted strings
	re = regexp.MustCompile(`"[^"]*"`)
	normalized = re.ReplaceAllString(normalized, `"STR"`)

	// Remove single-quoted strings
	re = regexp.MustCompile(`'[^']*'`)
	normalized = re.ReplaceAllString(normalized, "'STR'")

	// Collapse whitespace
	re = regexp.MustCompile(`\s+`)
	normalized = re.ReplaceAllString(normalized, " ")

	return strings.TrimSpace(normalized)
}

// ExtractLocation extracts the primary error location from an error message.
func ExtractLocation(errorMsg string) string {
	// Common file:line patterns
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`File "([^"]+)", line \d+`),               // Python
		regexp.MustCompile(`([^\s:]+\.(go|py|js|ts|rs|java)):\d+`), // Go/Node/Rust/Java
		regexp.MustCompile(`at ([^\s]+)\(([^:]+):\d+:\d+\)`),        // JS stack
		regexp.MustCompile(`([^\s]+\.(go|py|js|ts|rs|java)):\d+:\d+`), // With column
	}

	for _, re := range patterns {
		matches := re.FindStringSubmatch(errorMsg)
		if len(matches) > 1 {
			return matches[1]
		}
	}

	// Fallback: look for any file extension pattern
	re := regexp.MustCompile(`([^\s:]+\.(go|py|js|ts|rs|java|rb|c|cpp|h))`)
	matches := re.FindStringSubmatch(errorMsg)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// ExtractStackHashes extracts stack trace hashes from an error message.
func ExtractStackHashes(errorMsg string) []uint64 {
	var hashes []uint64
	seen := make(map[uint64]bool)

	// Extract file:line pairs and hash them
	re := regexp.MustCompile(`([^\s:]+\.(go|py|js|ts|rs|java)):\d+`)
	matches := re.FindAllString(errorMsg, -1)
	for _, m := range matches {
		h := HashString(m)
		if !seen[h] {
			hashes = append(hashes, h)
			seen[h] = true
		}
	}
	return hashes
}

// ExtractChangeType identifies the type of change that caused an error.
func ExtractChangeType(changedFiles []string) string {
	if len(changedFiles) == 0 {
		return "unknown"
	}
	// Simple heuristic based on file extensions and names
	for _, f := range changedFiles {
		lower := strings.ToLower(f)
		if strings.Contains(lower, "test") {
			return "test"
		}
		if strings.Contains(lower, "config") || strings.Contains(lower, ".yaml") || strings.Contains(lower, ".toml") {
			return "config"
		}
		if strings.Contains(lower, "package") || strings.Contains(lower, "go.mod") || strings.Contains(lower, "requirements") {
			return "dependency"
		}
	}
	return "logic_fix"
}

// DistanceToError determines how far the error location is from the changes.
func DistanceToError(location string, changedFiles []string) int {
	if location == "" || len(changedFiles) == 0 {
		return 2 // different module
	}
	locLower := strings.ToLower(location)
	for _, f := range changedFiles {
		fLower := strings.ToLower(f)
		if fLower == locLower {
			return 0 // same file
		}
		// Check if same directory
		locParts := strings.Split(locLower, "/")
		fParts := strings.Split(fLower, "/")
		if len(locParts) > 1 && len(fParts) > 1 {
			if strings.Join(locParts[:len(locParts)-1], "/") == strings.Join(fParts[:len(fParts)-1], "/") {
				return 1 // same module
			}
		}
	}
	return 2 // different module
}
