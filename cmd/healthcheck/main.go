// healthcheck is a standalone binary run by Docker's HEALTHCHECK. The bot's
// final image is FROM scratch with no shell, curl, or wget, so the health
// check has to be a static Go binary that makes the request itself.
package main

import (
	"net/http"
	"os"
	"time"
)

func main() {
	client := &http.Client{Timeout: 2 * time.Second}

	resp, err := client.Get("http://localhost:8080/health")
	if err != nil {
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
}
