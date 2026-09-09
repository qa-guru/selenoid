package session

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMinImageCapabilityError(t *testing.T) {
	t.Run("full image allows VNC video and HAR", func(t *testing.T) {
		caps := Caps{Version: "152.0", VNC: true, Video: true, HAR: true}
		require.NoError(t, caps.MinImageCapabilityError())
	})

	t.Run("min image without extras is allowed", func(t *testing.T) {
		caps := Caps{W3CVersion: "152.0-min"}
		require.NoError(t, caps.MinImageCapabilityError())
	})

	t.Run("min image rejects video and VNC", func(t *testing.T) {
		caps := Caps{Version: "152.0-min", VNC: true, Video: true}
		err := caps.MinImageCapabilityError()
		require.Error(t, err)
		require.Contains(t, err.Error(), "152.0-min")
		require.Contains(t, err.Error(), "enableVNC")
		require.Contains(t, err.Error(), "enableVideo")
		require.Contains(t, err.Error(), "headless CI image")
	})

	t.Run("min image rejects HAR", func(t *testing.T) {
		caps := Caps{Version: "1.62.1-min", HAR: true}
		err := caps.MinImageCapabilityError()
		require.Error(t, err)
		require.Contains(t, err.Error(), "enableHAR")
	})
}
