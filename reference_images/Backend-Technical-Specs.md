# Backend Server Technical Specs

## Business Goal:

A language learning school wants to build a prototype of
learning portal which will act as three things:
- Inventory of possible vocabulary that can be learned
- Act as a Learning record store (LRS), providing correct and
wrong score on practice vocabulary
- A unified launchpad to launch different learning apps

## Business Goals

Additional goals should include:
- Track user's learning progress over time
- Provide analytics on learning effectiveness
- Support multiple learning activities/games
- Enable flexible grouping of vocabulary words
- Support spaced repetition learning

## Technical Requirements

- The backend will be built using Go
- The database will be SQLite3
- The API will be built using Gin
- Mage is a task runner for Go.
- The API will always return JSON
- There will no authentication or authorization
- Everything be treated as a single user

Additional requirements:
- API versioning (e.g., /api/v1/...)
- Error handling standards and response formats
- Rate limiting considerations
- Logging requirements
- Performance requirements (response times)
- Data validation rules
- Backup and recovery procedures
- Development environment setup requirements

## Directory Structure

```
backend_go/
├── cmd/                    # Application entry points
│   └── server/            # Main server application
├── internal/              # Private application code
│   ├── api/              # API handlers
│   ├── middleware/       # HTTP middleware
│   ├── models/           # Database models
│   ├── repository/       # Database operations
│   └── service/          # Business logic
├── migrations/           # Database migrations
├── seeds/                # Seed data
├── config/              # Configuration files
├── pkg/                 # Public library code
├── scripts/             # Utility scripts
└── test/               # Test files
```

## Database Schema

Table: `words`
| Column    | Type    | Constraints       | Description                    |
|-----------|---------|------------------|--------------------------------|
| id        | INTEGER | PRIMARY KEY      | Unique identifier             |
| japanese  | TEXT    | NOT NULL         | Japanese word/phrase          |
| romaji    | TEXT    | NOT NULL         | Romanized Japanese            |
| english   | TEXT    | NOT NULL         | English translation           |
| parts     | JSON    |                  | Word components/metadata      |
| created_at| DATETIME| NOT NULL         | Creation timestamp            |
| updated_at| DATETIME| NOT NULL         | Last update timestamp         |

Table: `words_groups`
| Column    | Type    | Constraints                | Description              |
|-----------|---------|---------------------------|--------------------------|
| id        | INTEGER | PRIMARY KEY              | Unique identifier        |
| word_id   | INTEGER | FOREIGN KEY, NOT NULL    | Reference to words      |
| group_id  | INTEGER | FOREIGN KEY, NOT NULL    | Reference to groups     |
| created_at| DATETIME| NOT NULL                 | Creation timestamp      |

Table: `groups`
| Column     | Type    | Constraints           | Description              |
|------------|---------|----------------------|--------------------------|
| id         | INTEGER | PRIMARY KEY          | Unique identifier        |
| name       | TEXT    | NOT NULL, UNIQUE     | Group name              |
| created_at | DATETIME| NOT NULL             | Creation timestamp      |
| updated_at | DATETIME| NOT NULL             | Last update timestamp   |

We have the following tables:
- words - stored vocabulary words
- id integer
- japasese string
- romaji string
- english string
- parts json
- words_groups - join table for words and groups many-to-many
- id integer
- word_id integer
- group_id integer
- groups - thematic groups of words
- id integer
- name string
- study_sessions - records of study sessions grouping word_review_items
- id integer
- group_id integer
- created_at datetime
- study_activity_id integer
- study_activities - a specific study activity, linking a study session to group
- id integer
- study_session_id integer
- group_id integer
- created_at datetime
- word_review_items - a record of word practice, determining if the word was correct or not
- word_id integer
- study_session_id integer
- correct boolean
- created_at datetime

## API Endpoints

### GET /api/v1/dashboard/last_study_session

**Description:** Returns information about the most recent study session.

**Query Parameters:**
None

**Success Response (200 OK):**
```json
{
    "id": 123,
    "group_id": 456,
    "created_at": "2025-02-08T17:20:23-05:00",
    "study_activity_id": 789,
    "group_name": "Basic Greetings"
}
```

**Error Responses:**
- 404 Not Found: No study sessions exist
- 500 Internal Server Error: Server error

### GET /api/dashboard/study_progress
Returns study progress statistics.
Please note that the frontend will determine progress bar
basedon total words studied and total available words.

#### JSON Response

```json

"total_words_studied": 3,
"total_available_words": 124,

### GET /api/dashboard/quick-stats

Returns quick overview statistics.

#### JSON Response
json

"success_rate": 80.0,
"total_study_sessions": 4,
"total_active_groups": 3,
"study_streak_days": 4
}
```

###GET /api/study_activities/:id

#### JSON Response
```json
{
"id": 1,
"name": "Vocabulary Quiz",
"thumbnail_url": "https://example.com/thumbnail.jpg",
"description": "Practice your vocabulary with flashcards"

}

### GET /api/study_activities/:id/study_sessions

- pagination with 100 items per page

```json

"items": [
{

"id": 123,
"activity_name": "Vocabulary Quiz",
"group_name": "Basic Greetings",
"start_time": "2025-02-08T17:20:23-05:00",
"end_time": "2025-02-08T17:30:23-05:00",
"review_items_count": 20
}
"pagination": {
"current_page": 1,
"total_pages": 5,
"total_items": 100,
"items_per_page": 20
}
}
```

### POST /api/study_activities

#### Request Params
- group_id integer
- study_activity_id integer

#### JSON Response

"id": 124,
"group_id": 123

1

}

### GET /api/words

- pagination with 100 items per page

#### JSON Response
```json
{

"items": [
{
"id": 123,
"activity_name": "Vocabulary Quiz",
"group_name": "Basic Greetings",
"start_time": "2025-02-08T17:20:23-05:00",
"end_time": "2025-02-08T17:30:23-05:00",
"review_items_count": 20
}
],
"pagination": {
"current_page": 1,
"total_pages": 5,
"total_items": 100,
"items_per_page": 20

}
```
### POST /api/study_activities



### POST /api/study_activities

#### Request Params
- group_id integer
- study_activity_id integer

#### JSON Response
{
    "id": 124,
"group_id": 123
}

### GET /api/words

- pagination with 100 items per page

#### JSON Response
```json
"id": 124,
"group_id": 123
}

### GET /api/words

- pagination with 100 items per page

#### JSON Response
```json
{
"items": [
{
"japanese": "2/(c5(#",
"romaji": "konnichiwa",
"english": "hello",
"correct_count": 5,
"wrong_count": 2
}
],
"pagination": {
"current_page": 1,
"total_pages": 5,
"total_items": 500,
"items_per_page": 100
}
}
```

### GET /api/words/:id
#### JSON Response
json

"japanese": "/(c5(a",
"romaji": "konnichiwa",
"english": "hello",
"stats": {
"correct_count": 5,
"wrong_count": 2

"groups": [
{
"id": 1,
"name": "Basic Greetings"
}
]
}
```

### GET /api/groups
- pagination with 100 items per page
#### JSON Response
```json

"items": [

"id": 1,
"name": "Basic Greetings",
"word_count": 20

"items": [

"id": 1,
"name": "Basic Greetings",
"word_count": 20
}
],
"pagination": {
"current_page": 1,
"total_pages": 1,
"total_items": 10,
"items_per_page": 100
}
}
```

### GET /api/groups/:id
#### JSON Response
json
{
"id": 1,
"name": "Basic Greetings",
"stats": {

### GET /api/groups/:id
#### JSON Response
json

"id": 1,
"name": "Basic Greetings",
"stats": {
"total_word_count": 20
}
}
```

### GET /api/groups/:id/words
#### JSON Response
```json

"items": [

"japanese":"こんにちは",
"romaji": "konnichiwa",
"english": "hello",
"correct_count": 5,
"wrong_count": 2
}
],
"pagination": {
"current_page": 1,
"total_pages": 1,
"total_items": 20,
"items_per_page": 100
}
}
```

### GET /api/groups/:id/study_sessions
#### JSON Response
```json

"items": [

"id": 123,
"activity_name": "Vocabulary Quiz",
"group_name": "Basic Greetings",
"start_time": "2025-02-08T17:20:23-05:00",
"end_time": "2025-02-08T17:30:23-05:00",
"review_items_count": 20
}
],
"pagination": {
"current_page": 1,
"total_pages": 1,
"total_items": 5,
"items_per_page": 100
}
}
```

### GET /api/study_sessions
- pagination with 100 items per page
#### JSON Response
```json
{
"items": [

"id": 123,
"activity_name": "Vocabulary Quiz",
"group_name": "Basic Greetings",
"start_time": "2025-02-08T17:20:23-05:00",
"end_time": "2025-02-08T17:30:23-05:00",
"review_items_count": 20

"items": [

"id": 123,
"activity_name": "Vocabulary Quiz",
"group_name": "Basic Greetings",
"start_time": "2025-02-08T17:20:23-05:00",
"end_time": "2025-02-08T17:30:23-05:00",


],
"pagination": {
"current_page": 1,
"total_pages": 5,
"total_items": 100,
"items_per_page": 100
}
}
```

### GET /api/study_sessions/:id
#### JSON Response
```json

"id": 123,
"activity_name": "Vocabulary Quiz",
"group_name": "Basic Greetings",
"start_time": "2025-02-08T17:20:23-05:00",
"end time": "2025-07-A8T17:30:23-05:00"

### GET /api/study_sessions/:id
#### JSON Response
`json
{
"id": 123,
"activity_name": "Vocabulary Quiz",
"group_name": "Basic Greetings",
"start_time": "2025-02-08T17:20:23-05:00",
"end_time": "2025-02-08T17:30:23-05:00",
"review_items_count": 20
}
```
### GET /api/study_sessions/:id/words
- pagination with 100 items per page
#### JSON Response
```json
{
"items": [
{
"japanese":"こんにちは"」
"romaji": "konnichiwa",
"english": "hello",
"correct_count": 5,
"wrong_count": 2
}
],
"pagination": {
"current_page": 1,
"total_pages": 1,
"total_items": 20,
"items_per_page": 100
}
}
```

### POST /api/reset_history
#### JSON Response
```json
{
"success": true,
"message": "Study history has been reset"
"current_page": 1,
"total_pages": 1,
"total_items": 20,
"items_per_page": 100
}
```

### POST /api/full_reset
#### JSON Response
```json
{
"success": true,
"message": "System has been fully reset"

### POST /ani/study sessions/:id/words/:word id/review
#### JSON Response

### POST /api/study_sessions/:id/words/:word_id/review
#### Request Params
- id (study_session_id) integer
- word_id integer
- correct boolean

#### Request Payload
```json
{
"correct": true
}
```

#### JSON Response
```json
{
"success": true,
"word_id": 1,
"study_session_id": 123,
"correct": true,
"created_at": "2025-02-08T17:33:07-05:00"
}
```

## Task Runner Tasks

Mage is a task runner for Go.
Lets list out possible tasks we need for our lang portal.

### Initialize Database
This task will initialize the sqlite database called `words.db

### Migrate Database
This task will run a series of migrations sql files on the
database

Migrations live in the `migrations' folder.
The migration files will be run in order of their file name.
The file names should looks like this:

```sql
0001_init.sql
0002_create_words_table.sql
```

### Seed Data
This task will import json files and transform them into target
data for our database.

All seed files live in the `seeds' folder.

In our task we should have DSL to specific each seed file and
its expected group word name.

```json
[
{
"kanji": "#43",
"romaji": "harau",
"english": "to pay",
},
```
]
```

## New Sections to Add

### Error Handling
Standard error response format:
```json
{
    "error": {
        "code": "RESOURCE_NOT_FOUND",
        "message": "The requested resource was not found",
        "details": {}
    }
}
```

### Data Validation Rules
- Japanese text must be valid UTF-8
- Romaji must only contain ASCII characters
- Word lengths must be between 1 and 100 characters
- etc.

### Performance Requirements
- API response time: < 200ms for 95% of requests
- Database query time: < 100ms for 95% of queries
- Maximum concurrent users: 1000

### Security Considerations
- Input sanitization
- SQL injection prevention
- XSS prevention
- Rate limiting rules

### Monitoring and Logging
- Request logging format
- Error logging requirements
- Performance metrics to track
- Health check endpoints

## API Standards

### Response Format
All API responses should follow this structure:
```json
{
    "status": "success" | "error",
    "data": {
        // Response data here
    },
    "meta": {
        "timestamp": "2024-03-14T12:00:00Z",
        "version": "1.0"
    },
    "pagination": {
        // Pagination info if applicable
    }
}
```

### Error Response Format
```json
{
    "status": "error",
    "error": {
        "code": "ERROR_CODE",
        "message": "Human readable message",
        "details": {
            // Additional error details
        }
    },
    "meta": {
        "timestamp": "2024-03-14T12:00:00Z",
        "version": "1.0"
    }
}
```

### HTTP Status Codes
- 200: Successful request
- 201: Resource created
- 400: Bad request
- 404: Resource not found
- 422: Validation error
- 429: Too many requests
- 500: Server error

## Configuration Management

### Environment Variables
```
DATABASE_PATH=./words.db
API_PORT=8080
LOG_LEVEL=info
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_DURATION=60
MAX_PAGE_SIZE=100
```

### Configuration Files
Location: `config/`
```yaml
# config/default.yaml
server:
  port: 8080
  timeout: 30s
  
database:
  path: ./words.db
  max_connections: 10
  
logging:
  level: info
  format: json
```

## Development Setup

### Prerequisites
- Go 1.21 or higher
- SQLite 3.x
- Make (optional)

### Local Development
```bash
# Clone repository
git clone <repository-url>

# Install dependencies
go mod download

# Setup database
mage initdb
mage migrate
mage seed

# Run development server
go run cmd/server/main.go

# Run tests
go test ./...
```

### Development Tools
- golangci-lint for code linting
- mockgen for generating mocks
- swag for API documentation

## Testing Strategy

### Test Types
1. Unit Tests
   - Coverage target: 80%
   - Location: Next to implementation files
   - Naming: `*_test.go`

2. Integration Tests
   - Coverage target: 70%
   - Location: `test/integration`
   - Requires test database

3. API Tests
   - Coverage target: 90%
   - Location: `test/api`
   - Tests all API endpoints

### Test Data
- Test fixtures in `test/fixtures`
- Separate test database for integration tests
- Mock external dependencies

## Deployment

### Build Process
```bash
# Build binary
go build -o bin/server cmd/server/main.go

# Docker build
docker build -t lang-portal-api .
```

### Deployment Requirements
- Linux x64 environment
- 1GB RAM minimum
- 1GB disk space
- SQLite file permissions
- Network access for API

### Health Checks
Endpoint: `/health`
```json
{
    "status": "healthy",
    "version": "1.0.0",
    "timestamp": "2024-03-14T12:00:00Z",
    "checks": {
        "database": "up",
        "api": "up"
    }
}
```

## Documentation

### API Documentation
- OpenAPI/Swagger specification required
- Generated using swag
- Available at `/swagger/index.html`
- Must include:
  - Request/response examples
  - Schema definitions
  - Error responses
  - Authentication details

### Code Documentation
- All packages must have package documentation
- Public functions require documentation
- Complex algorithms need detailed explanations
- Generated documentation hosted via godoc

### Database Documentation
- Entity Relationship Diagram (ERD)
- Indexes and constraints
- Migration history
- Backup/restore procedures

## Data Migration

### Migration Files
```sql
-- Example migration file: 0001_init.sql
CREATE TABLE words (
    id INTEGER PRIMARY KEY,
    japanese TEXT NOT NULL,
    romaji TEXT NOT NULL,
    english TEXT NOT NULL,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
);

-- Add indexes
CREATE INDEX idx_words_romaji ON words(romaji);
CREATE INDEX idx_words_english ON words(english);
```

### Migration Process
1. Backup existing database
2. Run new migrations
3. Verify data integrity
4. Update application version
5. Roll back procedure if needed

### Data Seeding
- Seed files in JSON format
- Validation before insertion
- Environment-specific seeds
- Version control for seed data

## Caching

### Cache Levels
1. In-Memory Cache
   - Frequently accessed data
   - Cache size: 100MB max
   - TTL: 5 minutes

2. Response Cache
   - GET requests only
   - Cache-Control headers
   - Vary by query parameters

### Cache Keys
- Format: `{entity}:{id}:{field}`
- Example: `word:123:translations`

### Cache Invalidation
- On write operations
- Scheduled cleanup
- Manual purge endpoint

## Error Handling Details

### Error Categories
1. Validation Errors (400)
```json
{
    "status": "error",
    "error": {
        "code": "VALIDATION_ERROR",
        "message": "Invalid input data",
        "details": {
            "japanese": "required field",
            "romaji": "invalid characters"
        }
    }
}
```

2. Business Logic Errors (422)
```json
{
    "status": "error",
    "error": {
        "code": "BUSINESS_RULE_VIOLATION",
        "message": "Cannot delete word in active study session",
        "details": {
            "session_id": "123",
            "word_id": "456"
        }
    }
}
```

### Error Logging
- Log all 5xx errors
- Include stack traces
- Correlation IDs
- Client information

## Performance Optimization

### Database Optimization
1. Indexes
   - Primary keys
   - Foreign keys
   - Search fields
   - Sort fields

2. Query Optimization
   - Use EXPLAIN
   - Limit result sets
   - Proper JOIN types
   - Avoid N+1 queries

### API Optimization
1. Response Compression
   - gzip encoding
   - Minimum size: 1KB

2. Connection Pooling
   - Max connections: 100
   - Idle timeout: 60s

3. Resource Limits
   - Request body: 1MB
   - File uploads: 5MB
   - Request timeout: 30s

## Monitoring and Alerting

### Metrics
1. System Metrics
   - CPU usage
   - Memory usage
   - Disk space
   - Network I/O

2. Application Metrics
   - Request rate
   - Error rate
   - Response times
   - Active sessions

### Alert Thresholds
- CPU > 80% for 5 minutes
- Memory > 90%
- Error rate > 5%
- Response time > 1s

### Logging Format
```json
{
    "timestamp": "2024-03-14T12:00:00Z",
    "level": "info",
    "service": "api",
    "trace_id": "abc123",
    "message": "Request processed",
    "details": {
        "method": "GET",
        "path": "/api/v1/words",
        "duration_ms": 45,
        "status": 200
    }
}
```

## API Rate Limiting

### Rate Limit Rules
1. Anonymous Access
   - 100 requests per minute per IP
   - Burst: 20 requests

2. Endpoints Specific Limits
   - POST operations: 30 requests per minute
   - GET operations: 60 requests per minute
   - Study session operations: 50 requests per minute

### Rate Limit Response
```json
{
    "status": "error",
    "error": {
        "code": "RATE_LIMIT_EXCEEDED",
        "message": "Too many requests",
        "details": {
            "retry_after": 30,
            "limit": 100,
            "remaining": 0,
            "reset": "2024-03-14T12:01:00Z"
        }
    }
}
```

## Data Backup and Recovery

### Backup Strategy
1. Database Backups
   - Full backup daily
   - Incremental backup every 6 hours
   - Retention: 30 days

2. Backup Storage
   - Local disk backup
   - Compressed format
   - Encrypted at rest

### Recovery Procedures
1. Point-in-time Recovery
   - Using SQLite WAL
   - Transaction logs
   - Recovery time objective: 1 hour

2. Disaster Recovery
   - Recovery point objective: 6 hours
   - Documented recovery steps
   - Regular recovery testing

## Security Implementation

### Input Validation
1. Request Validation
   - Sanitize all inputs
   - Validate content types
   - Max length restrictions
   - Character set restrictions

2. SQL Injection Prevention
   - Prepared statements
   - Parameter binding
   - Input escaping
   - Query builder usage

### Security Headers
```yaml
headers:
  X-Content-Type-Options: nosniff
  X-Frame-Options: DENY
  X-XSS-Protection: 1; mode=block
  Content-Security-Policy: default-src 'self'
  Strict-Transport-Security: max-age=31536000
```

## Development Workflow

### Git Workflow
1. Branch Strategy
   - main: production ready code
   - develop: integration branch
   - feature/*: new features
   - bugfix/*: bug fixes
   - release/*: release preparation

2. Commit Guidelines
   - Conventional commits
   - Signed commits required
   - Linear history preferred

### Code Review Process
1. Pull Request Requirements
   - Tests included
   - Documentation updated
   - Linting passed
   - No security vulnerabilities
   - Coverage maintained

2. Review Checklist
   - Code style compliance
   - Error handling
   - Performance impact
   - Security considerations
   - API compatibility

## API Versioning Strategy

### Version Control
1. URL Versioning
   - Format: `/api/v{major}/`
   - Example: `/api/v1/words`

2. Version Compatibility
   - Breaking changes: new major version
   - Additions: minor version update
   - Bug fixes: patch version update

### Version Lifecycle
1. Active Versions
   - Latest version (v1)
   - One previous version
   - 6 months deprecation notice

2. Version Documentation
   - Changelog maintained
   - Migration guides
   - Breaking changes listed

## Data Validation and Sanitization

### Japanese Text Validation
1. Character Sets
   - Hiragana (ひらがな)
   - Katakana (カタカナ)
   - Kanji (漢字)
   - Basic Latin (ASCII)

2. Length Constraints
   - Minimum: 1 character
   - Maximum: 100 characters
   - Whitespace rules

### Romaji Validation
1. Character Rules
   - ASCII only
   - Hepburn romanization system
   - Allowed special characters: - '
   - Case sensitivity rules

### English Translation
1. Text Rules
   - ASCII only
   - Max length: 200 characters
   - Multiple translations separator: ";"

## Business Logic Implementation

### Study Session Rules
1. Session Creation
   - Maximum words per session: 50
   - Minimum words per session: 5
   - Session timeout: 30 minutes
   - Auto-save interval: 1 minute

2. Progress Tracking
   - Correct/incorrect counting
   - Time spent per word
   - Session completion criteria
   - Progress persistence

### Word Group Management
1. Group Rules
   - Maximum words per group: 500
   - Minimum words per group: 1
   - Maximum groups: 100
   - Nested groups: Not supported

2. Word Assignment
   - Words can belong to multiple groups
   - Default group assignment
   - Group hierarchy rules

## Performance Benchmarks

### Response Time Targets
1. API Endpoints
   | Endpoint Type | Target (95th percentile) | Maximum |
   |--------------|-------------------------|---------|
   | GET requests | 100ms                  | 500ms   |
   | POST requests| 200ms                  | 1000ms  |
   | Bulk operations| 500ms                | 2000ms  |

2. Database Operations
   | Operation Type | Target (95th percentile) | Maximum |
   |---------------|-------------------------|---------|
   | Simple queries| 50ms                   | 200ms   |
   | Complex joins | 100ms                  | 500ms   |
   | Transactions  | 150ms                  | 750ms   |

## Dependency Management

### External Dependencies
1. Core Dependencies
   ```go
   require (
       github.com/gin-gonic/gin v1.9.1
       github.com/mattn/go-sqlite3 v1.14.17
       github.com/magefile/mage v1.15.0
       github.com/go-playground/validator/v10 v10.14.0
   )
   ```

2. Development Dependencies
   ```go
   require (
       github.com/stretchr/testify v1.8.4
       github.com/golang/mock v1.6.0
       github.com/swaggo/swag v1.16.2
   )
   ```

### Dependency Update Policy
1. Update Schedule
   - Security updates: Immediate
   - Minor versions: Monthly
   - Major versions: Quarterly
   - Breaking changes: Planned migrations

## Error Codes and Messages

### Standard Error Codes
```go
const (
    ErrInvalidInput        = "INVALID_INPUT"
    ErrResourceNotFound    = "RESOURCE_NOT_FOUND"
    ErrDatabaseOperation   = "DATABASE_ERROR"
    ErrInternalServer      = "INTERNAL_SERVER_ERROR"
    ErrRateLimitExceeded   = "RATE_LIMIT_EXCEEDED"
    ErrValidationFailed    = "VALIDATION_FAILED"
    ErrSessionExpired      = "SESSION_EXPIRED"
    ErrDuplicateEntry      = "DUPLICATE_ENTRY"
)
```

### Localized Error Messages
```json
{
    "error_messages": {
        "INVALID_INPUT": {
            "en": "Invalid input provided",
            "ja": "無効な入力です"
        },
        "RESOURCE_NOT_FOUND": {
            "en": "Requested resource was not found",
            "ja": "リソースが見つかりません"
        }
    }
}
```

## Application Metrics and Analytics

### Learning Analytics
1. User Progress Metrics
   ```json
   {
       "daily_stats": {
           "words_studied": 50,
           "success_rate": 0.85,
           "study_time_minutes": 45,
           "completion_rate": 0.90
       },
       "learning_curve": {
           "day_1": 0.65,
           "day_7": 0.75,
           "day_30": 0.85
       }
   }
   ```

2. Word Difficulty Analysis
   - Success rate per word
   - Average time spent per word
   - Common mistake patterns
   - Spaced repetition intervals

### System Analytics
1. Performance Metrics
   ```go
   type SystemMetrics struct {
       RequestLatency    []float64  // in milliseconds
       DatabaseLatency   []float64  // in milliseconds
       MemoryUsage      float64    // in MB
       ActiveSessions    int
       ErrorRate        float64    // percentage
   }
   ```

## Data Export and Import

### Export Formats
1. Study History
   ```json
   {
       "format_version": "1.0",
       "export_date": "2024-03-14T12:00:00Z",
       "study_sessions": [
           {
               "id": "session_123",
               "date": "2024-03-14T10:00:00Z",
               "words": [
                   {
                       "japanese": "こんにちは",
                       "result": "correct",
                       "time_spent": 2.5
                   }
               ]
           }
       ]
   }
   ```

2. Word Lists
   ```csv
   japanese,romaji,english,tags
   こんにちは,konnichiwa,hello,"greeting,basic"
   ありがとう,arigatou,thank you,"courtesy,basic"
   ```

## Database Maintenance

### Optimization Tasks
1. Regular Maintenance
   ```sql
   -- Rebuild indexes
   REINDEX words_romaji_idx;
   
   -- Analyze tables
   ANALYZE words;
   ANALYZE study_sessions;
   
   -- Vacuum database
   VACUUM;
   ```

2. Data Cleanup
   ```sql
   -- Remove old sessions
   DELETE FROM study_sessions 
   WHERE created_at < datetime('now', '-90 days');
   
   -- Clean orphaned records
   DELETE FROM word_review_items 
   WHERE study_session_id NOT IN (
       SELECT id FROM study_sessions
   );
   ```

## API Rate Limiting Implementation

### Rate Limiter Configuration
```go
type RateLimiterConfig struct {
    WindowSize      time.Duration
    MaxRequests     int
    BurstSize      int
    Strategy       string  // "token-bucket" or "sliding-window"
}

var EndpointLimits = map[string]RateLimiterConfig{
    "/api/v1/words": {
        WindowSize:  time.Minute,
        MaxRequests: 60,
        BurstSize:   10,
        Strategy:    "token-bucket",
    },
    "/api/v1/study-sessions": {
        WindowSize:  time.Minute,
        MaxRequests: 30,
        BurstSize:   5,
        Strategy:    "sliding-window",
    },
}
```

## Service Health Management

### Health Check Implementation
```go
type HealthCheck struct {
    Component   string    `json:"component"`
    Status      string    `json:"status"`
    LastChecked time.Time `json:"last_checked"`
    Details     struct {
        Latency    int64  `json:"latency_ms"`
        ErrorCount int    `json:"error_count"`
    } `json:"details"`
}

type HealthResponse struct {
    Status      string        `json:"status"`
    Version     string        `json:"version"`
    Environment string        `json:"environment"`
    Checks      []HealthCheck `json:"checks"`
}
```

### Circuit Breaker Configuration
```go
type CircuitBreakerConfig struct {
    MaxFailures      int           `json:"max_failures"`
    ResetTimeout     time.Duration `json:"reset_timeout"`
    HalfOpenTimeout  time.Duration `json:"half_open_timeout"`
    FailureThreshold float64       `json:"failure_threshold"`
}

var ServiceBreakers = map[string]CircuitBreakerConfig{
    "database": {
        MaxFailures:      5,
        ResetTimeout:     time.Second * 30,
        HalfOpenTimeout:  time.Second * 5,
        FailureThreshold: 0.5,
    },
}
```

## Data Retention and Privacy

### Data Lifecycle
1. Active Data
   - Study sessions: 90 days
   - Word history: 1 year
   - User progress: Indefinite
   - System logs: 30 days

2. Data Archival
   ```go
   type ArchivalConfig struct {
       DataType     string        `json:"data_type"`
       RetentionPeriod time.Duration `json:"retention_period"`
       ArchiveFormat   string        `json:"archive_format"`
       Compression     bool          `json:"compression"`
   }
   
   var RetentionPolicies = map[string]ArchivalConfig{
       "study_sessions": {
           DataType:        "activity",
           RetentionPeriod: time.Hour * 24 * 90,
           ArchiveFormat:   "json",
           Compression:     true,
       },
   }
   ```

## Internationalization (i18n)

### Language Support
1. Interface Languages
   ```go
   var SupportedLanguages = []struct {
       Code string
       Name string
       RTL  bool
   }{
       {"en", "English", false},
       {"ja", "日本語", false},
       {"zh", "中文", false},
   }
   ```

2. Translation Format
   ```json
   {
       "messages": {
           "study.session.start": {
               "en": "Study session started",
               "ja": "学習セッションを開始しました",
               "zh": "学习会话已开始"
           },
           "study.session.complete": {
               "en": "Session completed! You learned %d words",
               "ja": "%d個の単語を学習しました！",
               "zh": "会话完成！您学习了%d个单词"
           }
       }
   }
   ```

## Study Algorithm Implementation

### Spaced Repetition System
```go
type ReviewSchedule struct {
    InitialInterval  time.Duration `json:"initial_interval"`
    IntervalModifier float64      `json:"interval_modifier"`
    EaseFactors     []float64     `json:"ease_factors"`
}

type WordReviewData struct {
    WordID          int       `json:"word_id"`
    LastReviewed    time.Time `json:"last_reviewed"`
    NextReview      time.Time `json:"next_review"`
    CurrentInterval int       `json:"current_interval"`
    EaseFactor      float64   `json:"ease_factor"`
    ConsecutiveCorrect int    `json:"consecutive_correct"`
}
```

### Learning Progress Algorithm
```go
type ProgressCalculator struct {
    WeightFactors struct {
        RecencyWeight     float64 `json:"recency_weight"`
        CorrectnessWeight float64 `json:"correctness_weight"`
        SpeedWeight      float64 `json:"speed_weight"`
    }
    ThresholdLevels struct {
        Beginner     float64 `json:"beginner"`
        Intermediate float64 `json:"intermediate"`
        Advanced     float64 `json:"advanced"`
        Mastered     float64 `json:"mastered"`
    }
}
```

## Performance Testing Framework

### Load Test Scenarios
```yaml
scenarios:
  - name: "Basic Usage"
    duration: "5m"
    users: 50
    ramp_up: "30s"
    endpoints:
      - path: "/api/v1/words"
        method: "GET"
        weight: 0.4
      - path: "/api/v1/study-sessions"
        method: "POST"
        weight: 0.3

  - name: "Heavy Load"
    duration: "10m"
    users: 200
    ramp_up: "1m"
    endpoints:
      - path: "/api/v1/words/search"
        method: "GET"
        weight: 0.5
```

### Performance Metrics Collection
```go
type PerformanceMetrics struct {
    Timestamp   time.Time `json:"timestamp"`
    Environment string    `json:"environment"`
    Metrics     struct {
        ResponseTimes []struct {
            Endpoint    string  `json:"endpoint"`
            P50ms      float64 `json:"p50_ms"`
            P95ms      float64 `json:"p95_ms"`
            P99ms      float64 `json:"p99_ms"`
        } `json:"response_times"`
        ResourceUsage struct {
            CPUPercent     float64 `json:"cpu_percent"`
            MemoryUsageMB  float64 `json:"memory_usage_mb"`
            DiskIOOps      int64   `json:"disk_io_ops"`
            NetworkIOBytes int64   `json:"network_io_bytes"`
        } `json:"resource_usage"`
    } `json:"metrics"`
}
```

## Database Indexing Strategy

### Primary Indexes
```sql
-- Words table indexes
CREATE INDEX idx_words_created_at ON words(created_at);
CREATE INDEX idx_words_updated_at ON words(updated_at);
CREATE INDEX idx_words_search ON words(japanese, romaji, english);

-- Study sessions indexes
CREATE INDEX idx_study_sessions_user ON study_sessions(user_id, created_at);
CREATE INDEX idx_study_sessions_group ON study_sessions(group_id, created_at);
```

### Index Maintenance
1. Index Statistics
   - Weekly analysis
   - Usage monitoring
   - Size tracking
   - Performance impact

2. Index Optimization Rules
   - Remove unused indexes
   - Combine overlapping indexes
   - Monitor index fragmentation
   - Regular ANALYZE runs

## API Request/Response Lifecycle

### Request Processing
1. Request Flow
   ```mermaid
   graph TD
       A[Client Request] --> B[Rate Limiter]
       B --> C[Input Validation]
       C --> D[Authentication]
       D --> E[Business Logic]
       E --> F[Database Operation]
       F --> G[Response Formatting]
       G --> H[Client Response]
   ```

2. Middleware Chain
   ```go
   type MiddlewareChain struct {
       Order []string `json:"order"`
       Middlewares map[string]struct {
           Enabled bool          `json:"enabled"`
           Config  interface{}   `json:"config"`
       }
   }
   
   var DefaultChain = MiddlewareChain{
       Order: []string{
           "recover",
           "logger",
           "cors",
           "ratelimit",
           "metrics",
       },
   }
   ```

## Study Session Management

### Session State Machine
```go
type SessionState string

const (
    SessionStateCreated   SessionState = "created"
    SessionStateActive    SessionState = "active"
    SessionStatePaused    SessionState = "paused"
    SessionStateCompleted SessionState = "completed"
    SessionStateExpired   SessionState = "expired"
)

type SessionTransition struct {
    FromState SessionState
    ToState   SessionState
    Validator func(session *StudySession) bool
    Action    func(session *StudySession) error
}
```

### Progress Calculation
```go
type ProgressMetrics struct {
    WordsLearned      int     `json:"words_learned"`
    WordsMastered     int     `json:"words_mastered"`
    AccuracyRate      float64 `json:"accuracy_rate"`
    CompletionRate    float64 `json:"completion_rate"`
    TimeSpentMinutes  float64 `json:"time_spent_minutes"`
    StreakDays        int     `json:"streak_days"`
    Level             string  `json:"level"`
    NextLevelProgress float64 `json:"next_level_progress"`
}
```

## Word Difficulty Algorithm

### Difficulty Calculation
```go
type WordDifficulty struct {
    Factors struct {
        CharacterCount     float64 `json:"character_count"`
        UniqueKanji       float64 `json:"unique_kanji"`
        HistoricalErrors  float64 `json:"historical_errors"`
        AverageTimeSpent  float64 `json:"average_time_spent"`
        UsageFrequency    float64 `json:"usage_frequency"`
    }
    
    Weights struct {
        CharacterWeight    float64 `json:"character_weight"`
        KanjiWeight       float64 `json:"kanji_weight"`
        ErrorWeight       float64 `json:"error_weight"`
        TimeWeight        float64 `json:"time_weight"`
        FrequencyWeight   float64 `json:"frequency_weight"`
    }
}
```

### Dynamic Difficulty Adjustment
```go
type DifficultyAdjustment struct {
    InitialLevel     int     `json:"initial_level"`
    CurrentLevel     int     `json:"current_level"`
    SuccessRate      float64 `json:"success_rate"`
    AdjustmentFactor float64 `json:"adjustment_factor"`
    MinLevel         int     `json:"min_level"`
    MaxLevel         int     `json:"max_level"`
    
    ThresholdUp      float64 `json:"threshold_up"`    // e.g., 0.85
    ThresholdDown    float64 `json:"threshold_down"`  // e.g., 0.60
}
```

## Database Transaction Management

### Transaction Patterns
```go
type TransactionConfig struct {
    IsolationLevel sql.IsolationLevel
    Timeout       time.Duration
    RetryCount    int
    BackoffPolicy string // "linear", "exponential"
}

// Common transaction patterns
var TransactionPatterns = map[string]TransactionConfig{
    "study_session": {
        IsolationLevel: sql.LevelReadCommitted,
        Timeout:       time.Second * 5,
        RetryCount:    3,
        BackoffPolicy: "exponential",
    },
    "word_update": {
        IsolationLevel: sql.LevelSerializable,
        Timeout:       time.Second * 3,
        RetryCount:    2,
        BackoffPolicy: "linear",
    },
}
```

## API Response Caching Strategy

### Cache Configuration
```go
type CacheConfig struct {
    Type           string        // "memory", "redis"
    TTL            time.Duration
    MaxSize        int64        // bytes
    EvictionPolicy string       // "LRU", "LFU"
}

var CacheRules = map[string]CacheConfig{
    "/api/v1/words": {
        Type:           "memory",
        TTL:            time.Minute * 15,
        MaxSize:        1024 * 1024 * 10, // 10MB
        EvictionPolicy: "LRU",
    },
    "/api/v1/groups": {
        Type:           "memory",
        TTL:            time.Minute * 30,
        MaxSize:        1024 * 1024 * 5,  // 5MB
        EvictionPolicy: "LFU",
    },
}
```

## Study Session Analytics

### Session Metrics
```go
type SessionAnalytics struct {
    SessionID      string    `json:"session_id"`
    StartTime      time.Time `json:"start_time"`
    Duration       float64   `json:"duration_minutes"`
    
    WordMetrics struct {
        TotalWords      int     `json:"total_words"`
        CompletedWords  int     `json:"completed_words"`
        CorrectWords    int     `json:"correct_words"`
        AverageTimePerWord float64 `json:"avg_time_per_word"`
        DifficultWords  []struct {
            WordID    int     `json:"word_id"`
            Attempts  int     `json:"attempts"`
            TimeSpent float64 `json:"time_spent"`
        } `json:"difficult_words"`
    } `json:"word_metrics"`
    
    LearningEfficiency struct {
        OverallScore    float64 `json:"overall_score"`
        SpeedScore      float64 `json:"speed_score"`
        AccuracyScore   float64 `json:"accuracy_score"`
        RetentionScore  float64 `json:"retention_score"`
    } `json:"learning_efficiency"`
}
```

## Word Recommendation Engine

### Recommendation Algorithm
```go
type RecommendationConfig struct {
    Weights struct {
        DifficultyWeight   float64 `json:"difficulty_weight"`
        RelevanceWeight    float64 `json:"relevance_weight"`
        ProgressWeight     float64 `json:"progress_weight"`
        RepetitionWeight   float64 `json:"repetition_weight"`
    }
    
    Thresholds struct {
        MinConfidence     float64 `json:"min_confidence"`
        MaxDifficulty     float64 `json:"max_difficulty"`
        OptimalInterval   time.Duration `json:"optimal_interval"`
    }
}

type WordRecommendation struct {
    WordID          int     `json:"word_id"`
    Score           float64 `json:"score"`
    RecommendReason string  `json:"recommend_reason"`
    OptimalTime     time.Time `json:"optimal_time"`
    ExpectedDifficulty float64 `json:"expected_difficulty"`
}
```

## Background Task Management

### Task Scheduler
```go
type ScheduledTask struct {
    TaskID      string        `json:"task_id"`
    Schedule    string        `json:"schedule"` // Cron expression
    TaskType    string        `json:"task_type"`
    Priority    int          `json:"priority"`
    Timeout     time.Duration `json:"timeout"`
    RetryPolicy struct {
        MaxRetries  int           `json:"max_retries"`
        BackoffTime time.Duration `json:"backoff_time"`
    } `json:"retry_policy"`
}

var MaintenanceTasks = []ScheduledTask{
    {
        TaskID:   "cleanup_old_sessions",
        Schedule: "0 0 * * *",  // Daily at midnight
        TaskType: "maintenance",
        Priority: 1,
        Timeout:  time.Minute * 30,
    },
    {
        TaskID:   "update_word_statistics",
        Schedule: "0 */6 * * *", // Every 6 hours
        TaskType: "analytics",
        Priority: 2,
        Timeout:  time.Minute * 15,
    },
}
```

### Task Queue Management
```go
type TaskQueue struct {
    QueueName    string `json:"queue_name"`
    MaxSize      int    `json:"max_size"`
    Workers      int    `json:"workers"`
    
    Priorities struct {
        High   int `json:"high"`
        Normal int `json:"normal"`
        Low    int `json:"low"`
    } `json:"priorities"`
    
    Monitoring struct {
        QueueLength    int           `json:"queue_length"`
        ProcessingTime time.Duration `json:"processing_time"`
        ErrorRate      float64       `json:"error_rate"`
        LastProcessed  time.Time     `json:"last_processed"`
    } `json:"monitoring"`
}
```

## Data Synchronization

### Sync Strategy
```go
type SyncConfig struct {
    Mode           string        `json:"mode"`      // "full" or "incremental"
    Interval       time.Duration `json:"interval"`
    BatchSize      int          `json:"batch_size"`
    MaxRetries     int          `json:"max_retries"`
    ConflictPolicy string       `json:"conflict_policy"` // "client-wins" or "server-wins"
}

type SyncState struct {
    LastSyncTime time.Time `json:"last_sync_time"`
    Status       string    `json:"status"`
    Progress     float64   `json:"progress"`
    Errors       []string  `json:"errors"`
}
```

## Study Session Timeout Management

### Timeout Configuration
```go
type TimeoutConfig struct {
    SessionTimeout      time.Duration `json:"session_timeout"`
    WarningTime        time.Duration `json:"warning_time"`
    GracePeriod        time.Duration `json:"grace_period"`
    AutoSaveInterval   time.Duration `json:"auto_save_interval"`
}

var DefaultTimeouts = TimeoutConfig{
    SessionTimeout:    30 * time.Minute,
    WarningTime:      25 * time.Minute,
    GracePeriod:      5 * time.Minute,
    AutoSaveInterval: 1 * time.Minute,
}
```

## Word Grouping and Tagging System

### Tag Management
```go
type TagConfig struct {
    MaxTagsPerWord     int      `json:"max_tags_per_word"`
    MaxTagLength       int      `json:"max_tag_length"`
    ReservedTags      []string `json:"reserved_tags"`
    AutoTagRules      []struct {
        Pattern string `json:"pattern"`
        Tags    []string `json:"tags"`
    } `json:"auto_tag_rules"`
}

type TagStats struct {
    TagName     string `json:"tag_name"`
    WordCount   int    `json:"word_count"`
    UsageCount  int    `json:"usage_count"`
    LastUsed    time.Time `json:"last_used"`
    Difficulty  float64   `json:"difficulty"`
}
```

## Learning Path Generation

### Path Configuration
```go
type LearningPathConfig struct {
    Levels []struct {
        Name           string  `json:"name"`
        WordCount     int     `json:"word_count"`
        Difficulty    float64 `json:"difficulty"`
        Prerequisites []string `json:"prerequisites"`
    } `json:"levels"`
    
    Progression struct {
        MinSuccessRate float64 `json:"min_success_rate"`
        MinWordCount   int     `json:"min_word_count"`
        ReviewInterval time.Duration `json:"review_interval"`
    } `json:"progression"`
}

type PathProgress struct {
    CurrentLevel    string    `json:"current_level"`
    StartDate      time.Time `json:"start_date"`
    CompletedWords int       `json:"completed_words"`
    NextMilestone  struct {
        Type        string  `json:"type"`
        Target      int     `json:"target"`
        Progress    float64 `json:"progress"`
    } `json:"next_milestone"`
}
```

## Performance Profiling

### Profiling Configuration
```go
type ProfilingConfig struct {
    Enabled      bool          `json:"enabled"`
    SampleRate   float64       `json:"sample_rate"`
    MaxProfiles  int          `json:"max_profiles"`
    StoragePath  string       `json:"storage_path"`
    
    Triggers struct {
        SlowRequests    time.Duration `json:"slow_requests"`
        HighMemory      float64       `json:"high_memory"`
        HighCPU         float64       `json:"high_cpu"`
    } `json:"triggers"`
}

type ProfileMetadata struct {
    ProfileID   string    `json:"profile_id"`
    Type        string    `json:"type"`
    StartTime   time.Time `json:"start_time"`
    Duration    time.Duration `json:"duration"`
    TriggerType string    `json:"trigger_type"`
    Metrics     map[string]float64 `json:"metrics"`
}
```

## API Documentation Generation

### Documentation Config
```go
type SwaggerConfig struct {
    Title          string `json:"title"`
    Version        string `json:"version"`
    Description    string `json:"description"`
    BasePath       string `json:"base_path"`
    
    Security struct {
        Enabled     bool     `json:"enabled"`
        Schemes     []string `json:"schemes"`
        Definitions map[string]interface{} `json:"definitions"`
    } `json:"security"`
    
    UIConfig struct {
        Theme       string `json:"theme"`
        Language    string `json:"language"`
        Plugins     []string `json:"plugins"`
    } `json:"ui_config"`
}

type EndpointDoc struct {
    Path        string `json:"path"`
    Method      string `json:"method"`
    Summary     string `json:"summary"`
    Description string `json:"description"`
    Parameters  []struct {
        Name        string `json:"name"`
        Type        string `json:"type"`
        Required    bool   `json:"required"`
        Description string `json:"description"`
    } `json:"parameters"`
}
```

## Production Deployment Strategy

### Container Configuration
```yaml
# Docker configuration
services:
  api:
    build:
      context: .
      dockerfile: Dockerfile
    environment:
      - GO_ENV=production
      - API_PORT=8080
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
    resources:
      limits:
        memory: 512M
        cpu: "1.0"
      reservations:
        memory: 256M
        cpu: "0.5"
```

### Production Environment Variables
```bash
# Required environment variables for production
ENVIRONMENT=production
LOG_LEVEL=info
API_PORT=8080
DATABASE_PATH=/data/words.db
BACKUP_PATH=/backup
MAX_REQUEST_SIZE=10485760  # 10MB
GRACEFUL_SHUTDOWN_TIMEOUT=30s
CORS_ALLOWED_ORIGINS=https://app.example.com
```

## Database Backup Strategy

### Backup Configuration
```go
type BackupConfig struct {
    Schedule struct {
        FullBackup      string `json:"full_backup"`      // "0 0 * * *"
        IncrementalBackup string `json:"incremental_backup"` // "0 */6 * * *"
    }
    Retention struct {
        FullBackups      int `json:"full_backups"`       // 30 days
        IncrementalBackups int `json:"incremental_backups"` // 7 days
    }
    Compression struct {
        Algorithm string `json:"algorithm"`  // "gzip"
        Level     int    `json:"level"`     // 6
    }
    Storage struct {
        Path        string `json:"path"`
        MaxSize     int64  `json:"max_size"`  // 100GB
        AlertThreshold float64 `json:"alert_threshold"` // 0.85
    }
}
```

## Production Monitoring

### Prometheus Metrics
```go
type MetricsConfig struct {
    Endpoints []struct {
        Path    string `json:"path"`
        Labels  map[string]string `json:"labels"`
        Buckets []float64 `json:"buckets"`
    }
    CustomMetrics []struct {
        Name        string `json:"name"`
        Type        string `json:"type"` // counter, gauge, histogram
        Description string `json:"description"`
    }
}

var ProductionMetrics = []string{
    "http_request_duration_seconds",
    "http_requests_total",
    "database_connection_pool_size",
    "database_query_duration_seconds",
    "memory_usage_bytes",
    "goroutines_count",
    "study_sessions_active",
}
```

## Graceful Shutdown

### Shutdown Procedure
```go
type ShutdownConfig struct {
    Timeout     time.Duration `json:"timeout"`
    Procedures  []struct {
        Name     string        `json:"name"`
        Priority int           `json:"priority"`
        Timeout  time.Duration `json:"timeout"`
    }
}

var ShutdownSequence = ShutdownConfig{
    Timeout: 30 * time.Second,
    Procedures: []struct{
        {
            Name:     "http_server",
            Priority: 1,
            Timeout:  10 * time.Second,
        },
        {
            Name:     "active_sessions",
            Priority: 2,
            Timeout:  15 * time.Second,
        },
        {
            Name:     "database",
            Priority: 3,
            Timeout:  5 * time.Second,
        },
    },
}
```

## Production Security Measures

### Security Configuration
```go
type SecurityConfig struct {
    RateLimiting struct {
        Enabled     bool          `json:"enabled"`
        IPWhitelist []string      `json:"ip_whitelist"`
        BlockedIPs  []string      `json:"blocked_ips"`
    }
    Headers struct {
        HSTS              bool   `json:"hsts"`
        HSTSMaxAge        int    `json:"hsts_max_age"`
        FrameOptions      string `json:"frame_options"`
        ContentTypeOptions string `json:"content_type_options"`
    }
    SQLInjection struct {
        SanitizePatterns []string `json:"sanitize_patterns"`
        BlockedKeywords  []string `json:"blocked_keywords"`
    }
}
```

## Error Recovery Procedures

### Recovery Strategies
```go
type RecoveryProcedure struct {
    ErrorType    string   `json:"error_type"`
    Severity     int      `json:"severity"` // 1-5
    Actions      []string `json:"actions"`
    Notification struct {
        Channels []string `json:"channels"` // slack, email, pager
        Priority int      `json:"priority"`
    }
    Fallback struct {
        Enabled bool   `json:"enabled"`
        Mode    string `json:"mode"` // readonly, degraded
    }
}

var ProductionRecovery = map[string]RecoveryProcedure{
    "database_connection_lost": {
        Severity: 5,
        Actions: []string{
            "attempt_reconnect",
            "use_backup_db",
            "notify_admin",
        },
        Notification: {
            Channels: []string{"slack", "pager"},
            Priority: 1,
        },
    },
}
```

## Production Logging

### Log Configuration
```go
type LogConfig struct {
    Format     string   `json:"format"` // json
    Level      string   `json:"level"`  // info
    OutputPath []string `json:"output_path"`
    
    Retention struct {
        MaxSize    int      `json:"max_size"`    // MB
        MaxAge     int      `json:"max_age"`     // days
        MaxBackups int      `json:"max_backups"`
    }
    
    Sensitive []string `json:"sensitive_fields"` // fields to mask
}

var ProductionLogging = LogConfig{
    Format: "json",
    Level:  "info",
    OutputPath: []string{
        "/var/log/app/api.log",
        "stdout",
    },
    Retention: struct{
        MaxSize:    100,    // 100MB
        MaxAge:     30,     // 30 days
        MaxBackups: 5,
    },
    Sensitive: []string{
        "password",
        "token",
        "api_key",
    },
}
```
