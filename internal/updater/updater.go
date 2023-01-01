package updater

import (
	"context"
	"time"
)

// Release - the release model
type Release struct {
	Version   string
	Notes     string
	PubDate   time.Time
	Platforms map[string]PlatformAsset
}

// PlatformAsset - platform asset model
type PlatformAsset struct {
	Name        string
	ApiURL      string
	URL         string
	ContentType string
	Size        int64
}

// UpdaterStore - the updater interface
type UpdaterStore interface {
	GetLatest(context.Context) (Release, error)
}

// GetLatest - gets the latest release from the store
func (s *Service) GetLatest(ctx context.Context) (Release, error) {
	r, err := s.Store.GetLatest(ctx)
	if err != nil {
		return Release{}, err
	}
	return r, nil
}
