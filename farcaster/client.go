package farcaster

import (
	"checkingsocial/pkg/cache"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// CheckFollow checks if a user (userID) is a follower of the target FID by querying Redis cache
// userID is the FID of the user we want to check
func CheckFollow(userID string) (bool, error) {
	// Load environment variables from .env file
	_ = godotenv.Load()

	targetFIDStr := os.Getenv("TARGET_FID")
	if targetFIDStr == "" {
		return false, fmt.Errorf("TARGET_FID environment variable not set or empty")
	}

	// Parse userID as int64
	userFID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return false, fmt.Errorf("invalid userID format: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if userID is in the followers set in Redis
	isFollower, err := cache.IsFollower(ctx, targetFIDStr, userFID)
	if err != nil {
		log.Printf("Error checking follower in Redis: %v", err)
		return false, err
	}

	return isFollower, nil
}
