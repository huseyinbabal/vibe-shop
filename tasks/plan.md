# vibe-shop — Plan

> Kaynak: [SPEC.md](../SPEC.md). Her dilim kendi bölümünde; tamamlanan dilimler silinmez,
> geçmiş referans olarak kalır.

## Dilim 1 — İskelet + `GET /health` ✅ tamamlandı

> Kapsam: **boş iskelet + çalışan `GET /health`**. Ürün/sepet/sipariş bu dilimde YOK.

### Bağımlılık Grafiği

```
go.mod  (T1)
  └──> internal/health/handler.go        (T2)  ── testi: handler_test.go (T3, aynı dilim)
  └──> internal/http/router.go           (T4)  ← health.Handler'a bağlanır
          └──> cmd/server/main.go         (T5)  ← router'ı 8080'de dinler
                  └──> uçtan uca doğrulama (T6)
```

### Fazlar ve Checkpoint'ler

- **Faz 1:** T1 `go.mod` → **CHECKPOINT A:** `go build ./...` başarılı.
- **Faz 2:** T2 `internal/health/handler.go`, T3 `handler_test.go` → **CHECKPOINT B:** `go test ./internal/health/` yeşil.
- **Faz 3:** T4 `internal/http/router.go`, T5 `cmd/server/main.go` → **CHECKPOINT C:** `go run ./cmd/server` + `curl localhost:8080/health` → `{"status":"ok"}`.
- **Faz 4:** T6 Kalite doğrulaması → **CHECKPOINT D (final):** İnsan onayı.

Detaylı görev listesi: [todo.md](./todo.md).

---

## Dilim 2 — Ürün Okuma API'si ← *şu anki dilim*

> Kaynak: [SPEC.md §7](../SPEC.md). Kapsam: **local Postgres (Docker) + GORM ile
> `GET /api/products` ve `GET /api/products/{id}`**. Sepet/sipariş bu dilimde YOK.

### Mimari Kararlar

- **`NewRouter` imzası değişiyor:** `NewRouter() http.Handler` → `NewRouter(products *product.Handler) http.Handler`.
  Health rotası davranışsal olarak değişmiyor; router artık product handler'ı da alıyor.
  `router_test.go`'daki health testi, sahte in-memory bir `product.Repository` üzerinden
  kurulan bir `*product.Handler` ile çağrılacak — gerçek DB gerekmez.
- **`product.Handler` struct + method seti:** `product.NewHandler(repo Repository) *Handler`
  ile `List(w, r)` ve `GetByID(w, r)` metodları (`http.HandlerFunc` imzasında). Handler
  doğrudan `*gorm.DB` değil, `Repository` arayüzü kullanır (SPEC §7.4).
- **Repository arayüzü:** `List(ctx) ([]Product, error)`, `GetByID(ctx, id uint) (Product, error)`.
  Bulunamama durumu `gorm.ErrRecordNotFound`'u sarmalayan `ErrNotFound` ile döner; handler
  `errors.Is` ile yakalayıp 404'e çevirir.
- **Price alanı:** `NUMERIC(10,2)` DB'de, Go tarafında `float64` (`gorm:"type:numeric(10,2)"`).
  Ek bağımlılık gerektirmiyor; hassasiyet notu SPEC'te zaten var.
- **id path param:** Go 1.26 `http.ServeMux` `{id}` pattern'i + `r.PathValue("id")` +
  `strconv.Atoi` — geçersizse 400.
- **Testcontainers:** `testcontainers-go/modules/postgres` resmi modülü, `WithInitScripts`
  ile `migrations/0001_create_products.sql` otomatik uygulanır; seed veri test kodu içinde
  ayrı bir SQL insert ile eklenir (migration dosyasına karışmaz).

### Bağımlılık Grafiği

```
docker-compose.yml (T7)          migrations/0001_create_products.sql (T8)
        │                                    │
        └──────────────┬─────────────────────┘
                        ▼
              CHECKPOINT E (manuel doğrulama: container + şema)

go.mod (gorm, pgx driver, testcontainers-go)  (T9)
        │
        ▼
internal/db/db.go — Connect(dsn) (T10)
        │
        ▼
              CHECKPOINT F (go build ./... temiz)

internal/product/model.go (T11)
        │
        ▼
internal/product/repository.go (T12)  ← db.go'nun ürettiği *gorm.DB tipini kullanır
        │
        ▼
internal/product/handler.go (T13)  ← Repository arayüzüne bağımlı
        │
        ▼
internal/product/handler_test.go (T14)  ← testcontainers ile gerçek DB
        │
        ▼
              CHECKPOINT G (go test ./internal/product/ yeşil, Docker gerekli)

internal/http/router.go güncelle (T15)  ← product.Handler'ı bağlar
        │
        ▼
cmd/server/main.go güncelle (T16)  ← DATABASE_URL okur, db.Connect + repo + handler kurar
        │
        ▼
              CHECKPOINT H (uçtan uca: docker compose + migration + go run + curl)

Kalite kapısı (T17)
        │
        ▼
              CHECKPOINT I (final, insan onayı)
```

### Dikey Dilimleme Yaklaşımı

Katman katman değil (önce tüm altyapı, sonra tüm handler'lar) — tek bir uçtan-uca yol:
**istek → router → product handler → repository → Postgres → JSON yanıt**. Local Postgres
altyapısı (T7-T8) ve Go bağımlılıkları (T9-T10) önkoşul olarak en başta; ardından ürün
domain'i test-first inşa edilir (T11-T14); son olarak kablolama (T15-T16) ve kalite kapısı (T17).

### Fazlar ve Checkpoint'ler

#### Faz 1 — Local Postgres altyapısı
- **T7** `docker-compose.yml` + `.env.example` — `postgres:16-alpine`, `5432:5432`,
  `POSTGRES_USER/PASSWORD/DB=vibeshop`, named volume.
- **T8** `migrations/0001_create_products.sql` — `products(id SERIAL PK, name TEXT NOT NULL, price NUMERIC(10,2) NOT NULL)`.
- **CHECKPOINT E:** `docker compose up -d` healthy + migration `psql` ile uygulanır, `\d products` şemayı gösterir.

#### Faz 2 — Go bağımlılıkları ve DB bağlantısı
- **T9** `go.mod` — `gorm.io/gorm`, `gorm.io/driver/postgres`, `testcontainers-go` (+ `modules/postgres`).
- **T10** `internal/db/db.go` — `Connect(dsn string) (*gorm.DB, error)`.
- **CHECKPOINT F:** `go build ./...` temiz.

#### Faz 3 — Ürün domain'i (test-first)
- **T11** `internal/product/model.go` — `Product{ID, Name, Price}`, migration şemasıyla eşleşir.
- **T12** `internal/product/repository.go` — `Repository` arayüzü + GORM implementasyonu, `ErrNotFound`.
- **T13** `internal/product/handler.go` — `Handler.List`, `Handler.GetByID`; 400/404/500 davranışları.
- **T14** `internal/product/handler_test.go` — testcontainers ile 4 senaryo (200 liste, 200 tekil, 404, 400).
- **CHECKPOINT G:** `go test ./internal/product/` yeşil (Docker gerekli).

#### Faz 4 — Kablolama
- **T15** `internal/http/router.go` + `router_test.go` — `NewRouter(products *product.Handler)`,
  `/api/products`, `/api/products/{id}` rotaları; health testi sahte repository ile güncellenir.
- **T16** `cmd/server/main.go` — `DATABASE_URL` okunur, `db.Connect` → repo → handler → router → `:8080`.
- **CHECKPOINT H:** `docker compose up -d` + migration + `go run ./cmd/server` + `curl` ile
  `/api/products` ve `/api/products/1` gerçek JSON döner.

#### Faz 5 — Kalite kapısı
- **T17** Tüm doğrulamalar: `gofmt -l .` boş, `go vet ./...` temiz, `go test ./...` yeşil (Docker açık).
- **CHECKPOINT I (final):** İnsan onayı → dilim tamam.

### Kapsam Dışı (bilerek, SPEC §7.6 ile uyumlu)
- Yeni migration/şema değişiklikleri (bu dilim yalnızca T8'i içerir).
- GORM dışında ORM/query builder.
- `/api/products` dışında endpoint (sepet/sipariş).
- `docker-compose.yml`'i production için genişletme.

Detaylı görev listesi: [todo.md](./todo.md).
