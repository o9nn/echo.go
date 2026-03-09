package tray

import (
	"github.com/o9nn/echo.go/app/tray/commontray"
	"github.com/o9nn/echo.go/app/tray/wintray"
)

func InitPlatformTray(icon, updateIcon []byte) (commontray.OllamaTray, error) {
	return wintray.InitTray(icon, updateIcon)
}
