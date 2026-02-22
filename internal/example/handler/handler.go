package handler

import (
	"github.com/Alwanly/management-sport/internal/example/repository"
	"github.com/Alwanly/management-sport/internal/example/schema"
	"github.com/Alwanly/management-sport/internal/example/usecase"
	"github.com/Alwanly/management-sport/pkg/binding"
	"github.com/Alwanly/management-sport/pkg/deps"
	"github.com/Alwanly/management-sport/pkg/logger"
	"github.com/Alwanly/management-sport/pkg/validator"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const ContextName = "Internal.Book.Handler"

type (
	Handler struct {
		Logger    *zap.Logger
		Validator validator.IValidatorService
		UseCase   usecase.IUseCase
	}
)

func NewHandler(d *deps.App) *Handler {
	repository := repository.NewRepository(repository.Repository{
		DB:    d.DB,
		Redis: d.Redis,
	})
	usecase := usecase.NewUseCase(usecase.UseCase{
		Config:     d.Config,
		Logger:     d.Logger,
		Repository: repository,
	})
	handler := &Handler{
		Logger:    d.Logger,
		Validator: d.Validator,
		UseCase:   usecase,
	}

	e := d.Gin.Group("/books/v1")
	e.Use(d.Auth.JwtAuth())
	e.POST("/", handler.Create)
	e.GET("/", handler.List)
	e.GET("/:id", handler.Get)
	e.PUT("/:id", handler.Update)
	e.DELETE("/:id", handler.Delete)
	return handler
}

// Create godoc
// @Summary      Create a new book
// @Description  Create a new book
// @Tags         Books
// @Accept       json
// @Produce      json
// @Param        request body schema.RequestBookCreate true "Create book"
// @Success      201 {object} wrapper.JSONResult{data=schema.ResponseBookCreate}
// @Failure      400 {object} wrapper.JSONResult
// @Security     BearerAuth
// @Router       /books/v1 [post]
func (h *Handler) Create(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "Create")

	// bind model
	model := &schema.RequestBookCreate{}
	if err := binding.BindModel(l, c, model, binding.BindFromBody()); err != nil {
		perr := err.(*binding.ModelBindingError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	// validate model
	if err := validator.ValidateModel(l, h.Validator, model); err != nil {
		perr := err.(*validator.ModelValidationError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	// create a new book
	response := h.UseCase.Create(c.Request.Context(), model)
	c.JSON(response.Code, response)
}

// List godoc
// @Summary      List books
// @Description  List books with pagination
// @Tags         Books
// @Accept       json
// @Produce      json
// @Param        page query int false "Page number"
// @Param        page_size query int false "Page size"
// @Param        sort_by query string false "Sort by field"
// @Param        sort_order query string false "Sort order (asc|desc)"
// @Success      200 {object} wrapper.JSONResult{data=[]schema.ResponseBookGet}
// @Failure      400 {object} wrapper.JSONResult
// @Security     BearerAuth
// @Router       /books/v1 [get]
func (h *Handler) List(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "List")

	// bind model
	model := &schema.RequestBookList{
		Page:      1,
		PageSize:  10,
		SortBy:    "title",
		SortOrder: "desc",
	}
	if err := binding.BindModel(l, c, model, binding.BindFromQuery()); err != nil {
		perr := err.(*binding.ModelBindingError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	// validate model
	if err := validator.ValidateModel(l, h.Validator, model); err != nil {
		perr := err.(*validator.ModelValidationError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	// get list of books
	response := h.UseCase.List(c.Request.Context(), model)
	c.JSON(response.Code, response)
}

// Get godoc
// @Summary      Get book by id
// @Description  Retrieve a single book
// @Tags         Books
// @Accept       json
// @Produce      json
// @Param        id path string true "Book ID"
// @Success      200 {object} wrapper.JSONResult{data=schema.ResponseBookGet}
// @Failure      400 {object} wrapper.JSONResult
// @Security     BearerAuth
// @Router       /books/v1/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "Get")

	// bind model
	model := &schema.RequestBookGet{}
	if err := binding.BindModel(l, c, model, binding.BindFromParams()); err != nil {
		perr := err.(*binding.ModelBindingError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	// validate model
	if err := validator.ValidateModel(l, h.Validator, model); err != nil {
		perr := err.(*validator.ModelValidationError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	// get a book
	response := h.UseCase.Get(c.Request.Context(), model)
	c.JSON(response.Code, response)
}

// Update godoc
// @Summary      Update a book
// @Description  Update a book's information
// @Tags         Books
// @Accept       json
// @Produce      json
// @Param        id path string true "Book ID"
// @Param        request body schema.RequestBookUpdate true "Update book"
// @Success      200 {object} wrapper.JSONResult{data=schema.ResponseBookUpdate}
// @Failure      400 {object} wrapper.JSONResult
// @Security     BearerAuth
// @Router       /books/v1/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "Update")

	// bind model
	model := &schema.RequestBookUpdate{}
	if err := binding.BindModel(l, c, model, binding.BindFromParams(), binding.BindFromBody()); err != nil {
		perr := err.(*binding.ModelBindingError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	// validate model
	if err := validator.ValidateModel(l, h.Validator, model); err != nil {
		perr := err.(*validator.ModelValidationError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	// update a book
	response := h.UseCase.Update(c.Request.Context(), model)
	c.JSON(response.Code, response)
}

// Delete godoc
// @Summary      Delete a book
// @Description  Delete a book
// @Tags         Books
// @Accept       json
// @Produce      json
// @Param        id path string true "Book ID"
// @Success      200 {object} wrapper.JSONResult{data=schema.ResponseBookDelete}
// @Failure      400 {object} wrapper.JSONResult
// @Security     BearerAuth
// @Router       /books/v1/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	l := logger.WithID(h.Logger, ContextName, "Delete")

	// bind model
	model := &schema.RequestBookDelete{}
	if err := binding.BindModel(l, c, model, binding.BindFromParams()); err != nil {
		perr := err.(*binding.ModelBindingError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	// validate model
	if err := validator.ValidateModel(l, h.Validator, model); err != nil {
		perr := err.(*validator.ModelValidationError)
		c.JSON(perr.Code, perr.ResponseBody)
		return
	}

	// delete a book
	response := h.UseCase.Delete(c.Request.Context(), model)
	c.JSON(response.Code, response)
}
