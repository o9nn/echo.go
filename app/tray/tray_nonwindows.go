//go:build !windows

package tray

import (
	"errors"

	"github.com/cogpy/echo9llama/app/tray/commontray"
)

func InitPlatformTray(icon, updateIcon []byte) (commontray.OllamaTray, error) {
	return nil, errors.New("not implemented")
}
