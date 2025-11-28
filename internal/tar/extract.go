package tar

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractToDir extracts a tar archive to a directory with path sanitization
func ExtractToDir(tarData []byte, destDir string) error {
	reader := tar.NewReader(bytes.NewReader(tarData))

	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("reading tar: %w", err)
		}

		// Sanitize path - reject any path traversal attempts
		if err := validatePath(header.Name); err != nil {
			return err
		}

		// Build target path
		targetPath := filepath.Join(destDir, header.Name)

		// Security: ensure the path is still within destDir after joining
		if !strings.HasPrefix(filepath.Clean(targetPath), filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid path: %s (path traversal detected)", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return fmt.Errorf("creating directory %s: %w", targetPath, err)
			}

		case tar.TypeReg:
			// Create parent directory if needed
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("creating parent directory for %s: %w", targetPath, err)
			}

			// Create file
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("creating file %s: %w", targetPath, err)
			}

			// Copy file contents
			if _, err := io.Copy(outFile, reader); err != nil {
				outFile.Close()
				return fmt.Errorf("writing file %s: %w", targetPath, err)
			}

			outFile.Close()

		default:
			// Skip symlinks, devices, etc. for security
			continue
		}
	}

	return nil
}

// validatePath checks for path traversal attempts
func validatePath(path string) error {
	// Reject paths containing ..
	if strings.Contains(path, "..") {
		return fmt.Errorf("invalid path: %s (contains ..)", path)
	}

	// Reject absolute paths
	if filepath.IsAbs(path) {
		return fmt.Errorf("invalid path: %s (absolute path not allowed)", path)
	}

	// Reject paths starting with /
	if strings.HasPrefix(path, "/") {
		return fmt.Errorf("invalid path: %s (starts with /)", path)
	}

	return nil
}

// ListFiles lists all files in a tar archive (for debugging/validation)
func ListFiles(tarData []byte) ([]string, error) {
	reader := tar.NewReader(bytes.NewReader(tarData))
	var files []string

	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if header.Typeflag == tar.TypeReg {
			files = append(files, header.Name)
		}
	}

	return files, nil
}
