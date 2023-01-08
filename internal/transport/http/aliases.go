package http

import (
	"errors"

	"github.com/anfelo/go-updater/internal/helpers"
)

var aliases = map[string][]string{
	"darwin":         {"mac", "macos", "osx"},
	"darwin_arm64":   {"mac_arm64", "macos_arm64", "osx_arm64"},
	"exe":            {"win32", "windows", "win"},
	"exe_arm64":      {"win32_arm64", "windows_arm64", "win_arm64"},
	"deb":            {"debian"},
	"deb_arm64":      {"debian_arm64"},
	"rmp":            {"fedora"},
	"rmp_arm64":      {"fedora_arm64"},
	"AppImage":       {"appimage"},
	"AppImage_arm64": {"appimage_arm64"},
	"dmg":            {"dmg"},
	"dmg_arm64":      {"dmg_arm64"},
}

func getAlias(platform string) (string, error) {
	for key, platforms := range aliases {
		if helpers.Contains(platforms, platform) {
			return key, nil
		}
	}
	return "", errors.New("no platform found")
}
