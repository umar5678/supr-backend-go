package fraud

import (
	"context"
	"encoding/json"
	"time"

	"github.com/umar5678/go-backend/internal/models"
	"github.com/umar5678/go-backend/internal/modules/fraud/dto"
	notificationsmodule "github.com/umar5678/go-backend/internal/modules/notifications"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Service interface {
	DetectFraudPatterns(ctx context.Context, rideID string) error
	GetFraudPattern(ctx context.Context, patternID string) (*dto.FraudPatternResponse, error)
	ListFraudPatterns(ctx context.Context, req dto.ListFraudPatternsRequest) ([]*dto.FraudPatternListResponse, int64, error)
	ReviewFraudPattern(ctx context.Context, patternID, reviewerID string, req dto.ReviewFraudPatternRequest) error
	GetFraudStats(ctx context.Context) (*dto.FraudStatsResponse, error)
	CheckUserRiskScore(ctx context.Context, userID string) (int, error)
}

type service struct {
	repo          Repository
	eventProducer notificationsmodule.EventProducer
}

func NewService(repo Repository) Service {
	return NewServiceWithNotifications(repo, nil)
}

func NewServiceWithNotifications(repo Repository, eventProducer notificationsmodule.EventProducer) Service {
	return &service{
		repo:          repo,
		eventProducer: eventProducer,
	}
}

func (s *service) DetectFraudPatterns(ctx context.Context, rideID string) error {
	go s.detectFrequentCancellations(context.Background(), rideID)
	go s.detectCollusionPattern(context.Background(), rideID)
	go s.detectShortDistanceHighFare(context.Background(), rideID)
	go s.detectLocationGaming(context.Background(), rideID)

	return nil
}

func (s *service) detectFrequentCancellations(ctx context.Context, rideID string) {
	// Get ride details (in real implementation)
	// For now, we'll use a placeholder userID
	userID := "placeholder-user-id"

	count, err := s.repo.CheckFrequentCancellations(ctx, userID, 7)
	if err != nil {
		return
	}

	if count > 10 {
		riskScore := 60 + (count-10)*5
		if riskScore > 100 {
			riskScore = 100
		}

		details := map[string]interface{}{
			"cancellation_count": count,
			"period_days":        7,
			"threshold":          10,
		}
		detailsJSON, _ := json.Marshal(details)

		pattern := &models.FraudPattern{
			PatternType: "frequent_cancellation",
			UserID:      &userID,
			RideID:      &rideID,
			Details:     string(detailsJSON),
			RiskScore:   riskScore,
			Status:      "flagged",
		}

		s.repo.Create(ctx, pattern)

		logger.Warn("fraud pattern detected: frequent cancellations",
			"userID", userID,
			"count", count,
			"riskScore", riskScore,
		)
	}
}

func (s *service) detectCollusionPattern(ctx context.Context, rideID string) {
	// Get ride details
	riderID := "placeholder-rider-id"
	driverID := "placeholder-driver-id"

	count, err := s.repo.CheckSameRiderDriverPair(ctx, riderID, driverID, 30)
	if err != nil {
		return
	}

	if count > 15 {
		riskScore := 70 + (count-15)*3
		if riskScore > 100 {
			riskScore = 100
		}

		details := map[string]interface{}{
			"ride_count":  count,
			"period_days": 30,
			"threshold":   15,
		}
		detailsJSON, _ := json.Marshal(details)

		pattern := &models.FraudPattern{
			PatternType: "same_rider_driver",
			UserID:      &riderID,
			DriverID:    &driverID,
			RideID:      &rideID,
			Details:     string(detailsJSON),
			RiskScore:   riskScore,
			Status:      "flagged",
		}

		s.repo.Create(ctx, pattern)

		logger.Warn("fraud pattern detected: collusion",
			"riderID", riderID,
			"driverID", driverID,
			"count", count,
			"riskScore", riskScore,
		)
	}
}

func (s *service) detectShortDistanceHighFare(ctx context.Context, rideID string) {
	suspicious, err := s.repo.CheckShortDistanceHighFare(ctx, rideID)
	if err != nil || !suspicious {
		return
	}

	details := map[string]interface{}{
		"reason": "Short distance with unusually high fare",
	}
	detailsJSON, _ := json.Marshal(details)

	pattern := &models.FraudPattern{
		PatternType: "fake_trips",
		RideID:      &rideID,
		Details:     string(detailsJSON),
		RiskScore:   85,
		Status:      "flagged",
	}

	s.repo.Create(ctx, pattern)

	logger.Warn("fraud pattern detected: short distance high fare",
		"rideID", rideID,
	)
}

func (s *service) detectLocationGaming(ctx context.Context, rideID string) {
	userID := "placeholder-user-id"

	gaming, err := s.repo.CheckLocationGaming(ctx, userID, 7)
	if err != nil || !gaming {
		return
	}

	details := map[string]interface{}{
		"reason": "Multiple rides with identical locations",
	}
	detailsJSON, _ := json.Marshal(details)

	pattern := &models.FraudPattern{
		PatternType: "location_gaming",
		UserID:      &userID,
		RideID:      &rideID,
		Details:     string(detailsJSON),
		RiskScore:   75,
		Status:      "flagged",
	}

	s.repo.Create(ctx, pattern)

	logger.Warn("fraud pattern detected: location gaming",
		"userID", userID,
	)
}

func (s *service) GetFraudPattern(ctx context.Context, patternID string) (*dto.FraudPatternResponse, error) {
	pattern, err := s.repo.FindByID(ctx, patternID)
	if err != nil {
		return nil, response.NotFoundError("Fraud pattern")
	}

	return dto.ToFraudPatternResponse(pattern), nil
}

func (s *service) ListFraudPatterns(ctx context.Context, req dto.ListFraudPatternsRequest) ([]*dto.FraudPatternListResponse, int64, error) {
	req.SetDefaults()

	filters := make(map[string]interface{})
	if req.PatternType != "" {
		filters["patternType"] = req.PatternType
	}
	if req.Status != "" {
		filters["status"] = req.Status
	}
	if req.MinRiskScore > 0 {
		filters["minRiskScore"] = req.MinRiskScore
	}

	patterns, total, err := s.repo.List(ctx, filters, req.Page, req.Limit)
	if err != nil {
		return nil, 0, response.InternalServerError("Failed to list fraud patterns", err)
	}

	result := make([]*dto.FraudPatternListResponse, len(patterns))
	for i, pattern := range patterns {
		result[i] = dto.ToFraudPatternListResponse(pattern)
	}

	return result, total, nil
}

func (s *service) ReviewFraudPattern(ctx context.Context, patternID, reviewerID string, req dto.ReviewFraudPatternRequest) error {
	if err := s.repo.Review(ctx, patternID, reviewerID, req.Status, req.ReviewNotes); err != nil {
		return response.InternalServerError("Failed to review fraud pattern", err)
	}

	logger.Info("fraud pattern reviewed",
		"patternID", patternID,
		"reviewerID", reviewerID,
		"status", req.Status,
	)

	return nil
}

func (s *service) GetFraudStats(ctx context.Context) (*dto.FraudStatsResponse, error) {
	stats, err := s.repo.GetFraudStats(ctx)
	if err != nil {
		return nil, response.InternalServerError("Failed to get fraud stats", err)
	}

	response := &dto.FraudStatsResponse{
		TotalPatterns:      int(stats["total"].(int64)),
		FlaggedCount:       0,
		InvestigatingCount: 0,
		ConfirmedCount:     0,
		DismissedCount:     0,
		ByType:             make(map[string]int),
		HighRiskCount:      int(stats["highRisk"].(int64)),
	}

	if flagged, ok := stats["flagged"].(int64); ok {
		response.FlaggedCount = int(flagged)
	}
	if investigating, ok := stats["investigating"].(int64); ok {
		response.InvestigatingCount = int(investigating)
	}
	if confirmed, ok := stats["confirmed"].(int64); ok {
		response.ConfirmedCount = int(confirmed)
	}
	if dismissed, ok := stats["dismissed"].(int64); ok {
		response.DismissedCount = int(dismissed)
	}

	if byType, ok := stats["byType"].(map[string]int64); ok {
		for k, v := range byType {
			response.ByType[k] = int(v)
		}
	}

	return response, nil
}

func (s *service) CheckUserRiskScore(ctx context.Context, userID string) (int, error) {
	patterns, _, err := s.repo.List(ctx, map[string]interface{}{
		"userID": userID,
		"status": "flagged",
	}, 1, 100)
	if err != nil {
		return 0, response.InternalServerError("Failed to list fraud patterns for user", err)
	}

	if len(patterns) == 0 {
		return 0, nil
	}

	totalRisk := 0
	for _, p := range patterns {
		totalRisk += p.RiskScore
	}

	avgRisk := totalRisk / len(patterns)
	return avgRisk, nil
}

func (s *service) publishFraudEvent(ctx context.Context, eventType notificationsmodule.EventType, userID string, data map[string]interface{}) {
	if s.eventProducer == nil {
		logger.Debug("event producer not available, skipping fraud event publication", "eventType", eventType, "userID", userID)
		return
	}

	payload := map[string]interface{}{
		"user_id":   userID,
		"timestamp": time.Now().UTC(),
	}

	for k, v := range data {
		payload[k] = v
	}

	go func() {
		// Use background context to prevent cancellation when HTTP request completes
		bgCtx := context.Background()
		if err := s.eventProducer.PublishEventWithKey(bgCtx, eventType, userID, payload); err != nil {
			logger.Error("failed to publish fraud event",
				"error", err,
				"eventType", eventType,
				"userID", userID,
			)
		}
	}()
}
