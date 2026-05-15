package utils

import (
	"encoding/json"
	"fmt"
	"sync"

	ms "github.com/meilisearch/meilisearch-go"
)

var (
	mClient ms.ServiceManager
	once    sync.Once
)

func InitMeili() {
	once.Do(func() {
		mClient = ms.New("http://localhost:7700", ms.WithAPIKey("masterKey"))
		index := mClient.Index("locations")

		// 1. Sortable & Searchable: These strictly want *[]string
		sortable := []string{"_geo"}
		searchable := []string{"name", "district"}

		index.UpdateSortableAttributes(&sortable)
		index.UpdateSearchableAttributes(&searchable)

		// 2. Filterable: The compiler is forcing *[]interface{} here
		filterNames := []string{"category", "district"}
		filterable := make([]interface{}, len(filterNames))
		for i, v := range filterNames {
			filterable[i] = v
		}

		// This now matches the *[]interface{} signature
		index.UpdateFilterableAttributes(&filterable)

		fmt.Println("Meilisearch Service initialized successfully.")
	})
}

func SearchLocations(query string, userLat float64, userLng float64) ([]Location, error) {
	InitMeili()

	// Initialize the search request
	searchReq := &ms.SearchRequest{
		Limit: 10,
	}

	// Only apply Geo-sorting if coordinates are NOT zero
	// Note: Lat 0, Lng 0 is a valid coordinate, but for Pakistan rideshare,
	// it's a safe bet that 0.0 means "not provided".
	if userLat != 0 && userLng != 0 {
		searchReq.Sort = []string{
			fmt.Sprintf("_geoDistance(%f, %f):asc", userLat, userLng),
		}
	}

	res, err := mClient.Index("locations").Search(query, searchReq)
	if err != nil {
		return nil, err
	}

	var locations []Location
	for _, hit := range res.Hits {
		var loc Location
		rawBytes, err := json.Marshal(hit)
		if err != nil {
			continue
		}

		if err := json.Unmarshal(rawBytes, &loc); err == nil {
			locations = append(locations, loc)
		}
	}

	return locations, nil
}
