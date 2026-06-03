package monitoring

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"rideshare/pkgs/configuration"
	"sync"
	"time"
)

var StatsStoreInstance *StatsStore

type StatsStore struct {
	Mu    sync.RWMutex
	Stats map[string]*EndpointStats
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewStatsStore() {
	path := configuration.ConfigurationData.Alert.APIStatsPath

	// Check if file exists
	if _, err := os.Stat(path); err == nil {
		// file exists → rename with timestamp
		dir := filepath.Dir(path)
		base := filepath.Base(path)
		ext := filepath.Ext(base)
		name := base[:len(base)-len(ext)]

		timestamp := time.Now().Format("20060102_150405")

		newName := fmt.Sprintf("%s_%s%s", name, timestamp, ext)
		newPath := filepath.Join(dir, newName)

		if err := os.Rename(path, newPath); err != nil {
			// you may want to log this instead of panic
			fmt.Printf("failed to rename stats file: %v\n", err)
		}
	}

	StatsStoreInstance = &StatsStore{
		Stats: make(map[string]*EndpointStats),
	}
}
