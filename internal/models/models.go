package models

import "time"

// Rate Limiter Models
type RequestPayload struct {
	UserID  string      `json:"user_id"`
	Payload interface{} `json:"payload"`
}

type UserStats struct {
	AcceptedRequests int `json:"accepted_requests"`
	RejectedRequests int `json:"rejected_requests"`
}

// Product Models
type Product struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	SKU       string    `json:"sku"`
	ImageURLs []string  `json:"image_urls"`
	VideoURLs []string  `json:"video_urls"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateProductRequest struct {
	Name      string   `json:"name"`
	SKU       string   `json:"sku"`
	ImageURLs []string `json:"image_urls,omitempty"`
	VideoURLs []string `json:"video_urls,omitempty"`
}

type AddMediaRequest struct {
	ImageURLs []string `json:"image_urls,omitempty"`
	VideoURLs []string `json:"video_urls,omitempty"`
}

type ProductSummary struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	SKU          string `json:"sku"`
	ImageCount   int    `json:"image_count"`
	VideoCount   int    `json:"video_count"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

type PaginatedProducts struct {
	Data       []ProductSummary `json:"data"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	TotalCount int              `json:"total_count"`
}
