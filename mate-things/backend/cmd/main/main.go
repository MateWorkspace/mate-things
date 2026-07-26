package main

import (
	"os"

	compositionmain "github.com/ABA-Developer/nusapala-things/backend/internal/composition/main"
)

// @title Nusapala Things
// @version 1.0
// @description Nusapala Things is the public HTTP and MQTT entrypoint for IoT telemetry, action dispatch, and firmware OTA.
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Paste the access token. The API accepts "Bearer " followed by the access token.
func main() {
	os.Exit(compositionmain.Launch())
}
