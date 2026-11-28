package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"product-api/internal/models"
	"product-api/internal/repositories"
)

// fakeRepo is an in-memory ProductRepository used for testing.
type fakeRepo struct {
	products map[int64]models.Product
	nextID   int64
	err      error
}

func newFakeRepo() *fakeRepo { return &fakeRepo{products: map[int64]models.Product{}, nextID: 1} }

func (r *fakeRepo) Create(_ context.Context, p *models.Product) error {
	if r.err != nil {
		return r.err
	}
	p.ID = r.nextID
	r.nextID++
	r.products[p.ID] = *p
	return nil
}

func (r *fakeRepo) FindAll(_ context.Context, keyword string) ([]models.Product, error) {
	out := []models.Product{}
	for _, p := range r.products {
		if keyword == "" || strings.Contains(strings.ToLower(p.Name), strings.ToLower(keyword)) {
			out = append(out, p)
		}
	}
	return out, r.err
}

func (r *fakeRepo) FindByID(_ context.Context, id int64) (*models.Product, error) {
	p, ok := r.products[id]
	if !ok {
		return nil, repositories.ErrNotFound
	}
	return &p, nil
}

func (r *fakeRepo) Update(_ context.Context, p *models.Product) error {
	if _, ok := r.products[p.ID]; !ok {
		return repositories.ErrNotFound
	}
	r.products[p.ID] = *p
	return nil
}

func (r *fakeRepo) Delete(_ context.Context, id int64) error {
	if _, ok := r.products[id]; !ok {
		return repositories.ErrNotFound
	}
	delete(r.products, id)
	return nil
}

func floatPtr(v float64) *float64 { return &v }
func intPtr(v int) *int           { return &v }

func validInput() models.ProductInput {
	return models.ProductInput{Name: "Mechanical Keyboard", Description: "Wireless", Price: floatPtr(120.5), Quantity: intPtr(10)}
}

func TestValidateProductInput(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*models.ProductInput)
		wantMsg string
	}{
		{"valid", func(*models.ProductInput) {}, ""},
		{"zero quantity is allowed", func(in *models.ProductInput) { in.Quantity = intPtr(0) }, ""},
		{"empty name", func(in *models.ProductInput) { in.Name = "" }, "name is required"},
		{"blank name", func(in *models.ProductInput) { in.Name = "   " }, "name is required"},
		{"short name", func(in *models.ProductInput) { in.Name = "ab" }, "name must be at least 3 characters"},
		{"long name", func(in *models.ProductInput) { in.Name = strings.Repeat("a", 256) }, "name must be at most 255 characters"},
		{"missing price", func(in *models.ProductInput) { in.Price = nil }, "price is required"},
		{"zero price", func(in *models.ProductInput) { in.Price = floatPtr(0) }, "price must be greater than 0"},
		{"negative price", func(in *models.ProductInput) { in.Price = floatPtr(-1) }, "price must be greater than 0"},
		{"missing quantity", func(in *models.ProductInput) { in.Quantity = nil }, "quantity is required"},
		{"negative quantity", func(in *models.ProductInput) { in.Quantity = intPtr(-1) }, "quantity must be greater than or equal to 0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validInput()
			tt.modify(&in)
			err := validateProductInput(in)
			if tt.wantMsg == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			var vErr *ValidationError
			if !errors.As(err, &vErr) || vErr.Message != tt.wantMsg {
				t.Fatalf("expected validation error %q, got %v", tt.wantMsg, err)
			}
		})
	}
}

func TestProductServiceCRUD(t *testing.T) {
	ctx := context.Background()
	svc := NewProductService(newFakeRepo())

	created, err := svc.CreateProduct(ctx, validInput())
	if err != nil || created.ID != 1 {
		t.Fatalf("create: got %+v, %v", created, err)
	}

	got, err := svc.GetProduct(ctx, created.ID)
	if err != nil || got.Name != "Mechanical Keyboard" {
		t.Fatalf("get: got %+v, %v", got, err)
	}

	update := validInput()
	update.Name = "  Updated Keyboard  "
	updated, err := svc.UpdateProduct(ctx, created.ID, update)
	if err != nil || updated.Name != "Updated Keyboard" {
		t.Fatalf("update: got %+v, %v", updated, err)
	}

	list, err := svc.ListProducts(ctx, "updated")
	if err != nil || len(list) != 1 {
		t.Fatalf("list: got %d items, %v", len(list), err)
	}

	if err := svc.DeleteProduct(ctx, created.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := svc.GetProduct(ctx, created.ID); !errors.Is(err, ErrProductNotFound) {
		t.Fatalf("expected ErrProductNotFound after delete, got %v", err)
	}
}

func TestProductServiceNotFound(t *testing.T) {
	ctx := context.Background()
	svc := NewProductService(newFakeRepo())

	if _, err := svc.GetProduct(ctx, 99); !errors.Is(err, ErrProductNotFound) {
		t.Errorf("get: expected ErrProductNotFound, got %v", err)
	}
	if _, err := svc.UpdateProduct(ctx, 99, validInput()); !errors.Is(err, ErrProductNotFound) {
		t.Errorf("update: expected ErrProductNotFound, got %v", err)
	}
	if err := svc.DeleteProduct(ctx, 99); !errors.Is(err, ErrProductNotFound) {
		t.Errorf("delete: expected ErrProductNotFound, got %v", err)
	}
}

func TestProductServiceRepositoryError(t *testing.T) {
	repo := newFakeRepo()
	repo.err = errors.New("connection refused")
	svc := NewProductService(repo)

	_, err := svc.CreateProduct(context.Background(), validInput())
	var vErr *ValidationError
	if err == nil || errors.As(err, &vErr) || errors.Is(err, ErrProductNotFound) {
		t.Fatalf("expected a plain internal error, got %v", err)
	}
}
