# Farcaster Follower Cronjob Setup Guide

## Overview

This application implements a cronjob that:
1. Fetches Farcaster followers for a target FID every 5 minutes
2. Caches the follower FIDs in Redis
3. Provides an API endpoint to check if a user is a follower by querying Redis (instead of calling the API)

## Architecture

### Components

1. **Redis Cache** (`pkg/cache/redis.go`)
   - Manages Redis connection and operations
   - Stores followers as Redis Sets with key format: `farcaster:followers:{targetFID}`
   - Tracks last sync time for each target FID

2. **Farcaster Follower Fetcher** (`farcaster/follower_fetcher.go`)
   - Fetches all followers using pagination (cursor-based)
   - Handles API rate limiting with delays
   - Stores fetched FIDs in Redis cache

3. **Cronjob Scheduler** (`pkg/cronjob/scheduler.go`)
   - Runs every 5 minutes using `robfig/cron`
   - Supports multiple target FIDs (comma-separated in env)
   - Automatically fetches and caches followers

4. **Updated CheckFollow** (`farcaster/client.go`)
   - Now queries Redis cache instead of calling the API
   - Fast response times (Redis lookup)
   - No API rate limiting concerns

## Environment Variables

Create a `.env` file in the project root with the following variables:

```env
# Farcaster Configuration
FARCASTER_BEARER_TOKEN=your_bearer_token_here
TARGET_FID=1093215
TARGET_FIDS=1093215,1093216,1093217

# Redis Configuration
REDIS_ADDR=localhost:6379
REDIS_DB=0
REDIS_PASSWORD=

# Server Configuration
SERVER_PORT=8080
```

### Environment Variables Explanation

| Variable | Description | Example |
|----------|-------------|---------|
| `FARCASTER_BEARER_TOKEN` | Bearer token for Farcaster API authentication | `MK-CJD77WmZKAO300T9sW2Z...` |
| `TARGET_FID` | Primary target FID (used for CheckFollow) | `1093215` |
| `TARGET_FIDS` | Comma-separated list of FIDs to fetch followers for | `1093215,1093216,1093217` |
| `REDIS_ADDR` | Redis server address | `localhost:6379` |
| `REDIS_DB` | Redis database number (0-15) | `0` |
| `REDIS_PASSWORD` | Redis password (empty if no auth) | `` |
| `SERVER_PORT` | Server port | `8080` |

## Installation & Setup

### Prerequisites

- Go 1.24.0 or higher
- Redis 7.0 or higher
- Docker & Docker Compose (optional)

### Local Setup

1. **Install dependencies:**
   ```bash
   go mod download
   go mod tidy
   ```

2. **Start Redis:**
   ```bash
   # Using Docker
   docker run -d -p 6379:6379 redis:7-alpine

   # Or using local Redis installation
   redis-server
   ```

3. **Create `.env` file:**
   ```bash
   cp .env.example .env
   # Edit .env with your Farcaster bearer token and target FIDs
   ```

4. **Run the application:**
   ```bash
   go run main.go
   ```

### Docker Compose Setup

1. **Create `.env` file:**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

2. **Start services:**
   ```bash
   docker-compose up -d
   ```

3. **View logs:**
   ```bash
   docker-compose logs -f app
   ```

4. **Stop services:**
   ```bash
   docker-compose down
   ```

## API Usage

### Check if User is a Follower

**Endpoint:** `POST /api/v1/social-action`

**Request:**
```bash
curl -X POST 'http://localhost:8080/api/v1/social-action' \
  -H 'Content-Type: application/json' \
  -d '{
    "social": "farcaster",
    "action": "follow",
    "iduser": "1829782561419415552"
  }'
```

**Response (Success):**
```json
true
```

**Response (Not a Follower):**
```json
false
```

**Response (Error):**
```json
{
  "error": "invalid userID format: strconv.ParseInt: parsing \"invalid\": invalid syntax"
}
```

## Cronjob Behavior

The cronjob runs every 5 minutes and:

1. **Fetches followers** for each FID in `TARGET_FIDS`
2. **Clears old cache** before fetching new data
3. **Paginates through all followers** using cursor-based pagination
4. **Stores FIDs in Redis** as a Set for fast lookups
5. **Updates last sync time** for tracking

### Logs Example

```
2025-11-26 07:46:47 Starting to fetch followers for FID: 1093215
2025-11-26 07:46:47 Fetching page 1: https://client.farcaster.xyz/v2/followers?fid=1093215&limit=15
2025-11-26 07:46:48 Cached 15 followers (page 1)
2025-11-26 07:46:48 Fetching page 2: https://client.farcaster.xyz/v2/followers?cursor=...&fid=1093215&limit=15
2025-11-26 07:46:49 Cached 15 followers (page 2)
...
2025-11-26 07:47:05 Successfully fetched and cached 150 followers for FID: 1093215
```

## Redis Data Structure

### Followers Set
```
Key: farcaster:followers:{targetFID}
Type: Redis Set
Members: FIDs of followers (integers)

Example:
SMEMBERS farcaster:followers:1093215
1) "1406368"
2) "466033"
3) "1108383"
...
```

### Last Sync Time
```
Key: farcaster:sync:last:{targetFID}
Type: String (Unix timestamp)
Value: Unix timestamp of last sync

Example:
GET farcaster:sync:last:1093215
"1732605965"
```

## Redis CLI Commands

```bash
# Connect to Redis
redis-cli

# Check followers for a FID
SMEMBERS farcaster:followers:1093215

# Check if a specific FID is a follower
SISMEMBER farcaster:followers:1093215 1406368

# Get number of followers
SCARD farcaster:followers:1093215

# Get last sync time
GET farcaster:sync:last:1093215

# Clear followers cache
DEL farcaster:followers:1093215

# View all keys
KEYS farcaster:*
```

## Performance Considerations

1. **Redis Lookup**: O(1) average time complexity
2. **Follower Fetch**: Paginated with 15 followers per request
3. **Rate Limiting**: 500ms delay between API requests to avoid rate limiting
4. **Memory**: Each follower FID takes ~8 bytes in Redis

## Troubleshooting

### Redis Connection Error
```
Failed to connect to Redis: dial tcp localhost:6379: connect: connection refused
```
**Solution:** Ensure Redis is running on the specified address

### Invalid Bearer Token
```
API request failed with status 401
```
**Solution:** Check that `FARCASTER_BEARER_TOKEN` is correct in `.env`

### Cronjob Not Running
```
Warning: TARGET_FIDS environment variable not set, cronjob will not run
```
**Solution:** Add `TARGET_FIDS` to `.env` file

### Slow Follower Fetch
- Check network connectivity
- Verify API rate limits
- Monitor Redis performance

## File Structure

```
checkingsocial/
├── main.go                          # Entry point with Redis & cronjob init
├── go.mod                           # Go dependencies
├── docker-compose.yml               # Docker services (Redis + App)
├── .env.example                     # Environment variables template
├── pkg/
│   ├── cache/
│   │   └── redis.go                # Redis client & operations
│   └── cronjob/
│       └── scheduler.go            # Cronjob scheduler (5-minute interval)
├── farcaster/
│   ├── client.go                   # Updated CheckFollow (uses Redis)
│   └── follower_fetcher.go         # Fetches & caches followers
├── internal/
│   ├── handler/
│   │   └── social.go               # API handlers
│   ├── service/
│   │   └── social_checker.go       # Business logic
│   └── model/
│       └── social.go               # Data models
└── twitter/
    └── client.go                   # Twitter client (unchanged)
```

## Next Steps

1. Add `.env` file with your Farcaster bearer token
2. Start Redis (locally or via Docker)
3. Run the application
4. Monitor logs to see cronjob execution
5. Use the API endpoint to check followers

## Support

For issues or questions, check:
- Application logs for error messages
- Redis connection status
- Farcaster API bearer token validity
- Environment variables configuration

