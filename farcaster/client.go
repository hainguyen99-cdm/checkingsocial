package farcaster

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const farcasterAPIURL = "https://client.farcaster.xyz/v2/following"

// Structs for Farcaster API response
type APIResponse struct {
	Result Result `json:"result"`
	Next   Next   `json:"next"`
}

type Result struct {
	Users []User `json:"users"`
}

type User struct {
	Fid int `json:"fid"`
}

type Next struct {
	Cursor string `json:"cursor"`
}

// CheckFollow checks if a user is following the target FID.
func CheckFollow(userID string) (bool, error) {
	// Load environment variables from .env file
	_ = godotenv.Load()

	targetFIDStr := os.Getenv("TARGET_FID")
	if targetFIDStr == "" {
		return false, fmt.Errorf("TARGET_FID environment variable not set or empty")
	}

	targetFID, err := strconv.Atoi(targetFIDStr)
	if err != nil {
		return false, fmt.Errorf("invalid TARGET_FID: %w", err)
	}

	var cursor string

	for {
		url := fmt.Sprintf("%s?fid=%s&limit=15", farcasterAPIURL, userID)
		if cursor != "" {
			url = fmt.Sprintf("%s?cursor=%s&fid=%s&limit=15", farcasterAPIURL, cursor, userID)
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return false, err
		}

		// Set headers from the curl command
		req.Header.Set("accept", "*/*")
		req.Header.Set("accept-language", "en-US,en;q=0.9")
		bearerToken := os.Getenv("FARCASTER_BEARER_TOKEN")
		if bearerToken == "" {
			return false, fmt.Errorf("FARCASTER_BEARER_TOKEN environment variable not set or empty")
		}
		req.Header.Set("authorization", "Bearer "+bearerToken)
		req.Header.Set("content-type", "application/json; charset=utf-8")
		req.Header.Set("origin", "https://farcaster.xyz")
		req.Header.Set("referer", "https://farcaster.xyz/")
		req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return false, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}

		if resp.StatusCode != http.StatusOK {
			return false, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
		}

		var apiResp APIResponse
		if err := json.Unmarshal(body, &apiResp); err != nil {
			return false, err
		}

		for _, user := range apiResp.Result.Users {
			if user.Fid == targetFID {
				return true, nil
			}
		}

		if apiResp.Next.Cursor == "" {
			break // No more pages
		}
		cursor = apiResp.Next.Cursor
	}

	return false, nil
}
