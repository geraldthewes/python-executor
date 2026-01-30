package imports

import (
	"sort"
	"strings"
	"testing"
)

func TestParseImports(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected []string
	}{
		{
			name:     "simple import",
			code:     "import numpy",
			expected: []string{"numpy"},
		},
		{
			name:     "import with alias",
			code:     "import numpy as np",
			expected: []string{"numpy"},
		},
		{
			name:     "multiple imports on one line",
			code:     "import os, sys, json",
			expected: []string{"os", "sys", "json"},
		},
		{
			name:     "from import",
			code:     "from PIL import Image",
			expected: []string{"PIL"},
		},
		{
			name:     "from submodule import",
			code:     "from sklearn.model_selection import train_test_split",
			expected: []string{"sklearn"},
		},
		{
			name:     "multiple import statements",
			code:     "import numpy\nimport pandas\nfrom scipy import stats",
			expected: []string{"numpy", "pandas", "scipy"},
		},
		{
			name:     "import in comment should be ignored",
			code:     "# import fake_module\nimport real_module",
			expected: []string{"real_module"},
		},
		{
			name:     "import in string should be ignored",
			code:     "x = \"import fake_module\"\nimport real_module",
			expected: []string{"real_module"},
		},
		{
			name:     "import in triple-quoted string should be ignored",
			code:     "\"\"\"import fake_module\"\"\"\nimport real_module",
			expected: []string{"real_module"},
		},
		{
			name:     "indented import",
			code:     "if True:\n    import pandas",
			expected: []string{"pandas"},
		},
		{
			name:     "tab-indented import",
			code:     "if True:\n\timport pandas",
			expected: []string{"pandas"},
		},
		{
			name:     "complex sympy example",
			code: `import sympy as sp
from sympy.physics import constants as const
from sympy import symbols, sqrt, pi, Rational`,
			expected: []string{"sympy"},
		},
		{
			name:     "empty code",
			code:     "",
			expected: []string{},
		},
		{
			name:     "no imports",
			code:     "x = 1\ny = 2\nprint(x + y)",
			expected: []string{},
		},
		{
			name:     "deep nested module",
			code:     "from tensorflow.keras.layers import Dense",
			expected: []string{"tensorflow"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseImports(tt.code)
			sort.Strings(result)
			sort.Strings(tt.expected)

			if len(result) != len(tt.expected) {
				t.Errorf("ParseImports() returned %d modules, want %d\ngot: %v\nwant: %v",
					len(result), len(tt.expected), result, tt.expected)
				return
			}

			for i, mod := range result {
				if mod != tt.expected[i] {
					t.Errorf("ParseImports() module %d = %q, want %q", i, mod, tt.expected[i])
				}
			}
		})
	}
}

func TestIsStdlib(t *testing.T) {
	stdlibModules := []string{
		"os", "sys", "json", "re", "math", "datetime", "collections",
		"itertools", "functools", "typing", "pathlib", "subprocess",
		"asyncio", "threading", "multiprocessing", "urllib", "http",
		"socket", "ssl", "hashlib", "base64", "pickle", "sqlite3",
		"csv", "xml", "html", "email", "logging", "unittest", "io",
	}

	for _, mod := range stdlibModules {
		if !IsStdlib(mod) {
			t.Errorf("IsStdlib(%q) = false, want true", mod)
		}
	}

	thirdPartyModules := []string{
		"numpy", "pandas", "requests", "flask", "django", "tensorflow",
		"scipy", "matplotlib", "PIL", "cv2", "sklearn", "sympy",
	}

	for _, mod := range thirdPartyModules {
		if IsStdlib(mod) {
			t.Errorf("IsStdlib(%q) = true, want false", mod)
		}
	}
}

func TestGetPackageName(t *testing.T) {
	tests := []struct {
		module   string
		expected string
	}{
		{"PIL", "Pillow"},
		{"cv2", "opencv-python"},
		{"sklearn", "scikit-learn"},
		{"bs4", "beautifulsoup4"},
		{"yaml", "PyYAML"},
		{"dotenv", "python-dotenv"},
		{"dateutil", "python-dateutil"},
		{"psycopg2", "psycopg2-binary"},
		// Modules without special mapping should return as-is
		{"requests", "requests"},
		{"numpy", "numpy"},
		{"pandas", "pandas"},
		{"unknown_module", "unknown_module"},
	}

	for _, tt := range tests {
		t.Run(tt.module, func(t *testing.T) {
			result := GetPackageName(tt.module)
			if result != tt.expected {
				t.Errorf("GetPackageName(%q) = %q, want %q", tt.module, result, tt.expected)
			}
		})
	}
}

func TestDetectRequirements(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected []string // Expected packages (sorted)
	}{
		{
			name:     "single third-party import",
			code:     "import numpy",
			expected: []string{"numpy"},
		},
		{
			name:     "multiple third-party imports",
			code:     "import numpy\nimport pandas\nimport requests",
			expected: []string{"numpy", "pandas", "requests"},
		},
		{
			name:     "mixed stdlib and third-party",
			code:     "import os\nimport sys\nimport numpy\nimport json\nimport pandas",
			expected: []string{"numpy", "pandas"},
		},
		{
			name:     "only stdlib imports",
			code:     "import os\nimport sys\nimport json\nimport re",
			expected: []string{},
		},
		{
			name:     "module name mapping",
			code:     "from PIL import Image\nfrom sklearn.model_selection import train_test_split",
			expected: []string{"Pillow", "scikit-learn"},
		},
		{
			name: "complex real-world example",
			code: `import sympy as sp
from sympy.physics import constants as const
from sympy import symbols, sqrt, pi, Rational
import numpy as np
from scipy import stats
import os
import json`,
			expected: []string{"numpy", "scipy", "sympy"},
		},
		{
			name:     "empty code returns empty string",
			code:     "",
			expected: []string{},
		},
		{
			name:     "yaml mapping",
			code:     "import yaml",
			expected: []string{"PyYAML"},
		},
		{
			name:     "beautifulsoup mapping",
			code:     "from bs4 import BeautifulSoup",
			expected: []string{"beautifulsoup4"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectRequirements(tt.code)

			// Parse result into sorted slice
			var resultPkgs []string
			if result != "" {
				resultPkgs = strings.Split(result, "\n")
				sort.Strings(resultPkgs)
			}

			sort.Strings(tt.expected)

			if len(resultPkgs) != len(tt.expected) {
				t.Errorf("DetectRequirements() returned %d packages, want %d\ngot: %v\nwant: %v",
					len(resultPkgs), len(tt.expected), resultPkgs, tt.expected)
				return
			}

			for i, pkg := range resultPkgs {
				if pkg != tt.expected[i] {
					t.Errorf("DetectRequirements() package %d = %q, want %q", i, pkg, tt.expected[i])
				}
			}
		})
	}
}

func TestMergeRequirements(t *testing.T) {
	tests := []struct {
		name         string
		detected     string
		userProvided string
		wantContains []string
	}{
		{
			name:         "detected only",
			detected:     "numpy\npandas",
			userProvided: "",
			wantContains: []string{"numpy", "pandas"},
		},
		{
			name:         "user provided only",
			detected:     "",
			userProvided: "requests>=2.28.0",
			wantContains: []string{"requests>=2.28.0"},
		},
		{
			name:         "merge without overlap",
			detected:     "numpy\npandas",
			userProvided: "requests>=2.28.0",
			wantContains: []string{"requests>=2.28.0", "numpy", "pandas"},
		},
		{
			name:         "user provided takes precedence",
			detected:     "numpy\npandas",
			userProvided: "numpy==1.24.0",
			wantContains: []string{"numpy==1.24.0", "pandas"},
		},
		{
			name:         "case insensitive merge",
			detected:     "Flask\nPillow",
			userProvided: "flask>=2.0.0",
			wantContains: []string{"flask>=2.0.0", "Pillow"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeRequirements(tt.detected, tt.userProvided)

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("MergeRequirements() missing %q\ngot: %q", want, result)
				}
			}

			// Check that user-provided numpy==1.24.0 means detected numpy is not duplicated
			if tt.name == "user provided takes precedence" {
				lines := strings.Split(result, "\n")
				numpyCount := 0
				for _, line := range lines {
					if strings.HasPrefix(strings.ToLower(strings.TrimSpace(line)), "numpy") {
						numpyCount++
					}
				}
				if numpyCount != 1 {
					t.Errorf("MergeRequirements() has %d numpy entries, want 1\ngot: %q", numpyCount, result)
				}
			}
		})
	}
}

func TestExtractPackageName(t *testing.T) {
	tests := []struct {
		line     string
		expected string
	}{
		{"requests", "requests"},
		{"requests>=2.28.0", "requests"},
		{"numpy==1.24.0", "numpy"},
		{"pandas<2.0", "pandas"},
		{"flask!=1.0.0", "flask"},
		{"package[extra]", "package"},
		{"package; python_version >= '3.8'", "package"},
		{"  spaces  ", "spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			result := extractPackageName(tt.line)
			if result != tt.expected {
				t.Errorf("extractPackageName(%q) = %q, want %q", tt.line, result, tt.expected)
			}
		})
	}
}

func TestIsValidModuleName(t *testing.T) {
	validNames := []string{
		"numpy", "pandas", "PIL", "_private", "mod123", "my_module",
		"sklearn.model_selection", "tensorflow.keras",
	}

	for _, name := range validNames {
		if !isValidModuleName(name) {
			t.Errorf("isValidModuleName(%q) = false, want true", name)
		}
	}

	invalidNames := []string{
		"", "123module", "-dash", "space name", "special@char",
	}

	for _, name := range invalidNames {
		if isValidModuleName(name) {
			t.Errorf("isValidModuleName(%q) = true, want false", name)
		}
	}
}
