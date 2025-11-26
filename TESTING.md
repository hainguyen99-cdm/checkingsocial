# Testing Guide

## Prerequisites

- Application running on `http://localhost:8080`
- Redis running on `localhost:6379`
- Farcaster bearer token configured

## Manual Testing

### 1. Test Redis Connection

```bash
# Connect to Redis
redis-cli

# Check if connected
ping
# Expected: PONG

# View all keys
KEYS farcaster:*

# Exit
exit
```

### 2. Test Cronjob Execution

```bash
# Watch application logs
docker-compose logs -f app

# Or if running locally, check console output

# Look for:
# - "Cronjob scheduler started"
# - "Starting to fetch followers for FID"
# - "Successfully fetched and cached X followers"
```

### 3. Test Follower Cache

```bash
redis-cli

# Check followers for a FID
SMEMBERS farcaster:followers:1093215

# Check if specific user is a follower
SISMEMBER farcaster:followers:1093215 1406368
# Expected: 1 (true) or 0 (false)

# Get total follower count
SCARD farcaster:followers:1093215

# Check last sync time
GET farcaster:sync:last:1093215
```

## API Testing

### Test 1: Check Valid Follower

```bash
curl -X POST 'http://localhost:8080/api/v1/social-action' \
  -H 'Content-Type: application/json' \
  -d '{
    "social": "farcaster",
    "action": "follow",
    "iduser": "1406368"
  }'

# Expected: true (if 1406368 is in the followers list)
```

### Test 2: Check Non-Follower

```bash
curl -X POST 'http://localhost:8080/api/v1/social-action' \
  -H 'Content-Type: application/json' \
  -d '{
    "social": "farcaster",
    "action": "follow",
    "iduser": "9999999999"
  }'

# Expected: false (if 9999999999 is not in the followers list)
```

### Test 3: Invalid User ID Format

```bash
curl -X POST 'http://localhost:8080/api/v1/social-action' \
  -H 'Content-Type: application/json' \
  -d '{
    "social": "farcaster",
    "action": "follow",
    "iduser": "invalid_id"
  }'

# Expected: 500 error with message about invalid format
```

### Test 4: Missing Required Fields

```bash
curl -X POST 'http://localhost:8080/api/v1/social-action' \
  -H 'Content-Type: application/json' \
  -d '{
    "social": "farcaster"
  }'

# Expected: 400 error - missing required fields
```

### Test 5: Unsupported Social Platform

```bash
curl -X POST 'http://localhost:8080/api/v1/social-action' \
  -H 'Content-Type: application/json' \
  -d '{
    "social": "unsupported",
    "action": "follow",
    "iduser": "123456"
  }'

# Expected: 500 error - unsupported social or action
```

## Load Testing

### Using Apache Bench

```bash
# Install ab (Apache Bench)
# macOS: brew install httpd
# Ubuntu: sudo apt-get install apache2-utils

# Test with 100 requests, 10 concurrent
ab -n 100 -c 10 -p request.json -T application/json \
  'http://localhost:8080/api/v1/social-action'
```

### Create request.json for load test

```json
{
  "social": "farcaster",
  "action": "follow",
  "iduser": "1406368"
}
```

### Using wrk (Modern Load Testing)

```bash
# Install wrk
# macOS: brew install wrk
# Ubuntu: sudo apt-get install wrk

# Create script.lua
cat > script.lua << 'EOF'
request = function()
  wrk.method = "POST"
  wrk.headers["Content-Type"] = "application/json"
  wrk.body = '{"social":"farcaster","action":"follow","iduser":"1406368"}'
  return wrk.format(nil)
end
EOF

# Run test: 4 threads, 100 connections, 30 seconds
wrk -t4 -c100 -d30s -s script.lua http://localhost:8080/api/v1/social-action
```

## Performance Benchmarks

### Expected Response Times

| Operation | Time | Notes |
|-----------|------|-------|
| Redis lookup | < 1ms | Direct cache hit |
| API response | < 5ms | Including JSON parsing |
| Follower fetch | 1-5 min | Depends on follower count |

### Load Test Results Example

```
Running 100 requests with 10 concurrent connections:

Requests per second: ~1000
Mean response time: 10ms
Min response time: 2ms
Max response time: 50ms
Failed requests: 0
```

## Debugging

### Enable Verbose Logging

```bash
# Check application logs
docker-compose logs -f app

# Or for local run, output goes to console
```

### Monitor Redis Performance

```bash
redis-cli --stat

# Output shows:
# - Commands per second
# - Memory usage
# - Connected clients
# - Evicted keys
```

### Check Cronjob Status

```bash
# View last sync time
redis-cli GET farcaster:sync:last:1093215

# View follower count
redis-cli SCARD farcaster:followers:1093215

# View sample followers
redis-cli SRANDMEMBER farcaster:followers:1093215 5
```

## Troubleshooting Tests

### Test Fails: "Connection refused"

```bash
# Check if Redis is running
redis-cli ping

# Check if app is running
curl http://localhost:8080/api/v1/social-action

# Check ports
lsof -i :6379  # Redis
lsof -i :8080  # App
```

### Test Fails: "Unsupported social or action"

```bash
# Verify request format
# Must have: social, action, iduser
# social must be "farcaster"
# action must be "follow"
```

### Test Fails: "Invalid userID format"

```bash
# Verify iduser is a valid number
# iduser must be parseable as int64
# Example valid: "1406368"
# Example invalid: "abc123" or "1.5"
```

### Cronjob Not Running

```bash
# Check if TARGET_FIDS is set
grep TARGET_FIDS .env

# Check logs for cronjob initialization
docker-compose logs app | grep -i cronjob

# Verify Redis connection
redis-cli ping
```

## Continuous Testing

### Health Check Script

```bash
#!/bin/bash

echo "=== Health Check ==="

# Check Redis
echo -n "Redis: "
redis-cli ping > /dev/null 2>&1 && echo "✓" || echo "✗"

# Check API
echo -n "API: "
curl -s http://localhost:8080/api/v1/social-action \
  -X POST \
  -H 'Content-Type: application/json' \
  -d '{"social":"farcaster","action":"follow","iduser":"1"}' \
  > /dev/null 2>&1 && echo "✓" || echo "✗"

# Check followers cached
echo -n "Followers cached: "
COUNT=$(redis-cli SCARD farcaster:followers:1093215)
echo "$COUNT followers"

# Check last sync
echo -n "Last sync: "
TIMESTAMP=$(redis-cli GET farcaster:sync:last:1093215)
if [ -z "$TIMESTAMP" ]; then
  echo "Never"
else
  date -d @$TIMESTAMP
fi
```

### Run Health Check Every Minute

```bash
# macOS/Linux
watch -n 60 ./health_check.sh

# Or using cron
* * * * * /path/to/health_check.sh >> /var/log/health_check.log 2>&1
```

## Integration Testing

### Test Full Flow

```bash
#!/bin/bash

echo "1. Clear Redis cache"
redis-cli DEL farcaster:followers:1093215
redis-cli DEL farcaster:sync:last:1093215

echo "2. Wait for cronjob to run (5 minutes)"
sleep 300

echo "3. Check if followers cached"
COUNT=$(redis-cli SCARD farcaster:followers:1093215)
echo "Followers cached: $COUNT"

echo "4. Test API with known follower"
RESULT=$(curl -s -X POST 'http://localhost:8080/api/v1/social-action' \
  -H 'Content-Type: application/json' \
  -d '{"social":"farcaster","action":"follow","iduser":"1406368"}')
echo "API Result: $RESULT"

echo "5. Test API with unknown follower"
RESULT=$(curl -s -X POST 'http://localhost:8080/api/v1/social-action' \
  -H 'Content-Type: application/json' \
  -d '{"social":"farcaster","action":"follow","iduser":"9999999999"}')
echo "API Result: $RESULT"
```

## Performance Testing Checklist

- [ ] Redis connection stable
- [ ] API responds within 5ms
- [ ] Cronjob runs every 5 minutes
- [ ] Followers cached correctly
- [ ] No memory leaks after 1 hour
- [ ] Handles 100+ concurrent requests
- [ ] Graceful shutdown works
- [ ] Error handling works correctly

