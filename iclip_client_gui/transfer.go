package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"time"
)

// ClipboardMeta holds metadata for clipboard transfer
type ClipboardMeta struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Username  string `json:"username"`
	Encoding  string `json:"encoding"`
}

// TransferClipboard creates an archive and uploads it via SMB
func TransferClipboard(content string, config *Config) error {
	// Get current user
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}
	username := currentUser.Username

	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "iclip_")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Create clipboard_data subdirectory
	dataDir := filepath.Join(tempDir, "clipboard_data")
	if err := os.Mkdir(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data dir: %w", err)
	}

	// Create meta.json
	meta := ClipboardMeta{
		Type:      "text",
		Timestamp: time.Now().Format(time.RFC3339),
		Username:  username,
		Encoding:  "utf-8",
	}
	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal meta: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "meta.json"), metaData, 0644); err != nil {
		return fmt.Errorf("failed to write meta.json: %w", err)
	}

	// Create content.txt
	if err := os.WriteFile(filepath.Join(dataDir, "content.txt"), []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write content.txt: %w", err)
	}

	// Create tar.gz archive
	archiveName := fmt.Sprintf("clipboard_%s.tar.gz", username)
	archivePath := filepath.Join(tempDir, archiveName)
	if err := createTarGz(archivePath, dataDir); err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}

	// Upload via SMB
	if err := uploadViaSMB(archivePath, archiveName, config); err != nil {
		return fmt.Errorf("failed to upload: %w", err)
	}

	return nil
}

// createTarGz creates a tar.gz archive of a directory
func createTarGz(archivePath string, sourceDir string) error {
	file, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		// Set the name relative to the source directory, but include the base directory
		relPath, err := filepath.Rel(filepath.Dir(sourceDir), path)
		if err != nil {
			return err
		}
		header.Name = relPath

		// Handle directories
		if info.IsDir() {
			header.Name += "/"
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// Write file content
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tarWriter, file)
			return err
		}

		return nil
	})
}

// uploadViaSMB uploads a file using smbclient
func uploadViaSMB(localPath string, remoteName string, config *Config) error {
	// Build smbclient command
	share := fmt.Sprintf("//%s/%s", config.SMB.Server, config.SMB.Share)
	auth := fmt.Sprintf("%s%%%s", config.SMB.Username, config.SMB.Password)

	args := []string{share, "-U", auth}

	// Add workgroup if specified
	if config.SMB.Workgroup != "" {
		args = append(args, "-W", config.SMB.Workgroup)
	}

	// Build the command to execute
	smbCommand := fmt.Sprintf("cd \"%s\"; put \"%s\" \"%s\"",
		config.SMB.Path, localPath, remoteName)
	args = append(args, "-c", smbCommand)

	cmd := exec.Command("smbclient", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("smbclient failed: %s, output: %s", err, string(output))
	}

	return nil
}
