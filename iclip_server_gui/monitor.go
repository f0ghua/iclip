package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ClipboardData holds extracted clipboard data
type ClipboardData struct {
	Content     string
	ContentType string
	Username    string
	Timestamp   string
}

// ClipboardMeta holds metadata from meta.json
type ClipboardMeta struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Username  string `json:"username"`
	Encoding  string `json:"encoding"`
}

// FindClipboardFiles finds matching clipboard files in the watch directory
func FindClipboardFiles(watchPath string, pattern string) ([]string, error) {
	var files []string

	// Check if directory exists
	info, err := os.Stat(watchPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("watch path does not exist: %s", watchPath)
		}
		return nil, fmt.Errorf("cannot access watch path: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("watch path is not a directory: %s", watchPath)
	}

	// Read directory
	entries, err := os.ReadDir(watchPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read directory: %w", err)
	}

	// Match files against pattern
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		matched, err := filepath.Match(pattern, entry.Name())
		if err != nil {
			continue
		}

		if matched {
			files = append(files, filepath.Join(watchPath, entry.Name()))
		}
	}

	return files, nil
}

// ExtractClipboardData extracts data from a clipboard archive
func ExtractClipboardData(archivePath string) (*ClipboardData, error) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "iclip_extract_")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Open archive
	file, err := os.Open(archivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzReader)

	// Extract files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar: %w", err)
		}

		// Security check: prevent path traversal
		cleanName := filepath.Clean(header.Name)
		if strings.HasPrefix(cleanName, "..") || filepath.IsAbs(cleanName) {
			return nil, fmt.Errorf("invalid path in archive: %s", header.Name)
		}

		targetPath := filepath.Join(tempDir, cleanName)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return nil, fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return nil, fmt.Errorf("failed to create parent directory: %w", err)
			}

			outFile, err := os.Create(targetPath)
			if err != nil {
				return nil, fmt.Errorf("failed to create file: %w", err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return nil, fmt.Errorf("failed to write file: %w", err)
			}
			outFile.Close()
		}
	}

	// Read meta.json
	metaPath := filepath.Join(tempDir, "clipboard_data", "meta.json")
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read meta.json: %w", err)
	}

	var meta ClipboardMeta
	if err := json.Unmarshal(metaData, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse meta.json: %w", err)
	}

	// Read content.txt
	contentPath := filepath.Join(tempDir, "clipboard_data", "content.txt")
	contentData, err := os.ReadFile(contentPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read content.txt: %w", err)
	}

	return &ClipboardData{
		Content:     string(contentData),
		ContentType: meta.Type,
		Username:    meta.Username,
		Timestamp:   meta.Timestamp,
	}, nil
}

// DeleteFile safely deletes a file
func DeleteFile(filePath string) error {
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}
