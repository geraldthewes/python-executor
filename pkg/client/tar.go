package client

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// TarFromFiles creates an uncompressed tar archive from a list of file paths
func TarFromFiles(files []string) ([]byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	defer tw.Close()

	for _, filePath := range files {
		if err := addFileToTar(tw, filePath, ""); err != nil {
			return nil, fmt.Errorf("adding %s to tar: %w", filePath, err)
		}
	}

	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("closing tar: %w", err)
	}

	return buf.Bytes(), nil
}

// TarFromDirectory creates an uncompressed tar archive from a directory
func TarFromDirectory(dirPath string) ([]byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	defer tw.Close()

	// Walk the directory tree
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == dirPath {
			return nil
		}

		// Calculate relative path for tar
		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		return addFileToTar(tw, path, relPath)
	})

	if err != nil {
		return nil, fmt.Errorf("walking directory: %w", err)
	}

	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("closing tar: %w", err)
	}

	return buf.Bytes(), nil
}

// TarFromReader creates a tar archive from stdin or any reader (single file)
func TarFromReader(r io.Reader, filename string) ([]byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	defer tw.Close()

	// Read all content
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}

	// Write to tar
	header := &tar.Header{
		Name: filename,
		Mode: 0644,
		Size: int64(len(content)),
	}

	if err := tw.WriteHeader(header); err != nil {
		return nil, fmt.Errorf("writing tar header: %w", err)
	}

	if _, err := tw.Write(content); err != nil {
		return nil, fmt.Errorf("writing content: %w", err)
	}

	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("closing tar: %w", err)
	}

	return buf.Bytes(), nil
}

// TarFromMap creates a tar archive from a map of filename -> content
func TarFromMap(files map[string]string) ([]byte, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	defer tw.Close()

	for filename, content := range files {
		header := &tar.Header{
			Name: filename,
			Mode: 0644,
			Size: int64(len(content)),
		}

		if err := tw.WriteHeader(header); err != nil {
			return nil, fmt.Errorf("writing header for %s: %w", filename, err)
		}

		if _, err := tw.Write([]byte(content)); err != nil {
			return nil, fmt.Errorf("writing content for %s: %w", filename, err)
		}
	}

	if err := tw.Close(); err != nil {
		return nil, fmt.Errorf("closing tar: %w", err)
	}

	return buf.Bytes(), nil
}

// addFileToTar adds a single file to a tar writer
func addFileToTar(tw *tar.Writer, filePath string, tarPath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	// Use the provided tar path, or the file's base name
	if tarPath == "" {
		tarPath = filepath.Base(filePath)
	}

	// Normalize to forward slashes for tar
	tarPath = filepath.ToSlash(tarPath)

	// Handle directories
	if info.IsDir() {
		header := &tar.Header{
			Name:     tarPath + "/",
			Mode:     int64(info.Mode()),
			Typeflag: tar.TypeDir,
		}
		return tw.WriteHeader(header)
	}

	// Handle regular files
	header := &tar.Header{
		Name: tarPath,
		Mode: int64(info.Mode()),
		Size: info.Size(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(tw, file)
	return err
}

// DetectEntrypoint finds the entrypoint in a tar archive
func DetectEntrypoint(tarData []byte) (string, error) {
	reader := tar.NewReader(bytes.NewReader(tarData))

	var candidates []string
	var firstPy string

	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		if header.Typeflag == tar.TypeReg && strings.HasSuffix(header.Name, ".py") {
			basename := filepath.Base(header.Name)

			// Priority order
			if basename == "main.py" {
				return header.Name, nil
			}
			if basename == "__main__.py" {
				candidates = append(candidates, header.Name)
			}
			if firstPy == "" {
				firstPy = header.Name
			}
		}
	}

	// Return in priority order
	if len(candidates) > 0 {
		return candidates[0], nil
	}
	if firstPy != "" {
		return firstPy, nil
	}

	return "", fmt.Errorf("no Python files found in archive")
}
