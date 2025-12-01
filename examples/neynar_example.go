package main

import (
	"checkingsocial/farcaster"
	"checkingsocial/pkg/cache"
	"context"
	"log"
	"os"
)

// Example 1: Check if a user follows a target FID using Neynar API
func exampleCheckFollow() {
	log.Println("\n=== Example 1: Check if User Follows Target ===")

	// Set environment variables (or load from .env)
	os.Setenv("NEYNAR_API_KEY", "your_api_key_here")
	os.Setenv("USE_NEYNAR_API", "true")
	os.Setenv("TARGET_FIDS", "1112245")

	// Check if user 1093215 follows the target FID
	isFollower, err := farcaster.CheckFollow("1093215")
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	log.Printf("User 1093215 is follower: %v", isFollower)
}

// Example 2: Fetch bulk users with viewer context
func exampleFetchBulkUsers() {
	log.Println("\n=== Example 2: Fetch Bulk Users with Viewer Context ===")

	os.Setenv("NEYNAR_API_KEY", "your_api_key_here")

	client, err := farcaster.NewNeynarClient()
	if err != nil {
		log.Printf("Error creating client: %v", err)
		return
	}

	ctx := context.Background()

	// Fetch multiple users with viewer context
	// The viewer is FID 1112245, and we want to see their relationships to FIDs 1093215 and 1112245
	resp, err := client.FetchBulkUsers(ctx, []int64{1093215, 1112245}, 1112245)
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		return
	}

	log.Printf("Fetched %d users:", len(resp.Users))
	for _, user := range resp.Users {
		log.Printf("\nUser: %s (@%s)", user.DisplayName, user.Username)
		log.Printf("  FID: %d", user.Fid)

		if user.ViewerContext != nil {
			log.Printf("  Viewer is following: %v", user.ViewerContext.Following)
			log.Printf("  Viewer is followed by: %v", user.ViewerContext.FollowedBy)
			log.Printf("  Viewer has blocked: %v", user.ViewerContext.Blocked)
			log.Printf("  Viewer has muted: %v", user.ViewerContext.Muted)
		}
	}
}

// Example 3: Fetch followers with pagination
func exampleFetchFollowers() {
	log.Println("\n=== Example 3: Fetch Followers with Pagination ===")

	os.Setenv("NEYNAR_API_KEY", "your_api_key_here")

	client, err := farcaster.NewNeynarClient()
	if err != nil {
		log.Printf("Error creating client: %v", err)
		return
	}

	ctx := context.Background()

	// Fetch first page of followers for FID 1112245
	resp, err := client.FetchFollowers(ctx, "1112245", 10, "")
	if err != nil {
		log.Printf("Error fetching followers: %v", err)
		return
	}

	log.Printf("Fetched %d followers:", len(resp.Result.Users))
	for i, user := range resp.Result.Users {
		log.Printf("  %d. FID: %d", i+1, user.Fid)
	}

	if resp.Next != nil && resp.Next.Cursor != "" {
		log.Printf("\nNext cursor available: %s (for pagination)", resp.Next.Cursor)
	}
}

// Example 4: Fetch and cache all followers
func exampleFetchAndCacheFollowers() {
	log.Println("\n=== Example 4: Fetch and Cache All Followers ===")

	os.Setenv("NEYNAR_API_KEY", "your_api_key_here")
	os.Setenv("REDIS_ADDR", "localhost:6379")

	// Initialize Redis
	if err := cache.InitRedis(); err != nil {
		log.Printf("Error initializing Redis: %v", err)
		return
	}
	defer cache.Close()

	client, err := farcaster.NewNeynarClient()
	if err != nil {
		log.Printf("Error creating client: %v", err)
		return
	}

	ctx := context.Background()

	// Fetch all followers for FID 1112245 and cache them in Redis
	if err := client.FetchAndCacheFollowersUsingNeynar(ctx, "1112245"); err != nil {
		log.Printf("Error fetching and caching followers: %v", err)
		return
	}

	// Get follower count from cache
	count, err := cache.GetFollowerCount(ctx, "1112245")
	if err != nil {
		log.Printf("Error getting follower count: %v", err)
		return
	}

	log.Printf("Total followers cached: %d", count)

	// Check if a specific user is a follower
	isFollower, err := cache.IsFollower(ctx, "1112245", 1093215)
	if err != nil {
		log.Printf("Error checking follower: %v", err)
		return
	}

	log.Printf("Is FID 1093215 a follower: %v", isFollower)
}

// Example 5: Direct Neynar API check
func exampleDirectNeynarCheck() {
	log.Println("\n=== Example 5: Direct Neynar API Check ===")

	os.Setenv("NEYNAR_API_KEY", "your_api_key_here")

	client, err := farcaster.NewNeynarClient()
	if err != nil {
		log.Printf("Error creating client: %v", err)
		return
	}

	ctx := context.Background()

	// Check if user 1093215 follows user 1112245
	isFollowing, err := client.CheckFollowUsingNeynar(ctx, 1093215, 1112245)
	if err != nil {
		log.Printf("Error checking follow: %v", err)
		return
	}

	log.Printf("User 1093215 follows user 1112245: %v", isFollowing)
}

// Main function to run all examples
func main() {
	log.Println("Neynar SDK Integration Examples")
	log.Println("================================")

	// Note: You need to set NEYNAR_API_KEY in environment or .env file
	// Get your API key from https://dev.neynar.com/

	// Uncomment the examples you want to run:

	// exampleCheckFollow()
	// exampleFetchBulkUsers()
	// exampleFetchFollowers()
	// exampleFetchAndCacheFollowers()
	// exampleDirectNeynarCheck()

	log.Println("\nTo run examples:")
	log.Println("1. Set NEYNAR_API_KEY environment variable")
	log.Println("2. Uncomment the example function calls in main()")
	log.Println("3. Run: go run examples/neynar_example.go")
}

