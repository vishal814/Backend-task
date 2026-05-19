package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"backend-assignment/internal/models"
	"backend-assignment/internal/store"
)

type ProductHandler struct {
	store *store.ProductStore
}

func NewProductHandler(s *store.ProductStore) *ProductHandler {
	return &ProductHandler{store: s}
}

func (h *ProductHandler) HandleProducts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.createProduct(w, r)
	case http.MethodGet:
		h.listProducts(w, r)
	default:
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (h *ProductHandler) HandleProductDetail(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/products/")
	if id == "" || id == "/" {
		http.NotFound(w, r)
		return
	}

	// If path is /products/{id}/media, handle it
	if strings.HasSuffix(id, "/media") {
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		id = strings.TrimSuffix(id, "/media")
		if id == "" || strings.Contains(id, "/") {
			http.NotFound(w, r)
			return
		}
		h.addMedia(w, r, id)
		return
	}

	if strings.Contains(id, "/") {
		http.NotFound(w, r)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	h.getProduct(w, r, id)
}

func (h *ProductHandler) createProduct(w http.ResponseWriter, r *http.Request) {
	var req models.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.SKU = strings.TrimSpace(req.SKU)

	if req.Name == "" || req.SKU == "" {
		sendError(w, http.StatusBadRequest, "name and sku are required and cannot be empty or just whitespace")
		return
	}

	if len(req.ImageURLs) > 20 || len(req.VideoURLs) > 20 {
		sendError(w, http.StatusBadRequest, "maximum 20 URLs allowed per request")
		return
	}

	if !validateURLs(req.ImageURLs) || !validateURLs(req.VideoURLs) {
		sendError(w, http.StatusBadRequest, "invalid URLs provided")
		return
	}

	id := uuid.New().String()
	p, err := h.store.CreateProduct(id, req.Name, req.SKU, req.ImageURLs, req.VideoURLs)
	if err != nil {
		if err == store.ErrSKUConflict {
			sendError(w, http.StatusConflict, "sku already exists")
			return
		}
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func (h *ProductHandler) listProducts(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	limit := 20

	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	result := h.store.ListProducts(page, limit)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *ProductHandler) getProduct(w http.ResponseWriter, r *http.Request, id string) {
	p, err := h.store.GetProduct(id)
	if err != nil {
		if err == store.ErrNotFound {
			sendError(w, http.StatusNotFound, "product not found")
			return
		}
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *ProductHandler) addMedia(w http.ResponseWriter, r *http.Request, id string) {
	var req models.AddMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if len(req.ImageURLs) == 0 && len(req.VideoURLs) == 0 {
		sendError(w, http.StatusBadRequest, "at least one of image_urls or video_urls must be provided")
		return
	}

	if len(req.ImageURLs) > 20 || len(req.VideoURLs) > 20 {
		sendError(w, http.StatusBadRequest, "maximum 20 URLs allowed per array")
		return
	}

	if !validateURLs(req.ImageURLs) || !validateURLs(req.VideoURLs) {
		sendError(w, http.StatusBadRequest, "invalid URLs provided")
		return
	}

	err := h.store.AddMedia(id, req.ImageURLs, req.VideoURLs)
	if err != nil {
		if err == store.ErrNotFound {
			sendError(w, http.StatusNotFound, "product not found")
			return
		}
		if err.Error() == "maximum 50 URLs per product exceeded" {
			sendError(w, http.StatusBadRequest, err.Error())
			return
		}
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "media appended successfully"})
}

func validateURLs(urls []string) bool {
	for _, u := range urls {
		if len(u) > 2048 {
			return false
		}
		parsed, err := url.ParseRequestURI(u)
		if err != nil {
			return false
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return false
		}
	}
	return true
}

func sendError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
