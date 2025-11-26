# Architecture Overview

## System Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                     Application Startup                          │
└────────────────────────┬────────────────────────────────────────┘
                         │
         ┌───────────────┼───────────────┐
         │               │               │
         ▼               ▼               ▼
    ┌─────────┐    ┌──────────┐    ┌──────────────┐
    │ Load    │    │Initialize│    │ Initialize  │
    │ .env    │    │  Redis   │    │  Cronjob    │
    └────┬────┘    └────┬─────┘    └──────┬───────┘
         │              │                  │
         └──────────────┼──────────────────┘
                        │
                        ▼
            ┌───────────────────────┐
            │  Server Ready on      │
            │  Port 8080            │
            └───────────┬───────────┘
                        │
        ┌───────────────┼───────────────┐
        │               │               │
        ▼               ▼               ▼
   ┌─────────┐    ┌──────────┐    ┌──────────────┐
   │ Cronjob │    │ API      │    │ Graceful     │
   │ Runs    │    │ Endpoint │    │ Shutdown     │
   │ Every   │    │ Available│    │ Handler      │
   │ 5 min   │    │          │    │              │
   └────┬────┘    └────┬─────┘    └──────────────┘
        │              │
        ▼              ▼
   ┌─────────────────────────────────────────┐
   │  Farcaster Follower Fetcher             │
   │  - Fetch followers with pagination      │
   │  - Handle cursor-based navigation       │
   │  - Add rate limiting delays             │
   └────────────┬────────────────────────────┘
                │
                ▼
   ┌─────────────────────────────────────────┐
   │  Redis Cache                            │
   │  - Store FIDs in Set                    │
   │  - Track last sync time                 │
   │  - Fast O(1) lookups                    │
   └────────────┬────────────────────────────┘
                │
                ▼
   ┌─────────────────────────────────────────┐
   │  API Request: /api/v1/social-action     │
   │  - Query Redis for follower             │
   │  - Return true/false instantly          │
   │  - No API calls needed                  │
   └─────────────────────────────────────────┘
```

## Component Interaction

```
┌──────────────────────────────────────────────────────────────────┐
│                         main.go                                   │
│  - Initialize Redis connection                                   │
│  - Start cronjob scheduler                                       │
│  - Setup Gin router                                              │
│  - Handle graceful shutdown                                      │
└──────────────────────────────────────────────────────────────────┘
         │                          │                      │
         ▼                          ▼                      ▼
    ┌──────────────┐         ┌──────────────┐      ┌─────────────┐
    │ pkg/cache/   │         │ pkg/cronjob/ │      │ internal/   │
    │ redis.go     │         │ scheduler.go │      │ handler/    │
    │              │         │              │      │ social.go   │
    │ - Connect    │         │ - Schedule   │      │             │
    │ - Add FIDs   │         │ - Run every  │      │ - Handle    │
    │ - Check FID  │         │   5 minutes  │      │   requests  │
    │ - Clear      │         │ - Fetch all  │      │             │
    │   cache      │         │   target     │      │             │
    └──────┬───────┘         │   FIDs       │      └─────────────┘
           │                 └──────┬───────┘
           │                        │
           │                        ▼
           │              ┌──────────────────────┐
           │              │ farcaster/           │
           │              │ follower_fetcher.go  │
           │              │                      │
           │              │ - Fetch followers    │
           │              │ - Paginate with      │
           │              │   cursor             │
           │              │ - Rate limit         │
           │              │ - Cache in Redis     │
           │              └──────┬───────────────┘
           │                     │
           └─────────────────────┘
                    │
                    ▼
           ┌──────────────────┐
           │    Redis         │
           │                  │
           │ Set: followers   │
           │ String: sync_time│
           └──────────────────┘
```

## Data Flow: Cronjob Execution

```
Every 5 minutes:

1. Cronjob Triggered
   ↓
2. For each TARGET_FID in TARGET_FIDS:
   ├─ Clear old followers cache
   ├─ Initialize cursor = ""
   └─ Loop:
      ├─ Build API URL with cursor
      ├─ Make HTTP request to Farcaster
      ├─ Parse JSON response
      ├─ Extract FIDs from response
      ├─ Add FIDs to Redis Set
      ├─ Update cursor from response
      ├─ Wait 500ms (rate limiting)
      └─ If cursor exists, repeat; else break
   ├─ Update last sync time
   └─ Log completion
```

## Data Flow: API Request

```
User Request: POST /api/v1/social-action
{
  "social": "farcaster",
  "action": "follow",
  "iduser": "1406368"
}
   ↓
SocialHandler.SocialAction()
   ↓
SocialChecker.CheckSocialAction()
   ↓
farcaster.CheckFollow(userID)
   ├─ Parse userID to int64
   ├─ Get TARGET_FID from env
   └─ Query Redis: SISMEMBER farcaster:followers:{TARGET_FID} {userID}
      ├─ If member exists: return true
      └─ If not exists: return false
   ↓
Response: true or false
```

## Redis Data Structure

```
Key Format: farcaster:followers:{targetFID}
Type: Redis Set
TTL: None (persistent)

Example:
farcaster:followers:1093215
├─ 1406368
├─ 466033
├─ 1108383
├─ 1290249
├─ 1513624
└─ ... (more FIDs)

Key Format: farcaster:sync:last:{targetFID}
Type: String (Unix timestamp)
TTL: None (persistent)

Example:
farcaster:sync:last:1093215 = "1732605965"
```

## Sequence Diagram: First Run

```
Time  │ Application │ Redis │ Farcaster API │ Cronjob
──────┼─────────────┼───────┼───────────────┼─────────
  0   │ Start       │       │               │
  1   │ Init Redis  │ ✓     │               │
  2   │ Init Cron   │       │               │
  3   │ Server OK   │       │               │
  4   │             │       │               │ Trigger
  5   │             │       │ Fetch page 1  │
  6   │             │ Store │               │
  7   │             │ FIDs  │               │
  8   │             │       │ Fetch page 2  │
  9   │             │ Store │               │
 10   │             │ FIDs  │               │
 11   │             │       │ ... continue  │
 12   │             │ Done  │               │ Complete
 13   │ API Ready   │       │               │
```

## Sequence Diagram: API Check

```
Time  │ User Request │ Handler │ Redis │ Response
──────┼──────────────┼─────────┼───────┼──────────
  0   │ POST /api/   │         │       │
  1   │              │ Parse   │       │
  2   │              │ Check   │       │
  3   │              │         │ Query │
  4   │              │         │ Found │
  5   │              │ Return  │       │
  6   │              │         │       │ true
```

## Error Handling Flow

```
┌─────────────────────────────────────────┐
│         Error Scenarios                 │
└────────────────────┬────────────────────┘
                     │
        ┌────────────┼────────────┐
        │            │            │
        ▼            ▼            ▼
   ┌─────────┐  ┌─────────┐  ┌──────────┐
   │ Redis   │  │ API     │  │ Invalid  │
   │ Error   │  │ Error   │  │ Input    │
   └────┬────┘  └────┬────┘  └────┬─────┘
        │            │            │
        ▼            ▼            ▼
   Log Error    Log Error    Return Error
   Continue     Retry or     Response
   Cronjob      Skip         400 Bad Request
```

## Performance Characteristics

| Operation | Complexity | Time |
|-----------|-----------|------|
| Check if follower | O(1) | < 1ms |
| Fetch all followers | O(n) | 1-5 min |
| Store FID in cache | O(1) | < 1ms |
| Clear cache | O(n) | < 100ms |
| Get follower count | O(1) | < 1ms |

Where n = number of followers (typically 1000-10000)

## Scalability Considerations

1. **Multiple Target FIDs**: Cronjob handles multiple FIDs sequentially
2. **Redis Memory**: Each FID ~8 bytes, 10k followers = ~80KB
3. **API Rate Limiting**: 500ms delay between requests
4. **Concurrent Requests**: Gin handles multiple API requests
5. **Cronjob Timing**: 5-minute interval prevents API overload

