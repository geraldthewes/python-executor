package client

import (
	"archive/tar"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTarFromMap(t *testing.T) {
	files := map[string]string{
		"main.py":    "print('hello')",
		"utils.py":   "# utils",
		"README.md":  "# Project",
	}

	tarData, err := TarFromMap(files)
	require.NoError(t, err)

	// Verify tar contents
	tr := tar.NewReader(bytes.NewReader(tarData))
	found := make(map[string]bool)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		found[header.Name] = true

		// Read content and verify
		content, err := io.ReadAll(tr)
		require.NoError(t, err)
		assert.Equal(t, files[header.Name], string(content))
	}

	// Verify all files are in tar
	for name := range files {
		assert.True(t, found[name], "file %s not found in tar", name)
	}
}

func TestTarFromReader(t *testing.T) {
	content := "print('hello from stdin')"
	reader := strings.NewReader(content)

	tarData, err := TarFromReader(reader, "script.py")
	require.NoError(t, err)

	// Verify tar contents
	tr := tar.NewReader(bytes.NewReader(tarData))
	header, err := tr.Next()
	require.NoError(t, err)

	assert.Equal(t, "script.py", header.Name)

	actualContent, err := io.ReadAll(tr)
	require.NoError(t, err)
	assert.Equal(t, content, string(actualContent))
}

func TestTarFromFiles(t *testing.T) {
	// Create temp files
	tmpDir, err := os.MkdirTemp("", "test-tar-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	file1 := filepath.Join(tmpDir, "test1.py")
	file2 := filepath.Join(tmpDir, "test2.py")

	require.NoError(t, os.WriteFile(file1, []byte("content1"), 0644))
	require.NoError(t, os.WriteFile(file2, []byte("content2"), 0644))

	// Create tar
	tarData, err := TarFromFiles([]string{file1, file2})
	require.NoError(t, err)

	// Verify tar contents
	tr := tar.NewReader(bytes.NewReader(tarData))
	fileCount := 0

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		fileCount++

		// Files should be archived by basename
		assert.Contains(t, []string{"test1.py", "test2.py"}, header.Name)
	}

	assert.Equal(t, 2, fileCount)
}

func TestTarFromDirectory(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "test-dir-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create files
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.py"), []byte("main"), 0644))

	subDir := filepath.Join(tmpDir, "utils")
	require.NoError(t, os.Mkdir(subDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "helper.py"), []byte("helper"), 0644))

	// Create tar
	tarData, err := TarFromDirectory(tmpDir)
	require.NoError(t, err)

	// Verify tar contents
	tr := tar.NewReader(bytes.NewReader(tarData))
	found := make(map[string]bool)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		if header.Typeflag == tar.TypeReg {
			found[header.Name] = true
		}
	}

	assert.True(t, found["main.py"])
	assert.True(t, found["utils/helper.py"])
}

func TestDetectEntrypoint(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		expected string
	}{
		{
			name: "has main.py",
			files: map[string]string{
				"main.py":   "print('main')",
				"script.py": "print('script')",
			},
			expected: "main.py",
		},
		{
			name: "has __main__.py",
			files: map[string]string{
				"__main__.py": "print('main')",
				"script.py":   "print('script')",
			},
			expected: "__main__.py",
		},
		{
			name: "first .py file",
			files: map[string]string{
				"script.py": "print('script')",
			},
			expected: "script.py",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tarData, err := TarFromMap(tt.files)
			require.NoError(t, err)

			entrypoint, err := DetectEntrypoint(tarData)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, entrypoint)
		})
	}
}

func TestDetectEntrypoint_NoFiles(t *testing.T) {
	files := map[string]string{
		"README.md": "# Project",
	}

	tarData, err := TarFromMap(files)
	require.NoError(t, err)

	_, err = DetectEntrypoint(tarData)
	assert.Error(t, err)
}
