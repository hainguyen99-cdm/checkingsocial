package farcaster

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	neynarAPIBaseURL = "https://api.neynar.com/v2"
)

// NeynarClient wraps the Neynar API client
type NeynarClient struct {
	apiKey     string
	httpClient *http.Client
}

// NewNeynarClient creates a new Neynar API client
func NewNeynarClient() (*NeynarClient, error) {
	apiKey := os.Getenv("NEYNAR_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("NEYNAR_API_KEY environment variable not set")
	}

	return &NeynarClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// FetchBulkUsersResponse represents the response from Neynar's fetchBulkUsers endpoint
type FetchBulkUsersResponse struct {
	Users []NeynarUser `json:"users"`
}

// NeynarUser represents a user from Neynar API
type NeynarUser struct {
	Fid           int64          `json:"fid"`
	Username      string         `json:"username"`
	DisplayName   string         `json:"display_name"`
	ViewerContext *ViewerContext `json:"viewer_context"`
}

// ViewerContext represents the viewer's relationship to a user
type ViewerContext struct {
	FollowedBy bool `json:"followed_by"`
	Following  bool `json:"following"`
	Blocked    bool `json:"blocked"`
	BlockedBy  bool `json:"blocked_by"`
	Muted      bool `json:"muted"`
	MutedBy    bool `json:"muted_by"`
}

// FetchBulkUsersRequest represents the request to fetchBulkUsers endpoint
type FetchBulkUsersRequest struct {
	Fids      []int64 `json:"fids"`
	ViewerFid int64   `json:"viewer_fid,omitempty"`
}

// FetchBulkUsers fetches multiple users and their relationship to a viewer
// This is useful for checking if users follow a target FID
func (nc *NeynarClient) FetchBulkUsers(ctx context.Context, fids []int64, viewerFid int64) (*FetchBulkUsersResponse, error) {
	if len(fids) == 0 {
		return &FetchBulkUsersResponse{Users: []NeynarUser{}}, nil
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/farcaster/user/bulk", neynarAPIBaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	for _, fid := range fids {
		q.Add("fids", strconv.FormatInt(fid, 10))
	}
	if viewerFid > 0 {
		q.Add("viewer_fid", strconv.FormatInt(viewerFid, 10))
	}
	req.URL.RawQuery = q.Encode()

	// Set headers
	req.Header.Set("accept", "application/json")
	req.Header.Set("x-api-key", nc.apiKey)

	// Make request
	resp, err := nc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Log response for debugging (no secrets)
	// Note: API key is not logged. Body is truncated to avoid huge logs.
	raw := string(respBody)
	if len(raw) > 2000 {
		raw = raw[:2000] + "...(truncated)"
	}
	log.Printf("[Neynar][DEBUG] GET %s?%s status=%d body=%s", url, req.URL.RawQuery, resp.StatusCode, raw)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var result FetchBulkUsersResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return &result, nil
}

// FollowersResponse represents the response from Neynar's followers endpoint
type FollowersResponse struct {
	Result FollowersResult `json:"result"`
	Next   *NextCursor     `json:"next,omitempty"`
}

type FollowersResult struct {
	Users []FollowerUserInfo `json:"users"`
}

type FollowerUserInfo struct {
	Fid int64 `json:"fid"`
}

type NextCursor struct {
	Cursor string `json:"cursor"`
}

// FetchFollowers fetches followers for a target FID using Neynar API
// This replaces the direct API call to farcaster.xyz
func (nc *NeynarClient) FetchFollowers(ctx context.Context, targetFID string, limit int, cursor string) (*FollowersResponse, error) {
	if limit == 0 {
		limit = 100
	}

	url := fmt.Sprintf("%s/farcaster/followers", neynarAPIBaseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add query parameters
	q := req.URL.Query()
	q.Add("fid", targetFID)
	q.Add("limit", strconv.Itoa(limit))
	if cursor != "" {
		q.Add("cursor", cursor)
	}
	req.URL.RawQuery = q.Encode()

	// Set headers
	req.Header.Set("accept", "application/json")
	req.Header.Set("x-api-key", nc.apiKey)

	// Make request
	resp, err := nc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var result FollowersResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

// CheckFollowUsingNeynar checks if a user follows a target FID using Neynar API
// userFID: the FID to check
// targetFID: the FID we want to check if userFID follows
func (nc *NeynarClient) CheckFollowUsingNeynar(ctx context.Context, userFID int64, targetFID int64) (bool, error) {
	// Fetch the user with viewer context
	resp, err := nc.FetchBulkUsers(ctx, []int64{targetFID}, userFID)
	if err != nil {
		return false, fmt.Errorf("failed to fetch user: %w", err)
	}

	if len(resp.Users) == 0 {
		return false, fmt.Errorf("user not found")
	}

	user := resp.Users[0]
	if user.ViewerContext == nil {
		return false, nil
	}

	// Check if the viewer (userFID) is following the target
	return user.ViewerContext.Following, nil
}
