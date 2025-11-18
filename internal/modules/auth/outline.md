## Auth Module – Overview & Short Documentation

**Purpose**:  
The `auth` module is the complete authentication & user profile system for your Uber-like ride-hailing platform. It supports two very different authentication flows while keeping a single source of truth for users.

### Core Features

| Feature                      | Description                                                                                   | Who uses it                     |
|------------------------------|-----------------------------------------------------------------------------------------------|---------------------------------|
| Phone-based auth (OTP-less)  | Signup & login using only phone + name + role. First-time phone → auto-signup, next → login. | Riders & Drivers (main users)   |
| Email + Password auth        | Classic signup/login with password hashing.                                                   | Admin, Delivery, Handyman, etc. |
| JWT Access + Refresh tokens  | Short-lived access token + long-lived refresh token. Refresh endpoint + blacklist on logout.  | All users                       |
| Wallet auto-creation         | On first signup, riders get $1000 fake balance, drivers get $0.                           | Riders & Drivers                |
| Rider profile auto-creation  | When a rider signs up, it automatically calls `riderService.CreateProfile`.                 | Riders only                     |
| Profile caching (Redis)      | User profile cached for 5 min, invalidated on update/logout.                                 | All authenticated calls         |
| Token blacklisting           | Refresh tokens are blacklisted in Redis on logout/refresh.                                      | Security                        |

### Folder Structure (as per your modular design)

```
internal/modules/auth/
├── dto/              → Request/response structs + manual validation
├── handler.go        → Gin HTTP handlers + Swagger docs
├── repository.go     → GORM user & wallet operations
├── routes.go         → Route registration (public + protected)
├── service.go        → Business logic, token generation, wallet/profile creation
└── (interfaces)      → Repository & Service interfaces
```

### API Endpoints Summary

| Method | Path                     | Auth    | Description                              |
|-------|--------------------------|---------|------------------------------------------|
| POST  | `/auth/phone/signup`     | Public  | Rider/Driver signup (auto-login if exists) |
| POST  | `/auth/phone/login`      | Public  | Rider/Driver login                       |
| POST  | `/auth/email/signup`     | Public  | Admin/Delivery/etc signup                |
| POST  | `/auth/email/login`      | Public  | Email + password login                   |
| POST  | `/auth/refresh`          | Public  | Refresh access & refresh tokens          |
| POST  | `/auth/logout`           | Bearer  | Blacklist refresh token                  |
| GET   | `/auth/profile`          | Bearer  | Get own profile (cached)                 |
| PUT   | `/auth/profile`          | Bearer  | Update name/email/photo                  |

### Key Design Decisions & Highlights

- Phone auth is intentionally password-less and “magical” – perfect for mobile-first rider/driver experience.
- First-time phone signup creates the user + wallet + rider profile in one flow.
- Email accounts are completely separate (different roles, password required).
- Refresh-token rotation + blacklisting implemented via Redis.
- Profile caching reduces DB hits on every authenticated request.
- All validation is done in DTOs (`Validate()` method) + Gin binding tags.
- Clean separation: Handler → Service → Repository (easy to test/mock).
- Wallet creation is role-aware (different types & initial balances).
- Extensible – adding driver profile creation later is just injecting `driverService` and a few lines.

### Dependencies Used

- Gin + gin-gonic binding/validation
- GORM (PostgreSQL/MySQL)
- Redis (via internal cache utils) – for profile cache & token blacklist
- Custom JWT utils
- Bcrypt password hashing
- Your shared `response`, `logger`, `config` packages

### In One Sentence

> The auth module provides secure, dual-mode (phone vs email) authentication with automatic wallet/rider-profile provisioning, JWT token management, Redis-backed caching & blacklisting — forming the identity foundation for the entire ride-hailing platform.

Ready for the next module whenever you are!