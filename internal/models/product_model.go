// Package models defines the domain entities and API payloads.
package models

import "time"

// Product is a product stored in the inventory.
type Product struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Quantity    int       `json:"quantity"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ProductInput is the request body for creating or updating a product.
//
// Price and Quantity are pointers so the service can tell a missing field
// ("price is required") apart from an explicit zero value.
type ProductInput struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Price       *float64 `json:"price"`
	Quantity    *int     `json:"quantity"`
}

// MessageResponse is the JSON shape for errors and simple confirmations.
type MessageResponse struct {
	Message string `json:"message"`
}
