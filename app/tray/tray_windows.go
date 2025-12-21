package tray

import (
	"github.com/cogpy/echo9llama/app/tray/commontray"
	"github.com/cogpy/echo9llama/app/tray/wintray"
)

func InitPlatformTray(icon, updateIcon []byte) (commontray.OllamaTray, error) {
	return wintray.InitTray(icon, updateIcon)
}
