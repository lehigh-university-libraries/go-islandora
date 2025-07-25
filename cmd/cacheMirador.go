package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/spf13/cobra"
)

type URLItem struct {
	URL string `json:"url"`
}

var (
	endpoint string
	workers  int
)

// csvCmcacheMiradord represents the csv command
var cacheMirador = &cobra.Command{
	Use:   "cache-mirador",
	Short: "Make sure IIIF server has cached images",
	RunE: func(cmd *cobra.Command, args []string) error {
		urls, err := fetchURLs(endpoint)
		if err != nil {
			return fmt.Errorf("error fetching URLs: %v", err)
		}

		type job struct {
			URL string
		}

		jobs := make(chan job)

		var wg sync.WaitGroup

		for i := 0; i < workers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for j := range jobs {

					execCtx, cancelExec := chromedp.NewExecAllocator(context.Background(), chromedp.DefaultExecAllocatorOptions[:]...)

					ctx, cancelCtx := chromedp.NewContext(execCtx)
					err := warmURL(ctx, j.URL)
					cancelCtx()
					cancelExec()
					if err != nil {
						slog.Error("worker Failed", "url", j.URL, "err", err)
					}
				}
			}(i + 1)
		}

		for _, u := range urls {
			jobs <- job{URL: u}
		}
		close(jobs)

		wg.Wait()
		return nil
	},
}

func fetchURLs(endpoint string) ([]string, error) {
	resp, err := http.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("GET failed: %w", err)
	}
	defer resp.Body.Close()

	var items []URLItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}

	var urls []string
	for _, item := range items {
		if item.URL != "" {
			urls = append(urls, item.URL)
		}
	}
	return urls, nil
}

func warmURL(ctx context.Context, url string) error {
	slog.Info("Warming URL", "url", url)

	localCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tasks := chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.Poll(`
			(function($) {
					const id = $('.block-mirador[data-once*="mirador-viewer"]').attr('id');
					if (id == undefined) return false;
					const state = Drupal.IslandoraMirador.instances['#' + id].store.getState();
					if (!state.manifests || Object.keys(state.manifests).length === 0) return false;
					if (!state.windows || Object.keys(state.windows).length === 0) return false;
					const windows = Object.values(state.windows);
					return windows.some(window => window.canvasId && window.manifestId);
			})(jQuery);`, nil),
	}

	if err := chromedp.Run(localCtx, tasks); err != nil {
		return err
	}

	return nil
}

func init() {
	cacheMirador.Flags().StringVar(&endpoint, "endpoint", "", "Remote JSON endpoint returning array of {url: ...} (required)")
	cacheMirador.Flags().IntVar(&workers, "workers", 2, "Number of concurrent workers")
	err := cacheMirador.MarkFlagRequired("endpoint")
	if err != nil {
		slog.Error("Unable to mark endpoint flag as required for cache-mirador command")
		os.Exit(1)
	}
	rootCmd.AddCommand(cacheMirador)
}
