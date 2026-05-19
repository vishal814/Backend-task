# Source Asia Backend Assignment

This repository contains the backend assignment for Source Asia, implemented in **Go**.

## Requirements Met
- **Language**: Go
- **Storage**: In-memory (maps, slices) with safe concurrent access using `sync.RWMutex`.
- **Framework**: `net/http` standard library (Go 1.22+).
- **Concurrency**: Fully thread-safe rate limiter and product catalog.

## AI Tools Acknowledgment
AI tools were used during the development of this project to generate boilerplate code, aid in structuring the Go project according to standard practices, and help format this README.

---

## Part 1: Rate-limited API

### Design Decisions
- **Algorithm**: Implemented a **Fixed 1-minute window** algorithm by truncating the current `time.Now()` to the minute. If the truncated time changes, the accepted requests count resets.
- **POST `/request` Success Response**: Chosen **200 OK** returning a JSON message `{"message": "Request accepted"}`.
- **GET `/stats` Rejections**: Rejections are recorded **cumulatively** over the lifetime of the application per user, while accepted requests reflect only the **current 1-minute window**.

### Production Limitations (Part 1)
This implementation uses an in-memory map.
1. **Single Instance**: It only works correctly if deployed as a single instance. In a multi-instance deployment (like Kubernetes), the rate limit would effectively become `5 * N` instances because memory is not shared.
2. **State Loss**: If the server restarts, all rate limiting state (and stats) is lost.
3. **Memory Leak Risk**: Over a very long time, if millions of unique `user_id`s make requests, the map will grow unbounded. A production solution would require a TTL (Time-to-Live) cleanup mechanism.
4. **Production Solution**: In production, I would use **Redis** (e.g., using the INCR command and EXPIRE) to maintain a distributed rate limiter that works across multiple instances and persists (or at least shares state).

---

## Part 2: Product Catalog with Media

### Design Decisions
- **Data Model (Memory)**: 
  - `map[string]*Product` for O(1) lookups by ID.
  - `map[string]struct{}` for fast O(1) duplicate SKU validation on creation.
  - `[]string` array holding the sequential IDs to allow for stable pagination (offset/limit).
- **List vs Detail Queries**:
  - `GET /products` returns a lightweight `ProductSummary` struct. Instead of serializing all arrays of URLs, it only calculates the lengths (`image_count`, `video_count`) and grabs the first image as `thumbnail_url`. This ensures O(1) serialization time per product regarding media, keeping it extremely fast.
  - `GET /products/{id}` returns the full `Product` struct including all media URL arrays.
- **Duplicate SKU**: Chosen to return **409 Conflict**.

### Production Changes (PostgreSQL + CDN)
If moving to a real database like PostgreSQL and a CDN:
1. **Database Schema**: 
   - `products` table (id, name, sku, created_at).
   - `product_images` table (id, product_id, url, position).
   - `product_videos` table (id, product_id, url).
2. **List Endpoint Optimization**: 
   - Instead of fetching all media, the list query would `JOIN` with aggregations (e.g., `COUNT(image.id)`) and only select the first image using a `LATERAL JOIN` or a `thumbnail_url` cache column on the `products` table.
3. **CDN Integration**: 
   - The API would generate pre-signed upload URLs (e.g., S3/CloudFront) for clients to upload directly to the CDN. 
   - The API would store only the relative path or CDN URL in the database after successful upload.

---

## How to Run the Server

Requires Go 1.22 or higher.

```bash
# Run the server directly
go run cmd/server/main.go

# Or build and run
go build -o server.exe ./cmd/server/main.go
./server.exe
```

The server will start on `http://localhost:8080`.

---

## Example `curl` Commands

### Part 1: Rate Limiter

**Send a request (Run 6 times to see 429 Too Many Requests):**
```bash
curl -X POST http://localhost:8080/request \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user123", "payload": {"foo": "bar"}}'
```

**Get Stats:**
```bash
curl http://localhost:8080/stats
```

### Part 2: Product Catalog

**Create a Product:**
```bash
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{
        "name": "Widget A",
        "sku": "SKU-001",
        "image_urls": ["https://cdn.example.com/img1.jpg"],
        "video_urls": []
      }'
```

**List Products (Pagination):**
```bash
curl "http://localhost:8080/products?page=1&limit=20"
```

**Get Product Detail:**
```bash
# Replace with the UUID returned from the POST /products response
curl http://localhost:8080/products/{id}
```

**Append Media:**
```bash
curl -X POST http://localhost:8080/products/{id}/media \
  -H "Content-Type: application/json" \
  -d '{
        "image_urls": ["https://cdn.example.com/img2.jpg"]
      }'
```
