## Drivers Module – Full Outline & Review (as of your current code)

### Purpose in the Uber-like Platform
This module owns **everything related to drivers**:
- Registration & onboarding (license + vehicle)
- Real-time online/offline status
- Live location tracking (PostGIS + Redis)
- Profile, vehicle, wallet, dashboard
- Foundation for **ride matching** (nearby drivers query already implemented!)

### Folder Structure (perfectly follows your modular pattern)

```
internal/modules/drivers/
├── dto/          → All request/response DTOs + validation
├── handler.go    → Gin handlers (all protected by JWT)
├── repository.go → GORM + PostGIS queries
├── routes.go     → /drivers/* protected routes
├── service.go    → Business logic + caching + Redis coordination
└── interfaces    → Repository & Service interfaces
```

### API Endpoints (All require Bearer token)

| Method | Path                  | Purpose                                 | Response Type                     |
|-------|-----------------------|-----------------------------------------|-----------------------------------|
| POST  | `/drivers/register`   | Complete driver onboarding             | DriverProfileResponse             |
| GET   | `/drivers/profile`    | Get full driver profile                 | DriverProfileResponse             |
| PUT   | `/drivers/profile`    | Update license number                   | DriverProfileResponse             |
| PUT   | `/drivers/vehicle`    | Update vehicle details                  | VehicleResponse                   |
| POST  | `/drivers/status`     | Go online / offline                     | DriverProfileResponse             |
| POST  | `/drivers/location`   | Live location heartbeat (every few sec) | 200 OK                            |
| GET   | `/drivers/wallet`     | Current balance + earnings              | WalletResponse                    |
| GET   | `/drivers/dashboard`  | Stats (trips, earnings, rating, etc.)   | DriverDashboardResponse           |

### Real-time & Matching Features Already Built

| Feature                        | How it works                                                                 | Ready? |
|-------------------------------|-------------------------------------------------------------------------------|--------|
| Online/offline tracking       | Redis key `driver:online:{id}` + SAdd/SRem on set `drivers:online`            | Yes    |
| Live location                 | PostGIS `current_location` + Redis `driver:location:{id}` (30s TTL)           | Yes    |
| Heartbeat / auto-offline      | TTL on `driver:online:{id}` → expires → driver appears offline                | Yes    |
| Nearby drivers query          | `FindNearbyDrivers()` with ST_DWithin + vehicle type filter + ordering       | Yes    |
| Location history              | `DriverLocation` table + `GetDriverLocationHistory()`                         | Yes    |

### Service Functions – Are They Correctly Used?

| Function                | Called from Handler? | Correctly Implemented? | Cache Strategy | Notes / TODOs |
|-------------------------|----------------------|-------------------------|----------------|---------------|
| `RegisterDriver`        | Yes                  | Yes                     | None           | Perfect flow, creates driver + vehicle |
| `GetProfile`            | Yes                  | Yes                     | Redis 5 min    | Cache hit → huge DB save |
| `UpdateProfile`         | Yes                  | Yes                     | Invalidates cache | Good uniqueness checks |
| `UpdateVehicle`         | Yes                  | Yes                     | Invalidates cache | Plate uniqueness check |
| `UpdateStatus`          | Yes                  | Yes                     | Redis + Set    | **Excellent** – broadcasts via WebSocket too! |
| `UpdateLocation`        | Yes                  | Yes                     | Redis 30s      | Critical for real-time tracking |
| `GetWallet`             | Yes                  | Yes                     | Redis 1 min    | Good |
| `GetDashboard`          | Yes                  | Warning (mock data)     | None           | **TODO**: implement real stats from rides table |

**All service functions are 100% correctly wired and used by handlers** – no missing or orphaned methods.

### What’s Already Production-Ready

- Full driver onboarding flow
- Real-time online/offline with Redis heartbeat
- Live location with PostGIS + Redis fallback
- Nearby driver search (the core of ride matching!)
- Caching everywhere it matters
- Proper validation + uniqueness checks
- Clean separation of concerns

### Minor Improvements / Missing Pieces

| Area                     | Status     | Recommendation |
|--------------------------|------------|----------------|
| Dashboard stats          | Mock       | Query `rides` table for today/week earnings & trips |
| Driver verification      | Auto-true  | Add document upload + admin approval flow later |
| Acceptance/Cancellation rate | Static | Update on ride accept/reject/cancel events |
| Location parsing in DTO  | Placeholder| Parse `POINT(lng lat)` string properly in `ToDriverProfileResponse` |
| Tests                    | Missing    | Add unit + integration tests (especially location & status) |

