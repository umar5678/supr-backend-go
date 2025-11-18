Tracking Module – Outline & Short Documentation
Purpose:
The tracking module handles real-time driver location management, enabling core ride-hailing features like nearby driver search, live ETA calculation, and streaming updates to riders. It uses Redis for fast access and PostGIS for persistent geo-queries, serving as the geospatial backbone for matching and in-trip tracking.
Core Features



































FeatureDescriptionImplementation NotesDriver Location UpdatesPeriodic (e.g., every few sec) updates from driver app; validated and cached.Redis (30s TTL) + async DB save for history.Nearby Drivers SearchFinds online, verified drivers within radius, filtered by vehicle type.PostGIS ST_DWithin + distance ordering; defaults to 5km, 20 limit.Location RetrievalGet current or history; supports active ride checks.Cache-first for current; DB for history.Streaming to RiderReal-time push of driver location during active rides via WebSocket.Interval-based ticker; error-resilient.Batch & ValidationHandles bulk saves; strict lat/lon/heading validation.Prevents bad data; optimized for scale.
Folder Structure (Modular as Usual)
textinternal/modules/tracking/
├── dto/              → Requests/responses (e.g., UpdateLocationRequest, NearbyDriversResponse)
├── handler.go        → Gin handlers for updates and queries
├── repository.go     → PostGIS + GORM geo ops
├── routes.go         → /tracking group (mixed auth/public)
├── service.go        → Logic with cache, validation, streaming
└── interfaces        → Repository & Service
API Endpoints Summary





























MethodPathAuth?DescriptionPOST/tracking/locationYesDriver updates locationGET/tracking/driver/{driverId}NoGet driver's current locationGET/tracking/nearbyNoFind nearby drivers (query params)
Key Design Decisions & Highlights

Hybrid Storage: Redis for sub-second reads (current location); PostGIS DB for history and complex queries.
Real-Time Focus: Short TTLs (30s location, 5min online); async DB writes to not block app.
Validation-First: DTOs enforce geo bounds; defaults for search (5km, 20 limit, only available drivers).
Scalable Search: Radius + vehicle filter; calculates distance/ETA on-the-fly (using utils/location).
Streaming Safety: Background goroutines with tickers; logs errors but continues.
No Polyline Yet: Basic location points; ready for extension (see suggestions below).
Dependencies: Relies on Drivers module for profiles; integrates with WebSocket for pushes.

Dependencies Used

Gin + binding
GORM + PostGIS
Redis (cache service)
Custom utils: location (distance/ETA calc), logger, response
WebSocket utils for streaming

In One Sentence
The tracking module delivers fast, geo-accurate driver positioning with real-time caching, search, and streaming — the essential enabler for ride matching and live trip monitoring in your Uber-like platform.