package monitoring

func GetStats() map[string]*EndpointStats {
	StatsStoreInstance.Mu.RLock()
	defer StatsStoreInstance.Mu.RUnlock()

	result := make(map[string]*EndpointStats)
	for k, v := range StatsStoreInstance.Stats {
		result[k] = v
	}

	return result
}
