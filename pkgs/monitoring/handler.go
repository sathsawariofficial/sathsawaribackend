package monitoring

import (
	"net/http"
	"sync"
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
	StatsStoreInstance = &StatsStore{
		Stats: make(map[string]*EndpointStats),
	}
}
