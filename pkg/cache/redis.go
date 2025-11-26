package cache

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

// InitRedis khởi tạo kết nối Redis
func InitRedis() error {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisDB := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if db, err := strconv.Atoi(dbStr); err == nil {
			redisDB = db
		}
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		DB:       redisDB,
		Password: os.Getenv("REDIS_PASSWORD"),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Connected to Redis successfully")
	return nil
}

// GetRedisClient trả về Redis client
func GetRedisClient() *redis.Client {
	return redisClient
}

// AddFollowerFID thêm một FID vào set followers
func AddFollowerFID(ctx context.Context, targetFID string, fid int64) error {
	key := fmt.Sprintf("farcaster:followers:%s", targetFID)
	return redisClient.SAdd(ctx, key, fid).Err()
}

// AddFollowerFIDs thêm nhiều FID vào set followers (batch)
func AddFollowerFIDs(ctx context.Context, targetFID string, fids []int64) error {
	key := fmt.Sprintf("farcaster:followers:%s", targetFID)
	if len(fids) == 0 {
		return nil
	}

	args := make([]interface{}, len(fids))
	for i, fid := range fids {
		args[i] = fid
	}

	return redisClient.SAdd(ctx, key, args...).Err()
}

// IsFollower kiểm tra xem một FID có trong set followers không
func IsFollower(ctx context.Context, targetFID string, fid int64) (bool, error) {
	key := fmt.Sprintf("farcaster:followers:%s", targetFID)
	return redisClient.SIsMember(ctx, key, fid).Val(), nil
}

// GetFollowerCount lấy số lượng followers
func GetFollowerCount(ctx context.Context, targetFID string) (int64, error) {
	key := fmt.Sprintf("farcaster:followers:%s", targetFID)
	return redisClient.SCard(ctx, key).Result()
}

// ClearFollowers xóa tất cả followers của một FID
func ClearFollowers(ctx context.Context, targetFID string) error {
	key := fmt.Sprintf("farcaster:followers:%s", targetFID)
	return redisClient.Del(ctx, key).Err()
}

// SetLastSyncTime lưu thời gian sync cuối cùng
func SetLastSyncTime(ctx context.Context, targetFID string, timestamp time.Time) error {
	key := fmt.Sprintf("farcaster:sync:last:%s", targetFID)
	return redisClient.Set(ctx, key, timestamp.Unix(), 0).Err()
}

// GetLastSyncTime lấy thời gian sync cuối cùng
func GetLastSyncTime(ctx context.Context, targetFID string) (time.Time, error) {
	key := fmt.Sprintf("farcaster:sync:last:%s", targetFID)
	val, err := redisClient.Get(ctx, key).Result()
	if err == redis.Nil {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, err
	}

	timestamp, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(timestamp, 0), nil
}

// Close đóng kết nối Redis
func Close() error {
	if redisClient != nil {
		return redisClient.Close()
	}
	return nil
}

