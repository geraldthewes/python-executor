package tar

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractToDir(t *testing.T) {
	// Create a test tar archive
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	// Add a file
	content := []byte("print('hello')")
	header := &tar.Header{
		Name: "main.py",
		Mode: 0644,
		Size: int64(len(content)),
	}
	require.NoError(t, tw.WriteHeader(header))
	_, err := tw.Write(content)
	require.NoError(t, err)

	// Add a directory
	dirHeader := &tar.Header{
		Name:     "subdir/",
		Mode:     0755,
		Typeflag: tar.TypeDir,
	}
	require.NoError(t, tw.WriteHeader(dirHeader))

	// Add a file in subdirectory
	subContent := []byte("# utils")
	subHeader := &tar.Header{
		Name: "subdir/utils.py",
		Mode: 0644,
		Size: int64(len(subContent)),
	}
	require.NoError(t, tw.WriteHeader(subHeader))
	_, err = tw.Write(subContent)
	require.NoError(t, err)

	require.NoError(t, tw.Close())

	// Extract to temp directory
	tmpDir, err := os.MkdirTemp("", "test-extract-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	err = ExtractToDir(buf.Bytes(), tmpDir)
	require.NoError(t, err)

	// Verify files exist
	mainPath := filepath.Join(tmpDir, "main.py")
	assert.FileExists(t, mainPath)

	mainData, err := os.ReadFile(mainPath)
	require.NoError(t, err)
	assert.Equal(t, content, mainData)

	// Verify subdirectory file
	utilsPath := filepath.Join(tmpDir, "subdir", "utils.py")
	assert.FileExists(t, utilsPath)
}

func TestValidatePath_RejectsTraversal(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"normal path", "main.py", false},
		{"nested path", "src/main.py", false},
		{"parent traversal", "../etc/passwd", true},
		{"hidden parent", "foo/../../../etc/passwd", true},
		{"absolute path", "/etc/passwd", true},
		{"starts with slash", "/main.py", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePath(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListFiles(t *testing.T) {
	// Create a test tar
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	files := []string{"main.py", "utils.py", "data/config.json"}
	for _, name := range files {
		content := []byte("test1")
		header := &tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(content)),
		}
		require.NoError(t, tw.WriteHeader(header))
		_, err := tw.Write(content)
		require.NoError(t, err)
	}

	require.NoError(t, tw.Close())

	// List files
	listed, err := ListFiles(buf.Bytes())
	require.NoError(t, err)

	assert.ElementsMatch(t, files, listed)
}
