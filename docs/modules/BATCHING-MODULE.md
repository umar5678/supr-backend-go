# Batching Module Documentation

## Overview

The Batching Module handles background job processing, batch operations, scheduled tasks, and asynchronous job execution. It provides a framework for running time-consuming operations outside the request-response cycle.

## Key Responsibilities

- Background job queue management
- Job scheduling and cron tasks
- Batch processing of large data sets
- Periodic maintenance and cleanup tasks
- Report generation
- Data exports and imports
- Settlement and payment batch processing
- Email/SMS delivery queues
- Daily reconciliation and reporting

## Architecture

### Handler Layer (`batching/handler.go`)

Handles HTTP requests for job management and monitoring.

**Key Endpoints:**

```go
POST /api/v1/jobs/submit               // Submit a batch job
GET /api/v1/jobs                       // List jobs (admin)
GET /api/v1/jobs/:id                   // Get job details
PUT /api/v1/jobs/:id/cancel            // Cancel a job
GET /api/v1/jobs/:id/progress          // Get job progress
GET /api/v1/jobs/:id/results           // Get job results
POST /api/v1/scheduled-tasks           // Create scheduled task (admin)
GET /api/v1/scheduled-tasks            // List scheduled tasks (admin)
PUT /api/v1/scheduled-tasks/:id        // Update scheduled task
DELETE /api/v1/scheduled-tasks/:id     // Delete scheduled task
```

### Service Layer (`batching/service.go`)

Contains business logic for job processing.

**Key Methods:**

```go
func (s *BatchingService) SubmitJob(ctx context.Context, req *JobRequest) (*Job, error)
func (s *BatchingService) GetJob(ctx context.Context, jobID string) (*Job, error)
func (s *BatchingService) ListJobs(ctx context.Context, filters *JobFilters) ([]*Job, error)
func (s *BatchingService) CancelJob(ctx context.Context, jobID string) error
func (s *BatchingService) GetJobProgress(ctx context.Context, jobID string) (*JobProgress, error)
func (s *BatchingService) GetJobResults(ctx context.Context, jobID string) (interface{}, error)
func (s *BatchingService) ProcessJob(ctx context.Context, jobID string) error
func (s *BatchingService) ScheduleTask(ctx context.Context, req *ScheduledTaskRequest) (*ScheduledTask, error)
func (s *BatchingService) GetScheduledTasks(ctx context.Context) ([]*ScheduledTask, error)
func (s *BatchingService) ExecuteScheduledTasks(ctx context.Context) error
```

### Repository Layer (`batching/repository.go`)

Manages database operations.

**Key Methods:**

```go
func (r *BatchingRepository) CreateJob(ctx context.Context, job *Job) error
func (r *BatchingRepository) GetJob(ctx context.Context, jobID string) (*Job, error)
func (r *BatchingRepository) UpdateJob(ctx context.Context, jobID string, updates map[string]interface{}) error
func (r *BatchingRepository) DeleteJob(ctx context.Context, jobID string) error
func (r *BatchingRepository) GetPendingJobs(ctx context.Context) ([]*Job, error)
func (r *BatchingRepository) UpdateJobProgress(ctx context.Context, jobID string, progress *JobProgress) error
func (r *BatchingRepository) SaveJobResults(ctx context.Context, jobID string, results interface{}) error
func (r *BatchingRepository) CreateScheduledTask(ctx context.Context, task *ScheduledTask) error
func (r *BatchingRepository) GetScheduledTasks(ctx context.Context) ([]*ScheduledTask, error)
```

## Data Models

### Job

```go
type Job struct {
    ID                  string     `db:"id" json:"id"`
    JobType             string     `db:"job_type" json:"job_type"`                  // SETTLEMENT, EXPORT, IMPORT, REPORT, CLEANUP, etc.
    Status              string     `db:"status" json:"status"`                      // PENDING, PROCESSING, COMPLETED, FAILED, CANCELLED
    Priority            string     `db:"priority" json:"priority"`                  // LOW, MEDIUM, HIGH, URGENT
    
    // Input and Configuration
    Parameters          string     `db:"parameters" json:"parameters"`              // JSON encoded job parameters
    InputData           *string    `db:"input_data" json:"input_data"`              // JSON or S3 URL
    
    // Progress and Status
    Progress            float64    `db:"progress" json:"progress"`                  // 0-100 percentage
    CurrentItem         int        `db:"current_item" json:"current_item"`          // For batch progress
    TotalItems          int        `db:"total_items" json:"total_items"`
    
    // Results
    ResultStatus        *string    `db:"result_status" json:"result_status"`        // SUCCESS, PARTIAL_FAILURE, FAILURE
    ResultData          *string    `db:"result_data" json:"result_data"`            // JSON results or S3 URL
    OutputPath          *string    `db:"output_path" json:"output_path"`            // S3 or file path for large results
    ErrorMessage        *string    `db:"error_message" json:"error_message"`
    
    // Metadata
    CreatedBy           string     `db:"created_by" json:"created_by"`              // User ID who submitted
    CreatedAt           time.Time  `db:"created_at" json:"created_at"`
    StartedAt           *time.Time `db:"started_at" json:"started_at"`
    CompletedAt         *time.Time `db:"completed_at" json:"completed_at"`
    Duration            *int       `db:"duration" json:"duration"`                  // Seconds
    
    // Retry
    RetryCount          int        `db:"retry_count" json:"retry_count"`
    MaxRetries          int        `db:"max_retries" json:"max_retries"`
    NextRetryAt         *time.Time `db:"next_retry_at" json:"next_retry_at"`
    
    // Scheduling
    ScheduledFor        *time.Time `db:"scheduled_for" json:"scheduled_for"`        // For deferred execution
    TimeoutSeconds      int        `db:"timeout_seconds" json:"timeout_seconds"`    // Job timeout limit
}
```

### ScheduledTask

```go
type ScheduledTask struct {
    ID                  string     `db:"id" json:"id"`
    Name                string     `db:"name" json:"name"`
    Description         *string    `db:"description" json:"description"`
    JobType             string     `db:"job_type" json:"job_type"`
    
    // Schedule Configuration
    CronExpression      string     `db:"cron_expression" json:"cron_expression"`    // Standard cron format
    // Examples:
    // "0 2 * * *"     - Daily at 2 AM
    // "0 */1 * * *"   - Every hour
    // "0 0 * * MON"   - Every Monday at midnight
    
    Timezone            string     `db:"timezone" json:"timezone"`                  // America/New_York, etc.
    IsEnabled           bool       `db:"is_enabled" json:"is_enabled"`
    
    // Job Parameters
    Parameters          string     `db:"parameters" json:"parameters"`              // JSON encoded parameters
    
    // Execution History
    LastExecutedAt      *time.Time `db:"last_executed_at" json:"last_executed_at"`
    NextExecutionAt     *time.Time `db:"next_execution_at" json:"next_execution_at"`
    LastStatus          *string    `db:"last_status" json:"last_status"`            // SUCCESS, FAILED
    LastErrorMessage    *string    `db:"last_error_message" json:"last_error_message"`
    ExecutionCount      int        `db:"execution_count" json:"execution_count"`
    
    // Configuration
    MaxConcurrentJobs   int        `db:"max_concurrent_jobs" json:"max_concurrent_jobs"` // Prevent overload
    TimeoutSeconds      int        `db:"timeout_seconds" json:"timeout_seconds"`
    Retries             int        `db:"retries" json:"retries"`
    
    CreatedBy           string     `db:"created_by" json:"created_by"`
    CreatedAt           time.Time  `db:"created_at" json:"created_at"`
    UpdatedAt           time.Time  `db:"updated_at" json:"updated_at"`
}
```

### JobProgress

```go
type JobProgress struct {
    JobID               string     `json:"job_id"`
    Status              string     `json:"status"`
    Progress            float64    `json:"progress"`                                 // 0-100
    CurrentItem         int        `json:"current_item"`
    TotalItems          int        `json:"total_items"`
    ItemsProcessed      int        `json:"items_processed"`
    ItemsSuccessful     int        `json:"items_successful"`
    ItemsFailed         int        `json:"items_failed"`
    CurrentStep         string     `json:"current_step"`                            // Description of what's being processed
    EstimatedTimeLeft   int        `json:"estimated_time_left"`                     // Seconds
    Message             string     `json:"message"`
}
```

## DTOs (Data Transfer Objects)

### JobRequest

```go
type JobRequest struct {
    JobType             string                 `json:"job_type" binding:"required"`
    Priority            string                 `json:"priority"`                     // LOW, MEDIUM, HIGH
    Parameters          map[string]interface{} `json:"parameters"`
    ScheduledFor        *time.Time             `json:"scheduled_for"`
    TimeoutSeconds      int                    `json:"timeout_seconds"`              // Default: 3600 (1 hour)
}
```

### ScheduledTaskRequest

```go
type ScheduledTaskRequest struct {
    Name                string                 `json:"name" binding:"required"`
    Description         *string                `json:"description"`
    JobType             string                 `json:"job_type" binding:"required"`
    CronExpression      string                 `json:"cron_expression" binding:"required"`
    Timezone            string                 `json:"timezone"`
    Parameters          map[string]interface{} `json:"parameters"`
    IsEnabled           bool                   `json:"is_enabled"`
    MaxConcurrentJobs   int                    `json:"max_concurrent_jobs"`
    Retries             int                    `json:"retries"`
}
```

### JobResponse

```go
type JobResponse struct {
    ID                  string                 `json:"id"`
    JobType             string                 `json:"job_type"`
    Status              string                 `json:"status"`
    Progress            float64                `json:"progress"`
    CreatedAt           time.Time              `json:"created_at"`
    StartedAt           *time.Time             `json:"started_at"`
    CompletedAt         *time.Time             `json:"completed_at"`
    Duration            *int                   `json:"duration"`
    ErrorMessage        *string                `json:"error_message"`
}
```

## Database Schema

### jobs table

```sql
CREATE TABLE jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED', 'CANCELLED')),
    priority VARCHAR(50) DEFAULT 'MEDIUM',
    parameters JSONB,
    input_data TEXT,
    progress DECIMAL(5, 2) DEFAULT 0,
    current_item INTEGER DEFAULT 0,
    total_items INTEGER DEFAULT 0,
    result_status VARCHAR(50),
    result_data TEXT,
    output_path TEXT,
    error_message TEXT,
    created_by UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    duration INTEGER,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    next_retry_at TIMESTAMP,
    scheduled_for TIMESTAMP,
    timeout_seconds INTEGER DEFAULT 3600
);

CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_job_type ON jobs(job_type);
CREATE INDEX idx_jobs_created_at ON jobs(created_at);
CREATE INDEX idx_jobs_created_by ON jobs(created_by);
CREATE INDEX idx_jobs_scheduled_for ON jobs(scheduled_for);
```

### scheduled_tasks table

```sql
CREATE TABLE scheduled_tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(200) NOT NULL UNIQUE,
    description TEXT,
    job_type VARCHAR(100) NOT NULL,
    cron_expression VARCHAR(50) NOT NULL,
    timezone VARCHAR(100) DEFAULT 'UTC',
    is_enabled BOOLEAN DEFAULT TRUE,
    parameters JSONB,
    last_executed_at TIMESTAMP,
    next_execution_at TIMESTAMP,
    last_status VARCHAR(50),
    last_error_message TEXT,
    execution_count INTEGER DEFAULT 0,
    max_concurrent_jobs INTEGER DEFAULT 1,
    timeout_seconds INTEGER DEFAULT 3600,
    retries INTEGER DEFAULT 1,
    created_by UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_scheduled_tasks_is_enabled ON scheduled_tasks(is_enabled);
CREATE INDEX idx_scheduled_tasks_next_execution_at ON scheduled_tasks(next_execution_at);
```

### job_logs table (Optional)

```sql
CREATE TABLE job_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    level VARCHAR(50),                                               -- DEBUG, INFO, WARNING, ERROR
    message TEXT,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_job_logs_job_id ON job_logs(job_id);
CREATE INDEX idx_job_logs_timestamp ON job_logs(timestamp);
```

## Common Job Types

### Settlement Processing
```
- Collects all earnings for drivers/providers
- Calculates deductions and commissions
- Processes bank transfers
- Updates wallet balances
- Generates settlement reports
```

### Data Export
```
- Exports user data for GDPR requests
- Packages ride history, transactions, profile
- Generates ZIP with JSON/CSV files
- Encrypts and sends to user email
```

### Report Generation
```
- Daily/monthly revenue reports
- Driver performance reports
- Fraud analytics reports
- User growth metrics
- Payment reconciliation reports
```

### Cleanup Tasks
```
- Delete expired OTPs
- Remove old audit logs (90+ days)
- Clean up temporary files
- Archive old transactions
- Cleanup failed job logs
```

### Batch Notifications
```
- Send daily digest emails
- Send promotional bulk emails
- Send reminders to inactive users
- Process SMS queues
```

## Use Cases

### Use Case 1: Submit Data Export Job

```
1. User requests data export via profile page
2. System creates Job with type=EXPORT
3. Job queued with status=PENDING
4. Background worker picks up job
5. Collects user profile, ride history, wallet data
6. Generates JSON/CSV files
7. Zips and uploads to S3
8. Sends download link via email
9. Job status updated to COMPLETED
10. User can download within 7 days
```

### Use Case 2: Scheduled Daily Settlement

```
1. ScheduledTask configured: "Daily Settlement" at 2 AM
2. Cron job runs at scheduled time
3. Triggers Settlement job creation
4. Job processes all driver earnings
5. Calculates commissions and payouts
6. Creates bank transfer batch
7. Updates wallet balances
8. Generates settlement report
9. Job status updated to COMPLETED
10. Admin receives completion notification
```

### Use Case 3: Long-Running Export with Progress

```
1. Admin exports 1 million user records
2. Job created with total_items=1000000
3. Background worker processes in batches
4. Updates progress every 10,000 records
5. Admin can check progress via GET /jobs/{id}/progress
6. Progress returned: {"progress": 45, "current_item": 450000}
7. Estimated time left calculated
8. Job continues until all records processed
```

### Use Case 4: Job Failure with Retry

```
1. Database export job fails (connection timeout)
2. Status updated to FAILED
3. Error message logged
4. System queues retry (retry_count < max_retries)
5. next_retry_at set to 30 seconds later
6. Background worker picks up at retry time
7. Job re-executes
8. This time succeeds
9. Status updated to COMPLETED
```

## Common Operations

### Submit a Batch Job

```go
handler := func(c *gin.Context) {
    var req JobRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
        return
    }

    userID := c.GetString("user_id")
    
    job, err := s.batchingService.SubmitJob(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    // Queue job for async processing
    go s.processJobAsync(job.ID)

    c.JSON(http.StatusAccepted, job)
}
```

### Get Job Progress

```go
handler := func(c *gin.Context) {
    jobID := c.Param("id")
    
    progress, err := s.batchingService.GetJobProgress(c.Request.Context(), jobID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(http.StatusOK, progress)
}
```

### Schedule a Recurring Task

```go
handler := func(c *gin.Context) {
    var req ScheduledTaskRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
        return
    }

    task, err := s.batchingService.ScheduleTask(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(http.StatusCreated, task)
}
```

### Background Job Worker Loop

```go
func (s *BatchingService) StartWorker() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for range ticker.C {
        // Get pending jobs
        jobs, _ := s.repo.GetPendingJobs(context.Background())
        
        for _, job := range jobs {
            // Process job
            go s.ProcessJob(context.Background(), job.ID)
        }
        
        // Check and execute scheduled tasks
        s.ExecuteScheduledTasks(context.Background())
    }
}
```

## Error Handling

| Error | Status Code | Message |
|-------|------------|---------|
| Job not found | 404 | "Job not found" |
| Invalid job type | 400 | "Unknown job type" |
| Invalid cron expression | 400 | "Invalid cron expression" |
| Job already processing | 409 | "Job is already processing" |
| Job timeout | 504 | "Job exceeded timeout limit" |
| Insufficient permissions | 403 | "You don't have permission for this operation" |
| Task limit exceeded | 429 | "Too many concurrent jobs" |

## Performance Optimization

### Job Processing
- Use worker pool pattern (5-10 workers)
- Process jobs in priority order
- Implement job locking (prevent duplicate processing)
- Use batch operations for database updates

### Scheduled Tasks
- Use cron library for accurate scheduling
- Calculate next execution time after completion
- Track execution history for monitoring
- Implement health checks for failed tasks

### Database
- Index on status for quick pending job lookup
- Partition jobs table by date (old jobs archived)
- Index on created_at for date range queries
- Vacuum regularly to maintain performance

## Security Considerations

### Job Security
- Validate job parameters
- Implement permission checks (who can submit/cancel)
- Audit all job executions
- Encrypt sensitive output data

### Data Protection
- PII should be masked in logs
- Results stored securely
- Implement access control on results
- Cleanup sensitive data after completion

## Testing Strategy

### Unit Tests
- Job validation and creation
- Cron expression parsing
- Progress calculation
- Retry logic

### Integration Tests
- End-to-end job workflow
- Scheduled task execution
- Error handling and retry
- Database state after completion

## Integration Points

### With Wallet Module
- Settlement job processes driver payouts

### With Notifications Module
- Sends job completion notifications

### With Fraud Module
- Fraud detection batch processing

### With Ratings Module
- Rating cleanup and archival

## Common Pitfalls

1. **Not implementing job locking** - Can process same job twice
2. **Missing timeout handling** - Jobs can run forever
3. **No progress tracking** - Users don't know job status
4. **Poor error messages** - Admins can't debug failures
5. **Not retrying failed jobs** - One-time failures block processing
6. **Blocking request threads** - Jobs should be async only

---

**Module Status:** Fully Documented
**Last Updated:** February 22, 2026
**Version:** 1.0
