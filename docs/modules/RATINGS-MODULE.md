# Ratings Module Development Guide

## Overview

The Ratings Module manages user ratings and reviews for rides, services, and providers. It provides rating submission, aggregation, statistics, and fraud detection capabilities.

## Module Structure

```
ratings/
├── handler.go         # HTTP request handlers
├── service.go         # Business logic
├── repository.go      # Database operations
├── routes.go          # Route definitions
└── dto/
    ├── requests.go    # Request payloads
    └── responses.go   # Response structures
```

## Key Responsibilities

1. Rating Submission - Collect ratings and reviews
2. Rating Aggregation - Calculate average ratings
3. Review Management - Store and retrieve reviews
4. Fraud Detection - Identify suspicious ratings
5. Statistics - Generate rating analytics

## Data Transfer Objects

### SubmitRatingRequest

```go
type SubmitRatingRequest struct {
    RideID          string `json:"ride_id" binding:"required"`
    ProviderID      string `json:"provider_id" binding:"required"`
    Stars           int    `json:"stars" binding:"required,min=1,max=5"`
    Review          string `json:"review,omitempty"`
    Cleanliness     int    `json:"cleanliness,omitempty"`
    Behavior        int    `json:"behavior,omitempty"`
    Communication   int    `json:"communication,omitempty"`
    Photos          []string `json:"photos,omitempty"`
}
```

### RatingResponse

```go
type RatingResponse struct {
    ID              string    `json:"id"`
    RideID          string    `json:"ride_id"`
    ProviderID      string    `json:"provider_id"`
    RiderID         string    `json:"rider_id"`
    Stars           int       `json:"stars"`
    Review          string    `json:"review,omitempty"`
    Cleanliness     int       `json:"cleanliness,omitempty"`
    Behavior        int       `json:"behavior,omitempty"`
    Communication   int       `json:"communication,omitempty"`
    PhotoURLs       []string  `json:"photo_urls,omitempty"`
    Helpful         int       `json:"helpful"` // count of helpful votes
    IsFlagged       bool      `json:"is_flagged"`
    CreatedAt       time.Time `json:"created_at"`
}
```

### ProviderRatingResponse

```go
type ProviderRatingResponse struct {
    AverageRating       float64            `json:"average_rating"`
    TotalRatings        int64              `json:"total_ratings"`
    Distribution        RatingDistribution `json:"distribution"`
    RecentRatings       []RatingResponse   `json:"recent_ratings"`
    RatingsByCategory   map[string]float64 `json:"ratings_by_category"`
    ResponseRate        float64            `json:"response_rate"`
    CompletionRate      float64            `json:"completion_rate"`
}

type RatingDistribution struct {
    FiveStar   int64 `json:"five_star"`
    FourStar   int64 `json:"four_star"`
    ThreeStar  int64 `json:"three_star"`
    TwoStar    int64 `json:"two_star"`
    OneStar    int64 `json:"one_star"`
}
```

## Handler Methods

```
SubmitRating(c *gin.Context)           // POST /ratings/submit
GetRating(c *gin.Context)              // GET /ratings/{id}
GetProviderRatings(c *gin.Context)     // GET /ratings/provider/{id}
UpdateRating(c *gin.Context)           // PUT /ratings/{id}
DeleteRating(c *gin.Context)           // DELETE /ratings/{id}
FlagInappropriate(c *gin.Context)      // POST /ratings/{id}/flag
GetHelpful(c *gin.Context)             // GET /ratings/{id}/helpful
MarkHelpful(c *gin.Context)            // POST /ratings/{id}/helpful
GetAverageRating(c *gin.Context)       // GET /ratings/average/{providerID}
```

## Service Methods

```
SubmitRating(ctx context.Context, req SubmitRatingRequest) error
GetRating(ctx context.Context, ratingID string) (*RatingResponse, error)
GetProviderRatings(ctx context.Context, providerID string) (*ProviderRatingResponse, error)
UpdateRating(ctx context.Context, ratingID string, updates map[string]interface{}) error
DeleteRating(ctx context.Context, ratingID string) error
FlagInappropriate(ctx context.Context, ratingID, reason string) error
MarkHelpful(ctx context.Context, ratingID, userID string) error
CalculateAverageRating(ctx context.Context, providerID string) (float64, error)
DetectFraudRating(ctx context.Context, req SubmitRatingRequest) (bool, string, error)
GetRatingStats(ctx context.Context, providerID string) (map[string]interface{}, error)
```

## Fraud Detection Algorithm

Ratings are analyzed for fraud using:

1. Pattern Detection
   - Same user rating same provider multiple times
   - All ratings at same time
   - Similar review text

2. Behavior Analysis
   - Extreme ratings (1 or 5 stars) without review
   - Ratings contradicting previous ones
   - Ratings for completed rides only

3. Content Analysis
   - Spam keywords detected
   - Generic reviews
   - Abusive language
   - Suspicious links

4. Temporal Patterns
   - Rating given too quickly after completion
   - Cluster of ratings in short time
   - Unusual rating frequency

## Typical Use Cases

### 1. Submit Rating

Request:
```
POST /ratings/submit
{
    "ride_id": "ride-123",
    "provider_id": "driver-456",
    "stars": 5,
    "review": "Great driver, very professional!",
    "cleanliness": 5,
    "behavior": 5,
    "communication": 4
}
```

Flow:
1. Validate ride exists and is completed
2. Check user hasn't already rated
3. Detect fraud patterns
4. Store rating if legitimate
5. Update provider average rating
6. Update provider statistics
7. Return confirmation

### 2. Get Provider Ratings

Request:
```
GET /ratings/provider/driver-456
```

Response:
```json
{
    "average_rating": 4.85,
    "total_ratings": 342,
    "distribution": {
        "five_star": 280,
        "four_star": 45,
        "three_star": 12,
        "two_star": 3,
        "one_star": 2
    },
    "recent_ratings": [
        {
            "id": "rating-1",
            "stars": 5,
            "review": "Excellent!",
            "created_at": "2024-02-20T10:30:00Z"
        }
    ],
    "ratings_by_category": {
        "cleanliness": 4.8,
        "behavior": 4.9,
        "communication": 4.7
    }
}
```

### 3. Flag Inappropriate Rating

Request:
```
POST /ratings/{ratingID}/flag
{
    "reason": "Abusive language"
}
```

Flow:
1. Find rating by ID
2. Create flag record
3. Count flags for this rating
4. Auto-hide if flags exceed threshold
5. Notify admin for review
6. Return confirmation

### 4. Mark Rating as Helpful

Request:
```
POST /ratings/{ratingID}/helpful
```

Flow:
1. Record helpful vote
2. Check for duplicate votes
3. Update helpful count
4. Return updated count
5. Hide rating if unhelpful votes exceed threshold

## Fraud Detection Thresholds

```
Confidence Scores:
- 0-30: Legitimate
- 30-70: Review needed
- 70-100: Likely fraud

Auto-hide if:
- Fraud score > 80
- Flag count > 10
- Helpful votes > 5x unhelpful votes
```

## Rating Visibility Rules

```
Show:
- Verified ratings
- Non-fraudulent ratings
- Ratings by verified users

Hide:
- Flagged ratings (pending review)
- Suspected fraud ratings
- Deleted ratings

Delete after:
- User request (24 hour grace period)
- Provider dispute (admin review)
- System cleanup (after 5 years)
```

## Database Schema

### Ratings Table

```sql
CREATE TABLE ratings (
    id VARCHAR(36) PRIMARY KEY,
    ride_id VARCHAR(36),
    provider_id VARCHAR(36) NOT NULL,
    rider_id VARCHAR(36) NOT NULL,
    stars INT,
    review TEXT,
    cleanliness INT,
    behavior INT,
    communication INT,
    fraud_score DECIMAL(5, 2),
    is_flagged BOOLEAN DEFAULT false,
    flag_reason VARCHAR(255),
    helpful_count INT DEFAULT 0,
    unhelpful_count INT DEFAULT 0,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    FOREIGN KEY (provider_id) REFERENCES users(id),
    FOREIGN KEY (rider_id) REFERENCES users(id),
    INDEX (provider_id, created_at)
);
```

### Rating Flags Table

```sql
CREATE TABLE rating_flags (
    id VARCHAR(36) PRIMARY KEY,
    rating_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36),
    reason VARCHAR(255),
    created_at TIMESTAMP,
    FOREIGN KEY (rating_id) REFERENCES ratings(id)
);
```

## Integration Points

1. Rides Module - For ride information
2. Drivers Module - For driver rating aggregation
3. Messages Module - For notifications
4. Admin Module - For fraud review

---
