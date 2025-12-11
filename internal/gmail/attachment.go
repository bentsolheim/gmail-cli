package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DownloadAttachment downloads an attachment by its ID.
func (c *Client) DownloadAttachment(ctx context.Context, messageID, attachmentID string) ([]byte, error) {
	attachment, err := c.service.Users.Messages.Attachments.Get(c.userID, messageID, attachmentID).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to download attachment: %w", err)
	}

	data, err := base64.URLEncoding.DecodeString(attachment.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode attachment data: %w", err)
	}

	return data, nil
}

// SaveAttachment saves attachment data to a file in the specified directory.
// Returns the full path to the saved file.
func SaveAttachment(data []byte, outputDir, filename string) (string, error) {
	// Sanitize filename
	filename = sanitizeFilename(filename)
	if filename == "" {
		filename = "attachment"
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Handle duplicate filenames
	fullPath := filepath.Join(outputDir, filename)
	fullPath = uniquePath(fullPath)

	// Write file
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write attachment: %w", err)
	}

	return fullPath, nil
}

// sanitizeFilename removes or replaces characters that are unsafe for filenames.
func sanitizeFilename(name string) string {
	// Replace path separators and other unsafe characters
	unsafe := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\x00"}
	for _, char := range unsafe {
		name = strings.ReplaceAll(name, char, "_")
	}

	// Trim leading/trailing spaces and dots
	name = strings.Trim(name, " .")

	// Limit length
	if len(name) > 200 {
		ext := filepath.Ext(name)
		name = name[:200-len(ext)] + ext
	}

	return name
}

// uniquePath returns a unique file path by appending a number if the file already exists.
func uniquePath(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}

	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)

	for i := 1; ; i++ {
		newPath := fmt.Sprintf("%s_%d%s", base, i, ext)
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}
}
