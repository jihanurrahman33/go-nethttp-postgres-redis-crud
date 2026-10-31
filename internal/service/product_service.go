package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"go-crud-backend/internal/models"
	"go-crud-backend/internal/repository"
)

// ProductService defines the business logic operations for products.
type ProductService interface {
	CreateProduct(ctx context.Context, req models.CreateProductRequest) (*models.Product, error)
	GetProduct(ctx context.Context, id int64) (*models.Product, error)
	ListProducts(ctx context.Context, limit, offset int) ([]models.Product, error)
	UpdateProduct(ctx context.Context, id int64, req models.UpdateProductRequest) (*models.Product, error)
	DeleteProduct(ctx context.Context, id int64) error
}

type productService struct {
	repo     repository.ProductRepository
	redis    *redis.Client
	cacheTTL time.Duration
}

// NewProductService initializes a new ProductService instance.
func NewProductService(repo repository.ProductRepository, rdb *redis.Client, cacheTTL time.Duration) ProductService {
	return &productService{
		repo:     repo,
		redis:    rdb,
		cacheTTL: cacheTTL,
	}
}

func (s *productService) cacheKey(id int64) string {
	return fmt.Sprintf("product:%d", id)
}

func (s *productService) CreateProduct(ctx context.Context, req models.CreateProductRequest) (*models.Product, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	product := &models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, err
	}

	// Cache newly created product (best effort)
	if data, err := json.Marshal(product); err == nil {
		_ = s.redis.Set(ctx, s.cacheKey(product.ID), data, s.cacheTTL).Err()
	}

	return product, nil
}

func (s *productService) GetProduct(ctx context.Context, id int64) (*models.Product, error) {
	key := s.cacheKey(id)

	// 1. Try cache hit from Redis
	cachedVal, err := s.redis.Get(ctx, key).Result()
	if err == nil {
		var product models.Product
		if err := json.Unmarshal([]byte(cachedVal), &product); err == nil {
			log.Printf("[CACHE HIT] Fetched product %d from Redis", id)
			return &product, nil
		}
	} else if !errors.Is(err, redis.Nil) {
		log.Printf("[CACHE ERROR] Redis lookup failed for %s: %v", key, err)
	}

	// 2. Cache miss -> query Postgres database
	log.Printf("[CACHE MISS] Fetching product %d from Postgres", id)
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 3. Write back to Redis with TTL
	if data, err := json.Marshal(product); err == nil {
		if setErr := s.redis.Set(ctx, key, data, s.cacheTTL).Err(); setErr != nil {
			log.Printf("[CACHE ERROR] Failed setting Redis cache key %s: %v", key, setErr)
		}
	}

	return product, nil
}

func (s *productService) ListProducts(ctx context.Context, limit, offset int) ([]models.Product, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetAll(ctx, limit, offset)
}

func (s *productService) UpdateProduct(ctx context.Context, id int64, req models.UpdateProductRequest) (*models.Product, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if product exists first
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	existing.Name = req.Name
	existing.Description = req.Description
	existing.Price = req.Price
	existing.Stock = req.Stock

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	// Invalidate cache in Redis
	key := s.cacheKey(id)
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		log.Printf("[CACHE ERROR] Failed deleting Redis cache key %s: %v", key, err)
	}

	return existing, nil
}

func (s *productService) DeleteProduct(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate cache in Redis
	key := s.cacheKey(id)
	if err := s.redis.Del(ctx, key).Err(); err != nil {
		log.Printf("[CACHE ERROR] Failed deleting Redis cache key %s: %v", key, err)
	}

	return nil
}
