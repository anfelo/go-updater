package datasource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anfelo/go-updater/internal/updater"
	"github.com/anfelo/go-updater/internal/helpers"
	log "github.com/sirupsen/logrus"
)

// GithubRelease - github's release schema
type GithubRelease struct {
	ID          int64                `json:"id"`
	URL         string               `json:"url"`
	Draft       bool                 `json:"draft"`
	Prerelease  bool                 `json:"prerelease"`
	Body        string               `json:"body"`
	Name        string               `json:"name"`
	TagName     string               `json:"tag_name"`
	HTMLURL     string               `json:"html_url"`
	AssetsURL   string               `json:"assets_url"`
	UploadURL   string               `json:"upload_url"`
	TarballURL  string               `json:"tarball_url"`
	ZipballURL  string               `json:"zipball_url"`
	CreatedAt   time.Time            `json:"created_at"`
	PublishedAt time.Time            `json:"published_at"`
	Assets      []GithubReleaseAsset `json:"assets"`
}

// GithubReleaseAsset - github's release asset schema
type GithubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	URL                string `json:"url"`
	ContentType        string `json:"content_type"`
	Size               int64  `json:"size"`
}

// GetLatest - Gets the latest release from github
func (d *Datasource) GetLatest(ctx context.Context) (updater.Release, error) {
	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/releases/latest",
		os.Getenv("GH_USERNAME"),
		os.Getenv("GH_REPO"),
	)

	log.Info(fmt.Sprintf("Requesting latest release from: %s", url))

	resp, err := d.Client.R().
		SetQueryParams(map[string]string{
			"per_page": "100",
		}).
		SetHeader("Accept", "application/vnd.github+json").
		SetAuthToken(os.Getenv("GH_TOKEN")).
		Get(url)

	if err != nil {
		log.Error(err)
		return updater.Release{}, fmt.Errorf("error fetching the latest release from github: %w", err)
	}

	if resp.StatusCode() != 200 {
		return updater.Release{}, errors.New("error fetching the latest release from github")
	}

	var gr GithubRelease
	if err := json.Unmarshal(resp.Body(), &gr); err != nil {
		log.Error(err)
		return updater.Release{}, fmt.Errorf("error fetching the latest release from github: %w", err)
	}

	return mapToRelease(gr), nil
}

func mapToRelease(gr GithubRelease) updater.Release {
	r := updater.Release{
		Version:   gr.TagName,
		Notes:     gr.Body,
		PubDate:   gr.PublishedAt,
		Platforms: map[string]updater.PlatformAsset{},
	}

	for _, asset := range gr.Assets {
		platform := checkPlatform(asset.Name)
		if platform == "" {
			continue
		}

		r.Platforms[platform] = updater.PlatformAsset{
			Name:        asset.Name,
			ApiURL:      asset.URL,
			URL:         asset.BrowserDownloadURL,
			ContentType: asset.ContentType,
			Size:        int64(math.Round(float64(asset.Size) / 1000000)),
		}
	}

	return r
}

func checkPlatform(fileName string) string {
	ext := filepath.Ext(fileName)
	_, ext, _ = strings.Cut(ext, ".")

	arch := ""
	if strings.Contains(fileName, "arm64") || strings.Contains(fileName, "aarch64") {
		arch = "_arm64"
	}

	if (strings.Contains(fileName, "mac") || strings.Contains(fileName, "darwin")) && ext == "zip" {
		return "darwin" + arch
	}

	directCache := []string{"exe", "dmg", "rmp", "deb", "AppImage"}
	if helpers.Contains(directCache, ext) {
		return ext + arch
	}
	return ""
}
