// Package services contains business logic and input validation.
package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"product-api/internal/models"
	"product-api/internal/repositories"
)

// Errors the handler layer maps to HTTP status codes.
var (
	ErrProductNotFound = errors.New("product not found")
)

// ValidationError is returned when the client sent invalid input.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

func newValidationError(msg string) error { return &ValidationError{Message: msg} }

// ProductService defines the product use cases.
type ProductService interface {
	CreateProduct(ctx context.Context, input models.ProductInput) (*models.Product, error)
	ListProducts(ctx context.Context, keyword string) ([]models.Product, error)
	GetProduct(ctx context.Context, id int64) (*models.Product, error)
	UpdateProduct(ctx context.Context, id int64, input models.ProductInput) (*models.Product, error)
	DeleteProduct(ctx context.Context, id int64) error
}

type productService struct {
	repo repositories.ProductRepository
}

// NewProductService creates a ProductService backed by the given repository.
func NewProductService(repo repositories.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) CreateProduct(ctx context.Context, input models.ProductInput) (*models.Product, error) {
	if err := validateProductInput(input); err != nil {
		return nil, err
	}
	product := buildProduct(input)
	if err := s.repo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("create product: %w", err)
	}
	return product, nil
}

func (s *productService) ListProducts(ctx context.Context, keyword string) ([]models.Product, error) {
	products, err := s.repo.FindAll(ctx, strings.TrimSpace(keyword))
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	return products, nil
}

func (s *productService) GetProduct(ctx context.Context, id int64) (*models.Product, error) {
	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, translateRepoError("get product", err)
	}
	return product, nil
}

func (s *productService) UpdateProduct(ctx context.Context, id int64, input models.ProductInput) (*models.Product, error) {
	if err := validateProductInput(input); err != nil {
		return nil, err
	}
	product := buildProduct(input)
	product.ID = id
	if err := s.repo.Update(ctx, product); err != nil {
		return nil, translateRepoError("update product", err)
	}
	return product, nil
}

func (s *productService) DeleteProduct(ctx context.Context, id int64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return translateRepoError("delete product", err)
	}
	return nil
}

func buildProduct(input models.ProductInput) *models.Product {
	return &models.Product{
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Price:       *input.Price,
		Quantity:    *input.Quantity,
	}
}

func translateRepoError(op string, err error) error {
	if errors.Is(err, repositories.ErrNotFound) {
		return ErrProductNotFound
	}
	return fmt.Errorf("%s: %w", op, err)
}

const maxNameLength = 255

func validateProductInput(input models.ProductInput) error {
	name := strings.TrimSpace(input.Name)
	switch {
	case name == "":
		return newValidationError("name is required")
	case utf8.RuneCountInString(name) < 3:
		return newValidationError("name must be at least 3 characters")
	case utf8.RuneCountInString(name) > maxNameLength:
		return newValidationError("name must be at most 255 characters")
	case input.Price == nil:
		return newValidationError("price is required")
	case *input.Price <= 0:
		return newValidationError("price must be greater than 0")
	case *input.Price >= 1e10:
		return newValidationError("price is too large")
	case input.Quantity == nil:
		return newValidationError("quantity is required")
	case *input.Quantity < 0:
		return newValidationError("quantity must be greater than or equal to 0")
	}
	return nil
}
