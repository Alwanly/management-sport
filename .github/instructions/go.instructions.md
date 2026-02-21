---
description: 'Instructions for writing Go code following idiomatic Go practices and community standards'
applyTo: '**/*.go,**/go.mod,**/go.sum'
---

# Go Development Instructions

Follow idiomatic Go practices and community standards when writing Go code. These instructions are based on [Effective Go](https://go.dev/doc/effective_go), [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments), and [Google's Go Style Guide](https://google.github.io/styleguide/go/).

## General Instructions

- Write simple, clear, and idiomatic Go code
- Favor clarity and simplicity over cleverness
- Follow the principle of least surprise
- Keep the happy path left-aligned (minimize indentation)
- Return early to reduce nesting
- Prefer early return over if-else chains; use `if condition { return }` pattern to avoid else blocks
- Make the zero value useful
- Write self-documenting code with clear, descriptive names
- Document exported types, functions, methods, and packages
- Use Go modules for dependency management
- Leverage the Go standard library instead of reinventing the wheel (e.g., use `strings.Builder` for string concatenation, `filepath.Join` for path construction)
- Prefer standard library solutions over custom implementations when functionality exists
- Write comments when nessacry or complex logic 
- Write comments in English by default; translate only upon user request
- Avoid using emoji in code and comments

## Naming Conventions

### Packages

- Use lowercase, single-word package names
- Avoid underscores, hyphens, or mixedCaps
- Choose names that describe what the package provides, not what it contains
- Avoid generic names like `util`, `common`, or `base`
- Package names should be singular, not plural

#### Package Declaration Rules (CRITICAL):
- **NEVER duplicate `package` declarations** - each Go file must have exactly ONE `package` line
- When editing an existing `.go` file:
  - **PRESERVE** the existing `package` declaration - do not add another one
  - If you need to replace the entire file content, start with the existing package name
- When creating a new `.go` file:
  - **BEFORE writing any code**, check what package name other `.go` files in the same directory use
  - Use the SAME package name as existing files in that directory
  - If it's a new directory, use the directory name as the package name
  - Write **exactly one** `package <name>` line at the very top of the file
- When using file creation or replacement tools:
  - **ALWAYS verify** the target file doesn't already have a `package` declaration before adding one
  - If replacing file content, include only ONE `package` declaration in the new content
  - **NEVER** create files with multiple `package` lines or duplicate declarations

### Variables and Functions

- Use mixedCaps or MixedCaps (camelCase) rather than underscores
- Keep names short but descriptive
- Use single-letter variables only for very short scopes (like loop indices)
- Exported names start with a capital letter
- Unexported names start with a lowercase letter
- Avoid stuttering (e.g., avoid `http.HTTPServer`, prefer `http.Server`)

### Interfaces

- Name interfaces with -er suffix when possible (e.g., `Reader`, `Writer`, `Formatter`)
- Single-method interfaces should be named after the method (e.g., `Read` → `Reader`)
- Keep interfaces small and focused

### Constants

- Use MixedCaps for exported constants
- Use mixedCaps for unexported constants
- Group related constants using `const` blocks
- Consider using typed constants for better type safety

## Code Style and Formatting

### Formatting

- Always use `gofmt` to format code
- Use `goimports` to manage imports automatically
- Keep line length reasonable (no hard limit, but consider readability)
- Add blank lines to separate logical groups of code

### Comments

- Strive for self-documenting code; prefer clear variable names, function names, and code structure over comments
- Write comments only when necessary to explain complex logic, business rules, or non-obvious behavior
- Write comments in complete sentences in English by default
- Translate comments to other languages only upon specific user request
- Start sentences with the name of the thing being described
- Package comments should start with "Package [name]"
- Use line comments (`//`) for most comments
- Use block comments (`/* */`) sparingly, mainly for package documentation
- Document why, not what, unless the what is complex
- Avoid emoji in comments and code

### Error Handling

- Check errors immediately after the function call
- Don't ignore errors using `_` unless you have a good reason (document why)
- Wrap errors with context using `fmt.Errorf` with `%w` verb
- Create custom error types when you need to check for specific errors
- Place error returns as the last return value
- Name error variables `err`
- Keep error messages lowercase and don't end with punctuation

## Architecture and Project Structure

This project follows Clean Architecture and Domain-Driven Design (DDD) principles:

### Layers

```
┌─────────────────────────────────────────┐
│           Handler Layer                  │  HTTP handlers (presentation)
├─────────────────────────────────────────┤
│           UseCase Layer                  │  Business logic
├─────────────────────────────────────────┤
│         Repository Layer                 │  Data access
├─────────────────────────────────────────┤
│           Model Layer                    │  Database models
└─────────────────────────────────────────┘
```

### Dependency Flow

- **Handler** → **UseCase** → **Repository** → **Model**
- Higher layers can depend on lower layers
- Lower layers should NOT depend on higher layers

## Adding New Features

### Example: Adding a "Product" Feature

#### Step 1: Create Domain Structure

Create the following directory structure:

```
internal/
└── product/
    ├── handler/
    │   └── handler.go
    ├── repository/
    │   └── repository.go
    ├── usecase/
    │   └── usecase.go
    └── schema/
        ├── request.go
        ├── response.go
        └── constant.go
```

#### Step 2: Define the Model

Create `model/product.go`:

```go
package model

import "time"

type Product struct {
	ID          string    `gorm:"primaryKey;column:id;type:varchar(255);not null"`
	Name        string    `gorm:"column:name;type:varchar(255);not null"`
	Description string    `gorm:"column:description;type:text"`
	Price       float64   `gorm:"column:price;type:decimal(10,2);not null"`
	Stock       int       `gorm:"column:stock;type:integer;not null"`
	CreatedAt   time.Time `gorm:"column:created_at;type:timestamptz;not null"`
	CreatedBy   string    `gorm:"column:created_by;type:varchar(255);not null"`
	UpdatedAt   time.Time `gorm:"column:updated_at;type:timestamptz;not null"`
	UpdatedBy   string    `gorm:"column:updated_by;type:varchar(255);not null"`
}

func (Product) TableName() string {
	return "products"
}

type Products []Product
```

#### Step 3: Define Schemas

Create `internal/product/schema/request.go`:

```go
package schema

type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Stock       int     `json:"stock" validate:"required,gte=0"`
}

type UpdateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price" validate:"omitempty,gt=0"`
	Stock       int     `json:"stock" validate:"omitempty,gte=0"`
}
```

Create `internal/product/schema/response.go`:

```go
package schema

import "time"

type ProductResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
```

#### Step 4: Create Repository Interface

Create `internal/product/repository/repository.go`:

```go
package repository

import (
	"context"

	"github.com/Alwanly/management-sport/model"
	"github.com/Alwanly/management-sport/pkg/database"
	"github.com/Alwanly/management-sport/pkg/redis"
)

type IRepository interface {
	Create(ctx context.Context, product *model.Product) error
	FindByID(ctx context.Context, id string) (*model.Product, error)
	FindAll(ctx context.Context, limit, offset int) (model.Products, int64, error)
	Update(ctx context.Context, product *model.Product) error
	Delete(ctx context.Context, id string) error
}

type Repository struct {
	DB    *database.DBService
	Redis *redis.Service
}

func NewRepository(r Repository) IRepository {
	return &r
}

func (r *Repository) Create(ctx context.Context, product *model.Product) error {
	return r.DB.Gorm.WithContext(ctx).Create(product).Error
}

func (r *Repository) FindByID(ctx context.Context, id string) (*model.Product, error) {
	var product model.Product
	err := r.DB.Gorm.WithContext(ctx).Where("id = ?", id).First(&product).Error
	return &product, err
}

func (r *Repository) FindAll(ctx context.Context, limit, offset int) (model.Products, int64, error) {
	var products model.Products
	var count int64

	err := r.DB.Gorm.WithContext(ctx).Model(&model.Product{}).Count(&count).Error
	if err != nil {
		return nil, 0, err
	}

	err = r.DB.Gorm.WithContext(ctx).Limit(limit).Offset(offset).Find(&products).Error
	return products, count, err
}

func (r *Repository) Update(ctx context.Context, product *model.Product) error {
	return r.DB.Gorm.WithContext(ctx).Save(product).Error
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	return r.DB.Gorm.WithContext(ctx).Delete(&model.Product{}, "id = ?", id).Error
}
```

#### Step 5: Create UseCase

Create `internal/product/usecase/usecase.go`:

```go
package usecase

import (
	"context"
	"time"

	"github.com/Alwanly/management-sport/config"
	"github.com/Alwanly/management-sport/internal/product/repository"
	"github.com/Alwanly/management-sport/internal/product/schema"
	"github.com/Alwanly/management-sport/model"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type IUseCase interface {
	Create(ctx context.Context, req *schema.CreateProductRequest, userID string) (*schema.ProductResponse, error)
	GetByID(ctx context.Context, id string) (*schema.ProductResponse, error)
	GetAll(ctx context.Context, page, limit int) ([]*schema.ProductResponse, int64, error)
	Update(ctx context.Context, id string, req *schema.UpdateProductRequest, userID string) error
	Delete(ctx context.Context, id string) error
}

type UseCase struct {
	Config     *config.GlobalConfig
	Logger     *zap.Logger
	Repository repository.IRepository
}

func NewUseCase(u UseCase) IUseCase {
	return &u
}

func (u *UseCase) Create(ctx context.Context, req *schema.CreateProductRequest, userID string) (*schema.ProductResponse, error) {
	product := &model.Product{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CreatedAt:   time.Now(),
		CreatedBy:   userID,
		UpdatedAt:   time.Now(),
		UpdatedBy:   userID,
	}

	if err := u.Repository.Create(ctx, product); err != nil {
		u.Logger.Error("Failed to create product", zap.Error(err))
		return nil, err
	}

	return u.toResponse(product), nil
}

func (u *UseCase) GetByID(ctx context.Context, id string) (*schema.ProductResponse, error) {
	product, err := u.Repository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return u.toResponse(product), nil
}

func (u *UseCase) GetAll(ctx context.Context, page, limit int) ([]*schema.ProductResponse, int64, error) {
	offset := (page - 1) * limit
	products, count, err := u.Repository.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*schema.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = u.toResponse(&product)
	}

	return responses, count, nil
}

func (u *UseCase) Update(ctx context.Context, id string, req *schema.UpdateProductRequest, userID string) error {
	product, err := u.Repository.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.Price > 0 {
		product.Price = req.Price
	}
	if req.Stock >= 0 {
		product.Stock = req.Stock
	}
	product.UpdatedAt = time.Now()
	product.UpdatedBy = userID

	return u.Repository.Update(ctx, product)
}

func (u *UseCase) Delete(ctx context.Context, id string) error {
	return u.Repository.Delete(ctx, id)
}

func (u *UseCase) toResponse(product *model.Product) *schema.ProductResponse {
	return &schema.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
}
```

#### Step 6: Create Handler

Create `internal/product/handler/handler.go`:

```go
package handler

import (
	"strconv"

	"github.com/Alwanly/management-sport/internal/product/repository"
	"github.com/Alwanly/management-sport/internal/product/schema"
	"github.com/Alwanly/management-sport/internal/product/usecase"
	"github.com/Alwanly/management-sport/pkg/binding"
	"github.com/Alwanly/management-sport/pkg/contract"
	"github.com/Alwanly/management-sport/pkg/deps"
	"github.com/Alwanly/management-sport/pkg/wrapper"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type Handler struct {
	Logger  *zap.Logger
	UseCase usecase.IUseCase
}

func NewHandler(d *deps.App) *Handler {
	repo := repository.NewRepository(repository.Repository{
		DB:    d.DB,
		Redis: d.Redis,
	})
	uc := usecase.NewUseCase(usecase.UseCase{
		Config:     d.Config,
		Logger:     d.Logger,
		Repository: repo,
	})

	handler := &Handler{
		Logger:  d.Logger,
		UseCase: uc,
	}

	// Register routes
	e := d.Fiber.Group("/products/v1", d.Auth.JwtAuth())
	e.Post("/", handler.Create)
	e.Get("/", handler.List)
	e.Get("/:id", handler.Get)
	e.Put("/:id", handler.Update)
	e.Delete("/:id", handler.Delete)

	return handler
}

// Create godoc
// @Summary Create product
// @Tags products
// @Accept json
// @Produce json
// @Param request body schema.CreateProductRequest true "Product data"
// @Success 201 {object} wrapper.JSONResult{data=schema.ProductResponse}
// @Router /products/v1/ [post]
// @Security Bearer
func (h *Handler) Create(c *fiber.Ctx) error {
	var req schema.CreateProductRequest
	if err := binding.Bind(c, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(wrapper.ResponseFailed(
			fiber.StatusBadRequest,
			contract.StatusCodeBadRequest,
			"Invalid request",
			nil,
		))
	}

	// Get user ID from context (set by JWT middleware)
	userID := c.Locals("user_id").(string)

	result, err := h.UseCase.Create(c.Context(), &req, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(wrapper.ResponseFailed(
			fiber.StatusInternalServerError,
			contract.StatusCodeInternalServerError,
			"Failed to create product",
			nil,
		))
	}

	return c.Status(fiber.StatusCreated).JSON(wrapper.ResponseSuccess(fiber.StatusCreated, result))
}

// List godoc
// @Summary List products
// @Tags products
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} wrapper.JSONResult{data=[]schema.ProductResponse}
// @Router /products/v1/ [get]
// @Security Bearer
func (h *Handler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	products, count, err := h.UseCase.GetAll(c.Context(), page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(wrapper.ResponseFailed(
			fiber.StatusInternalServerError,
			contract.StatusCodeInternalServerError,
			"Failed to fetch products",
			nil,
		))
	}

	return c.JSON(wrapper.ResponsePagination(page, limit, len(products), int(count), products, nil))
}

// Get godoc
// @Summary Get product by ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} wrapper.JSONResult{data=schema.ProductResponse}
// @Router /products/v1/{id} [get]
// @Security Bearer
func (h *Handler) Get(c *fiber.Ctx) error {
	id := c.Params("id")

	product, err := h.UseCase.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(wrapper.ResponseFailed(
			fiber.StatusNotFound,
			contract.StatusCodeNotFound,
			"Product not found",
			nil,
		))
	}

	return c.JSON(wrapper.ResponseSuccess(fiber.StatusOK, product))
}

// Update godoc
// @Summary Update product
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param request body schema.UpdateProductRequest true "Product data"
// @Success 200 {object} wrapper.JSONResult
// @Router /products/v1/{id} [put]
// @Security Bearer
func (h *Handler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	var req schema.UpdateProductRequest

	if err := binding.Bind(c, &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(wrapper.ResponseFailed(
			fiber.StatusBadRequest,
			contract.StatusCodeBadRequest,
			"Invalid request",
			nil,
		))
	}

	userID := c.Locals("user_id").(string)

	if err := h.UseCase.Update(c.Context(), id, &req, userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(wrapper.ResponseFailed(
			fiber.StatusInternalServerError,
			contract.StatusCodeInternalServerError,
			"Failed to update product",
			nil,
		))
	}

	return c.JSON(wrapper.ResponseSuccess(fiber.StatusOK, nil))
}

// Delete godoc
// @Summary Delete product
// @Tags products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} wrapper.JSONResult
// @Router /products/v1/{id} [delete]
// @Security Bearer
func (h *Handler) Delete(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.UseCase.Delete(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(wrapper.ResponseFailed(
			fiber.StatusInternalServerError,
			contract.StatusCodeInternalServerError,
			"Failed to delete product",
			nil,
		))
	}

	return c.JSON(wrapper.ResponseSuccess(fiber.StatusOK, nil))
}
```

#### Step 7: Register Handler in Bootstrap

Edit `cmd/main/bootstrap.go`:

```go
import (
	// ... other imports
	product_handler "github.com/Alwanly/management-sport/internal/product/handler"
)

// In Bootstrap function:
book_handler.NewHandler(inst)
product_handler.NewHandler(inst) // Add this line
```

#### Step 8: Generate Documentation

```bash
make docs
```

## Code Style Guidelines

### Follow Go Standards

- Use `gofmt` for formatting
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Follow [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)

### Naming Conventions

- **Packages**: lowercase, single word (e.g., `product`, `user`)
- **Files**: lowercase with underscores (e.g., `product_handler.go`)
- **Types**: PascalCase (e.g., `ProductService`)
- **Functions**: camelCase for private, PascalCase for exported
- **Constants**: PascalCase (e.g., `MaxRetries`)

### Documentation

- Document all exported types and functions
- Use godoc format
- Include Swagger annotations for API endpoints

```go
// ProductService handles product-related operations.
// It implements business logic for product management.
type ProductService struct {
    repo IProductRepository
}

// Create creates a new product.
// It validates the input and returns the created product or an error.
func (s *ProductService) Create(ctx context.Context, req *CreateProductRequest) (*Product, error) {
    // Implementation
}
```

### Package Organization

- Follow standard Go project layout conventions
- Keep `main` packages in `cmd/` directory
- Put reusable packages in `pkg/` or `internal/`
- Use `internal/` for packages that shouldn't be imported by external projects
- Group related functionality into packages
- Avoid circular dependencies

### Dependency Management

- Use Go modules (`go.mod` and `go.sum`)
- Keep dependencies minimal
- Regularly update dependencies for security patches
- Use `go mod tidy` to clean up unused dependencies
- Vendor dependencies only when necessary

## Type Safety and Language Features

### Type Definitions

- Define types to add meaning and type safety
- Use struct tags for JSON, XML, database mappings
- Prefer explicit type conversions
- Use type assertions carefully and check the second return value
- Prefer generics over unconstrained types; when an unconstrained type is truly needed, use the predeclared alias `any` instead of `interface{}` (Go 1.18+)

### Pointers vs Values

- Use pointer receivers for large structs or when you need to modify the receiver
- Use value receivers for small structs and when immutability is desired
- Use pointer parameters when you need to modify the argument or for large structs
- Use value parameters for small structs and when you want to prevent modification
- Be consistent within a type's method set
- Consider the zero value when choosing pointer vs value receivers

### Interfaces and Composition

- Accept interfaces, return concrete types
- Keep interfaces small (1-3 methods is ideal)
- Use embedding for composition
- Define interfaces close to where they're used, not where they're implemented
- Don't export interfaces unless necessary

## Concurrency

### Goroutines

- Be cautious about creating goroutines in libraries; prefer letting the caller control concurrency
- If you must create goroutines in libraries, provide clear documentation and cleanup mechanisms
- Always know how a goroutine will exit
- Use `sync.WaitGroup` or channels to wait for goroutines
- Avoid goroutine leaks by ensuring cleanup

### Channels

- Use channels to communicate between goroutines
- Don't communicate by sharing memory; share memory by communicating
- Close channels from the sender side, not the receiver
- Use buffered channels when you know the capacity
- Use `select` for non-blocking operations

### Synchronization

- Use `sync.Mutex` for protecting shared state
- Keep critical sections small
- Use `sync.RWMutex` when you have many readers
- Choose between channels and mutexes based on the use case: use channels for communication, mutexes for protecting state
- Use `sync.Once` for one-time initialization
- WaitGroup usage by Go version:
	- If `go >= 1.25` in `go.mod`, use the new `WaitGroup.Go` method ([documentation](https://pkg.go.dev/sync#WaitGroup)):
		```go
		var wg sync.WaitGroup
		wg.Go(task1)
		wg.Go(task2)
		wg.Wait()
		```
	- If `go < 1.25`, use the classic `Add`/`Done` pattern

## Error Handling Patterns

### Creating Errors

- Use `errors.New` for simple static errors
- Use `fmt.Errorf` for dynamic errors
- Create custom error types for domain-specific errors
- Export error variables for sentinel errors
- Use `errors.Is` and `errors.As` for error checking

### Error Propagation

- Add context when propagating errors up the stack
- Don't log and return errors (choose one)
- Handle errors at the appropriate level
- Consider using structured errors for better debugging

## API Design

### HTTP Handlers

- Use `http.HandlerFunc` for simple handlers
- Implement `http.Handler` for handlers that need state
- Use middleware for cross-cutting concerns
- Set appropriate status codes and headers
- Handle errors gracefully and return appropriate error responses
- Router usage by Go version:
	- If `go >= 1.22`, prefer the enhanced `net/http` `ServeMux` with pattern-based routing and method matching
	- If `go < 1.22`, use the classic `ServeMux` and handle methods/paths manually (or use a third-party router when justified)

### JSON APIs

- Use struct tags to control JSON marshaling
- Validate input data
- Use pointers for optional fields
- Consider using `json.RawMessage` for delayed parsing
- Handle JSON errors appropriately

### HTTP Clients

- Keep the client struct focused on configuration and dependencies only (e.g., base URL, `*http.Client`, auth, default headers). It must not store per-request state
- Do not store or cache `*http.Request` inside the client struct, and do not persist request-specific state across calls; instead, construct a fresh request per method invocation
- Methods should accept `context.Context` and input parameters, assemble the `*http.Request` locally (or via a short-lived builder/helper created per call), then call `c.httpClient.Do(req)`
- If request-building logic is reused, factor it into unexported helper functions or a per-call builder type; never keep `http.Request` (URL params, body, headers) as fields on the long-lived client
- Ensure the underlying `*http.Client` is configured (timeouts, transport) and is safe for concurrent use; avoid mutating `Transport` after first use
- Always set headers on the request instance you’re sending, and close response bodies (`defer resp.Body.Close()`), handling errors appropriately

## Performance Optimization

### Memory Management

- Minimize allocations in hot paths
- Reuse objects when possible (consider `sync.Pool`)
- Use value receivers for small structs
- Preallocate slices when size is known
- Avoid unnecessary string conversions

### I/O: Readers and Buffers

- Most `io.Reader` streams are consumable once; reading advances state. Do not assume a reader can be re-read without special handling
- If you must read data multiple times, buffer it once and recreate readers on demand:
	- Use `io.ReadAll` (or a limited read) to obtain `[]byte`, then create fresh readers via `bytes.NewReader(buf)` or `bytes.NewBuffer(buf)` for each reuse
	- For strings, use `strings.NewReader(s)`; you can `Seek(0, io.SeekStart)` on `*bytes.Reader` to rewind
- For HTTP requests, do not reuse a consumed `req.Body`. Instead:
	- Keep the original payload as `[]byte` and set `req.Body = io.NopCloser(bytes.NewReader(buf))` before each send
	- Prefer configuring `req.GetBody` so the transport can recreate the body for redirects/retries: `req.GetBody = func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(buf)), nil }`
- To duplicate a stream while reading, use `io.TeeReader` (copy to a buffer while passing through) or write to multiple sinks with `io.MultiWriter`
- Reusing buffered readers: call `(*bufio.Reader).Reset(r)` to attach to a new underlying reader; do not expect it to “rewind” unless the source supports seeking
- For large payloads, avoid unbounded buffering; consider streaming, `io.LimitReader`, or on-disk temporary storage to control memory

- Use `io.Pipe` to stream without buffering the whole payload:
	- Write to `*io.PipeWriter` in a separate goroutine while the reader consumes
	- Always close the writer; use `CloseWithError(err)` on failures
	- `io.Pipe` is for streaming, not rewinding or making readers reusable

- **Warning:** When using `io.Pipe` (especially with multipart writers), all writes must be performed in strict, sequential order. Do not write concurrently or out of order—multipart boundaries and chunk order must be preserved. Out-of-order or parallel writes can corrupt the stream and result in errors.

- Streaming multipart/form-data with `io.Pipe`:
	- `pr, pw := io.Pipe()`; `mw := multipart.NewWriter(pw)`; use `pr` as the HTTP request body
	- Set `Content-Type` to `mw.FormDataContentType()`
	- In a goroutine: write all parts to `mw` in the correct order; on error `pw.CloseWithError(err)`; on success `mw.Close()` then `pw.Close()`
	- Do not store request/in-flight form state on a long-lived client; build per call
	- Streamed bodies are not rewindable; for retries/redirects, buffer small payloads or provide `GetBody`

### Profiling

- Use built-in profiling tools (`pprof`)
- Benchmark critical code paths
- Profile before optimizing
- Focus on algorithmic improvements first
- Consider using `testing.B` for benchmarks

## Testing

### Test Organization

- Keep tests in the same package (white-box testing)
- Use `_test` package suffix for black-box testing
- Name test files with `_test.go` suffix
- Place test files next to the code they test

### Writing Tests

- Use table-driven tests for multiple test cases
- Name tests descriptively using `Test_functionName_scenario`
- Use subtests with `t.Run` for better organization
- Test both success and error cases
- Consider using `testify` or similar libraries when they add value, but don't over-complicate simple tests

### Test Helpers

- Mark helper functions with `t.Helper()`
- Create test fixtures for complex setup
- Use `testing.TB` interface for functions used in tests and benchmarks
- Clean up resources using `t.Cleanup()`

## Security Best Practices

### Input Validation

- Validate all external input
- Use strong typing to prevent invalid states
- Sanitize data before using in SQL queries
- Be careful with file paths from user input
- Validate and escape data for different contexts (HTML, SQL, shell)

### Cryptography

- Use standard library crypto packages
- Don't implement your own cryptography
- Use crypto/rand for random number generation
- Store passwords using bcrypt, scrypt, or argon2 (consider golang.org/x/crypto for additional options)
- Use TLS for network communication

## Documentation

### Code Documentation

- Prioritize self-documenting code through clear naming and structure
- Document all exported symbols with clear, concise explanations
- Start documentation with the symbol name
- Write documentation in English by default
- Use examples in documentation when helpful
- Keep documentation close to code
- Update documentation when code changes
- Avoid emoji in documentation and comments

### README and Documentation Files

- Include clear setup instructions
- Document dependencies and requirements
- Provide usage examples
- Document configuration options
- Include troubleshooting section

## Tools and Development Workflow

### Essential Tools

- `go fmt`: Format code
- `go vet`: Find suspicious constructs
- `golangci-lint`: Additional linting (golint is deprecated)
- `go test`: Run tests
- `go mod`: Manage dependencies
- `go generate`: Code generation

### Development Practices

- Run tests before committing
- Use pre-commit hooks for formatting and linting
- Keep commits focused and atomic
- Write meaningful commit messages
- Review diffs before committing

## Common Pitfalls to Avoid

- Not checking errors
- Ignoring race conditions
- Creating goroutine leaks
- Not using defer for cleanup
- Modifying maps concurrently
- Not understanding nil interfaces vs nil pointers
- Forgetting to close resources (files, connections)
- Using global variables unnecessarily
- Over-using unconstrained types (e.g., `any`); prefer specific types or generic type parameters with constraints. If an unconstrained type is required, use `any` rather than `interface{}`
- Not considering the zero value of types
- **Creating duplicate `package` declarations** - this is a compile error; always check existing files before adding package declarations