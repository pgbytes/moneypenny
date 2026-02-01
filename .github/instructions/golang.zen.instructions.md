---
description: "Enforced Go coding conventions and standards for AI suggestions"
applyTo: "**/*.go"
version: "0.0.1"
---

# Go Coding Conventions and Standards

**IMPORTANT**: These are **mandatory coding conventions** for this project. When writing, reviewing, or suggesting Go code, you MUST follow these principles. They are not optional guidelines—they define the expected code quality and style for this codebase.

## How to Use These Instructions

1. **Always validate** your code suggestions against these principles before presenting them
2. **Explain deviations** if you must break a rule (with justification)
3. **Reference specific principles** when reviewing code or explaining suggestions
4. **Prioritize** these conventions over personal preferences or alternative patterns
5. **Enforce** these standards in all code generation, refactoring, and review activities

## Core Go Principles (Based on Go Proverbs and The Zen of Go)

### Simplicity and Clarity

* **Clear is better than clever** - Write code that is easy to understand over code that is impressive
* **Gofmt's style is no one's favorite, yet gofmt is everyone's favorite** - Use standard Go formatting consistently
* **Moderation is a virtue** - Don't overuse Go's features just because they exist
* **Prefer simpler approaches over clever ones** - Choose clarity and maintainability over sophistication

**Do**:
````go
// GOOD: This function is explicit and easy to read.
// The logic flows sequentially, and the intent of each step is obvious.
func GetConfigValue(m map[string]string, key, defaultValue string) string {
	val, ok := m[key]
	if!ok {
		// The key does not exist.
		return defaultValue
	}

	if val == "" {
		// The key exists but the value is empty.
		return defaultValue
	}

	return val
}
````

**Don't**:
````go
// BAD: This function uses a comma-ok assertion inside a map lookup
// and combines it with a boolean check in a single, dense line.
// It's functional but requires the reader to mentally parse multiple operations.
func GetConfigValue(m map[string]string, key, defaultValue string) string {
	if val, ok := m[key]; ok && val!= "" {
		return val
	}
	return defaultValue
}
````

### Package Design and Organization

* **A good package starts with a good name** - Package names should be nouns that describe what they provide
* **Every Go package should have a single purpose** - Focus on clear, focused responsibilities
* **Replace components rather than enhance them** - When requirements change, consider replacement over modification
* **Design the architecture, name the components, document the details** - Focus on high-level design and clear naming
* **Avoid using generic package** names such as utils or commons - Packages should have clear defined purpose

#### Avoid using generic package names such as utils or commons - Examples
**Do**:
````bash
myproject/
├── cmd/
│   └── server/
│       └── main.go
└── internal/
    ├── user/
    │   ├── user.go
    │   └── email.go      // Contains validateEmail (unexported)
    ├── order/
    │   └── order.go
    ├── server/
    │   └── jsonutil.go   // Contains ParseJSON, WriteJSON
    └── platform/
        └── database/
            └── postgres.go // Contains ConnectToDB
````

In this structure, the utils package has low cohesion. A function for validating emails is a piece of domain logic related to users, while a function for connecting to a database is a piece of infrastructure logic. Grouping them together provides no meaningful architectural boundary and creates a magnet for future unrelated code

**Don't**:
````bash
myproject/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── user/
│   │   └── user.go
│   └── order/
│       └── order.go
└── pkg/
    └── utils/
        └── utils.go // Contains ParseJSON, ConnectToDB, ValidateEmail
````

In this improved structure:
- validateEmail is an unexported helper within the user package, where it belongs.
- ParseJSON is part of a server package, as it relates to handling HTTP requests.
- ConnectToDB is part of a dedicated database package under a platform directory, clearly separating application logic from infrastructure concerns.
-This approach creates small, focused packages that do one thing well, aligning with Go's core philosophy

#### Design the architecture, name the components, document the details - Examples

**Do**:
````go
// GOOD: An interface defines the contract, and separate structs
// provide concrete implementations. New formats can be added without
// modifying existing code.
type ReportGenerator interface {
    Generate() (byte, error)
}

type PDFGenerator struct { /*... fields... */ }
func (g *PDFGenerator) Generate() (byte, error) {
    //... logic to generate PDF
}

type CSVGenerator struct { /*... fields... */ }
func (g *CSVGenerator) Generate() (byte, error) {
    //... logic to generate CSV
}

// A factory function can be used to select the implementation.
func NewReportGenerator(format string) (ReportGenerator, error) {
    switch format {
    case "pdf":
        return &PDFGenerator{}, nil
    case "csv":
        return &CSVGenerator{}, nil
    default:
        return nil, fmt.Errorf("unsupported format: %s", format)
    }
}
````

**Don't**:
````go
// BAD: This struct is continuously modified to support new formats.
// Its complexity grows with each new requirement.
type ReportGenerator struct {
    //... fields
}

func (g *ReportGenerator) Generate(format string) (byte, error) {
    if format == "pdf" {
        //... logic to generate PDF
    } else if format == "csv" {
        //... logic to generate CSV
    } else if format == "json" {
        //... logic to generate JSON
    } else {
        return nil, errors.New("unsupported format")
    }
    //...
}
````

### Interfaces and Abstractions

* **The bigger the interface, the weaker the abstraction** - Keep interfaces small and focused
* **interface{} says nothing** - Use specific types instead of empty interfaces when possible
* **Make the zero value useful** - Design structs so their zero value is meaningful and safe to use

#### The bigger the interface, the weaker the abstraction - Examples

**Do**:
````go
// GOOD: The UserService defines the exact, minimal interface it needs.
package user

type UserFinder interface {
    FindUserByID(id int) (*User, error)
}

type Service struct {
    users UserFinder // The dependency is narrow and focused.
}

func (s *Service) GetUser(id int) (*User, error) {
    return s.users.FindUserByID(id)
}

// In another package, a concrete implementation can satisfy this interface.
package postgres

type Store struct { /*... */ }

// The Store struct doesn't need to know about the UserFinder interface.
// It just needs to have the right method.
func (s *Store) FindUserByID(id int) (*user.User, error) {
    //... database logic...
}

// These small interfaces can be composed if a larger contract is needed.
type UserUpdater interface {
    UpdateUser(u *User) error
}

type UserStore interface {
    UserFinder
    UserUpdater
}
````

**Don't**:
````go
// BAD: This interface is too large. A component that only needs to fetch a user
// is now forced to know about Orders and Products, creating unnecessary coupling.
// It is also difficult to create a mock for testing.
type Datastore interface {
    GetUser(id int) (*User, error)
    UpdateUser(u *User) error
    CreateOrder(o *Order) error
    GetOrder(id int) (*Order, error)
    DeleteProduct(id int) error
    //... and many more methods
}

// A service that only needs to find users.
type UserService struct {
    store Datastore // This dependency is too broad.
}

func (s *UserService) FindUser(id int) (*User, error) {
    return s.store.GetUser(id)
}
````

#### The Empty Interface and Zero Values - Examples

**Do**:
````go
// GOOD: An interface defines the required behavior.
type Processable interface {
    Process()
}

type StringData string
func (s StringData) Process() { fmt.Println("Processing string:", s) }

type IntData int
func (i IntData) Process() { fmt.Println("Processing int:", i * 2) }

// This function is now type-safe. The compiler ensures that only
// types that satisfy the Processable interface can be passed.
func Process(p Processable) {
    p.Process()
}
````

**Don't**:
````go
// BAD: This function relies on a type switch to handle different types.
// It is not type-safe at compile time. A caller could pass a float64
// and it would do nothing, which may be an unexpected behavior.
func Process(data interface{}) {
    switch v := data.(type) {
    case string:
        fmt.Println("Processing string:", v)
    case int:
        fmt.Println("Processing int:", v * 2)
    default:
        // What happens here? The behavior is implicit.
    }
}
````

#### Make the zero value useful - Examples

**Do**:
````go
// GOOD: The Increment method is written to be "nil-safe".
// It lazily initializes the map on first use.
type Counter struct {
    counts map[string]int
}

func (c *Counter) Increment(key string) {
    if c.counts == nil {
        c.counts = make(map[string]int)
    }
    c.counts[key]++
}

func main() {
    var c Counter
    c.Increment("requests") // Works perfectly.
    fmt.Println(c.counts["requests"]) // Prints 1
}
````

**Don't**:
````go
// BAD: The zero value of this Counter is not useful.
// Calling Increment on it will cause a panic because the map is nil.
type Counter struct {
    counts map[string]int
}

func (c *Counter) Increment(key string) {
    c.counts[key]++ // This will panic if c.counts is nil.
}

func main() {
    var c Counter
    c.Increment("requests") // PANIC!
}
````

### State Management and Dependencies

* **Avoid package level state** - Provide dependencies as fields on types rather than using package variables
* **Explicit is better than implicit** - Make coupling and dependencies visible and clear
* **A little copying is better than a little dependency** - Prefer small amounts of duplication over unnecessary dependencies
* **Reduce spooky action at a distance** - Minimize hidden dependencies and global state

**Do**:
````go
// GOOD: Dependencies are explicit and passed during construction.
package main

import (
    "database/sql"
    "net/http"
)

// Server holds all dependencies for the application.
type Server struct {
    db *sql.DB
    // router, logger, etc. could also be fields here.
}

// NewServer is a constructor that wires up the application.
func NewServer(db *sql.DB) *Server {
    return &Server{db: db}
}

// GetUserHandler is now a method on the Server.
// It accesses dependencies via the receiver `s`.
func (s *Server) GetUserHandler(w http.ResponseWriter, r *http.Request) {
    // The dependency on the database is explicit.
    rows, err := s.db.QueryContext(r.Context(), "SELECT name FROM users WHERE id = $1", 1)
    //...
}

func main() {
    db, err := sql.Open("postgres", "...")
    if err!= nil {
        log.Fatal(err)
    }

    server := NewServer(db)

    // The handler is now tied to the server instance.
    http.HandleFunc("/users", server.GetUserHandler)
    http.ListenAndServe(":8080", nil)
}
````

**Don't**:
````go
// BAD: The handler has a hidden, implicit dependency on the global db variable.
package storage

import "database/sql"

// Global database connection.
var DB *sql.DB

func InitDB(dataSourceName string) {
    var err error
    DB, err = sql.Open("postgres", dataSourceName)
    if err!= nil {
        log.Fatal(err)
    }
}

package main

import "net/http"
import "myapp/storage"

func GetUserHandler(w http.ResponseWriter, r *http.Request) {
    // This handler implicitly depends on storage.DB.
    // Testing this requires manipulating a global variable.
    rows, err := storage.DB.QueryContext(r.Context(), "SELECT name FROM users WHERE id = $1", 1)
    //...
}

func main() {
    storage.InitDB("...")
    http.HandleFunc("/users", GetUserHandler)
    http.ListenAndServe(":8080", nil)
}
````


### Control Flow and Structure

* **Return early rather than nesting deeply** - Use guard clauses and keep the happy path at the left edge
* **Flat is better than nested** - Minimize indentation levels and complex nested structures
* **Line of sight coding** - Keep important code visible without horizontal scrolling (max nesting levels to 3)

**Do**:
````go
// GOOD: This version is flat and easy to read.
// Each precondition is checked at the top, and the function returns early.
// The happy path is clear and unindented.
func ProcessPayment(p *Payment, u *User) error {
    if u == nil {
        return errors.New("user cannot be nil")
    }
    if!u.IsActive {
        return errors.New("user is not active")
    }
    if p.Amount <= 0 {
        return errors.New("payment amount must be positive")
    }

    if err := p.Submit(); err!= nil {
        log.Printf("failed to submit payment: %v", err)
        return err // Note: error wrapping would be even better here.
    }

    p.Status = "COMPLETED"
    return nil
}
````

**Don't**:
````go
// BAD: This function has a deeply nested structure that is hard to follow.
// The "happy path" logic is indented multiple levels deep.
func ProcessPayment(p *Payment, u *User) error {
    if u!= nil {
        if u.IsActive {
            if p.Amount > 0 {
                err := p.Submit()
                if err!= nil {
                    log.Printf("failed to submit payment: %v", err)
                    return err
                } else {
                    // This is the successful path, buried deep inside.
                    p.Status = "COMPLETED"
                    return nil
                }
            } else {
                return errors.New("payment amount must be positive")
            }
        } else {
            return errors.New("user is not active")
        }
    } else {
        return errors.New("user cannot be nil")
    }
}
````

### Error Handling

* **Errors are values** - Treat errors as first-class values, not exceptions
* **Don't just check errors, handle them gracefully** - Provide meaningful error handling and recovery
* **Don't panic** - Use panic only for truly exceptional situations; prefer returning errors
* **Always wrap errors when bubbling up** - When returning an error from an internal function, always wrap it with contextual information using `fmt.Errorf("<context>: %w", err)` where `<context>` describes what the current function was trying to do (e.g., `"get user %d"`, `"parse config"`, `"connect to database"`). This preserves the error chain and provides a clear trail for debugging
* **Log or return, but not both** - Avoid logging and returning the same error, as this causes duplicate logs up the call stack. Either handle (log) the error, or wrap and return it to the caller—not both
  * **Exception for API boundaries**: At service boundaries—both when handling incoming requests (HTTP/gRPC handlers) and when calling upstream services—it is acceptable to log detailed error information for observability while returning a sanitized or sentinel error to the caller. For example, an upstream client may log the full REST response from an external service but return a simplified sentinel error to its caller
  * **For gRPC**: Use status errors with appropriate codes and details at API boundaries. Log the full error for observability, but return a `status.Error` with only the information the client needs

#### Log or Return - Examples

**Do**:
````go
// GOOD: Internal function wraps and returns the error without logging.
// The caller up the stack decides how to handle it.
func (s *Store) GetUser(ctx context.Context, id int) (*User, error) {
    user, err := s.db.QueryUser(ctx, id)
    if err != nil {
        // Always wrap with context when bubbling up - preserves error chain
        return nil, fmt.Errorf("get user %d: %w", id, err)
    }
    return user, nil
}

// GOOD: At the HTTP API boundary, log the detailed error and return a sanitized response.
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    user, err := h.store.GetUser(r.Context(), userID)
    if err != nil {
        // Log the full error with context for observability (single log at transport layer)
        h.logger.Error("failed to get user",
            "userID", userID,
            "error", err,
            "requestID", middleware.GetRequestID(r.Context()),
        )
        // Return a sanitized error to the client - no internal details
        if errors.Is(err, ErrNotFound) {
            http.Error(w, "user not found", http.StatusNotFound)
            return
        }
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }
    // ... handle success
}

// GOOD: At the gRPC API boundary, log details and return a status error.
func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    user, err := s.store.GetUser(ctx, req.GetUserId())
    if err != nil {
        // Log the full error chain for observability
        s.logger.Error("failed to get user",
            "userID", req.GetUserId(),
            "error", err,
        )
        // Return a status error with appropriate code - no internal details exposed
        if errors.Is(err, ErrNotFound) {
            return nil, status.Error(codes.NotFound, "user not found")
        }
        return nil, status.Error(codes.Internal, "internal error")
    }
    return &pb.GetUserResponse{User: toProto(user)}, nil
}
````

**Don't**:
````go
// BAD: This function logs AND returns the error.
// Every caller that also logs will create duplicate log entries.
func (s *Store) GetUser(ctx context.Context, id int) (*User, error) {
    user, err := s.db.QueryUser(ctx, id)
    if err != nil {
        // Logging here AND returning causes duplicate logs
        log.Printf("failed to get user %d: %v", id, err)
        return nil, fmt.Errorf("get user %d: %w", id, err)
    }
    return user, nil
}

// BAD: Now the handler also logs, resulting in duplicate log entries.
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
    user, err := h.store.GetUser(r.Context(), userID)
    if err != nil {
        log.Printf("handler: failed to get user: %v", err) // Duplicate!
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }
    // ...
}
````

#### Don't Panic - Examples

**Do**:
````go
// GOOD: The function returns an error for invalid input, allowing the
// caller to handle it gracefully.
func GetUserByID(id string) (*User, error) {
    if id == "" {
        return nil, errors.New("user ID cannot be empty")
    }
    //... logic to fetch user
}
````

**Don't**:
````go
// BAD: This function panics on invalid user input. This is a recoverable
// error and should be handled by returning an error value.
func GetUserByID(id string) *User {
    if id == "" {
        panic("user ID cannot be empty")
    }
    //... logic to fetch user
}
````

### Concurrency

* **Don't communicate by sharing memory, share memory by communicating** - Prefer channels over shared variables with mutexes
* **Channels orchestrate; mutexes serialize** - Use channels for orchestration, mutexes for protecting data
* **Before you launch a goroutine, know when it will stop** - Have a clear termination strategy
* **Leave concurrency to the caller** - Let library users control goroutine lifecycle
* **Always provide a way to terminate** a go rountine from outside the routine, by passing a context or providing a termination method

### Advanced Features (Use Sparingly)

* **Reflection is never clear** - Avoid reflection unless absolutely necessary
* **With the unsafe package there are no guarantees** - Avoid unsafe operations
* **Cgo is not Go** - avoid cgo usage

### Maintainability Focus

* **Maintainability counts above all** - Optimize for the person who will maintain your code
* **Documentation is for users** - Write documentation that helps users understand what, not how
* **Clear to the reader** - The reader, not the writer, is who matters most

**Do**:
````go
// GOOD: This godoc comment is clear, concise, and focuses on the contract.
// It follows the standard format of "FuncName..." and describes what the
// function does, its parameters, and its return values.

// GetUser retrieves a user by their unique identifier.
// If the user is not found, it returns a ErrNotFound error.
func (s *Store) GetUser(ctx context.Context, id int) (*User, error) {
    //... implementation...
}

// An accompanying testable example provides even greater clarity.
func ExampleStore_GetUser() {
    // Assume newTestStore() creates a store with a mock database.
    store, cleanup := newTestStore()
    defer cleanup()

    user, err := store.GetUser(context.Background(), 1)
    if err!= nil {
        log.Fatal(err)
    }
    fmt.Println(user.Name)
    // Output: Alice
}
````

**Don't**:
````go
// BAD: This comment describes the implementation, not the contract.
// It is noisy and will become incorrect if the implementation changes.

// GetUser finds a user by their ID.
// It first checks the cache. If it's a cache hit, it decodes the JSON
// and returns the user. If it's a miss, it queries the database,
// then caches the result for 5 minutes before returning.
func (s *Store) GetUser(ctx context.Context, id int) (*User, error) {
    //... implementation...
}
````