package cronjob

import (
	"checkingsocial/farcaster"
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

var cronScheduler *cron.Cron

// InitCronScheduler khởi tạo cron scheduler
func InitCronScheduler() error {
	cronScheduler = cron.New()

	// Get target FIDs from environment variable (comma-separated)
	targetFIDsStr := os.Getenv("TARGET_FIDS")
	if targetFIDsStr == "" {
		log.Println("Warning: TARGET_FIDS environment variable not set, cronjob will not run")
		return nil
	}

	targetFIDs := strings.Split(targetFIDsStr, ",")
	for i := range targetFIDs {
		targetFIDs[i] = strings.TrimSpace(targetFIDs[i])
	}

	// Schedule the job to run every 5 minutes
	_, err := cronScheduler.AddFunc("*/5 * * * *", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		for _, targetFID := range targetFIDs {
			if targetFID == "" {
				continue
			}

			log.Printf("Running cronjob to fetch followers for FID: %s", targetFID)
			if err := farcaster.FetchAndCacheFollowers(ctx, targetFID); err != nil {
				log.Printf("Error fetching followers for FID %s: %v", targetFID, err)
			}
		}
	})

	if err != nil {
		return err
	}

	cronScheduler.Start()
	log.Println("Cronjob scheduler started - will run every 5 minutes")
	return nil
}

// StopCronScheduler dừng cron scheduler
func StopCronScheduler() {
	if cronScheduler != nil {
		cronScheduler.Stop()
		log.Println("Cronjob scheduler stopped")
	}
}

// GetCronScheduler trả về cron scheduler
func GetCronScheduler() *cron.Cron {
	return cronScheduler
}

// FetchFollowersNow fetches followers immediately (used on startup)
func FetchFollowersNow() error {
	targetFIDsStr := os.Getenv("TARGET_FIDS")
	if targetFIDsStr == "" {
		return nil // Skip if not configured
	}

	targetFIDs := strings.Split(targetFIDsStr, ",")
	for i := range targetFIDs {
		targetFIDs[i] = strings.TrimSpace(targetFIDs[i])
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	for _, targetFID := range targetFIDs {
		if targetFID == "" {
			continue
		}

		log.Printf("Fetching followers on startup for FID: %s", targetFID)
		if err := farcaster.FetchAndCacheFollowers(ctx, targetFID); err != nil {
			log.Printf("Error fetching followers on startup for FID %s: %v", targetFID, err)
			return err
		}
	}

	return nil
}

