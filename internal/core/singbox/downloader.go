package singbox

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"singbox-web/internal/storage"
)

const (
	GitHubAPIURL = "https://api.github.com/repos/SagerNet/sing-box/releases/latest"
)

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func getHTTPClient() *http.Client {
	client := &http.Client{}

	proxyEnabled, _ := storage.GetSetting("download_proxy_enabled")
	proxyURL, _ := storage.GetSetting("download_proxy_url")

	if proxyEnabled == "true" && proxyURL != "" {
		proxy, err := url.Parse(proxyURL)
		if err == nil {
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxy),
			}
		}
	}

	return client
}

func GetLatestVersion() (string, error) {
	client := getHTTPClient()

	resp, err := client.Get(GitHubAPIURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch release info: status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to parse release info: %w", err)
	}

	return release.TagName, nil
}

func DownloadLatest(dataDir string) (string, error) {
	client := getHTTPClient()

	// Get release info
	resp, err := client.Get(GitHubAPIURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch release info: status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("failed to parse release info: %w", err)
	}

	// Find matching asset
	arch := runtime.GOARCH
	goos := runtime.GOOS

	// Build expected asset name pattern
	version := strings.TrimPrefix(release.TagName, "v")
	expectedName := fmt.Sprintf("sing-box-%s-%s-%s.tar.gz", version, goos, arch)

	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == expectedName {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return "", fmt.Errorf("no matching asset found for %s-%s, looking for: %s", goos, arch, expectedName)
	}

	// Download
	resp, err = client.Get(downloadURL)
	if err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download: status %d", resp.StatusCode)
	}

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "", err
	}

	// Save to temp file
	tmpFile := filepath.Join(dataDir, "sing-box.tar.gz")
	f, err := os.Create(tmpFile)
	if err != nil {
		return "", err
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return "", err
	}
	f.Close()

	// Extract
	binPath := filepath.Join(dataDir, "sing-box")
	if err := extractTarGz(tmpFile, dataDir); err != nil {
		os.Remove(tmpFile)
		return "", fmt.Errorf("failed to extract: %w", err)
	}

	// Cleanup
	os.Remove(tmpFile)

	// Make executable
	if err := os.Chmod(binPath, 0755); err != nil {
		return "", err
	}

	// Save path to settings
	storage.SetSetting("singbox_path", binPath)

	return release.TagName, nil
}

func extractTarGz(tarGzPath, destDir string) error {
	f, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Only extract the sing-box binary
		baseName := filepath.Base(header.Name)
		if baseName != "sing-box" {
			continue
		}

		target := filepath.Join(destDir, baseName)

		switch header.Typeflag {
		case tar.TypeReg:
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}

func GetInstalledVersion(dataDir string) (string, error) {
	binPath := filepath.Join(dataDir, "sing-box")

	// Check if binary exists
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		return "", fmt.Errorf("sing-box not installed")
	}

	// TODO: Get version by running sing-box version command
	return "installed", nil
}
