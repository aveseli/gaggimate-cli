// Package update handles self-updating the gaggimate-cli binary.
package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const repo = "aveseli/gaggimate-cli"

// CurrentVersion is set at build time via -ldflags.
var CurrentVersion = "dev"

// GitHubRelease represents the relevant fields of a GitHub release.
type GitHubRelease struct {
	TagName string `json:"tag_name"`
}

// CheckLatest fetches the latest release version from GitHub.
func CheckLatest(client *http.Client) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetching latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("parsing release info: %w", err)
	}

	// Strip leading "v" if present
	return strings.TrimPrefix(release.TagName, "v"), nil
}

// BinaryName returns the platform-specific binary name.
func BinaryName() string {
	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}
	return fmt.Sprintf("gaggimate-%s-%s%s", runtime.GOOS, runtime.GOARCH, ext)
}

// DownloadURL returns the download URL for a given version.
func DownloadURL(version string) string {
	return fmt.Sprintf("https://github.com/%s/releases/download/v%s/%s", repo, version, BinaryName())
}

// SelfPath returns the absolute path to the running binary.
func SelfPath() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolving executable path: %w", err)
	}

	// Resolve any symlinks
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path, nil // fall back to unresolved path
	}
	return resolved, nil
}

// Download fetches the binary for the given version and writes it to dst.
// It preserves the executable permission bit from the existing file at dst.
func Download(client *http.Client, version, dst string) error {
	url := DownloadURL(version)

	// Get current file permissions
	var perm os.FileMode = 0755
	if info, err := os.Stat(dst); err == nil {
		perm = info.Mode().Perm()
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("downloading binary: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	// Write to a temp file in the same directory for atomic rename
	tmpFile, err := os.CreateTemp(filepath.Dir(dst), ".gaggimate-update-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Clean up temp file on error
	success := false
	defer func() {
		if !success {
			os.Remove(tmpPath)
		}
	}()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		tmpFile.Close()
		return fmt.Errorf("writing binary: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	// Preserve permissions and make executable
	if err := os.Chmod(tmpPath, perm|0111); err != nil {
		return fmt.Errorf("setting permissions: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, dst); err != nil {
		return fmt.Errorf("replacing binary: %w", err)
	}

	success = true
	return nil
}
