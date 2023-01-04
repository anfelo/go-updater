package http

import (
	"errors"

	"github.com/anfelo/go-updater/internal/helpers"
)

var aliases = map[string][]string{
	"darwin":         []string{"mac", "macos", "osx"},
	"darwin_arm64":   []string{"mac_arm64", "macos_arm64", "osx_arm64"},
	"exe":            []string{"win32", "windows", "win"},
	"exe_arm64":      []string{"win32_arm64", "windows_arm64", "win_arm64"},
	"deb":            []string{"debian"},
	"deb_arm64":      []string{"debian_arm64"},
	"rmp":            []string{"fedora"},
	"rmp_arm64":      []string{"fedora_arm64"},
	"AppImage":       []string{"appimage"},
	"AppImage_arm64": []string{"appimage_arm64"},
	"dmg":            []string{"dmg"},
	"dmg_arm64":      []string{"dmg_arm64"},
}

func getAlias(platform string) (string, error) {
	for key, platforms := range aliases {
		if helpers.Contains(platforms, platform) {
			return key, nil
		}
	}
	return "", errors.New("no platform found")
}
