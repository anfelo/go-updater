package http

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/anfelo/go-updater/internal/updater"
	"github.com/dustin/go-humanize"
	"github.com/gorilla/mux"
	"github.com/mssola/user_agent"
	log "github.com/sirupsen/logrus"
	"golang.org/x/mod/semver"
)

type UpdaterService interface {
	GetLatest(ctx context.Context) (updater.Release, error)
}

type UpdateResponse struct {
	Name    string    `json:"name"`
	Notes   string    `json:"notes"`
	PubDate time.Time `json:"pub_date"`
	URL     string    `json:"url"`
}

type PlatformDisplay struct {
	Name        string
	ApiURL      string
	URL         string
	ContentType string
	SizeInMB    string
}

// DownloadLatest - downloads the latest release non-prerelease and non-draft
func (h *Handler) DownloadLatest(w http.ResponseWriter, r *http.Request) {
	ua := user_agent.New(r.Header.Get("user-agent"))
	isUpdate := r.URL.Query().Get("update") != ""

	platform := ""
	if ua.OSInfo().Name == "Mac OS X" && isUpdate {
		platform = "darwin"
	} else if ua.OSInfo().Name == "Mac OS X" && !isUpdate {
		platform = "dmg"
	} else if ua.OSInfo().Name == "Windows" {
		platform = "exe"
	} else if ua.OSInfo().Name == "Linux" {
		platform = "deb"
	}

	latest, err := h.Service.GetLatest(r.Context())
	if err != nil {
		log.Error(err)
		http.Error(w, "release not found", http.StatusNotFound)
		return
	}

	asset, exists := latest.Platforms[platform]
	if !exists {
		http.Error(w, "no download available for your platform", http.StatusNotFound)
		return
	}

	// TODO: Check if the token for private repos
	// proxy private download

	w.Header().Set("Location", asset.URL)
	w.WriteHeader(http.StatusFound)
}

func (h *Handler) DownloadPlatform(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	p := vars["platform"]
	if p == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	isUpdate := r.URL.Query().Get("update") != ""
	platform := p
	if p == "mac" && !isUpdate {
		platform = "dmg"
	}

	if platform == "mac_arm64" && !isUpdate {
		platform = "dmg_arm64"
	}

	latest, err := h.Service.GetLatest(r.Context())
	if err != nil {
		log.Error(err)
		http.Error(w, "release not found", http.StatusNotFound)
		return
	}

	platform, err = getAlias(platform)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	asset, exists := latest.Platforms[platform]
	if !exists {
		http.Error(w, "no download available for your platform", http.StatusNotFound)
		return
	}

	// TODO: Check if the token for private repos
	// proxy private download

	w.Header().Set("Location", asset.URL)
	w.WriteHeader(http.StatusFound)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	p := vars["platform"]
	version := vars["version"]

	if !semver.IsValid(version) {
		http.Error(w, "version is not valid", http.StatusBadRequest)
		return
	}

	platform, err := getAlias(p)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	latest, err := h.Service.GetLatest(r.Context())
	if err != nil {
		log.Error(err)
		http.Error(w, "release not found", http.StatusNotFound)
		return
	}

	asset, exists := latest.Platforms[platform]
	if !exists {
		http.Error(w, "no download available for your platform", http.StatusNotFound)
		return
	}

	if semver.Compare(latest.Version, version) != 0 {
		var url = asset.URL
		if os.Getenv("GH_TOKEN") == "" {
			url = fmt.Sprintf("%s/download/%s?update=true", os.Getenv("PRIVATE_BASE_URL"), platform)
		}

		updateRes := UpdateResponse{
			Name:    latest.Version,
			Notes:   latest.Notes,
			PubDate: latest.PubDate,
			URL:     url,
		}
		if err := json.NewEncoder(w).Encode(updateRes); err != nil {
			log.Error("error encoding to json")
			panic(err)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// Home - home route of the updater
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	type templateData struct {
		Account         string
		Repository      string
		Date            string
		Version         string
		Files           map[string]PlatformDisplay
		ReleaseNotes    string
		AllReleases     string
		Github          string
		UserPlatform    string
		UserPlatformURL string
	}

	ua := user_agent.New(r.Header.Get("user-agent"))

	platform := ""
	platformURL := ""
	if ua.OSInfo().Name == "Mac OS X" {
		platform = "dmg"
	} else if ua.OSInfo().Name == "Windows" {
		platform = "exe"
	} else if ua.OSInfo().Name == "Linux" {
		platform = "deb"
	}

	latest, err := h.Service.GetLatest(r.Context())
	if err != nil {
		log.Error(err)
		http.Error(w, "release not found", http.StatusNotFound)
		return
	}

	asset, exists := latest.Platforms[platform]
	if exists {
		platformURL = asset.URL
	}

	tmpl := template.Must(template.ParseFiles(
		"internal/templates/layout.html",
		"internal/templates/home.html",
	))

	tmpl.Execute(w, templateData{
		Account:    os.Getenv("GH_USERNAME"),
		Repository: os.Getenv("GH_REPO"),
		Date:       humanize.Time(latest.PubDate),
		Version:    latest.Version,
		Files:      mapToPlatformDisplay(latest.Platforms),
		ReleaseNotes: fmt.Sprintf(
			"https://github.com/%s/%s/releases/tag/%s",
			os.Getenv("GH_USERNAME"),
			os.Getenv("GH_REPO"),
			latest.Version,
		),
		AllReleases: fmt.Sprintf(
			"https://github.com/%s/%s/releases",
			os.Getenv("GH_USERNAME"),
			os.Getenv("GH_REPO"),
		),
		Github: fmt.Sprintf(
			"https://github.com/%s/%s",
			os.Getenv("GH_USERNAME"),
			os.Getenv("GH_REPO"),
		),
		UserPlatform: platform,
		UserPlatformURL: platformURL,
	})
}

func mapToPlatformDisplay(platforms map[string]updater.PlatformAsset) map[string]PlatformDisplay {
	pp := map[string]PlatformDisplay{}
	for platform, asset := range platforms {
		pp[platform] = PlatformDisplay{
			Name:        asset.Name,
			ApiURL:      asset.ApiURL,
			URL:         asset.URL,
			ContentType: asset.ContentType,
			SizeInMB:    humanize.Bytes(uint64(asset.Size)),
		}
	}
	return pp
}
