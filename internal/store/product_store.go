package store

import (
	"errors"
	"sync"
	"time"

	"backend-assignment/internal/models"
)

var (
	ErrSKUConflict = errors.New("sku already exists")
	ErrNotFound    = errors.New("product not found")
)

type ProductStore struct {
	mu       sync.RWMutex
	products map[string]*models.Product
	skus     map[string]struct{} // For fast duplicate checks
	ids      []string            // Maintain order for pagination
}

func NewProductStore() *ProductStore {
	return &ProductStore{
		products: make(map[string]*models.Product),
		skus:     make(map[string]struct{}),
		ids:      make([]string, 0),
	}
}

func (s *ProductStore) CreateProduct(id, name, sku string, images, videos []string) (*models.Product, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.skus[sku]; exists {
		return nil, ErrSKUConflict
	}

	p := &models.Product{
		ID:        id,
		Name:      name,
		SKU:       sku,
		ImageURLs: images,
		VideoURLs: videos,
		CreatedAt: time.Now(),
	}

	if p.ImageURLs == nil {
		p.ImageURLs = []string{}
	}
	if p.VideoURLs == nil {
		p.VideoURLs = []string{}
	}

	s.products[id] = p
	s.skus[sku] = struct{}{}
	s.ids = append(s.ids, id)

	return p, nil
}

func (s *ProductStore) GetProduct(id string) (*models.Product, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, exists := s.products[id]
	if !exists {
		return nil, ErrNotFound
	}
	return p, nil
}

func (s *ProductStore) AddMedia(id string, newImages, newVideos []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, exists := s.products[id]
	if !exists {
		return ErrNotFound
	}

	if len(p.ImageURLs)+len(newImages) > 50 || len(p.VideoURLs)+len(newVideos) > 50 {
		return errors.New("maximum 50 URLs per product exceeded")
	}

	p.ImageURLs = append(p.ImageURLs, newImages...)
	p.VideoURLs = append(p.VideoURLs, newVideos...)

	return nil
}

// ListProducts returns a paginated list of product summaries.
// This satisfies the requirement to NOT load all image URLs for lists.
func (s *ProductStore) ListProducts(page, limit int) models.PaginatedProducts {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalCount := len(s.ids)
	
	start := (page - 1) * limit
	if start < 0 {
		start = 0
	}
	if start > totalCount {
		start = totalCount
	}
	
	end := start + limit
	if end > totalCount {
		end = totalCount
	}

	summaries := make([]models.ProductSummary, 0, end-start)

	for _, id := range s.ids[start:end] {
		p := s.products[id]
		
		thumb := ""
		if len(p.ImageURLs) > 0 {
			thumb = p.ImageURLs[0]
		}
		
		summaries = append(summaries, models.ProductSummary{
			ID:           p.ID,
			Name:         p.Name,
			SKU:          p.SKU,
			ImageCount:   len(p.ImageURLs),
			VideoCount:   len(p.VideoURLs),
			ThumbnailURL: thumb,
		})
	}

	return models.PaginatedProducts{
		Data:       summaries,
		Page:       page,
		Limit:      limit,
		TotalCount: totalCount,
	}
}
