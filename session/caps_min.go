package session

import (
	"fmt"
	"strings"
)

// IsMinCatalogVersion reports Selenoid catalog suffixes like "152.0-min".
// Those images are headless CI nodes: no VNC (:5900), no DevTools proxy (:7070),
// no video framebuffer. Full tags (without -min) keep the desktop extras.
func IsMinCatalogVersion(version string) bool {
	return strings.HasSuffix(strings.TrimSpace(version), "-min")
}

// CatalogVersion is the version the user asked the hub for (W3C or JSONWP).
func (c Caps) CatalogVersion() string {
	v := strings.TrimSpace(c.Version)
	if v == "" {
		v = strings.TrimSpace(c.W3CVersion)
	}
	return v
}

// MinImageCapabilityError is a W3C invalid-argument reason when a -min image
// is requested with desktop extras. Nil means the combination is allowed.
func (c Caps) MinImageCapabilityError() error {
	version := c.CatalogVersion()
	if !IsMinCatalogVersion(version) {
		return nil
	}
	var unsupported []string
	if c.VNC {
		unsupported = append(unsupported, "enableVNC")
	}
	if c.Video {
		unsupported = append(unsupported, "enableVideo")
	}
	if c.HAR {
		unsupported = append(unsupported, "enableHAR")
	}
	if len(unsupported) == 0 {
		return nil
	}
	return fmt.Errorf("%s is a headless CI image and does not support %s — use the full image, or turn those options off", version, strings.Join(unsupported, ", "))
}
