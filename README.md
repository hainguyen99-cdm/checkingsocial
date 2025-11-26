# Farcaster Follower Cronjob & Cache System

A production-ready Go application that automatically fetches Farcaster followers every 5 minutes, caches them in Redis, and provides a fast API endpoint to check if a user is a follower.

## 🚀 Quick Start

```bash
# 1. Clone and setup
cp .env.example .env
# Edit .env with your Farcaster bearer token and target FIDs

# 2. Start with Docker
docker-compose up -d

# 3. Test the API
curl -X POST 'http://localhost:8080/api/v1/social-action' \
  -H 'Content-Type: application/json' \
  -d '{"social":"farcaster","action":"follow","iduser":"1406368"}'
```

## 📋 Features

- ✅ **Automatic Follower Fetching** - Runs every 5 minutes via cronjob
- ✅ **Redis Caching** - O(1) lookup time for follower checks
- ✅ **Fast API Checks** - Response time < 1ms (no API calls needed)
- ✅ **Pagination Support** - Handles cursor-based pagination automatically
- ✅ **Rate Limiting** - Built-in delays to avoid API throttling
- ✅ **Multiple Target FIDs** - Support for monitoring multiple accounts
- ✅ **Production Ready** - Error handling, logging, graceful shutdown
- ✅ **Docker Support** - Complete Docker Compose setup included
- ✅ **Comprehensive Docs** - Setup, testing, architecture guides included

## [object Object]
┌─────────────────────────────────────────┐
│         Cronjob (Every 5 min)           │
│  Fetches followers from Farcaster API   │
└────────────────┬────────────────────────┘
                 │
                 ▼
        ┌────────────────────┐
        │   Redis Cache      │
        │  (Followers Set)   │
        └────────────────────┘
                 │
                 ▼
        ┌────────────────────┐
        │   API Endpoint     │
        │  /social-action    │
        │  (< 1ms response)  │
        └────────────────────┘
```

## 📦 What's Included

### Core Components
- **Redis Cache Manager** - Connection, storage, and retrieval
- **Farcaster Follower Fetcher** - Pagination and caching logic
- **Cronjob Scheduler** - 5-minute interval execution
- **API Handler** - Fast follower verification endpoint

### Documentation
- **QUICKSTART.md** - 5-minute setup guide
- **SETUP.md** - Detailed configuration and installation
- **ARCHITECTURE.md** - System design with diagrams
- **TESTING.md** - Testing procedures and examples
- **IMPLEMENTATION_SUMMARY.md** - Complete implementation details

## 🔧 Configuration

### Required Environment Variables

```env
# Farcaster API
FARCASTER_BEARER_TOKEN=your_bearer_token_here

# Target FIDs to monitor (comma-separated)
TARGET_FID=1093215
TARGET_FIDS=1093215,1093216,1093217

# Redis Configuration
REDIS_ADDR=localhost:6379
REDIS_DB=0
REDIS_PASSWORD=
```

See `.env.example` for all available options.

## 📡 API Usage

### Check if User is a Follower

**Endpoint:** `POST /api/v1/social-action`

**Request:**
```bash
curl -X POST 'http://localhost:8080/api/v1/social-action' \
  -H 'Content-Type: application/json' \
  -d '{
    "social": "farcaster",
    "action": "follow",
    "iduser": "1406368"
  }'
```

**Response:**
```json
true
```

## 🐳 Docker Setup (Recommended)

```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop services
docker-compose down
```

## 💻 Local Setup

### Prerequisites
- Go 1.24.0+
- Redis 7.0+

### Installation

```bash
# Install dependencies
go mod download

# Start Redis
redis-server

# Run application
go run main.go
```

## 📊 How It Works

### Startup
1. Load environment variables
2. Connect to Redis
3. Initialize cronjob scheduler
4. Start HTTP server

### Cronjob (Every 5 Minutes)
1. Fetch followers from Farcaster API
2. Handle pagination with cursors
3. Store FIDs in Redis Set
4. Update last sync timestamp

### API Request
1. Receive POST request with user FID
2. Query Redis for follower
3. Return true/false instantly

## 🔍 Monitoring

### Check Cronjob Status

```bash
# View last sync time
redis-cli GET farcaster:sync:last:1093215

# View follower count
redis-cli SCARD farcaster:followers:1093215

# View sample followers
redis-cli SRANDMEMBER farcaster:followers:1093215 5
```

### View Application Logs

```bash
# Docker
docker-compose logs -f app

# Local
# Check console output
```

## 📈 Performance

| Operation | Time | Complexity |
|-----------|------|-----------|
| API Check | < 1ms | O(1) |
| Redis Lookup | < 1ms | O(1) |
| Follower Fetch | 1-5 min | O(n) |
| Cache Clear | < 100ms | O(n) |

## 🧪 Testing

### Quick Test

```bash
# Test API
curl -X POST 'http://localhost:8080/api/v1/social-action' \
  -H 'Content-Type: application/json' \
  -d '{"social":"farcaster","action":"follow","iduser":"1406368"}'

# Check Redis
redis-cli SISMEMBER farcaster:followers:1093215 1406368
```

### Load Testing

```bash
# Using Apache Bench
ab -n 100 -c 10 -p request.json -T application/json \
  'http://localhost:8080/api/v1/social-action'
```

See **TESTING.md** for comprehensive testing guide.

## [object Object]

### Redis Connection Error
```
Failed to connect to Redis: dial tcp localhost:6379: connect: connection refused
```
**Solution:** Ensure Redis is running on the specified address

### Bearer Token Error
```
API request failed with status 401
```
**Solution:** Verify `FARCASTER_BEARER_TOKEN` is correct

### Cronjob Not Running
```
Warning: TARGET_FIDS environment variable not set
```
**Solution:** Add `TARGET_FIDS` to `.env`

See **SETUP.md** for more troubleshooting tips.

## 📚 Documentation

- **[QUICKSTART.md](QUICKSTART.md)** - Get started in 5 minutes
- **[SETUP.md](SETUP.md)** - Detailed setup and configuration
- **[ARCHITECTURE.md](ARCHITECTURE.md)** - System design and diagrams
- **[TESTING.md](TESTING.md)** - Testing procedures and examples
- **[IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md)** - Implementation details

## 🔐 Security Considerations

1. **Bearer Token** - Keep `FARCASTER_BEARER_TOKEN` secret, use environment variables
2. **Redis Password** - Use `REDIS_PASSWORD` in production
3. **Network** - Restrict Redis access to application only
4. **Rate Limiting** - Built-in 500ms delay between API requests

## 📁 Project Structure

```
checkingsocial/
├── main.go                      # Entry point
├── go.mod                       # Dependencies
├── docker-compose.yml           # Docker services
├── .env.example                 # Environment template
├── pkg/
│   ├── cache/redis.go          # Redis client
│   └── cronjob/scheduler.go    # Cronjob scheduler
├── farcaster/
│   ├── client.go               # CheckFollow function
│   └── follower_fetcher.go     # Follower fetcher
├── internal/
│   ├── handler/social.go       # API handlers
│   ├── service/social_checker.go # Business logic
│   └── model/social.go         # Data models
└── docs/
    ├── QUICKSTART.md           # Quick start guide
    ├── SETUP.md                # Setup guide
    ├── ARCHITECTURE.md         # Architecture docs
    ├── TESTING.md              # Testing guide
    └── IMPLEMENTATION_SUMMARY.md # Implementation details
```

## 🚀 Deployment

### Docker Compose (Development)
```bash
docker-compose up -d
```

### Kubernetes (Production)
```bash
# Build image
docker build -t farcaster-cronjob:latest .

# Deploy to Kubernetes
kubectl apply -f k8s/deployment.yaml
```

### AWS/GCP/Azure
- Use managed Redis service
- Deploy app to container service
- Configure `REDIS_ADDR` to managed service endpoint

## 📝 API Response Examples

### Success - User is a Follower
```bash
$ curl -X POST 'http://localhost:8080/api/v1/social-action' \
  -H 'Content-Type: application/json' \
  -d '{"social":"farcaster","action":"follow","iduser":"1406368"}'

true
```

### Success - User is NOT a Follower
```bash
$ curl -X POST 'http://localhost:8080/api/v1/social-action' \
  -H 'Content-Type: application/json' \
  -d '{"social":"farcaster","action":"follow","iduser":"9999999999"}'

false
```

### Error - Invalid Request
```bash
$ curl -X POST 'http://localhost:8080/api/v1/social-action' \
  -H 'Content-Type: application/json' \
  -d '{"social":"farcaster"}'

{
  "error": "Key: 'SocialActionRequest.Action' Error:Field validation for 'Action' failed on the 'required' tag"
}
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## 📄 License

This project is provided as-is for use with Farcaster API.

## 📞 Support

For issues and questions:
1. Check **SETUP.md** for configuration issues
2. Check **TESTING.md** for testing procedures
3. Check **ARCHITECTURE.md** for system design questions
4. Review application logs for error messages

## 🎯 Next Steps

1. **Setup** - Follow [QUICKSTART.md](QUICKSTART.md)
2. **Configure** - Add your Farcaster bearer token to `.env`
3. **Deploy** - Use Docker Compose or local setup
4. **Test** - Verify API endpoint works
5. **Monitor** - Check logs and Redis cache

---

**Built with:** Go, Redis, Gin, Cron

**Last Updated:** November 2025

