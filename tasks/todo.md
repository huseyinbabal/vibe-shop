# vibe-shop — Todo

Detaylar: [plan.md](./plan.md) · Spec: [../SPEC.md](../SPEC.md)

## Dilim 1 — İskelet + `GET /health` ✅ tamamlandı

## Faz 1 — Derlenen iskelet
- [x] **T1 — go.mod oluştur**
  - Yapılacak: `module vibe-shop`, `go 1.26`, dış `require` yok.
  - Kabul: dosya var, modül yolu `vibe-shop`, dış bağımlılık yok.
  - Doğrulama: `go build ./...` hatasız.
- [x] **CHECKPOINT A** — `go build ./...` başarılı.

## Faz 2 — Health yolu (test-önce)
- [x] **T2 — internal/health/handler.go**
  - Yapılacak: exported `Handler(http.ResponseWriter, *http.Request)`; `Content-Type: application/json`, status 200, gövde `{"status":"ok"}` (encoding/json ile).
  - Kabul: handler stdlib imzasında, JSON elle string birleştirmeden üretiliyor.
  - Doğrulama: `go build ./internal/health/`.
- [x] **T3 — internal/health/handler_test.go**
  - Yapılacak: `httptest` ile 3 assert — status 200, content-type application/json, gövde `{"status":"ok"}`.
  - Kabul: test health handler'ı gerçek istekle çağırır.
  - Doğrulama: `go test ./internal/health/` yeşil.
- [x] **CHECKPOINT B** — `go test ./internal/health/` yeşil.

## Faz 3 — Kablolama ve çalıştırma
- [x] **T4 — internal/http/router.go**
  - Yapılacak: `http.ServeMux` kuran, `/health` → `health.Handler` bağlayan fonksiyon (ör. `NewRouter() http.Handler`).
  - Kabul: router health paketine bağlanıyor; iş mantığı içermiyor.
  - Doğrulama: `go build ./internal/http/`.
- [x] **T5 — cmd/server/main.go**
  - Yapılacak: router'ı al, `http.ListenAndServe(":8080", ...)`; başlangıç logu; hata dönerse `log.Fatal`.
  - Kabul: giriş noktası yalnızca kablolama yapıyor.
  - Doğrulama: `go build ./cmd/server`.
- [x] **CHECKPOINT C** — `go run ./cmd/server` + `curl -s localhost:8080/health` → `{"status":"ok"}`.

## Faz 4 — Kalite kapısı
- [x] **T6 — Kalite doğrulaması**
  - Doğrulama: `gofmt -l .` boş · `go vet ./...` temiz · `go test ./...` yeşil.
- [x] **CHECKPOINT D (final)** — İnsan onayı; dilim tamam.

## Dilim 2 — Ürün Okuma API'si ✅ tamamlandı (CHECKPOINT I bekliyor: insan onayı)

Detaylar: [plan.md](./plan.md#dilim-2--ürün-okuma-apisi--şu-anki-dilim) · Spec: [../SPEC.md](../SPEC.md) §7

### Faz 1 — Local Postgres altyapısı
- [x] **T7 — docker-compose.yml + .env.example**
  - Yapılacak: `postgres:16-alpine` image, `5432:5432` port eşlemesi, `POSTGRES_USER=vibeshop`,
    `POSTGRES_PASSWORD=vibeshop`, `POSTGRES_DB=vibeshop`, named volume. `.env.example` içinde
    eşleşen `DATABASE_URL`.
  - Kabul: `docker compose up -d` container'ı sağlıklı başlatır.
  - Doğrulama: `docker compose up -d && docker compose ps` → `healthy`.
- [x] **T8 — migrations/0001_create_products.sql**
  - Yapılacak: `products` tablosu — `id SERIAL PRIMARY KEY`, `name TEXT NOT NULL`,
    `price NUMERIC(10,2) NOT NULL`.
  - Kabul: script hatasız çalışır, şema migration'la birebir.
  - Doğrulama: `psql "$DATABASE_URL" -f migrations/0001_create_products.sql`; `\d products`.
- [x] **CHECKPOINT E** — Container ayakta + tablo oluşturulmuş (manuel `psql` doğrulaması).

### Faz 2 — Go bağımlılıkları ve DB bağlantısı
- [x] **T9 — go.mod güncelle**
  - Yapılacak: `gorm.io/gorm`, `gorm.io/driver/postgres`, `github.com/testcontainers/testcontainers-go`,
    `.../modules/postgres` eklenir.
  - Kabul: `go mod tidy` temiz.
  - Doğrulama: `go build ./...` hatasız.
- [x] **T10 — internal/db/db.go**
  - Yapılacak: `Connect(dsn string) (*gorm.DB, error)`; hata `fmt.Errorf("db: connect: %w", err)`
    ile sarmalanır.
  - Kabul: paket yalnızca bağlantı sorumluluğu taşır.
  - Doğrulama: `go build ./internal/db/`.
- [x] **CHECKPOINT F** — `go build ./...` temiz.

### Faz 3 — Ürün domain'i (test-first)
- [x] **T11 — internal/product/model.go**
  - Yapılacak: `Product{ID uint, Name string, Price float64}`, GORM tag'leriyle `products`
    tablosuna eşlenir.
  - Kabul: alanlar migration şemasıyla birebir eşleşir.
  - Doğrulama: `go build ./internal/product/`.
- [x] **T12 — internal/product/repository.go**
  - Yapılacak: `Repository` arayüzü (`List(ctx) ([]Product, error)`, `GetByID(ctx, id uint) (Product, error)`)
    + GORM implementasyonu (`NewRepository(db *gorm.DB) Repository`); bulunamama `ErrNotFound`
    ile sarmalanır.
  - Kabul: handler `errors.Is(err, ErrNotFound)` ile 404 kararı verebilir.
  - Doğrulama: `go build ./internal/product/`.
- [x] **T13 — internal/product/handler.go**
  - Yapılacak: `Handler{repo Repository}`, `NewHandler(repo Repository) *Handler`, `List(w, r)`
    ve `GetByID(w, r)` metodları. Geçersiz id → 400; `ErrNotFound` → 404; diğer hatalar → 500
    (hepsi JSON gövdeli).
  - Kabul: SPEC §7.4/§7.6 davranışlarıyla birebir uyum.
  - Doğrulama: `go build ./internal/product/`.
- [x] **T14 — internal/product/handler_test.go**
  - Yapılacak: `testcontainers-go/modules/postgres` ile gerçek container, `WithInitScripts`
    migration'ı uygular, test seed verisi ekler. 4 senaryo: liste (200), tekil (200), olmayan
    id (404), geçersiz id (400).
  - Kabul: mock repository yok, gerçek Postgres'e karşı.
  - Doğrulama: `go test ./internal/product/` (Docker gerekli) yeşil.
- [x] **CHECKPOINT G** — `go test ./internal/product/` yeşil (Docker açık olmalı).

### Faz 4 — Kablolama
- [x] **T15 — internal/http/router.go + router_test.go**
  - Yapılacak: `NewRouter(products *product.Handler) http.Handler` — `/health` (mevcut) +
    `GET /api/products` → `products.List`, `GET /api/products/{id}` → `products.GetByID`.
    Health testi sahte in-memory `product.Repository` ile kurulan `*product.Handler` kullanır.
  - Kabul: health davranışı değişmedi; imza değişikliği tüm çağıranlarda güncellendi.
  - Doğrulama: `go build ./internal/http/` ve `go test ./internal/http/` yeşil.
- [x] **T16 — cmd/server/main.go güncelle**
  - Yapılacak: `DATABASE_URL` env'den okunur (yoksa `log.Fatal`); `db.Connect` →
    `product.NewRepository` → `product.NewHandler` → `apphttp.NewRouter(handler)` → `:8080`.
  - Kabul: `main.go` yalnızca kablolama yapar.
  - Doğrulama: `go build ./cmd/server`; manuel `curl` ile uçtan uca test.
- [x] **CHECKPOINT H** — `docker compose up -d` + migration + `go run ./cmd/server` +
  `curl localhost:8080/api/products` ve `curl localhost:8080/api/products/1` gerçek JSON döner.

### Faz 5 — Kalite kapısı
- [x] **T17 — Kalite doğrulaması**
  - Doğrulama: `gofmt -l .` boş · `go vet ./...` temiz · `go test ./...` yeşil (Docker açık).
- [ ] **CHECKPOINT I (final)** — İnsan onayı; dilim tamam.

## Dilim 3 — Ürün Yazma API'si ← *şu anki dilim*

Detaylar: [plan.md](./plan.md#dilim-3--ürün-yazma-apisi--şu-anki-dilim) · Spec: [../SPEC.md](../SPEC.md) §8

### Faz 1 — Ürün domain'i: yazma yolları (test-first)
- [x] **T18 — internal/product/model.go: Input + Validate**
  - Yapılacak: `Input{Name string, Price float64}` tipi + `Validate() error` — `name` kırpılınca
    boş olamaz ve ≤ 200 karakter; `price > 0`, en fazla iki ondalık (`math.Round(price*100)`
    karşılaştırması), üst sınır `99_999_999.99`. Hata mesajları alan bazlı.
  - Kabul: POST ve PUT'un paylaşacağı tek doğrulama noktası bu; `Product` struct'ı değişmez.
  - Doğrulama: `go build ./internal/product/`.
- [x] **T19 — internal/product/repository.go: Create/Update/Delete**
  - Yapılacak: `Repository` arayüzüne `Create(ctx, Product) (Product, error)`,
    `Update(ctx, Product) (Product, error)`, `Delete(ctx, id uint) error` eklenir; GORM
    implementasyonunda Update/Delete `RowsAffected == 0` → `ErrNotFound`; hatalar
    `fmt.Errorf("product: ...: %w", err)` ile sarmalanır.
  - Kabul: read metodlarının deseniyle birebir aynı stil; ekstra SELECT yok.
  - Doğrulama: `go build ./internal/product/`.
- [x] **T20 — internal/product/handler.go: Create/Update/Delete metodları**
  - Yapılacak: `Create` (`201` + oluşan ürün), `Update` (`200` + son hali), `Delete` (`204`,
    boş gövde). Gövde `Input`'a decode edilir (`id` doğal olarak yok sayılır); geçersiz JSON →
    `400 "invalid JSON body"`; `Validate` hatası → `400` + mesaj; sayısal olmayan id → `400`;
    `ErrNotFound` → `404 "product not found"`; diğer → `500`. Mevcut `writeJSON`/`writeError`
    kullanılır.
  - Kabul: SPEC §8.1/§8.4 davranışlarıyla birebir uyum; yeni yanıt üretme yolu yok.
  - Doğrulama: `go build ./internal/product/`.
- [x] **T21 — internal/product/handler_test.go: yazma senaryoları**
  - Yapılacak: SPEC §8.5 senaryoları — POST geçerli → `201` + `id`, ardından GET aynı ürünü döner;
    POST geçersiz JSON / boş `name` / `price <= 0` / 200+ karakter `name` → `400`;
    PUT var olan id → `200` + DB güncel, doğrulama hataları POST ile aynı → `400`;
    PUT/DELETE olmayan id → `404`, sayısal olmayan id → `400`;
    DELETE var olan id → `204` + boş gövde, ardından GET → `404`.
    Yazma testleri kendi satırlarını oluşturur, seed satırlarına (id 1-2) dokunmaz,
    `t.Cleanup` ile temizler; `TestList_ReturnsSeededProducts` "tam 2" yerine
    "Widget ve Gadget mevcut" olarak gevşetilir.
  - Kabul: testler sıradan bağımsız; mock repository yok, gerçek Postgres'e karşı.
  - Doğrulama: `go test ./internal/product/` (Docker gerekli) yeşil.
- [x] **CHECKPOINT J** — `go test ./internal/product/` yeşil (Docker açık olmalı).

### Faz 2 — Kablolama
- [x] **T22 — internal/http/router.go + router_test.go**
  - Yapılacak: `"POST /api/products"` → `products.Create`, `"PUT /api/products/{id}"` →
    `products.Update`, `"DELETE /api/products/{id}"` → `products.Delete` rotaları eklenir;
    `NewRouter` imzası değişmez. Router testine yeni rotaların bağlandığını gösteren
    asgari kontroller eklenir (sahte repository ile, DB'siz).
  - Kabul: health ve GET rotaları davranış değiştirmez.
  - Doğrulama: `go build ./internal/http/` ve `go test ./internal/http/` yeşil.
- [x] **CHECKPOINT K** — Uçtan uca: `docker compose up -d` + `go run ./cmd/server`;
  `curl -X POST` → `201` + listede görünür; `curl -X PUT` → `200` + değişiklik yansır;
  `curl -X DELETE` → `204` + ardından `GET` → `404`; `/health` → `{"status":"ok"}`.

### Faz 3 — Kalite kapısı
- [x] **T23 — Kalite doğrulaması**
  - Doğrulama: `gofmt -l .` boş · `go vet ./...` temiz · `go test ./...` yeşil (Docker açık).
- [ ] **CHECKPOINT L (final)** — İnsan onayı; dilim tamam.
