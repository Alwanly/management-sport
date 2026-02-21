# Management Club Sport Application

**Branch:** `feature/management-club-sport`
**Description:** Migrate from Fiber to Gin framework and implement complete football team management system with teams, players, matches, goals, and reporting capabilities.

## Goal
Migrate the web framework from Fiber to Gin, then build a comprehensive management system for XYZ company's football club that enables admins to manage teams, players, match schedules, goal tracking, and generate match result reports with statistics. The system enforces business rules like unique shirt numbers per team, player-team validations, and maintains audit logs for all changes.

## Implementation Steps

### Step 0: Migrate from Fiber to Gin Framework
**Files:**
- `go.mod` (replace fiber with gin dependencies)
- `pkg/deps/App.go` (change Fiber to Gin field)
- `cmd/main/bootstrap.go` (initialize Gin instead of Fiber)
- `cmd/main/main.go` (update server lifecycle)
- `pkg/binding/binding.go` (change from fiber.Ctx to gin.Context)
- `pkg/binding/binding_test.go` (update tests for Gin)
- `pkg/middleware/authentication.go` (change to Gin middleware pattern)
- `pkg/middleware/recovery.go` (change to Gin recovery middleware)
- `internal/example/handler/handler.go` (update handlers to use gin.Context)
- `api/docs.go` (update for ginSwagger)

**What:** Migrate the web framework from Fiber to Gin. Replace `github.com/gofiber/fiber/v2` with `github.com/gin-gonic/gin` in go.mod. Update App struct to use `*gin.Engine` instead of `*fiber.App`. Rewrite binding package to work with `*gin.Context` instead of `*fiber.Ctx`. Update middleware to use Gin's middleware signature `func(*gin.Context)`. Change handler signatures from `func(*fiber.Ctx) error` to `func(*gin.Context)`. Update routing from Fiber's `Group()` pattern to Gin's `Group()`. Replace Fiber response methods (`c.Status().JSON()`) with Gin equivalents (`c.JSON(statusCode, data)`). Update swagger integration from gofiber/swagger to swaggo/gin-swagger. Maintain existing business logic unchanged.

**Testing:** Run existing example endpoints (books CRUD), verify all routes work with Gin, test JWT authentication middleware, validate request binding and validation still work correctly, ensure swagger documentation accessible, run test suite and fix any failures, verify error handling and response formatting consistent with previous behavior.

---

### Step 1: Database Schema & Models
**Files:** 
- `db/schema/codebase.hcl` (add Team, Player, Match, Goal, AuditLog tables)
- `model/team.go` (new)
- `model/player.go` (new)
- `model/match.go` (new)
- `model/goal.go` (new)
- `model/audit_log.go` (new)

**What:** Create database schema with Atlas HCL definitions and corresponding GORM models for all entities. Include proper constraints: UNIQUE(team_id, shirt_number) on Player, CHECK(home_team_id <> away_team_id) on Match, and foreign key relationships. Models follow existing pattern with UUID primary keys, timestamps, and audit fields (created_by, updated_by). Add `deleted_at` column to Team and Player models for soft delete. Enhance User model with role field (enum: admin, user). Store timestamps as timestamptz (UTC).

**Testing:** Run `make new` to generate migration, `make up` to apply, verify tables created with `psql` and constraints are enforced. Test GORM auto-migration in bootstrap. Verify soft delete column and role field are present.

---

### Step 2: Role-Based Access Control (RBAC)
**Files:**
- `pkg/authentication/jwt.go` (add role to claims)
- `pkg/middleware/authentication.go` (add role extraction, add AdminAuth middleware)
- `pkg/middleware/type.go` (add Role to AuthUserData)
- `model/user.go` (verify role field exists from Step 1)
- `cmd/main/bootstrap.go` (update middleware initialization if needed)

**What:** Extend JWT authentication with role-based authorization. Add role field to JWT claims structure. Modify JwtAuth middleware to extract role and store in AuthUserData (using Gin context with c.Set/c.Get). Create new middleware AdminAuth() that checks for admin role, aborts with 403 Forbidden if role is not admin (using `c.AbortWithStatusJSON()`). Update token generation to include role from User model. Define role constants (RoleAdmin, RoleUser). Middleware follows Gin pattern: `func(*gin.Context)`. This middleware will be used in subsequent steps to protect write operations.

**Testing:** Generate JWT tokens with different roles (admin, user), verify AdminAuth middleware allows admin role, rejects user role with 403, test that regular JwtAuth still works for authentication-only endpoints, verify role is available in handlers via binding.

---

### Step 3: Team Management API
**Files:**
- `internal/team/handler/handler.go` (new)
- `internal/team/usecase/usecase.go` (new)
- `internal/team/repository/repository.go` (new)
- `internal/team/schema/request.go` (new)
- `internal/team/schema/response.go` (new)
- `internal/team/schema/constant.go` (new)
- `cmd/main/bootstrap.go` (register team handler)

**What:** Implement complete CRUD operations for teams following the existing example pattern migrated to Gin. Endpoints: GET /teams (list), POST /teams (create), GET /teams/:id (detail - note Gin uses `:id` for params), PUT /teams/:id (update), DELETE /teams/:id (soft delete). Include pagination for list. Use JWT authentication (d.Auth.JwtAuth()) for read operations, admin middleware (d.Auth.AdminAuth()) for write operations. Handler signature: `func(*gin.Context)`, use `c.JSON(statusCode, data)` for responses. Validate all inputs (name required, founded_year must be valid year, city/address format). List endpoint filters out soft-deleted teams by default. Implement repository method to query excluding deleted_at IS NOT NULL.

**Testing:** Test with Swagger UI or Postman: create teams (admin only), list with pagination (any authenticated user, verify deleted teams excluded), get team detail, update team info (admin only), soft delete team (admin only, verify deleted_at set, team disappears from list), test validation errors, verify non-admin users can read but cannot write.

---

### Step 4: Player Management API with Validations
**Files:**
- `internal/player/handler/handler.go` (new)
- `internal/player/usecase/usecase.go` (new)
- `internal/player/repository/repository.go` (new)
- `internal/player/schema/request.go` (new)
- `internal/player/schema/response.go` (new)
- `internal/player/schema/constant.go` (new)
- `cmd/main/bootstrap.go` (register player handler)

**What:** Implement player CRUD with business rule enforcement. Endpoints: GET /players?team_id={id} (list with query param), POST /players (create), GET /players/:id (detail), PUT /players/:id (update), DELETE /players/:id (soft delete). Enforce unique shirt_number per team (considering only non-deleted players), validate position enum values, validate physical stats (height_cm > 0, weight_kg > 0). Support team transfers via update team_id. Use d.Auth.JwtAuth() for reads, d.Auth.AdminAuth() for writes. List excludes soft-deleted players. Use Gin's `c.Query()` for filtering and `c.Param()` for route parameters.

**Testing:** Create players for different teams (admin only), list players (any authenticated user), verify shirt number uniqueness per team (should fail when duplicate within same team, succeed for different teams), test team filter, update player to transfer teams (admin only, verify historical goals preserved), get player detail, validate all fields, soft delete player (admin only), verify admin role requirement on all write operations.

---

### Step 5: Match Scheduling & Status Management
**Files:**
- `internal/match/handler/handler.go` (new)
- `internal/match/usecase/usecase.go` (new)
- `internal/match/repository/repository.go` (new)
- `internal/match/schema/request.go` (new)
- `internal/match/schema/response.go` (new)
- `internal/match/schema/constant.go` (new - status enum: scheduled, ongoing, finished)
- `cmd/main/bootstrap.go` (register match handler)

**What:** Implement match scheduling and lifecycle management. Endpoints: GET /matches (list), POST /matches (create schedule - admin only), PUT /matches/:id (update schedule including re-opening finished matches - admin only), PATCH /matches/:id/start (start match), PATCH /matches/:id/end (finish match with scores). Validate home_team_id ≠ away_team_id, both teams exist and not soft-deleted, match_date is valid datetime (UTC), status transitions (scheduled → ongoing → finished). Allow editing finished matches for corrections. Store match_date and match_time as timestamptz in UTC. Use Gin routing with `:id` parameter.

**Testing:** Create match schedules, verify team validation (same team rejected, soft-deleted teams rejected), test status transitions, update schedules including re-opening finished matches, verify audit logging when correcting finished match scores, test timezone handling (store UTC), verify admin role requirement.

---

### Step 6: Goal Recording & Score Validation
**Files:**
- `internal/goal/handler/handler.go` (new)
- `internal/goal/usecase/usecase.go` (new)
- `internal/goal/repository/repository.go` (new)
- `internal/goal/schema/request.go` (new)
- `internal/goal/schema/response.go` (new)
- `internal/match/usecase/usecase.go` (update - add end match with goals logic)

**What:** Implement goal recording with cross-entity validation. Endpoint: POST /matches/:id/goals (record goal with player_id and minute_scored). Validate: player exists, player's team_id matches one of the match teams (home or away), minute_scored > 0 and reasonable (< 150), match status is ongoing or finished. When ending match via PATCH /matches/:id/end, accept array of goals and validate sum equals home_score + away_score. Use `c.Param("id")` to extract match ID.

**Testing:** Record goals during match, verify player must be from participating teams (reject if wrong team), test goal minute validation, end match with goals payload, verify score consistency (sum of goals per team = final score), test rejection of mismatched scores.

---

### Step 7: Match Results Reporting
**Files:**
- `internal/report/handler/handler.go` (new)
- `internal/report/usecase/usecase.go` (new)
- `internal/report/repository/repository.go` (new)
- `internal/report/schema/response.go` (new)
- `cmd/main/bootstrap.go` (register report handler)

**What:** Implement comprehensive match results report. Endpoint: GET /report/match-results?from_date={date}&to_date={date}&team_id={id} (optional filters). Returns: match details (date in UTC, time, teams, scores), status (Home Win/Away Win/Draw), top scorer per match (player with most goals), cumulative stats within the filtered date range (total home wins, away wins, draws up to each match in the range). Use complex SQL joins and aggregations for efficiency. If no date filter, show all matches. Cumulative calculations reset at start of filtered range.

**Testing:** Generate reports with various filters, verify status calculation (home_score vs away_score), confirm top scorers are correct (handle ties - return first by goal time), validate cumulative statistics accuracy within date ranges, test with no filters (all matches), verify reporting includes corrected match data (from re-opened matches), test performance with large datasets, verify empty result handling.

---

### Step 8: Audit Logging System
**Files:**
- `pkg/audit/audit.go` (new - audit logging service)
- `pkg/middleware/audit.go` (new - audit middleware)
- `internal/team/usecase/usecase.go` (add audit logging)
- `internal/player/usecase/usecase.go` (add audit logging)
- `internal/match/usecase/usecase.go` (add audit logging)
- `internal/goal/usecase/usecase.go` (add audit logging)

**What:** Implement audit logging for all create/update/delete operations. Track: entity type, entity ID, action (create/update/delete), old values (JSON), new values (JSON), timestamp, user ID from JWT, user role. Create audit service with methods LogCreate(), LogUpdate(), LogDelete(). Integrate into all domain usecases for data-changing operations. Add GET /audit-logs endpoint for admins to view audit trail. Implement scheduled cleanup job (cron or background worker) to delete audit logs older than configurable retention period (default 90 days from env AUDIT_RETENTION_DAYS).

**Testing:** Perform CRUD operations on all entities, verify audit logs are created with correct data, check old/new value JSON serialization, confirm user ID and role are captured from JWT, test audit log retrieval endpoint (verify admin-only access), test cleanup job with past dates, verify performance impact is minimal, test match score correction audit trail.

---

## Technical Considerations

**Architecture:**
- Follow existing clean architecture pattern (Handler → UseCase → Repository → Model)
- Use Gin framework routing with JWT middleware protection
- Apply binding and validation patterns from pkg/binding and pkg/validator (migrated to Gin in Step 0)
- Leverage pkg/wrapper for consistent response formats

**Database:**
- Use Atlas for migrations with HCL schema definitions
- Implement proper indexes for foreign keys and frequently queried fields
- Use GORM transactions for operations affecting multiple entities (e.g., ending match with goals)
- Consider soft deletes for teams/players (add deleted_at column) if business needs restoration

**Validation Layers:**
- Database constraints (PRIMARY KEY, FOREIGN KEY, UNIQUE, CHECK)
- GORM model validations (validate struct tags)
- UseCase business rule validations (cross-entity checks)
- API input validations (pkg/validator with go-playground/validator)

**Error Handling:**
- Use pkg/contract status codes consistently
- Return descriptive error messages via pkg/wrapper.ResponseFailed()
- Handle constraint violations gracefully (e.g., duplicate shirt number)
- Log errors with structured logging (pkg/logger)

**Testing Strategy:**
- Unit tests for usecase business logic (mock repositories)
- Integration tests for repositories (test database)
- API endpoint tests via test_spec.go pattern
- Validate constraints at database level

**Performance:**
- Index foreign keys and commonly filtered fields (team_id, match_id, match_date)
- Use pagination for all list endpoints
- Optimize report queries with proper JOINs and aggregations
- Consider Redis caching for frequently accessed reports

**Security:**
- All endpoints protected with JWT middleware (d.Auth.JwtAuth() - Gin middleware)
- Extract user ID and role from JWT claims for authorization
- Implement admin role middleware for write operations (Create/Update/Delete)
- Add role field to AuthUserData struct in middleware
- Read-only operations (GET) available to all authenticated users
- Validate authorization in usecase layer as well (defense in depth)
- Sanitize inputs to prevent SQL injection (GORM parameterization)
- Use `c.AbortWithStatusJSON()` to halt request processing on auth failures

**Documentation:**
- Add Swagger annotations to all handlers
- Run `make docs` to regenerate swagger files after each step
- Document business rules in code comments
- Update README.md with new endpoints

---

## Design Decisions

1. **Team Deletion:** Implement soft delete - add `deleted_at` column to teams table. Deleted teams are marked but data is preserved for historical integrity. Queries filter out soft-deleted teams by default.

2. **Player Transfers:** Allow player transfers (update team_id) at any time. Historical goals remain associated with the player and reflect in their goal statistics regardless of current team.

3. **Match Date/Time Timezone:** Store all timestamps as UTC in the database using `timestamptz` type. Client applications handle timezone conversion for display purposes.

4. **Goal Timing:** Use simple integer field for minute_scored (e.g., 93 for stoppage time). No separate tracking of regulation vs injury time.

5. **Report Date Range:** Cumulative statistics calculated only within the filtered date range. If no date filter applied, calculate across all matches.

6. **Authentication vs Authorization:** Implement admin role validation. Add role field to User model and JWT claims. Check for admin role in middleware for write operations (Create/Update/Delete).

7. **Audit Log Retention:** Implement automatic cleanup - add scheduled job to delete audit logs older than configurable period (default: 90 days). Make retention period configurable via environment variable.

8. **Match Re-scheduling:** Allow finished matches to be edited (re-opened) for score corrections. All edits are tracked in audit log with old/new values for accountability.

9. **Framework Choice:** Migrate from Fiber to Gin framework (Step 0) to align with industry standard patterns and broader ecosystem support. Gin provides better performance characteristics and more extensive middleware ecosystem while maintaining simplicity.
