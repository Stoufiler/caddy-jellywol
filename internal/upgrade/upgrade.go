package upgrade

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

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

	if err := applyUpdate(resp.Body, update.Options{}); err != nil {
		return err
	}

	configDownloadURL, err := getconfigDownloadURL(latestVersion)
	if err != nil {
		return err
	}

	fmt.Printf("Downloading from %s...\n", configDownloadURL)

	configResp, err := http.Get(configDownloadURL)
	if err != nil {
		return err
	}
	defer func() {
		if err := configResp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	if configResp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", configResp.StatusCode)
	}

	return compareConfigs(configResp.Body)
}

var (
	githubAPIURL   = "https://api.github.com/repos/Stoufiler/JellyWolProxy/releases/latest"
	applyUpdate    = update.Apply
	getDownloadURL = func(version string) (string, error) {
		return fmt.Sprintf("https://github.com/Stoufiler/JellyWolProxy/releases/download/%s/jellywolproxy_%s_%s", version, runtime.GOOS, runtime.GOARCH), nil
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
