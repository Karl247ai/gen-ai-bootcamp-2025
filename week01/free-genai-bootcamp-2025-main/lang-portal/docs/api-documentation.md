# API Documentation

## API Standards

### Request/Response Format
All API endpoints follow these conventions:

1. Base URL: `/api/v1`
2. Content-Type: `application/json`
3. Character Encoding: `UTF-8`

### Standard Response Structure
```json
{
    "status": "success" | "error",
    "data": {},
    "meta": {
        "timestamp": "2024-03-14T12:00:00Z",
        "version": "1.0"
    },
    "pagination": {
        "current_page": 1,
        "total_pages": 10,
        "total_items": 100,
        "items_per_page": 10
    }
}
```

## Endpoints Reference

### Words Management

#### GET /api/v1/words
Retrieve a paginated list of words.

**Query Parameters:**
- `page`: Page number (default: 1)
- `limit`: Items per page (default: 100)
- `group_id`: Filter by group ID (optional)

**Response:**
```json
{
    "status": "success",
    "data": {
        "items": [
            {
                "id": 1,
                "japanese": "こんにちは",
                "romaji": "konnichiwa",
                "english": "hello",
                "stats": {
                    "correct_count": 5,
                    "wrong_count": 2
                }
            }
        ]
    },
    "pagination": {
        "current_page": 1,
        "total_pages": 5,
        "total_items": 100,
        "items_per_page": 20
    }
}
```

### Study Sessions

#### POST /api/v1/study-sessions
Create a new study session.

**Request Body:**
```json
{
    "group_id": 123,
    "study_activity_id": 456
}
```

**Response:**
```json
{
    "status": "success",
    "data": {
        "session_id": 789,
        "start_time": "2024-03-14T12:00:00Z",
        "words": [
            {
                "id": 1,
                "japanese": "こんにちは"
            }
        ]
    }
}
```

## Data Models

### Word
```go
type Word struct {
    ID        int       `json:"id"`
    Japanese  string    `json:"japanese"`
    Romaji    string    `json:"romaji"`
    English   string    `json:"english"`
    Parts     JSON      `json:"parts,omitempty"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```

### StudySession
```go
type StudySession struct {
    ID              int       `json:"id"`
    GroupID         int       `json:"group_id"`
    ActivityID      int       `json:"study_activity_id"`
    CreatedAt       time.Time `json:"created_at"`
    CompletedAt     time.Time `json:"completed_at,omitempty"`
    ReviewItemCount int       `json:"review_items_count"`
}
```

## Performance Requirements

### Latency Requirements
- 95th percentile response time < 200ms
- 99th percentile response time < 500ms
- Maximum response time < 1s

### Throughput Requirements
- Minimum 100 requests/second per instance
- Support for up to 1000 concurrent users
- Linear scaling with additional instances

### Error Rate Requirements
- Error rate < 0.1% under normal load
- Error rate < 1% under peak load
- Graceful degradation under extreme load

## Error Handling

### Error Response Format
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": {
      "field": "Additional context"
    }
  }
}
```

### Common Error Codes
- `INVALID_INPUT`: Request validation failed
- `DB_ERROR`: Database operation failed
- `NOT_FOUND`: Requested resource not found
- `UNAUTHORIZED`: Authentication required
- `FORBIDDEN`: Permission denied
- `TIMEOUT`: Operation timed out
- `INTERNAL_ERROR`: Unexpected server error

## Monitoring and Metrics

### Available Metrics
- Request count by endpoint
- Response time distributions
- Error rates by type
- Database operation latencies
- Connection pool status

### Health Check Endpoint
GET /health
```json
{
  "status": "healthy",
  "details": {
    "database": "connected",
    "cache": "available",
    "memory": "normal"
  },
  "timestamp": "2023-07-20T10:30:00Z"
}
```

## Rate Limiting

- Default: 1000 requests per minute per IP
- Authenticated: 5000 requests per minute per user
- Bulk operations: 100 requests per minute

## Monitoring Endpoints

### Health Check
```http
GET /health
```

Returns the health status of the service and its dependencies.

**Response**
```json
{
    "status": "healthy",
    "details": {
        "database": "connected",
        "cache": "available",
        "memory": "normal"
    },
    "timestamp": "2023-07-20T10:30:00Z"
}
```

### Metrics
```http
GET /metrics
```

Returns Prometheus-formatted metrics for the service.

**Response Format**: Plain text (Prometheus format)

### Debug Endpoints

These endpoints require admin authentication.

#### Database Stats
```http
GET /debug/db/stats
```

Returns database connection pool statistics.

**Response**
```json
{
    "max_open_connections": 50,
    "open_connections": 10,
    "in_use": 5,
    "idle": 5,
    "wait_count": 0,
    "wait_duration": "0s",
    "max_idle_closed": 0,
    "max_lifetime_closed": 0
}
```

#### Cache Stats
```http
GET /debug/cache/stats
```

Returns cache statistics.

**Response**
```json
{
    "items": 1000,
    "hits": 5000,
    "misses": 100,
    "evictions": 50
}
```

#### Goroutine Dump
```http
GET /debug/goroutines
```

Returns a stack trace of all running goroutines.

**Response Format**: Plain text

#### Reset Metrics
```http
POST /debug/metrics/reset
```

Resets all metrics counters to zero.

**Response**: 200 OK

## Error Responses

All monitoring endpoints may return these errors:

- `401 Unauthorized`: Authentication required (for debug endpoints)
- `403 Forbidden`: Insufficient permissions
- `500 Internal Server Error`: Server error
- `503 Service Unavailable`: Service is starting up or shutting down 