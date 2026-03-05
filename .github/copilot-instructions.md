# GitHub Copilot instructions for _qris-optimizer_

This repository is intentionally small.  It consists of a single Go HTTP service that
connects to a local PostgreSQL database and two k6 performance tests.  There are no
frameworks, no build tools besides the Go toolchain, and no CI configuration.

Use these notes to help an AI coding agent get up to speed quickly.

---
## 📦 Big‑picture architecture

- **Binary**: `cmd/main.go` is the only Go source file and contains `package main`.
  All business logic, HTTP handlers, and database access live here.
- **Database**: a PostgreSQL 15 container defined in `docker-compose.yml`.
  Connection string is hard‑coded in `main()` and uses `github.com/lib/pq`.
  A global `var db *sql.DB` is used by every handler.
- **HTTP server** listens on `:8080` and registers handlers via
  `http.HandleFunc`.

### Endpoints

| Path               | Purpose                                             | Notes                              |
|--------------------|-----------------------------------------------------|------------------------------------|
| `/`                | health/heartbeat                                    | prints "QRIS Optimizer Running"   |
| `/qris/inquiry`    | SELECT * from `merchants` and dump rows to response | no pagination, fields hard‑coded   |
| `/qris/payment`    | INSERT a row to `transactions` (hard‑coded values)  | merchantID/amount/status are static |

- The handlers are simple; they perform a query/exec, handle errors with
  `http.Error`, and write plain‑text output.
- There are commented `time.Sleep` lines that were used to simulate latency.

### Data flow

1. HTTP request → handler function.
2. Handler uses global `db` to query/exec raw SQL.
3. Results are formatted with `fmt.Fprintf` and sent back to client.

There is no domain layer, models, or middleware in this project.

---
## 🛠 Developer workflows

1. **Start dependencies**
   ```sh
   docker-compose up -d           # runs postgres on localhost:5432
   ```
2. **Prepare database schema** (not versioned anywhere in repo):
   ```sql
   CREATE TABLE merchants (
       id serial PRIMARY KEY,
       name text,
       city text,
       status text
   );

   CREATE TABLE transactions (
       id serial PRIMARY KEY,
       merchant_id int REFERENCES merchants(id),
       amount int,
       status text
   );
   ```
   Agents should assume tables exist when adding queries.
3. **Build/Run the service**
   ```sh
   go run ./cmd                # or go build -o qris ./cmd && ./qris
   ```
4. **Exercise endpoints**
   ```sh
   curl http://localhost:8080/            # health check
   curl http://localhost:8080/qris/inquiry
   curl http://localhost:8080/qris/payment
   ```
5. **Load testing**
   ```sh
   k6 run k6/payment_test.js              # hits /qris/payment with 150 VUs
   ```
6. **Formatting / linting**
   - `go fmt`, `go vet` manually if needed.  No automated tooling is defined.

> ⚠️ There are no Go unit tests in the repo; adding a `_test.go` file and
> manually creating a temporary database (or mocking `db`) is up to the
> developer/agent.

---
## 📐 Project conventions & patterns

- **Single package**: everything is in `main`.  If additional packages are
  created, import them from `cmd/main.go` and keep the code flat.
- **Global `db` variable**: used by handlers; connection pooling parameters
  are set in `main()`.
- **Error handling**: checked immediately and returned via `http.Error`.
- **Configuration** is hard‑coded.  Agents expanding the project should
  convert the connection string (and other constants) to use `os.Getenv`
  or a simple config struct.
- **Dependencies**: only `github.com/lib/pq` is required; new third‑party
  libraries should be rare and justified.
- **New endpoints**: define a handler function with signature
  `func(w http.ResponseWriter, r *http.Request)` and register it in `main()`.
  Follow the style of existing handlers (immediate query, error check,
  fmt.Fprintf).

---
## 🔗 Integration points & external services

- PostgreSQL is the sole external service.  The connection string ("localhost
  port=5432 user=qris …") appears in `main.go`.
- `k6` is used only for the `payment_test.js` script; nothing else depends on
  it.
- There are no cloud, message queue, or third‑party API integrations.

---
## 🧠 Notes for AI agents

1. **Start by reading `cmd/main.go`**; it contains nearly the entire codebase.
2. **Database logic is primitive**; SQL is written inline.  If modifying or
   optimizing queries, update both the code and any documentation or tests.
3. **Keep changes minimal** unless the user asks for architecture overhaul.
4. **Follow existing styling and error‑handling patterns** (no panics, no
   logging library).  Use `fmt.Println`/`log.Fatal` only as shown.
5. **Use docker-compose** in examples to explain how to bring up Postgres.
6. **Mention the k6 file when discussing performance**; it's the canonical
   load‑test for `/qris/payment`.
7. **When adding features that require a new table**, note that no migration
   tooling exists; the user must create the table manually or via SQL script.

---
*This document was generated/updated by Copilot.  Feel free to suggest edits
if parts of the workflow are unclear or incomplete.*
