# SDD: Task 01 — Fix CLI Tool

## Summary

The CLI tool at `cmd/cli/main.go` currently calls `app.NewServer()` which fully bootstraps the HTTP server: loading config, connecting to DB, dialing RabbitMQ, setting up routes, and initializing the WebSocket hub. The CLI only needs to generate and persist an API key, which requires only a DB connection. This change extracts the DB initialization and repository setup into standalone, exported functions in the `app` package so the CLI can operate with only a DB connection — no RabbitMQ, no routes, no HTTP server. The existing full-server path remains completely unchanged.

## Files to Modify

| # | File | Action | Description |
|---|------|--------|-------------|
| 1 | `app/db.go` | modify | Add exported standalone `InitDB()` and `InitRepositories()`; refactor `connectDb()` and `initialiseRepositories()` to delegate to them |
| 2 | `app/cli.go` | modify | Replace `GenerateApiKeyHandler()` method on `*server` with standalone `GenerateApiKeyCLI()` function |
| 3 | `app/server.go` | modify | Refactor `connectDb()` and `initialiseRepositories()` to use the new standalone functions |
| 4 | `cmd/cli/main.go` | modify | Call `config.LoadConfig()`, `app.InitDB()`, and `app.GenerateApiKeyCLI()` directly instead of `app.NewServer()` |

## Current State

### 1. `app/db.go` (full file, 58 lines)

```go
package app

import (
	"fmt"
	"notice-me-server/pkg/auth"
	"notice-me-server/pkg/notification"
	"notice-me-server/pkg/repository"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func (s *server) connectDb() {
	var err error

	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", s.c.Db.User, s.c.Db.Pwd, s.c.Db.Host, s.c.Db.Port, s.c.Db.Name)

	conn, err := gorm.Open(mysql.Open(connection), &gorm.Config{})

	if err != nil {
		panic(err)
	}

	sqlDB, _ := conn.DB()

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(30)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	s.db = conn

	err = s.db.AutoMigrate(&notification.Notification{}, &auth.ApiKey{})
	if err != nil {
		panic(err)
	}

	s.db.Exec("ALTER DATABASE " + s.c.Db.Name + " character set utf8mb4 collate utf8mb4_unicode_ci;")
}

func (s *server) initialiseRepositories() {
	if s.db == nil {
		s.connectDb()
	}

	s.repositories = make(map[string]interface{})

	s.repositories[notification.RepositoryKey] = repository.NewGorm[notification.Notification](s.db)
	s.repositories[auth.RepositoryKey] = repository.NewGorm[auth.ApiKey](s.db)
}

func (s *server) getRepository(name string) interface{} {
	if r, ok := s.repositories[name]; ok {
		return r
	}

	return nil
}
```

### 2. `app/cli.go` (full file, 12 lines)

```go
package app

import (
	"notice-me-server/pkg/auth"
	"notice-me-server/pkg/repository"
)

func (s *server) GenerateApiKeyHandler() (*auth.ApiKey, error) {
	repo := s.getRepository(auth.RepositoryKey).(repository.Repository[auth.ApiKey])

	return auth.GenerateApiKey(repo)
}
```

### 3. `app/server.go` (relevant sections)

```go
type server struct {
	r            *mux.Router
	ws           hub.HubInterface
	rabbit       rabbit.RabbitInterface
	repositories map[string]interface{}
	db           *gorm.DB
	c            *config.Config
}

func NewServer() *server {
	s := server{
		c:  config.LoadConfig(),
		ws: hub.NewHub(),
		r:  mux.NewRouter(),
	}

	s.initialiseRepositories()
	s.initialiseRabbit()

	s.routes()

	return &s
}
```

### 4. `cmd/cli/main.go` (full file, 19 lines)

```go
package main

import (
	"github.com/en-vee/alog"
	"notice-me-server/app"
)

func main() {
	s := app.NewServer()

	apiKey, err := s.GenerateApiKeyHandler()

	if err != nil {
		alog.Error("something went wrong: %v", err)
		return
	}

	alog.Info("Api key generated successfully: " + apiKey.Value)
}
```

## Detailed Changes

### 1. `app/db.go`

**Changes:**
- Add exported standalone function `InitDB(cfg *config.Config) (*gorm.DB, error)`.
- Add exported standalone function `InitRepositories(db *gorm.DB) map[string]interface{}`.
- Refactor `connectDb()` to call `InitDB()` and panic on error (preserving existing behavior).
- Refactor `initialiseRepositories()` to call `InitRepositories()`.
- **Keep `getRepository()` unchanged** — it is used by many handlers and middleware.

**After (full file):**

```go
package app

import (
	"fmt"
	"notice-me-server/pkg/auth"
	"notice-me-server/pkg/notification"
	"notice-me-server/pkg/repository"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// InitDB connects to the database using the provided config, runs auto-migration,
// and returns the *gorm.DB handle. Callers must check the error.
func InitDB(cfg *config.Config) (*gorm.DB, error) {
	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Db.User, cfg.Db.Pwd, cfg.Db.Host, cfg.Db.Port, cfg.Db.Name)

	db, err := gorm.Open(mysql.Open(connection), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(30)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	if err := db.AutoMigrate(&notification.Notification{}, &auth.ApiKey{}); err != nil {
		return nil, err
	}

	db.Exec("ALTER DATABASE " + cfg.Db.Name + " character set utf8mb4 collate utf8mb4_unicode_ci;")

	return db, nil
}

// InitRepositories creates the standard repository map (auth + notification)
// backed by the provided *gorm.DB. This is the lightweight initialization path
// for tools (like the CLI) that do not need the full HTTP/RabbitMQ stack.
func InitRepositories(db *gorm.DB) map[string]interface{} {
	repos := make(map[string]interface{})
	repos[notification.RepositoryKey] = repository.NewGorm[notification.Notification](db)
	repos[auth.RepositoryKey] = repository.NewGorm[auth.ApiKey](db)
	return repos
}

func (s *server) connectDb() {
	db, err := InitDB(s.c)
	if err != nil {
		panic(err)
	}
	s.db = db
}

func (s *server) initialiseRepositories() {
	if s.db == nil {
		s.connectDb()
	}
	s.repositories = InitRepositories(s.db)
}

func (s *server) getRepository(name string) interface{} {
	if r, ok := s.repositories[name]; ok {
		return r
	}
	return nil
}
```

**Key details:**
- `InitDB()` signature: `func InitDB(cfg *config.Config) (*gorm.DB, error)` — takes config, returns DB handle or error.
- `InitDB()` uses the **same connection string format** as before.
- `InitDB()` sets **identical pool settings** (MaxIdleConns=10, MaxOpenConns=30, ConnMaxLifetime=5min).
- `InitDB()` runs the **same auto-migration** (`notification.Notification`, `auth.ApiKey`).
- `InitRepositories()` signature: `func InitRepositories(db *gorm.DB) map[string]interface{}` — takes DB, returns repository map.
- The method `connectDb()` now delegates to `InitDB()` and panics on error (preserving the original panic behavior for the HTTP server path).
- The method `initialiseRepositories()` now delegates to `InitRepositories()`.

### 2. `app/cli.go`

**Changes:**
- Remove the `GenerateApiKeyHandler()` method from `*server`.
- Add exported standalone function `GenerateApiKeyCLI(db *gorm.DB) (*auth.ApiKey, error)`.
- New import needed: no new imports — `gorm` is not directly imported; `repository` is already imported.

**After (full file):**

```go
package app

import (
	"notice-me-server/pkg/auth"
	"notice-me-server/pkg/repository"

	"gorm.io/gorm"
)

// GenerateApiKeyCLI generates an API key using only a *gorm.DB connection.
// It creates an auth repository, generates a key, persists it, and returns it.
// This function does not require RabbitMQ, HTTP routes, or any other server infrastructure.
func GenerateApiKeyCLI(db *gorm.DB) (*auth.ApiKey, error) {
	repo := repository.NewGorm[auth.ApiKey](db)
	return auth.GenerateApiKey(repo)
}
```

**Key details:**
- `GenerateApiKeyCLI()` signature: `func GenerateApiKeyCLI(db *gorm.DB) (*auth.ApiKey, error)`.
- Uses `repository.NewGorm[auth.ApiKey](db)` directly (no type assertion needed).
- New import: `"gorm.io/gorm"` for the `*gorm.DB` parameter type.

### 3. `app/server.go`

**Changes:**
- No structural changes needed. The `connectDb()` and `initialiseRepositories()` methods now delegate to the standalone functions.
- `NewServer()` continues to work exactly as before: `c: config.LoadConfig()`, `ws: hub.NewHub()`, `r: mux.NewRouter()`, then `s.initialiseRepositories()` (which now calls `InitDB` + `InitRepositories` internally), then `s.initialiseRabbit()`, then `s.routes()`.

**After (full file with changes highlighted):**

```go
package app

import (
	"encoding/json"
	"github.com/en-vee/alog"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
	"net/http"
	"notice-me-server/pkg/config"
	"notice-me-server/pkg/hub"
	"notice-me-server/pkg/rabbit"
)

type server struct {
	r            *mux.Router
	ws           hub.HubInterface
	rabbit       rabbit.RabbitInterface
	repositories map[string]interface{}
	db           *gorm.DB
	c            *config.Config
}

func NewServer() *server {
	s := server{
		c:  config.LoadConfig(),
		ws: hub.NewHub(),
		r:  mux.NewRouter(),
	}

	s.initialiseRepositories()
	s.initialiseRabbit()

	s.routes()

	return &s
}

func (s *server) Run() error {

	defer func() {
		err := s.rabbit.Close()
		if err != nil {
			alog.Error("Error closing amqp connection: " + err.Error())
		}
	}()

	defer func() {
		dbInstance, _ := s.db.DB()
		err := dbInstance.Close()
		if err != nil {
			alog.Error("Error closing sql connection: " + err.Error())
		}
	}()

	go func(websocket hub.HubInterface) {
		websocket.Run()
	}(s.ws)

	s.startConsumers()

	headersOk := handlers.AllowedHeaders([]string{"Accept", "Accept-Language", "Content-Type", "Content-Language", "Origin"})
	originsOk := handlers.AllowedOrigins(s.c.Server.Cors)
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS", "DELETE"})

	handler := handlers.CORS(headersOk, originsOk, methodsOk)(handlers.RecoveryHandler()(s.r))

	log := newServerErrorLog()

	server := &http.Server{
		Addr:     ":" + s.c.Server.Port,
		ErrorLog: log,
		Handler:  handler,
	}

	if s.c.Server.Env == "production" {
		return server.ListenAndServeTLS(s.c.Server.TlsCert, s.c.Server.TlsKey)
	} else {
		return server.ListenAndServe()
	}
}

func (s *server) writeResponse(w http.ResponseWriter, response interface{}) {
	if response == nil {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	byteResponse, _ := json.Marshal(response)
	_, _ = w.Write(byteResponse)
}

func (s *server) writeErrorResponse(w http.ResponseWriter, err error, errorCode int) {
	response := make(map[string]interface{})

	response["error"] = err.Error()
	w.WriteHeader(errorCode)
	byteResponse, _ := json.Marshal(response)
	_, _ = w.Write(byteResponse)
}
```

**Note:** `server.go` itself requires **no edits** to its code lines — the behavior change is entirely in the methods defined in `app/db.go` (`connectDb`, `initialiseRepositories`) which already existed as methods on `*server`. No `server.go` lines change; it is listed only for context.

### 4. `cmd/cli/main.go`

**Changes:**
- Remove `app.NewServer()` call.
- Call `config.LoadConfig()` directly.
- Call `app.InitDB(cfg)` to get a DB handle (with error handling).
- Call `app.GenerateApiKeyCLI(db)` to generate and persist the key.
- Print the key value to stdout using `fmt.Println` (instead of `alog.Info` for a clean, machine-parseable output).
- Add import for `"notice-me-server/pkg/config"`.

**After (full file):**

```go
package main

import (
	"fmt"
	"notice-me-server/app"
	"notice-me-server/pkg/config"
)

func main() {
	cfg := config.LoadConfig()

	db, err := app.InitDB(cfg)
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		return
	}

	apiKey, err := app.GenerateApiKeyCLI(db)
	if err != nil {
		fmt.Printf("Error generating API key: %v\n", err)
		return
	}

	fmt.Println(apiKey.Value)
}
```

**Key details:**
- New imports: `"fmt"` (replaces `"github.com/en-vee/alog"`) and `"notice-me-server/pkg/config"`.
- Three-step flow: load config → init DB → generate key.
- Print only the key value to stdout (pipeline-friendly).
- Error messages go to stdout with descriptive prefixes.
- No `app.NewServer()`, no RabbitMQ, no routes, no HTTP server.

## Data Flow

```
cmd/cli/main.go
│
├─ 1. config.LoadConfig()
│      Reads embedded config.yaml → *config.Config
│      Config contains: db user/pwd/host/port/name
│
├─ 2. app.InitDB(cfg)
│      Opens GORM connection to MariaDB using mysql driver
│      Sets connection pool params (10 idle, 30 open, 5min lifetime)
│      Runs AutoMigrate for Notification + ApiKey tables
│      Runs ALTER DATABASE charset/collation
│      Returns *gorm.DB (or error if unreachable)
│
├─ 3. app.GenerateApiKeyCLI(db)
│      ├─ Creates repo: repository.NewGorm[auth.ApiKey](db)
│      ├─ Calls auth.GenerateApiKey(repo)
│      │     ├─ Creates new ApiKey entity (uuid ID, uuid Value, RequestCount=0)
│      │     ├─ repo.Create(ak) → INSERT into api_keys table
│      │     └─ Returns *ApiKey with plaintext Value
│      └─ Returns *auth.ApiKey (or error)
│
└─ 4. fmt.Println(apiKey.Value)
       Prints the plaintext UUID key to stdout
```

**Contrast with the HTTP server flow (unchanged):**
```
NewServer()
├─ config.LoadConfig()
├─ hub.NewHub()
├─ mux.NewRouter()
├─ s.initialiseRepositories()
│     ├─ s.connectDb() → InitDB(s.c) (same function, but panics on error)
│     └─ s.InitRepositories(s.db)
├─ s.initialiseRabbit()  ← NOT called in CLI path
├─ s.routes()            ← NOT called in CLI path
└─ return &server{...}
```

## Test Plan

### Existing tests to update
None. The existing `app/handler_test.go` creates a `server` struct directly via `initialiseMocks()` without calling `NewServer()`, so changes to `NewServer()`, `connectDb()`, `initialiseRepositories()`, and `db.go` do not affect existing tests.

### New test cases to add in `app/cli_test.go`

| # | Test Case | Input | Expected Output |
|---|-----------|-------|-----------------|
| 1 | `GenerateApiKeyCLI_creates_and_returns_key` | A valid `*gorm.DB` backed by an in-memory SQLite or a repository mock (if we create a test helper) | A non-nil `*auth.ApiKey` with a non-empty `Value`, persisted to DB |
| 2 | `GenerateApiKeyCLI_returns_error_on_nil_db` | `nil` for `*gorm.DB` | An error (the gorm repository will fail on Create) |

**Note on testability:** The repository layer uses GORM which requires a real or mock DB. The existing test pattern in `handler_test.go` uses `repo_mock.NewRepository[notification.Notification]()` from `pkg/repository/mock/`. For the CLI, we could similarly test `auth.GenerateApiKey` directly with a mock repository, since `GenerateApiKeyCLI` is a thin wrapper that creates a GORM repo and delegates. The acceptance criteria can be verified via integration test with a real DB, but unit testing the function is low-value since the logic lives in `pkg/auth/service.go`.

### Manual verification
1. Run `go run ./cmd/cli/main.go` with a running MariaDB and no RabbitMQ — should print a UUID.
2. Run `go run ./cmd/cli/main.go` with MariaDB unreachable — should print a clear connection error.
3. Run `go run ./cmd/server/main.go` — should still start the full HTTP server with RabbitMQ.

## Edge Cases & Failure Modes

### DB unreachable
- `InitDB()` returns an error from `gorm.Open()`. The CLI prints `"Error connecting to database: ..."` and exits with code 0 (Go default for `main()` returning). Error is descriptive (e.g., "dial tcp 127.0.0.1:3306: connect: connection refused").

### Config file missing/malformed
- `config.LoadConfig()` calls `log.Fatalln(err)` which terminates the process. This is inherited behavior and unchanged. The CLI will not proceed without valid config.

### DB connection lost after InitDB but before Create
- `auth.GenerateApiKey()` will receive an error from `repo.Create()`. The CLI prints `"Error generating API key: ..."` and exits. The error message comes from GORM, which is typically descriptive.

### Duplicate key (UUID collision, astronomically unlikely)
- `repo.Create()` would return a duplicate entry error. The CLI prints the error. No special handling needed.

### Concurrent access
- The CLI is a single-shot tool — it generates one key and exits. No concurrent access concerns.

### Invalid DB credentials in config
- `InitDB()` returns an error from `gorm.Open()` (e.g., "access denied for user ..."). The CLI prints the error and exits.

### Migration failure
- If the `ApiKey` or `Notification` table already exists with incompatible schema, `AutoMigrate` returns an error. `InitDB()` propagates it. The CLI prints the error and exits.

### Database name does not exist
- The connection string includes the database name. If it doesn't exist, `gorm.Open` may fail depending on driver. If it succeeds but the `ALTER DATABASE` Exec fails, `InitDB()` will **not** return an error for the Exec (the current code ignores the error from `s.db.Exec(...)`). The CLI will still work because the key table was created by `AutoMigrate`. This preserves existing behavior.

### CLI compiled without CGO (for SQLite-like scenarios)
- Not applicable — the project uses `gorm.io/driver/mysql` which is a pure Go driver (no CGO needed).
