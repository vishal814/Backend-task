$ErrorActionPreference = "Continue"

Write-Host "1. Testing Rate Limiter (POST /request)" -ForegroundColor Cyan
for ($i = 1; $i -le 6; $i++) {
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:8080/request" -Method Post -ContentType "application/json" -Body '{"user_id": "test_user", "payload": "hello"}' -UseBasicParsing
        Write-Host "Request $i - Status: $($response.StatusCode)"
    } catch {
        Write-Host "Request $i - Status: $($_.Exception.Response.StatusCode.value__)"
    }
}

Write-Host "`n2. Getting Stats (GET /stats)" -ForegroundColor Cyan
$stats = Invoke-RestMethod -Uri "http://localhost:8080/stats" -UseBasicParsing
$stats | ConvertTo-Json

Write-Host "`n3. Creating a Product (POST /products)" -ForegroundColor Cyan
$productBody = @{
    name = "Test Product"
    sku = "TEST-001"
    image_urls = @("http://example.com/1.jpg")
} | ConvertTo-Json
$product = Invoke-RestMethod -Uri "http://localhost:8080/products" -Method Post -ContentType "application/json" -Body $productBody -UseBasicParsing
$product | ConvertTo-Json
$productId = $product.id

Write-Host "`n4. Listing Products (GET /products?page=1&limit=10)" -ForegroundColor Cyan
$list = Invoke-RestMethod -Uri "http://localhost:8080/products?page=1&limit=10" -UseBasicParsing
$list | ConvertTo-Json -Depth 3

Write-Host "`n5. Getting Product Detail (GET /products/$productId)" -ForegroundColor Cyan
$detail = Invoke-RestMethod -Uri "http://localhost:8080/products/$productId" -UseBasicParsing
$detail | ConvertTo-Json

Write-Host "`n6. Appending Media (POST /products/$productId/media)" -ForegroundColor Cyan
$mediaBody = @{
    video_urls = @("http://example.com/demo.mp4")
} | ConvertTo-Json
$mediaResult = Invoke-RestMethod -Uri "http://localhost:8080/products/$productId/media" -Method Post -ContentType "application/json" -Body $mediaBody -UseBasicParsing
$mediaResult | ConvertTo-Json

Write-Host "`n7. Verifying Product Detail Again (GET /products/$productId)" -ForegroundColor Cyan
$detail2 = Invoke-RestMethod -Uri "http://localhost:8080/products/$productId" -UseBasicParsing
$detail2 | ConvertTo-Json
