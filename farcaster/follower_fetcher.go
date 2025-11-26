package farcaster

import (
	"checkingsocial/pkg/cache"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const followersAPIURL = "https://client.farcaster.xyz/v2/followers"

// FollowerResponse represents the API response structure for followers
type FollowerResponse struct {
	Result FollowerResult `json:"result"`
	Next   Next           `json:"next"`
}

type Next struct {
	Cursor string `json:"cursor"`
}

type FollowerResult struct {
	Users []FollowerUser `json:"users"`
}

type FollowerUser struct {
	Fid int64 `json:"fid"`
}

// FetchAndCacheFollowers fetches all followers for a target FID and caches them in Redis
func FetchAndCacheFollowers(ctx context.Context, targetFID string) error {
	log.Printf("Starting to fetch followers for FID: %s", targetFID)

	bearerToken := os.Getenv("FARCASTER_BEARER_TOKEN")
	if bearerToken == "" {
		return fmt.Errorf("FARCASTER_BEARER_TOKEN environment variable not set")
	}

	// Clear existing followers
	if err := cache.ClearFollowers(ctx, targetFID); err != nil {
		log.Printf("Warning: failed to clear existing followers: %v", err)
	}

	var cursor string
	totalFetched := 0
	pageCount := 0

	for {
		pageCount++
		url := fmt.Sprintf("%s?fid=%s&limit=15", followersAPIURL, targetFID)
		if cursor != "" {
			url = fmt.Sprintf("%s?cursor=%s&fid=%s&limit=15", followersAPIURL, cursor, targetFID)
		}

		log.Printf("Fetching page %d: %s", pageCount, url)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		setFollowerHeaders(req, bearerToken)

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to make request: %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}

		var apiResp FollowerResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}

		// Extract FIDs and cache them
		if len(apiResp.Result.Users) > 0 {
			fids := make([]int64, len(apiResp.Result.Users))
			for i, user := range apiResp.Result.Users {
				fids[i] = user.Fid
			}

			if err := cache.AddFollowerFIDs(ctx, targetFID, fids); err != nil {
				return fmt.Errorf("failed to cache followers: %w", err)
			}

			totalFetched += len(fids)
			log.Printf("Cached %d followers (page %d)", len(fids), pageCount)
		}

		// Check if there are more pages
		if apiResp.Next.Cursor == "" {
			break
		}
		cursor = apiResp.Next.Cursor

		// Add a small delay to avoid rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	// Update last sync time
	if err := cache.SetLastSyncTime(ctx, targetFID, time.Now()); err != nil {
		log.Printf("Warning: failed to set last sync time: %v", err)
	}

	log.Printf("Successfully fetched and cached %d followers for FID: %s", totalFetched, targetFID)
	return nil
}

// setFollowerHeaders sets the required headers for the API request
func setFollowerHeaders(req *http.Request, bearerToken string) {
	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "en-US,en;q=0.9")
	req.Header.Set("authorization", "Bearer "+bearerToken)
	req.Header.Set("content-type", "application/json; charset=utf-8")
	req.Header.Set("origin", "https://farcaster.xyz")
	req.Header.Set("referer", "https://farcaster.xyz/")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36")
}
