package farcaster

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// CheckFollow checks if a user (userID) follows the TARGET_FIDS using Neynar API only
// Redis and cronjob paths have been removed from this flow.
func CheckFollow(userID string) (bool, error) {
	// Load environment variables from .env file
	_ = godotenv.Load()

	targetFIDStr := os.Getenv("TARGET_FIDS")
	if targetFIDStr == "" {
		return false, fmt.Errorf("TARGET_FIDS environment variable not set or empty")
	}

	// Parse userID as int64
	userFID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return false, fmt.Errorf("invalid userID format: %w", err)
	}

	// Parse targetFID as int64
	targetFID, err := strconv.ParseInt(targetFIDStr, 10, 64)
	if err != nil {
		return false, fmt.Errorf("invalid targetFID format: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("[Neynar][DEBUG] Forcing Neynar API path for follow check TARGET_FIDS=%s userFID=%d", targetFIDStr, userFID)
	return CheckFollowUsingNeynar(ctx, userFID, targetFID)
}

// CheckFollowUsingNeynar checks if a user follows a target FID using Neynar API
func CheckFollowUsingNeynar(ctx context.Context, userFID int64, targetFID int64) (bool, error) {
	log.Printf("[Neynar][DEBUG] CheckFollowUsingNeynar userFID=%d targetFID=%d", userFID, targetFID)
	client, err := NewNeynarClient()
	if err != nil {
		return false, fmt.Errorf("failed to create Neynar client: %w", err)
	}

	res, err := client.CheckFollowUsingNeynar(ctx, userFID, targetFID)
	if err != nil {
		log.Printf("[Neynar][DEBUG] CheckFollowUsingNeynar error=%v", err)
		return false, err
	}
	log.Printf("[Neynar][DEBUG] CheckFollowUsingNeynar result=%v", res)
	return res, nil
}
