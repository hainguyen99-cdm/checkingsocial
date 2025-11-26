# Quick Start Guide

## 5-Minute Setup

### Step 1: Prepare Environment
```bash
# Copy example env file
cp .env.example .env

# Edit .env and add your Farcaster bearer token
# FARCASTER_BEARER_TOKEN=your_token_here
# TARGET_FIDS=1093215
```

### Step 2: Start with Docker Compose (Recommended)
```bash
docker-compose up -d
```

Or manually:

### Step 2 (Alternative): Local Setup
```bash
# Terminal 1: Start Redis
redis-server

# Terminal 2: Install dependencies and run app
go mod download
go run main.go
```

### Step 3: Test the API
```bash
# Check if a user is a follower
curl -X POST 'http://localhost:8080/api/v1/social-action' \
  -H 'Content-Type: application/json' \
  -d '{
    "social": "farcaster",
    "action": "follow",
    "iduser": "1406368"
  }'

# Expected response: true or false
```

## What Happens Automatically

1. ✅ Redis connects on startup
2. ✅ Cronjob starts (runs every 5 minutes)
3. ✅ **First fetch happens immediately on startup** (no waiting!)
4. ✅ Followers cached in Redis
5. ✅ API endpoint ready to use

## Verify It's Working

### Check Logs
```bash
# Docker
docker-compose logs -f app

# Local
# Check console output for "Cronjob scheduler started"
```

### Check Redis
```bash
# Connect to Redis
redis-cli

# View followers
SMEMBERS farcaster:followers:1093215

# Check if specific user is follower
SISMEMBER farcaster:followers:1093215 1406368
```

## Common Issues

| Issue | Solution |
|-------|----------|
| Redis connection refused | Start Redis: `redis-server` or `docker run -d -p 6379:6379 redis:7-alpine` |
| Bearer token error | Update `FARCASTER_BEARER_TOKEN` in `.env` |
| Cronjob not running | Ensure `TARGET_FIDS` is set in `.env` |
| Port 8080 in use | Change `SERVER_PORT` in `.env` or stop other services |

## Next Steps

- Read [SETUP.md](SETUP.md) for detailed configuration
- Check [API documentation](#api-usage) for endpoint details
- Monitor Redis performance with `redis-cli --stat`

