package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/anfelo/go-updater/internal/release"
	log "github.com/sirupsen/logrus"
)

// UpdaterService - interface of the updater service
type UpdaterService interface {
	GetReleases(ctx context.Context) (release.Release, error)
}

// GetReleases - handles the retrieval of latest releases
func (h *Handler) GetReleases(w http.ResponseWriter, r *http.Request) {
	resp, err := h.Client.R().
		SetQueryParams(map[string]string{
			"per_page": "100",
		}).
		SetHeader("Accept", "application/vnd.github+json").
		SetAuthToken(os.Getenv("GH_TOKEN")).
		Get(fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", "anfelo", "hello-tauri"))

	if err != nil {
		log.Error(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	log.Info(resp)

	if err := json.NewEncoder(w).Encode(map[string]string{"message": "ok"}); err != nil {
		log.Error("error encoding to json")
		panic(err)
	}
}
