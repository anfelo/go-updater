package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/anfelo/go-updater/internal/updater"
	"github.com/mssola/user_agent"
	log "github.com/sirupsen/logrus"
)

// UpdaterService - interface of the updater service
type UpdaterService interface {
	GetLatest(ctx context.Context) (updater.Release, error)
}

// ReleaseResponse - latest release response model
type ReleaseResponse struct {
	Version   string                           `json:"version"`
	Notes     string                           `json:"notes"`
	PubDate   time.Time                        `json:"pub_date"`
	Platforms map[string]PlatformAssetResponse `json:"platforms"`
}

// PlatformAssetResponse - platform asset model
type PlatformAssetResponse struct {
	Name        string `json:"name"`
	ApiURL      string `json:"api_url"`
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

// DownloadLatest - downloads the latest release non-prerelease and non-draft
func (h *Handler) DownloadLatest(w http.ResponseWriter, r *http.Request) {
	ua := user_agent.New(r.Header.Get("user-agent"))
	vars := mux.Vars(r)
	isUpdate := vars["update"] != ""

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

func mapToReleaseResponse(ur updater.Release) ReleaseResponse {
	r := ReleaseResponse{
		Version:   ur.Version,
		Notes:     ur.Notes,
		PubDate:   ur.PubDate,
		Platforms: map[string]PlatformAssetResponse{},
	}

	for platform, asset := range ur.Platforms {
		r.Platforms[platform] = PlatformAssetResponse{
			Name:        asset.Name,
			ApiURL:      asset.ApiURL,
			URL:         asset.URL,
			ContentType: asset.ContentType,
			Size:        asset.Size,
		}
	}

	return r
}
