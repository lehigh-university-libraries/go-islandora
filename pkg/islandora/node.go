package islandora

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/lehigh-university-libraries/go-islandora/api"
)

type Nid struct {
	Nid string `json:"nid"`
}

const cacheDir = "/tmp/islandora"

// getCacheFilename creates a unique filename based on URL
func getCacheFilename(url string) string {
	hash := md5.Sum([]byte(url))
	return filepath.Join(cacheDir, fmt.Sprintf("%x.json", hash))
}

// isCacheValid checks if cache file exists and is less than 1 day old
func isCacheValid(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return time.Since(info.ModTime()) < 24*time.Hour
}

// ensureCacheDir creates cache directory if it doesn't exist
func ensureCacheDir() error {
	return os.MkdirAll(cacheDir, 0755)
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

func getRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	user := os.Getenv("ISLANDORA_WORKBENCH_USERNAME")
	pass := os.Getenv("ISLANDORA_WORKBENCH_PASSWORD")
	if user != "" && pass != "" {
		req.SetBasicAuth(user, pass)
	}

	return req, nil
}

func decodeJsonResponse(resp *http.Response, obj any) error {
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code for %s: %s", resp.Request.URL, resp.Status)
	}

	err := json.NewDecoder(resp.Body).Decode(&obj)
	if err != nil {
		return fmt.Errorf("error decoding JSON response %s: %v", resp.Request.URL, err)
	}

	return nil
}
