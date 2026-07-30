# Implementation & Architecture Notes

## 1. The Seed Script Bug

- **Issue & Discovery:** During initial verification of `cmd/seed/main.go`, we analyzed transaction handling, database schema constraints (`task_status` PostgreSQL enum type: `pending`, `in_progress`, `done`), and data seeding logic.
- **Resolution & Idempotency:** The reset function executes `TRUNCATE task_tags, tasks, users RESTART IDENTITY CASCADE` prior to inserting seed data. This guarantees that running `go run ./cmd/seed` multiple times generates clean, deterministic ID sequences without primary key or foreign key constraint collisions.

---

## 2. Library & Structure Choices

- **GraphQL Library (`gqlgen`):** 
  - Chosen because `gqlgen` provides strong typing in Go generated directly from GraphQL schema definitions (`schema.graphqls`).
  - Auto-generation of model structs and resolver interfaces keeps schema definitions single-source-of-truth and avoids runtime reflection overhead.
- **Package Layout:**
  - `cmd/server/main.go`: Application entry point setting up HTTP handler, CORS middleware, and GraphQL playground/query endpoints.
  - `cmd/seed/main.go`: Database seeding CLI utility.
  - `internal/repository/db.go`: Data layer encapsulating raw SQL queries against PostgreSQL.
  - `graph/`: GraphQL schema files, generated executable schema, and resolver implementations (`schema.resolvers.go`).

---

## 3. Exposed Fields & Frontend Integration

- **Selected Fields:** Exposed `dueDate` and `description` end-to-end.
- **Rationale:** `dueDate` and `description` provide high utility to task management users. 
- **Implementation:**
  1. Updated `graph/schema.graphqls` and generated Go models (`Task.dueDate`, `Task.description`).
  2. Implemented nested field resolvers for `User.tasks` and `Task.user` in `graph/schema.resolvers.go`.
  3. Integrated GraphQL query and HTML rendering in `web/app.js` with formatted status pills, due dates, and task descriptions.

---

## 4. Tradeoffs & Future Work

If allocated additional time, the following improvements would be prioritized:
1. **DataLoader Integration:** Implement `dataloader` batching to optimize N+1 queries when resolving `User.tasks` across multiple users.
2. **Structured Logging:** Add standard `log/slog` structured JSON logger to record API request latency and error tracebacks.
3. **Automated Testing:** Add suite of unit/integration tests for `repository` methods and GraphQL resolvers using `testcontainers-go`.
4. **Containerized API Service:** Extend `docker-compose.yml` to include a multi-stage Docker build for `cmd/server`.

---

## 5. How to Run Everything

From a fresh clone, run the following commands in order:

1. **Start Postgres Container:**
   ```bash
   docker compose up -d
   ```

2. **Seed the Database:**
   ```bash
   go run ./cmd/seed
   ```

3. **Start the GraphQL Server:**
   ```bash
   go run ./cmd/server
   ```
   *The server listens on `http://localhost:8080/query` (with GraphQL Playground available at `http://localhost:8080/`).*

4. **Serve and View Web UI:**
   ```bash
   python3 -m http.server 8081 --directory web
   ```
   *Open `http://localhost:8081` in any web browser to view the Todo board.*

---

Reviewed by: MK
