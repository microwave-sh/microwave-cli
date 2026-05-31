package config

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// UpdateCheck is the cached state written to ~/.config/microwave/.update-check.
type UpdateCheck struct {
	LastCheck     time.Time `json:"last_check"`
	LatestVersion string    `json:"latest_version"`
}

// CheckForUpdate checks GitHub releases for a newer version of the CLI.
// It caches the result for 24 hours and prints a one-line notice to stderr
// when an update is available. The check is non-blocking (runs in a goroutine
// when the cache is stale). Respects MICROWAVE_NO_UPDATE_CHECK=1 for CI.
func CheckForUpdate(currentVersion string) {
	if os.Getenv("MICROWAVE_NO_UPDATE_CHECK") != "" {
		return
	}

	checkFile := filepath.Join(GlobalConfigDir(), ".update-check")

	// Return early if cache is fresh (< 24 h).
	if data, err := os.ReadFile(checkFile); err == nil {
		var check UpdateCheck
		if json.Unmarshal(data, &check) == nil {
			if time.Since(check.LastCheck) < 24*time.Hour {
				if check.LatestVersion != "" && check.LatestVersion != currentVersion {
					printUpdateNotice(currentVersion, check.LatestVersion)
				}
				return
			}
		}
	}

	// Fetch latest release from GitHub in the background (best-effort).
	go func() {
		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get(
			"https://api.github.com/repos/microwave-sh/microwave-cli/releases/latest")
		if err != nil {
			return
		}
		defer func() { _ = resp.Body.Close() }()

		var release struct {
			TagName string `json:"tag_name"`
		}
		if json.NewDecoder(resp.Body).Decode(&release) != nil {
			return
		}

		latest := release.TagName
		if len(latest) > 0 && latest[0] == 'v' {
			latest = latest[1:]
		}

		check := UpdateCheck{
			LastCheck:     time.Now(),
			LatestVersion: latest,
		}
		data, _ := json.Marshal(check)
		os.MkdirAll(GlobalConfigDir(), 0700) //nolint:errcheck
		os.WriteFile(checkFile, data, 0600)  //nolint:errcheck

		if latest != currentVersion {
			printUpdateNotice(currentVersion, latest)
		}
	}()
}

func printUpdateNotice(current, latest string) {
	fmt.Fprintf(os.Stderr,
		"\nA new version of microwave is available: %s → %s\n"+
			"Update: brew upgrade microwave  or  curl -sSL https://microwave.sh/install.sh | sh\n\n",
		current, latest)
}
