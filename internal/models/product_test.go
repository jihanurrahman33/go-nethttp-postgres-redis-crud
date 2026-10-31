package models_test

import (
	"testing"

	"go-crud-backend/internal/models"
)

func TestCreateProductRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     models.CreateProductRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: models.CreateProductRequest{
				Name:        "Test Product",
				Description: "Sample description",
				Price:       49.99,
				Stock:       10,
			},
			wantErr: false,
		},
		{
			name: "empty name",
			req: models.CreateProductRequest{
				Name:        "   ",
				Description: "Sample",
				Price:       10.0,
				Stock:       5,
			},
			wantErr: true,
		},
		{
			name: "negative price",
			req: models.CreateProductRequest{
				Name:        "Product",
				Description: "Sample",
				Price:       -5.0,
				Stock:       5,
			},
			wantErr: true,
		},
		{
			name: "negative stock",
			req: models.CreateProductRequest{
				Name:        "Product",
				Description: "Sample",
				Price:       10.0,
				Stock:       -1,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateProductRequest_Validate(t *testing.T) {
	req := models.UpdateProductRequest{
		Name:  "Updated Product",
		Price: 19.99,
		Stock: 2,
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("expected valid update request, got: %v", err)
	}

	invalidReq := models.UpdateProductRequest{
		Name: "",
	}
	if err := invalidReq.Validate(); err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}
