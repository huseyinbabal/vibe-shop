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

## Dilim 3 — Ürün Yazma API'si ✅ tamamlandı

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

---

## Dilim 4 — Auth + Sepet + Sipariş ← *şu anki dilim*

Detaylar: [plan.md](./plan.md#dilim-4--auth--sepet--sipariş--şu-anki-dilim) · Spec: [../SPEC.md](../SPEC.md) §9

### Faz 0 — Paylaşılan altyapı
- [x] **T24 — internal/httpx: WriteJSON/WriteError**
  - Yapılacak: yeni `internal/httpx/json.go` — `WriteJSON(w, status, body any)` ve
    `WriteError(w, status, message string)` (exported), mevcut ürün yardımcılarıyla **birebir aynı**
    davranış (`Content-Type: application/json`, `{"error":"..."}`). `internal/product/handler.go`
    bunları çağıracak şekilde güncellenir; yerel `writeJSON`/`writeError` kaldırılır.
  - Kabul: ürün uçlarının davranışı **değişmez**; yardımcılar diğer paketlerce import edilebilir.
  - Doğrulama: `go build ./...` · `go test ./internal/product/` yeşil (Docker) · `gofmt`/`vet` temiz.
  - Dosyalar: `internal/httpx/json.go` (yeni), `internal/product/handler.go`. **Kapsam: S**

### Faz 1 — Auth dikey dilimi
- [x] **T25 — users şeması + auth repository**
  - Yapılacak: `migrations/0002_create_users.sql` — `users(id, email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL, created_at)`. `internal/auth/model.go` — `User` (GORM,
    `TableName()="users"`, `password_hash` alanı `json:"-"`). `internal/auth/repository.go` —
    `Repository{Create(ctx,User)(User,error); GetByEmail(ctx,email)(User,error)}`; GORM impl;
    unique ihlali → `ErrEmailTaken`, bulunamazsa → `ErrNotFound`.
  - Kabul: migration uygulanır; create + get-by-email çalışır; duplicate email ayrı hata döner.
  - Doğrulama: `go build ./internal/auth/` · `internal/auth/repository_test.go` (testcontainers, users migration) yeşil.
  - Dosyalar: `migrations/0002_create_users.sql`, `internal/auth/model.go`, `repository.go`, `repository_test.go`. **Kapsam: M**
  - Bağımlılık: T24.
- [x] **T26 — JWT + auth middleware**
  - Yapılacak: `internal/auth/token.go` — golang-jwt/jwt/v5 ile secret+ttl'e bağlı `Issue(userID uint)
    (string,error)` ve `Parse(tokenStr string)(uint,error)` (HS256, `sub`, `exp`).
    `internal/auth/middleware.go` — `RequireAuth`: `Authorization: Bearer` oku → `Parse` → başarıda
    `user_id`'yi context'e koy, aksi halde `httpx.WriteError(401)`. Exported
    `UserIDFromContext(ctx)(uint,bool)`. `go get` ile jwt bağımlılığı eklenir.
  - Kabul: token round-trip aynı `userID`; süresi geçmiş/bozuk/yanlış secret → hata; token'sız/bozuk istek → `401`; geçerli istek context'te doğru `user_id`.
  - Doğrulama: `token_test.go` (container'sız — round-trip, expired, tampered, malformed) + `middleware_test.go` (header yok→401, bozuk→401, geçerli→context+next) yeşil.
  - Dosyalar: `internal/auth/token.go`, `token_test.go`, `middleware.go`, `middleware_test.go`. **Kapsam: M**
  - Bağımlılık: T24.
- [x] **T27 — register/login handler + kablolama**
  - Yapılacak: `internal/auth/handler.go` — `registerInput{Email,Password}`, `loginInput{...}`.
    `Register`: email boş değil/`@` içerir, parola ≥ 8 (aksi `400`); bcrypt hash; `repo.Create`;
    duplicate → `409 {"error":"email already registered"}`; başarı `201 {"id","email"}`.
    `Login`: email'e göre getir + bcrypt compare; başarı → JWT `200 {"token"}`; yanlış/eksik →
    `401 {"error":"invalid credentials"}`. `go get golang.org/x/crypto/bcrypt`. `router.go`'ya
    public `POST /api/register`, `POST /api/login`; `main.go` `JWT_SECRET` okur (boşsa fatal),
    token issuer/parser + auth handler kurar, `NewRouter` çağrısını günceller.
  - Kabul: register `201` (parola dönmez), dup `409`, geçersiz email/kısa parola `400`; login geçerli JWT, yanlış kimlik `401`; boot'ta `JWT_SECRET` zorunlu.
  - Doğrulama: `internal/auth/handler_test.go` (testcontainers) tüm senaryolar yeşil · `go build ./...`.
  - Dosyalar: `internal/auth/handler.go`, `handler_test.go`, `internal/http/router.go`, `cmd/server/main.go`. **Kapsam: M**
  - Bağımlılık: T25, T26.
- [x] **CHECKPOINT M** — Uçtan uca: `make start` (veya `docker compose up -d` + migrationlar +
  `JWT_SECRET=dev-secret ... go run ./cmd/server`); `curl register` → `201`; `curl login` →
  `{"token":...}`; token'sız `POST /api/cart` (henüz yoksa geçici korumalı bir uçla) → `401`.

### Faz 2 — Sepet dikey dilimi
- [x] **T28 — cart şeması + cart repository**
  - Yapılacak: `migrations/0003_create_cart.sql` — `cart(id, user_id INT NOT NULL REFERENCES
    users(id), product_id INT NOT NULL REFERENCES products(id), quantity INT NOT NULL CHECK
    (quantity>0), UNIQUE(user_id, product_id))`. `internal/cart/model.go` — `CartItem` +
    `GET` için ürün ad/fiyat/satır-toplamı taşıyan görünüm tipi. `internal/cart/repository.go` —
    `AddOrIncrement(ctx,userID,productID,qty)` upsert (`ON CONFLICT (user_id,product_id) DO UPDATE
    quantity = cart.quantity + EXCLUDED.quantity`); olmayan ürün → `ErrProductNotFound`;
    `ListByUser(ctx,userID)` ürünle join; `ClearByUser(ctx,userID)`.
  - Kabul: ekleme satır oluşturur; tekrar ekleme `quantity` artırır (yeni satır yok); olmayan ürün ayrı hata; list yalnızca o kullanıcının satırları.
  - Doğrulama: `repository_test.go` (testcontainers: products+users+cart, seed ürün+kullanıcı) — increment, izolasyon, olmayan ürün — yeşil.
  - Dosyalar: `migrations/0003_create_cart.sql`, `internal/cart/model.go`, `repository.go`, `repository_test.go`. **Kapsam: M**
  - Bağımlılık: T25 (users), 0001_products.
- [ ] **T29 — cart handler + rotalar**
  - Yapılacak: `internal/cart/handler.go` — `cartInput{ProductID uint, Quantity int}`; `Add`:
    `userID` context'ten, `quantity>0` değilse `400`, olmayan ürün → `404`, başarı `201` + kalem.
    `Get`: `userID` context'ten, `repo.ListByUser`, `200` + `{items,total}`. `router.go`'ya auth
    mw arkasında `POST /api/cart`, `GET /api/cart`; `main.go` cart handler'ı kablolar.
  - Kabul: `POST` `201`; tekrar ekleme artırır; kötü qty `400`; olmayan ürün `404`; token'sız `401`; `GET` doğru toplam; A, B'nin sepetini görmez.
  - Doğrulama: `handler_test.go` (testcontainers, iki kullanıcı/token dahil izolasyon) yeşil · `go build ./...`.
  - Dosyalar: `internal/cart/handler.go`, `handler_test.go`, `internal/http/router.go`, `cmd/server/main.go`. **Kapsam: M**
  - Bağımlılık: T26 (mw), T28.
- [ ] **CHECKPOINT N** — Uçtan uca: register→login→`POST /api/cart`→`GET /api/cart` doğru toplam;
  ikinci kullanıcı izolasyonu; token'sız `401`.

### Faz 3 — Sipariş dikey dilimi
- [ ] **T30 — orders/order_items şeması + order repository (transaction + snapshot)**
  - Yapılacak: `migrations/0004_create_orders.sql` — `orders(id, user_id REFERENCES users(id),
    total NUMERIC(10,2), created_at)`. `migrations/0005_create_order_items.sql` —
    `order_items(id, order_id REFERENCES orders(id), product_id REFERENCES products(id),
    quantity INT, unit_price NUMERIC(10,2))`. `internal/order/model.go` — `Order`+`OrderItem`.
    `internal/order/repository.go` — `CreateFromCart(ctx,userID)(Order,error)`: gorm transaction —
    sepeti o anki ürün fiyatıyla oku; boşsa `ErrCartEmpty`; `orders` ekle; her kalemi
    `order_items`'a `unit_price` = o anki fiyat snapshot'la; `total` = Σ `qty*unit_price`; sepeti sil.
  - Kabul: order+items oluşur; `unit_price` = sipariş anındaki fiyat; `total` doğru; sepet boşalır; boş sepet → `ErrCartEmpty`; atomik (hata → rollback, kısmi sipariş yok).
  - Doğrulama: `repository_test.go` (testcontainers, 5 migration, seed) — snapshot ürün fiyatı sonradan değişse de sabit; sepet boşalır; boş-sepet hatası; izolasyon — yeşil.
  - Dosyalar: `migrations/0004_create_orders.sql`, `0005_create_order_items.sql`, `internal/order/model.go`, `repository.go`, `repository_test.go`. **Kapsam: M**
  - Bağımlılık: T28 (cart), T25.
- [ ] **T31 — order handler + rota**
  - Yapılacak: `internal/order/handler.go` — `Create`: `userID` context'ten; `repo.CreateFromCart`;
    `ErrCartEmpty` → `400 {"error":"cart is empty"}`; başarı `201` + sipariş (items+total).
    `router.go`'ya auth mw arkasında `POST /api/orders`; `main.go` order handler'ı kablolar.
  - Kabul: `201` + order+items+total; boş sepet `400`; token'sız `401`; fiyat snapshot; sonrasında sepet boş; izolasyon.
  - Doğrulama: `handler_test.go` (testcontainers) yeşil · `go build ./...`.
  - Dosyalar: `internal/order/handler.go`, `handler_test.go`, `internal/http/router.go`, `cmd/server/main.go`. **Kapsam: M**
  - Bağımlılık: T26 (mw), T30.
- [ ] **CHECKPOINT O** — Uçtan uca: register→login→sepete ekle→`POST /api/orders` `201`→sepet boş
  (`GET /api/cart` boş); boş sepette `400`; ikinci kullanıcı izolasyonu.

### Faz 4 — Kalite kapısı
- [ ] **T32 — Kalite doğrulaması**
  - Yapılacak: `go mod tidy` (yeni bağımlılıklar düzgün kaydolsun); istenirse `api.http`'ye yeni
    uçların örnekleri eklenir.
  - Doğrulama: `gofmt -l .` boş · `go vet ./...` temiz · `go test ./...` yeşil (Docker açık).
- [ ] **CHECKPOINT P (final)** — İnsan onayı; dilim tamam.
