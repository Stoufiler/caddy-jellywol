package upgrade

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/inconshreveable/go-update"
	"github.com/sergi/go-diff/diffmatchpatch"
)

func RunUpgrade(currentVersion string) {
	fmt.Println("Checking for updates...")
	latestVersion, err := getLatestVersion()
	if err != nil {
		fmt.Printf("Error checking for updates: %v\n", err)
		return
	}

	if currentVersion == latestVersion {
		fmt.Println("You are already on the latest version.")
		return
	}

	fmt.Printf("A new version (%s) is available. You are on version %s.\n", latestVersion, currentVersion)
	fmt.Println("Upgrading...")

	if err := doUpgrade(latestVersion); err != nil {
		fmt.Printf("Error upgrading: %v\n", err)
		return
	}

	fmt.Println("Upgrade successful.")
}

func getLatestVersion() (string, error) {
	resp, err := http.Get(githubAPIURL)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return release.TagName, nil
}

func doUpgrade(latestVersion string) error {
	downloadURL, err := getDownloadURL(latestVersion)
	if err != nil {
		return err
	}

	fmt.Printf("Downloading from %s...\n", downloadURL)

	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Read archive content
	archiveBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Extract binary from archive
	var binaryReader io.Reader
	if runtime.GOOS == "windows" {
		// Extract from .zip
		binaryReader, err = extractFromZip(archiveBytes)
	} else {
		// Extract from .tar.gz
		binaryReader, err = extractFromTarGz(archiveBytes)
	}
	if err != nil {
		return fmt.Errorf("failed to extract binary: %v", err)
	}

	if err := applyUpdate(binaryReader, update.Options{}); err != nil {
		return err
	}

	// Try to download and compare config.json.example (optional)
	configDownloadURL, err := getconfigDownloadURL(latestVersion)
	if err != nil {
		fmt.Printf("Warning: could not get config download URL: %v\n", err)
		return nil
	}

	fmt.Printf("Downloading from %s...\n", configDownloadURL)

	configResp, err := http.Get(configDownloadURL)
	if err != nil {
		fmt.Printf("Warning: could not download config file: %v\n", err)
		return nil
	}
	defer func() {
		if err := configResp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	if configResp.StatusCode != http.StatusOK {
		fmt.Printf("Warning: config file not found in release (status %d)\n", configResp.StatusCode)
		return nil
	}

	if err := compareConfigs(configResp.Body); err != nil {
		fmt.Printf("Warning: could not compare configs: %v\n", err)
	}

	return nil
}

// extractFromTarGz extracts the binary from a .tar.gz archive
func extractFromTarGz(archiveBytes []byte) (io.Reader, error) {
	gzr, err := gzip.NewReader(bytes.NewReader(archiveBytes))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = gzr.Close()
	}()

	tarReader := tar.NewReader(gzr)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		// Find the binary file (should be the only file or starts with jellywolproxy)
		if header.Typeflag == tar.TypeReg && strings.Contains(header.Name, "jellywolproxy") {
			// Read entire binary into memory
			buf := new(bytes.Buffer)
			if _, err := io.Copy(buf, tarReader); err != nil {
				return nil, err
			}
			return buf, nil
		}
	}

	return nil, fmt.Errorf("binary not found in archive")
}

// extractFromZip extracts the binary from a .zip archive
func extractFromZip(archiveBytes []byte) (io.Reader, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(archiveBytes), int64(len(archiveBytes)))
	if err != nil {
		return nil, err
	}

	for _, file := range zipReader.File {
		// Find the .exe file
		if strings.HasSuffix(file.Name, ".exe") && strings.Contains(file.Name, "jellywolproxy") {
			rc, err := file.Open()
			if err != nil {
				return nil, err
			}
			defer func() {
				_ = rc.Close()
			}()

			// Read entire binary into memory
			buf := new(bytes.Buffer)
			if _, err := io.Copy(buf, rc); err != nil {
				return nil, err
			}
			return buf, nil
		}
	}

	return nil, fmt.Errorf("binary not found in archive")
}

var (
	githubAPIURL   = "https://api.github.com/repos/Stoufiler/JellyWolProxy/releases/latest"
	applyUpdate    = update.Apply
	getDownloadURL = func(version string) (string, error) {
		archiveExt := ".tar.gz"
		if runtime.GOOS == "windows" {
			archiveExt = ".zip"
		}
		return fmt.Sprintf("https://github.com/Stoufiler/JellyWolProxy/releases/download/%s/jellywolproxy-%s-%s%s", version, runtime.GOOS, runtime.GOARCH, archiveExt), nil
	}
	getconfigDownloadURL = func(version string) (string, error) {
		return fmt.Sprintf("https://github.com/Stoufiler/JellyWolProxy/releases/download/%s/config.json.example", version), nil
	}
	readFile       = os.ReadFile
	compareConfigs = func(newConfig io.Reader) error {
		newConfigBytes, err := io.ReadAll(newConfig)
		if err != nil {
			return err
		}

		oldConfigBytes, err := readFile("config.json.example")
		if err != nil {
			return err
		}

		if !bytes.Equal(newConfigBytes, oldConfigBytes) {
			fmt.Println("Configuration has been updated. Please check your config.json file.")
			fmt.Println("Here are the differences:")
			dmp := diffmatchpatch.New()
			diffs := dmp.DiffMain(string(oldConfigBytes), string(newConfigBytes), false)
			fmt.Println(dmp.DiffPrettyText(diffs))
		}

		return nil
	}
)
