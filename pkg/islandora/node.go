package islandora

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/lehigh-university-libraries/go-islandora/api"
)

type Nid struct {
	Nid string `json:"nid"`
}

func FetchMembers(url string) ([]Nid, error) {
	cacheFile := getCacheFilename(url)

	// Try to read from cache first
	if isCacheValid(cacheFile) {
		data, err := os.ReadFile(cacheFile)
		if err == nil {
			var nids []Nid
			if json.Unmarshal(data, &nids) == nil {
				return nids, nil
			}
		}
	}

	// Cache miss or invalid - fetch from API
	req, err := getRequest(url)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var nids []Nid
	err = decodeJsonResponse(resp, &nids)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := ensureCacheDir(); err == nil {
		if data, err := json.Marshal(nids); err == nil {
			err = os.WriteFile(cacheFile, data, 0644)
			if err != nil {
				slog.Error("Unable to write file", "file", cacheFile, "err", err)
			}
		}
	}

	return nids, nil
}

func FetchNode(url string) (*api.IslandoraObject, error) {
	cacheFile := getCacheFilename(url)

	// Try to read from cache first
	if isCacheValid(cacheFile) {
		data, err := os.ReadFile(cacheFile)
		if err == nil {
			var obj api.IslandoraObject
			if json.Unmarshal(data, &obj) == nil {
				return &obj, nil
			}
		}
	}

	// Cache miss or invalid - fetch from API
	req, err := getRequest(url)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	var obj api.IslandoraObject
	err = decodeJsonResponse(resp, &obj)
	if err != nil {
		return nil, err
	}

	// Cache the result
	if err := ensureCacheDir(); err == nil {
		if data, err := json.Marshal(obj); err == nil {
			err = os.WriteFile(cacheFile, data, 0644)
			if err != nil {
				slog.Error("Unable to write file", "file", cacheFile, "err", err)
			}
		}
	}

	return &obj, nil
}

// breadth first search of all descendants for a node
func FetchNodes(baseUrl string, nid int) ([]*api.IslandoraObject, error) {
	var allNodes []*api.IslandoraObject
	queue := []string{strconv.Itoa(nid)}

	for len(queue) > 0 {
		currentNid := queue[0]
		queue = queue[1:]

		url := fmt.Sprintf("%s/node/%s?_format=json", baseUrl, currentNid)
		node, err := FetchNode(url)
		if err != nil {
			return nil, err
		}
		allNodes = append(allNodes, node)

		// Fetch children and add to queue
		url = fmt.Sprintf("%s/node/%s/members?_format=json", baseUrl, currentNid)
		children, err := FetchMembers(url)
		if err != nil {
			return nil, err
		}
		for _, child := range children {
			queue = append(queue, child.Nid)
		}
	}

	return allNodes, nil
}
