package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/redis/go-redis/v9"
	"go-crud-backend/internal/models"
	"go-crud-backend/internal/repository"
	"go-crud-backend/internal/service"
)

// ProductHandler handles HTTP requests for products.
type ProductHandler struct {
	service service.ProductService
	db      *sql.DB
	rdb     *redis.Client
}

// NewProductHandler creates a new ProductHandler.
func NewProductHandler(svc service.ProductService, db *sql.DB, rdb *redis.Client) *ProductHandler {
	return &ProductHandler{
		service: svc,
		db:      db,
		rdb:     rdb,
	}
}

// RegisterRoutes registers product routes and health check on the provided ServeMux.
func (h *ProductHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("POST /api/v1/products", h.Create)
	mux.HandleFunc("GET /api/v1/products", h.List)
	mux.HandleFunc("GET /api/v1/products/{id}", h.GetByID)
	mux.HandleFunc("PUT /api/v1/products/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/products/{id}", h.Delete)
}

// Health verifies connectivity to Postgres and Redis.
func (h *ProductHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	status := map[string]string{
		"postgres": "ok",
		"redis":    "ok",
	}
	hasError := false

	if err := h.db.PingContext(ctx); err != nil {
		status["postgres"] = err.Error()
		hasError = true
	}

	if err := h.rdb.Ping(ctx).Err(); err != nil {
		status["redis"] = err.Error()
		hasError = true
	}

	statusCode := http.StatusOK
	if hasError {
		statusCode = http.StatusServiceUnavailable
	}

	respondJSON(w, statusCode, map[string]any{
		"status":   map[bool]string{true: "healthy", false: "unhealthy"}[!hasError],
		"services": status,
	})
}

// Create handles product creation.
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	product, err := h.service.CreateProduct(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, product)
}

// GetByID handles fetching a single product by ID.
func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	product, err := h.service.GetProduct(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "product not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to get product")
		return
	}

	respondJSON(w, http.StatusOK, product)
}

// List handles listing products with optional pagination.
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	limit := 10
	if l := query.Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0
	if o := query.Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	products, err := h.service.ListProducts(r.Context(), limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list products")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"data":   products,
		"limit":  limit,
		"offset": offset,
	})
}

// Update handles full/partial product update.
func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	var req models.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	product, err := h.service.UpdateProduct(r.Context(), id, req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "product not found")
			return
		}
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, product)
}

// Delete handles removing a product.
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		respondError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	err = h.service.DeleteProduct(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			respondError(w, http.StatusNotFound, "product not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to delete product")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
