package http

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/anfelo/go-updater/internal/release"
	log "github.com/sirupsen/logrus"
)

// UpdaterService - interface of the updater service
type UpdaterService interface {
	GetReleases(ctx context.Context) (release.Release, error)
	GetLatest(ctx context.Context) (release.Release, error)
}

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

// PlatformAsset - platform asset model
type PlatformAsset struct {
	Name        string `json:"name"`
	ApiURL      string `json:"api_url"`
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

// ReleaseResponse - latest release response model
type ReleaseResponse struct {
	Version   string                   `json:"version"`
	Notes     string                   `json:"notes"`
	PubDate   time.Time                `json:"pub_date"`
	Platforms map[string]PlatformAsset `json:"platforms"`
}

// GetReleases - handles the retrieval of latest releases
func (h *Handler) GetReleases(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/releases",
		os.Getenv("GH_USERNAME"),
		os.Getenv("GH_REPO"),
	)

	log.Info(fmt.Sprintf("Requesting releases from: %s", url))

	resp, err := h.Client.R().
		SetQueryParams(map[string]string{
			"per_page": "100",
		}).
		SetHeader("Accept", "application/vnd.github+json").
		SetAuthToken(os.Getenv("GH_TOKEN")).
		Get(url)

	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if resp.StatusCode() != 200 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var allReleases []GithubRelease
	if err := json.Unmarshal(resp.Body(), &allReleases); err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var releases []GithubRelease
	for _, rel := range allReleases {
		if !rel.Prerelease && !rel.Draft {
			releases = append(releases, rel)
		}
	}

	if err := json.NewEncoder(w).Encode(releases); err != nil {
		log.Error("error encoding to json")
		panic(err)
	}
}

// GetLatest - gets the latest release non-prerelease and non-draft
func (h *Handler) GetLatest(w http.ResponseWriter, r *http.Request) {
	url := fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/releases/latest",
		os.Getenv("GH_USERNAME"),
		os.Getenv("GH_REPO"),
	)

	log.Info(fmt.Sprintf("Requesting latest release from: %s", url))

	resp, err := h.Client.R().
		SetQueryParams(map[string]string{
			"per_page": "100",
		}).
		SetHeader("Accept", "application/vnd.github+json").
		SetAuthToken(os.Getenv("GH_TOKEN")).
		Get(url)

	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if resp.StatusCode() != 200 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var release GithubRelease
	if err := json.Unmarshal(resp.Body(), &release); err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(mapToRelease(release)); err != nil {
		log.Error("error encoding to json")
		panic(err)
	}
}

func mapToRelease(gr GithubRelease) ReleaseResponse {
	r := ReleaseResponse{
		Version:   gr.TagName,
		Notes:     gr.Body,
		PubDate:   gr.PublishedAt,
		Platforms: map[string]PlatformAsset{},
	}

	for _, asset := range gr.Assets {
		platform := checkPlatform(asset.Name)
		if platform == "" {
			continue
		}

		r.Platforms[platform] = PlatformAsset{
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
	if contains(directCache, ext) {
		return ext + arch
	}
	return ""
}

func contains(slice []string, name string) bool {
	for _, v := range slice {
		if v == name {
			return true
		}
	}
	return false
}
