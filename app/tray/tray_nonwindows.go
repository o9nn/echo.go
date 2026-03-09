//go:build !windows

package tray

import (
	"errors"

	"github.com/o9nn/echo.go/app/tray/commontray"
)

func InitPlatformTray(icon, updateIcon []byte) (commontray.OllamaTray, error) {
	return nil, errors.New("not implemented")
}
