package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-crud-backend/internal/handler"
	"go-crud-backend/internal/models"
	"go-crud-backend/internal/repository"
)

type mockProductService struct {
	createFn func(ctx context.Context, req models.CreateProductRequest) (*models.Product, error)
	getFn    func(ctx context.Context, id int64) (*models.Product, error)
	listFn   func(ctx context.Context, limit, offset int) ([]models.Product, error)
	updateFn func(ctx context.Context, id int64, req models.UpdateProductRequest) (*models.Product, error)
	deleteFn func(ctx context.Context, id int64) error
}

func (m *mockProductService) CreateProduct(ctx context.Context, req models.CreateProductRequest) (*models.Product, error) {
	if m.createFn != nil {
		return m.createFn(ctx, req)
	}
	return nil, nil
}

func (m *mockProductService) GetProduct(ctx context.Context, id int64) (*models.Product, error) {
	if m.getFn != nil {
		return m.getFn(ctx, id)
	}
	return nil, nil
}

func (m *mockProductService) ListProducts(ctx context.Context, limit, offset int) ([]models.Product, error) {
	if m.listFn != nil {
		return m.listFn(ctx, limit, offset)
	}
	return nil, nil
}

func (m *mockProductService) UpdateProduct(ctx context.Context, id int64, req models.UpdateProductRequest) (*models.Product, error) {
	if m.updateFn != nil {
		return m.updateFn(ctx, id, req)
	}
	return nil, nil
}

func (m *mockProductService) DeleteProduct(ctx context.Context, id int64) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

func TestProductHandler_CreateAndGet(t *testing.T) {
	mockSvc := &mockProductService{
		createFn: func(ctx context.Context, req models.CreateProductRequest) (*models.Product, error) {
			return &models.Product{
				ID:          1,
				Name:        req.Name,
				Description: req.Description,
				Price:       req.Price,
				Stock:       req.Stock,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}, nil
		},
		getFn: func(ctx context.Context, id int64) (*models.Product, error) {
			if id == 1 {
				return &models.Product{
					ID:    1,
					Name:  "Test Item",
					Price: 25.0,
				}, nil
			}
			return nil, repository.ErrNotFound
		},
	}

	h := handler.NewProductHandler(mockSvc, nil, nil)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// 1. Test POST /api/v1/products
	payload := `{"name":"Test Item","description":"Desc","price":25.0,"stock":10}`
	req := httptest.NewRequest("POST", "/api/v1/products", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var created models.Product
	if err := json.Unmarshal(w.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if created.ID != 1 || created.Name != "Test Item" {
		t.Fatalf("unexpected product data: %+v", created)
	}

	// 2. Test GET /api/v1/products/1
	getReq := httptest.NewRequest("GET", "/api/v1/products/1", nil)
	getRec := httptest.NewRecorder()
	mux.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getRec.Code)
	}

	// 3. Test GET /api/v1/products/999 (Not Found)
	nfReq := httptest.NewRequest("GET", "/api/v1/products/999", nil)
	nfRec := httptest.NewRecorder()
	mux.ServeHTTP(nfRec, nfReq)

	if nfRec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, nfRec.Code)
	}
}
