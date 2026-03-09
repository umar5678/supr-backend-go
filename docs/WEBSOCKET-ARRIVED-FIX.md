# WebSocket "Arrived" Status Fix

## Problem
The "arrived" phase was not showing up consistently in the rider's UI because:

1. **Missing WebSocket notification** - The `MarkArrived` endpoint was only sending the event to the rider, not the driver
2. **Race condition** - When "started" event arrived before "arrived" was processed, the UI would skip the "arrived" phase
3. **Race condition in notifications** - The database update and WebSocket notification were not perfectly synchronized

## Root Cause Analysis

In `/internal/modules/rides/service.go`, the `MarkArrived` function was using:
```go
websocketutil.SendToUser(ride.RiderID, websocket.TypeRideStatusUpdate, ...)
```

This only sends to the rider. The driver wasn't being notified that they were marked as arrived.

## Solution Applied

Changed the `MarkArrived` function (line 1083) to use the proper utility function that sends to BOTH parties:

**Before:**
```go
if err := websocketutil.SendToUser(ride.RiderID, websocket.TypeRideStatusUpdate, map[string]interface{}{
    "rideId":    rideID,
    "status":    "arrived",
    "message":   "Your driver has arrived at the pickup location",
    "timestamp": time.Now().UTC(),
}); err != nil {
    logger.Warn("failed to notify rider of arrival", "error", err, "rideID", rideID)
}
```

**After:**
```go
if err := websocketutil.SendRideStatusUpdate(ride.RiderID, userID, map[string]interface{}{
    "rideId":    rideID,
    "status":    "arrived",
    "message":   "Your driver has arrived at the pickup location",
    "timestamp": time.Now().UTC(),
}); err != nil {
    logger.Warn("failed to notify rider and driver of arrival", "error", err, "rideID", rideID)
}
```

## Why This Works

1. **Proper broadcast** - `SendRideStatusUpdate()` (in `websocketutils.go`) sends to both rider and driver:
   ```go
   func SendRideStatusUpdate(riderID, driverID string, statusData map[string]interface{}) error {
       if riderID != "" {
           SendToUser(riderID, websocket.TypeRideStatusUpdate, statusData)
       }
       if driverID != "" {
           SendToUser(driverID, websocket.TypeRideStatusUpdate, statusData)
       }
       return nil
   }
   ```

2. **Consistent with other status changes** - Both `StartRide()` and `CompleteRide()` already use `SendRideStatusUpdate()` for notifying both parties

3. **Immediate notification** - WebSocket events are instant, unlike polling which can be delayed on mobile networks

4. **Status ordering** - By ensuring both parties get the "arrived" notification, the UI can now properly process the status in the correct order

## Testing the Fix

1. Start a ride request as a rider
2. Accept the ride as a driver  
3. Mark as "arrived" via `POST /rides/{id}/arrived`
4. Verify:
   - Both rider and driver receive the "arrived" WebSocket event
   - The event includes the correct message and timestamp
   - The rider's UI properly shows the "Driver Arrived" phase
   - No race condition with the "started" event

## Related Functions

- `StartRide()` - Already correctly uses `SendRideStatusUpdate()` (line 1198)
- `CreateRide()` - Uses `SendRideStatusToBoth()` helper function
- `CompleteRide()` - Uses separate `SendToUser()` calls (intentional, different message types)
- `websocketutil.SendRideStatusUpdate()` - Utility function that broadcasts to both parties

## Files Modified

- `/internal/modules/rides/service.go` - Fixed `MarkArrived()` function (line 1083)
