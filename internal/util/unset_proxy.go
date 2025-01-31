package util

import (
	"os"
	"slices"

	"gov.gsa.fac.cgov-util/internal/environments"
	"gov.gsa.fac.cgov-util/internal/logging"
)

func UnsetProxy() {
	if slices.Contains([]string{environments.SANDBOX, environments.DEVELOPMENT, environments.PREVIEW, environments.STAGING, environments.PRODUCTION}, os.Getenv("ENV")) {
		if IsDebugLevel("DEBUG") {
			logging.Logger.Println("Proxy:", os.Getenv("https_proxy"))
		}
		logging.Logger.Printf("Unsetting https_proxy variable...")
		os.Unsetenv("https_proxy")
		logging.Logger.Println("DEBUG - Proxy after Unsetenv():", os.Getenv("https_proxy"))
	}
}
