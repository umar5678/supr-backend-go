package rides

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/umar5678/go-backend/internal/modules/messages"
	"github.com/umar5678/go-backend/internal/modules/rides/dto"
	"github.com/umar5678/go-backend/internal/utils/logger"
	"github.com/umar5678/go-backend/internal/utils/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// CreateRide godoc
// @Summary Create a new ride request
// @Tags rides
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateRideRequest true "Ride request data"
// @Success 201 {object} response.Response{data=dto.RideResponse}
// @Router /rides [post]
func (h *Handler) CreateRide(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req dto.CreateRideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	ride, err := h.service.CreateRide(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, ride, "Ride requested successfully")
}

// GetRide godoc
// @Summary Get ride details
// @Tags rides
// @Security BearerAuth
// @Produce json
// @Param id path string true "Ride ID"
// @Success 200 {object} response.Response{data=dto.RideResponse}
// @Router /rides/{id} [get]
func (h *Handler) GetRide(c *gin.Context) {
	userID, _ := c.Get("userID")
	rideID := c.Param("id")

	ride, err := h.service.GetRide(c.Request.Context(), userID.(string), rideID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, ride, "Ride retrieved successfully")
}

// ListRides godoc
// @Summary List user's rides
// @Tags rides
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param status query string false "Filter by status"
// @Param role query string true "User role (rider or driver)"
// @Success 200 {object} response.Response{data=[]dto.RideListResponse}
// @Router /rides [get]
func (h *Handler) ListRides(c *gin.Context) {
	userID, _ := c.Get("userID")
	role := c.Query("role") // "rider" or "driver"

	if role != "rider" && role != "driver" {
		c.Error(response.BadRequest("Role must be 'rider' or 'driver'"))
		return
	}

	var req dto.ListRidesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(response.BadRequest("Invalid query parameters"))
		return
	}

	rides, total, err := h.service.ListRides(c.Request.Context(), userID.(string), role, req)
	if err != nil {
		c.Error(err)
		return
	}

	pagination := response.NewPaginationMeta(total, req.Page, req.Limit)
	response.Paginated(c, rides, pagination, "Rides retrieved successfully")
}

// AcceptRide godoc
// @Summary Accept a ride request (Driver)
// @Tags rides
// @Security BearerAuth
// @Param id path string true "Ride ID"
// @Success 200 {object} response.Response{data=dto.RideResponse}
// @Router /rides/{id}/accept [post]
func (h *Handler) AcceptRide(c *gin.Context) {
	userID, _ := c.Get("userID") // ✅ This is user ID from JWT token, NOT driver profile ID
	rideID := c.Param("id")

	// ✅ Service will fetch driver profile using this userID
	ride, err := h.service.AcceptRide(c.Request.Context(), userID.(string), rideID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, ride, "Ride accepted successfully")
}

// RejectRide godoc
// @Summary Reject a ride request (Driver)
// @Tags rides
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Ride ID"
// @Param request body dto.RejectRideRequest true "Rejection data"
// @Success 200 {object} response.Response
// @Router /rides/{id}/reject [post]
func (h *Handler) RejectRide(c *gin.Context) {
	userID, _ := c.Get("userID") // ✅ User ID from JWT
	rideID := c.Param("id")

	var req dto.RejectRideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest(err.Error()))
		return
	}

	err := h.service.RejectRide(c.Request.Context(), userID.(string), rideID, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Ride rejected successfully")
}

// MarkArrived godoc
// @Summary Mark driver as arrived at pickup (Driver)
// @Tags rides
// @Security BearerAuth
// @Param id path string true "Ride ID"
// @Success 200 {object} response.Response{data=dto.RideResponse}
// @Router /rides/{id}/arrived [post]
func (h *Handler) MarkArrived(c *gin.Context) {
	userID, _ := c.Get("userID") // ✅ User ID from JWT
	rideID := c.Param("id")

	ride, err := h.service.MarkArrived(c.Request.Context(), userID.(string), rideID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, ride, "Marked as arrived")
}

// StartRide godoc
// @Summary Start the ride (Driver)
// @Tags rides
// @Security BearerAuth
// @Param id path string true "Ride ID"
// @Param request body dto.StartRideRequest true "Rider PIN"
// @Success 200 {object} response.Response{data=dto.RideResponse}
// @Router /rides/{id}/start [post]
func (h *Handler) StartRide(c *gin.Context) {
	userID, _ := c.Get("userID") // ✅ User ID from JWT
	rideID := c.Param("id")
	logger.Info("StartRide handler called", "userID", userID, "rideID", rideID)

	var req dto.StartRideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("failed to bind request body", "error", err)
		c.Error(response.BadRequest("Invalid request body"))
		return
	}
	logger.Info("request body parsed", "rideID", rideID)

	logger.Info("calling service.StartRide", "rideID", rideID)
	ride, err := h.service.StartRide(c.Request.Context(), userID.(string), rideID, req)
	if err != nil {
		logger.Error("service.StartRide returned error", "error", err, "rideID", rideID)
		c.Error(err)
		return
	}
	logger.Info("service.StartRide returned successfully", "rideID", rideID)

	logger.Info("calling response.Success", "rideID", rideID)
	response.Success(c, ride, "Ride started successfully")
	logger.Info("response.Success completed", "rideID", rideID)
}

// CompleteRide godoc
// @Summary Complete the ride (Driver)
// @Tags rides
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Ride ID"
// @Param request body dto.CompleteRideRequest true "Completion data"
// @Success 200 {object} response.Response{data=dto.RideResponse}
// @Router /rides/{id}/complete [post]
func (h *Handler) CompleteRide(c *gin.Context) {
	userID, _ := c.Get("userID") // ✅ User ID from JWT
	rideID := c.Param("id")

	var req dto.CompleteRideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest(err.Error()))
		return
	}

	ride, err := h.service.CompleteRide(c.Request.Context(), userID.(string), rideID, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, ride, "Ride completed successfully")
}

// CancelRide godoc
// @Summary Cancel a ride
// @Tags rides
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Ride ID"
// @Param request body dto.CancelRideRequest true "Cancellation data"
// @Success 200 {object} response.Response
// @Router /rides/{id}/cancel [post]
func (h *Handler) CancelRide(c *gin.Context) {
	userID, _ := c.Get("userID")
	rideID := c.Param("id")

	var req dto.CancelRideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req = dto.CancelRideRequest{}
	}

	if err := h.service.CancelRide(c.Request.Context(), userID.(string), rideID, req); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "Ride cancelled successfully")
}

// TriggerSOS godoc
// @Summary Trigger SOS alert during active ride (Rider)
// @Tags rides
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Ride ID"
// @Param request body map[string]interface{} true "Location data with latitude and longitude"
// @Success 200 {object} response.Response
// @Router /rides/{id}/emergency [post]
func (h *Handler) TriggerSOS(c *gin.Context) {
	userID, _ := c.Get("userID")
	rideID := c.Param("id")

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	// Extract location from request
	var latitude, longitude float64
	if lat, ok := req["latitude"].(float64); ok {
		latitude = lat
	} else {
		c.Error(response.BadRequest("Latitude is required"))
		return
	}

	if lon, ok := req["longitude"].(float64); ok {
		longitude = lon
	} else {
		c.Error(response.BadRequest("Longitude is required"))
		return
	}

	if err := h.service.TriggerSOS(c.Request.Context(), userID.(string), rideID, latitude, longitude); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, nil, "SOS alert triggered - Help is on the way")
}

// GetAvailableCars godoc
// @Summary Get available cars near the rider
// @Tags rides
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.AvailableCarRequest true "Location and radius"
// @Success 200 {object} response.Response{data=dto.AvailableCarsListResponse}
// @Router /rides/available-cars [post]
func (h *Handler) GetAvailableCars(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req dto.AvailableCarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	// Set default radius if not provided
	if req.RadiusKm == 0 {
		req.RadiusKm = 5.0
	}

	cars, err := h.service.GetAvailableCars(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, cars, "Available cars fetched successfully")
}

// GetVehiclesWithDetails godoc
// @Summary Get available vehicles with complete pricing and driver details
// @Description Returns nearby online drivers with their vehicles, including pricing estimates, surge multipliers, and demand information
// @Tags rides
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.VehicleDetailsRequest true "Pickup and dropoff locations"
// @Success 200 {object} response.Response{data=dto.VehiclesWithDetailsListResponse}
// @Router /rides/vehicles-with-details [post]
func (h *Handler) GetVehiclesWithDetails(c *gin.Context) {
	userID, _ := c.Get("userID")

	var req dto.VehicleDetailsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(response.BadRequest("Invalid request body"))
		return
	}

	// Set default radius if not provided
	if req.RadiusKm == 0 {
		req.RadiusKm = 5.0
	}

	vehicles, err := h.service.GetVehiclesWithDetails(c.Request.Context(), userID.(string), req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, vehicles, "Vehicles with details fetched successfully")
}

// ====================================================
//  Ride Messaging
// ====================================================

// RideMessageManager handles real-time messaging for rides
type RideMessageManager struct {
	connections map[string]map[string]*websocket.Conn // rideID -> userID -> conn
	mu          sync.RWMutex
	msgService  messages.Service
	broadcast   chan *BroadcastMessage
}

type BroadcastMessage struct {
	RideID    string
	UserID    string
	EventType string
	Data      interface{}
	ExcludeID string
}

type WSChatMessage struct {
	Type     string                 `json:"type"` // "message", "typing", "read"
	RideID   string                 `json:"rideId"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
}

func NewRideMessageManager(msgService messages.Service) *RideMessageManager {
	rm := &RideMessageManager{
		connections: make(map[string]map[string]*websocket.Conn),
		msgService:  msgService,
		broadcast:   make(chan *BroadcastMessage, 100),
	}
	go rm.handleBroadcast()
	return rm
}

func (rm *RideMessageManager) RegisterConnection(rideID, userID string, conn *websocket.Conn) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.connections[rideID] == nil {
		rm.connections[rideID] = make(map[string]*websocket.Conn)
	}

	rm.connections[rideID][userID] = conn
	logger.Info("user connected to ride chat", "rideID", rideID, "userID", userID)

	// Notify other participants that user is online
	rm.broadcast <- &BroadcastMessage{
		RideID:    rideID,
		UserID:    userID,
		EventType: "user_online",
		Data:      map[string]string{"userId": userID, "status": "online"},
		ExcludeID: userID,
	}
}

func (rm *RideMessageManager) UnregisterConnection(rideID, userID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if conns, exists := rm.connections[rideID]; exists {
		delete(conns, userID)
		if len(conns) == 0 {
			delete(rm.connections, rideID)
		}
	}

	logger.Info("user disconnected from ride chat", "rideID", rideID, "userID", userID)

	// Notify other participants that user is offline
	rm.broadcast <- &BroadcastMessage{
		RideID:    rideID,
		UserID:    userID,
		EventType: "user_offline",
		Data:      map[string]string{"userId": userID, "status": "offline"},
		ExcludeID: userID,
	}
}

func (rm *RideMessageManager) BroadcastMessage(rideID, userID, senderType, content string, metadata map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Save message to database
	msgResp, err := rm.msgService.SendMessage(ctx, rideID, userID, senderType, content, metadata)
	if err != nil {
		return err
	}

	// Broadcast to all connected users in this ride
	rm.broadcast <- &BroadcastMessage{
		RideID:    rideID,
		UserID:    userID,
		EventType: "message",
		Data:      msgResp,
	}

	return nil
}

func (rm *RideMessageManager) MarkAsRead(messageID, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return rm.msgService.MarkAsRead(ctx, messageID, userID)
}

func (rm *RideMessageManager) BroadcastTyping(rideID, userID, senderType string) {
	rm.broadcast <- &BroadcastMessage{
		RideID:    rideID,
		UserID:    userID,
		EventType: "typing",
		Data: map[string]string{
			"userId":     userID,
			"senderType": senderType,
		},
		ExcludeID: userID,
	}
}

func (rm *RideMessageManager) handleBroadcast() {
	for msg := range rm.broadcast {
		rm.mu.RLock()
		connections := rm.connections[msg.RideID]
		rm.mu.RUnlock()

		response := map[string]interface{}{
			"type":      msg.EventType,
			"data":      msg.Data,
			"timestamp": time.Now(),
		}

		data, _ := json.Marshal(response)

		for userID, conn := range connections {
			// Skip sender for certain events
			if msg.ExcludeID != "" && userID == msg.ExcludeID {
				continue
			}

			go func(c *websocket.Conn) {
				c.SetWriteDeadline(time.Now().Add(10 * time.Second))
				if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
					logger.Error("failed to write message", "error", err)
				}
			}(conn)
		}
	}
}

func (rm *RideMessageManager) HandleConnection(conn *websocket.Conn, rideID, userID, senderType string) {
	rm.RegisterConnection(rideID, userID, conn)
	defer rm.UnregisterConnection(rideID, userID)

	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var msg WSChatMessage
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("websocket error", "error", err)
			}
			return
		}

		msg.RideID = rideID

		switch msg.Type {
		case "message":
			if err := rm.BroadcastMessage(rideID, userID, senderType, msg.Content, msg.Metadata); err != nil {
				logger.Error("failed to broadcast message", "error", err)
			}

		case "typing":
			rm.BroadcastTyping(rideID, userID, senderType)

		case "read":
			if messageID, ok := msg.Metadata["messageId"].(string); ok {
				if err := rm.MarkAsRead(messageID, userID); err != nil {
					logger.Error("failed to mark as read", "error", err)
				}
			}
		}

		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	}
}
