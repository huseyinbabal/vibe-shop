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

## Dilim 2 — Ürün Okuma API'si ✅ tamamlandı

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
- [x] **CHECKPOINT I (final)** — İnsan onayı alındı (2026-07-13); dilim tamam.

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
- [x] **CHECKPOINT L (final)** — İnsan onayı alındı (2026-07-13); dilim tamam.

---

## Dilim 4 — Auth + Sepet + Sipariş ✅ tamamlandı

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
- [x] **T29 — cart handler + rotalar**
  - Yapılacak: `internal/cart/handler.go` — `cartInput{ProductID uint, Quantity int}`; `Add`:
    `userID` context'ten, `quantity>0` değilse `400`, olmayan ürün → `404`, başarı `201` + kalem.
    `Get`: `userID` context'ten, `repo.ListByUser`, `200` + `{items,total}`. `router.go`'ya auth
    mw arkasında `POST /api/cart`, `GET /api/cart`; `main.go` cart handler'ı kablolar.
  - Kabul: `POST` `201`; tekrar ekleme artırır; kötü qty `400`; olmayan ürün `404`; token'sız `401`; `GET` doğru toplam; A, B'nin sepetini görmez.
  - Doğrulama: `handler_test.go` (testcontainers, iki kullanıcı/token dahil izolasyon) yeşil · `go build ./...`.
  - Dosyalar: `internal/cart/handler.go`, `handler_test.go`, `internal/http/router.go`, `cmd/server/main.go`. **Kapsam: M**
  - Bağımlılık: T26 (mw), T28.
- [x] **CHECKPOINT N** (entegrasyon testleriyle kanıtlandı; canlı smoke CHECKPOINT O ile) — Uçtan uca: register→login→`POST /api/cart`→`GET /api/cart` doğru toplam;
  ikinci kullanıcı izolasyonu; token'sız `401`.

### Faz 3 — Sipariş dikey dilimi
- [x] **T30 — orders/order_items şeması + order repository (transaction + snapshot)**
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
- [x] **T31 — order handler + rota**
  - Yapılacak: `internal/order/handler.go` — `Create`: `userID` context'ten; `repo.CreateFromCart`;
    `ErrCartEmpty` → `400 {"error":"cart is empty"}`; başarı `201` + sipariş (items+total).
    `router.go`'ya auth mw arkasında `POST /api/orders`; `main.go` order handler'ı kablolar.
  - Kabul: `201` + order+items+total; boş sepet `400`; token'sız `401`; fiyat snapshot; sonrasında sepet boş; izolasyon.
  - Doğrulama: `handler_test.go` (testcontainers) yeşil · `go build ./...`.
  - Dosyalar: `internal/order/handler.go`, `handler_test.go`, `internal/http/router.go`, `cmd/server/main.go`. **Kapsam: M**
  - Bağımlılık: T26 (mw), T30.
- [x] **CHECKPOINT O** — Uçtan uca: register→login→sepete ekle→`POST /api/orders` `201`→sepet boş
  (`GET /api/cart` boş); boş sepette `400`; ikinci kullanıcı izolasyonu.
  (Canlı curl akışıyla doğrulandı; local 5432 başka projede olduğundan Postgres geçici olarak 5455 portundan çalıştırıldı.)

### Faz 4 — Kalite kapısı
- [x] **T32 — Kalite doğrulaması**
  - Yapılacak: `go mod tidy` (yeni bağımlılıklar düzgün kaydolsun); istenirse `api.http`'ye yeni
    uçların örnekleri eklenir.
  - Doğrulama: `gofmt -l .` boş · `go vet ./...` temiz · `go test ./...` yeşil (Docker açık).
- [x] **CHECKPOINT P (final)** — İnsan onayı alındı (2026-07-13); dilim tamam.

---

## Dilim 5 — Keycloak'a Geçiş: Tek Kimlik Sağlayıcı ← *şu anki dilim*

Detaylar: [plan.md](./plan.md#dilim-5--keycloaka-geçiş-tek-kimlik-sağlayıcı--şu-anki-dilim) · Spec: [../SPEC.md](../SPEC.md) §10

### Faz 1 — Keycloak local altyapısı
- [x] **T33 — docker-compose'a Keycloak + realm import + .env.example**
  - Yapılacak: `docker-compose.yml`'e `keycloak` servisi — `quay.io/keycloak/keycloak:26.3`,
    komut `start-dev --import-realm`, port `8081:8080`, env `KC_BOOTSTRAP_ADMIN_USERNAME=admin`,
    `KC_BOOTSTRAP_ADMIN_PASSWORD=admin`, `./keycloak/` klasörü `/opt/keycloak/data/import/`'a
    mount. Yeni `keycloak/vibe-shop-realm.json` — realm `vibe-shop` (enabled), public client
    `vibe-shop-api` (`publicClient: true`, `directAccessGrantsEnabled: true`), **iki** kullanıcı:
    `testuser` ve `testuser2` (enabled, emailVerified, kalıcı parola `test1234`). `.env.example`'a
    `KEYCLOAK_ISSUER_URL=http://localhost:8081/realms/vibe-shop` eklenir.
  - Kabul: `docker compose up -d` tek komutla Postgres **ve** Keycloak'ı başlatır; elle admin
    konsolu adımı gerekmez; `http://localhost:8081/realms/vibe-shop` `200` döner; her iki
    kullanıcıyla token alınabilir; Postgres servisi/portu değişmez.
  - Doğrulama: SPEC §10.2'deki parola-grant curl'ü `access_token` içeren JSON döner.
  - Dosyalar: `docker-compose.yml`, `keycloak/vibe-shop-realm.json` (yeni), `.env.example`. **Kapsam: S**
- [x] **CHECKPOINT Q** — Keycloak ayakta, realm import edilmiş, curl ile token alınıyor (manuel).

### Faz 2 — Token doğrulama çekirdeği (test-first, container gerekmez)
- [x] **T34 — internal/auth/keycloak.go: KeycloakVerifier**
  - Yapılacak: `go get github.com/MicahParks/keyfunc/v3`. `NewKeycloakVerifier(issuerURL string)
    (*KeycloakVerifier, error)` — JWKS URL'i `<issuer>/protocol/openid-connect/certs`'ten türetir,
    keyfunc ile çeker (erişilemezse hata döner; `main.go` fatal'ler — T37).
    `Verify(tokenStr string) (string, error)` — `jwt.Parse` +
    `jwt.WithValidMethods([]string{"RS256"})` + `iss == issuerURL` + boş olmayan `sub`; `exp`
    kütüphanece zorlanır; başarıda `sub` döner, her hata `ErrInvalidToken`'a katlanır
    (eski `TokenManager.Parse` deseni). Eski auth koduna bu görevde **dokunulmaz** (söküm T37).
  - Kabul: geçerli RS256 token doğru `sub`'ı döner; süresi geçmiş / yanlış `iss` / farklı
    anahtarla imzalı / HS256 imzalı / `alg=none` / boş `sub` / bilinmeyen `kid` / bozuk string →
    hepsi `ErrInvalidToken`.
  - Doğrulama: `keycloak_test.go` — testte üretilen RSA çifti + `httptest` JWKS sunucusu
    (gerçek imza doğrulaması, mock yok); `go test ./internal/auth/` yeşil.
  - Dosyalar: `internal/auth/keycloak.go`, `keycloak_test.go`, `go.mod`/`go.sum`. **Kapsam: M**
  - Bağımlılık: yok (T33'e ihtiyaç duymaz; paralel yürüyebilir).
- [x] **T35 — RequireAuth middleware + SubjectFromContext**
  - Yapılacak: `(v *KeycloakVerifier) RequireAuth(next http.HandlerFunc) http.HandlerFunc` —
    `Authorization: Bearer` okuma deseni eski middleware ile birebir; token yok/`Bearer` değil →
    `httpx.WriteError(401, "missing or malformed authorization header")`; `Verify` hatası →
    `httpx.WriteError(401, "invalid or expired token")`; başarıda `sub` context'e konur ve `next`
    çağrılır. Yeni erişimci `SubjectFromContext(ctx) (string, bool)` — kendi (yeni) context
    key'iyle; eski `UserIDFromContext` bu görevde yerinde kalır (çakışma yok, söküm T37).
  - Kabul: header yok / bozuk / geçersiz token → `401` + `{"error":"..."}`; geçerli token →
    `next` çağrılır ve `SubjectFromContext` doğru `sub`'ı döner.
  - Doğrulama: `keycloak_test.go`'ya middleware senaryoları; `go test ./internal/auth/` yeşil;
    `go build ./...` yeşil (eski kodla yan yana derlenir).
  - Dosyalar: `internal/auth/keycloak.go`, `keycloak_test.go`. **Kapsam: S**
  - Bağımlılık: T34.
- [x] **CHECKPOINT R** — `go test ./internal/auth/` yeşil (Docker/container gerekmez).

### Faz 3 — Kimlik geçişi: cart + order → Keycloak `sub`
- [x] **T36 — migration 0006 + cart/order paketlerinin string kimliğe geçişi**
  - Yapılacak: `migrations/0006_switch_to_keycloak_identity.sql` — `cart` ve `orders`
    tablolarındaki `users` FK constraint'leri düşürülür, `user_id` sütunları `TEXT`'e çevrilir,
    `users` tablosu drop edilir (`UNIQUE(user_id, product_id)` korunur). `internal/cart` ve
    `internal/order`: modellerde `UserID string`; repository metotları (`AddOrIncrement`,
    `ListByUser`, `ClearByUser`, `CreateFromCart`) `userID string` alır; handler'lar kimliği
    `auth.SubjectFromContext`'ten okur. Testler: token üretimi httptest JWKS'li gerçek verifier +
    test RSA anahtarına geçer; init script listesine `0006` eklenir (0001…0006 numara sırasıyla);
    izolasyon senaryoları iki farklı `sub` ile korunur. İş mantığı/doğrulama/yanıt gövdeleri
    **değişmez**.
  - Kabul: tüm Dilim 4 cart/order senaryoları (increment, toplamlar, snapshot, transaction,
    boş sepet `400`, izolasyon, token'sız `401`) yeni kimlik tipiyle aynen geçer.
  - Doğrulama: `go test ./internal/cart/ ./internal/order/` yeşil (Docker açık) · `go build ./...`.
  - Dosyalar: `migrations/0006_switch_to_keycloak_identity.sql` (yeni), `internal/cart/{model,repository,handler}.go`
    + testleri, `internal/order/{model,repository,handler}.go` + testleri. **Kapsam: L**
    (~5 dosya kuralı bilinçli aşılıyor: cart ve order aynı tabloyu paylaşır — bölmek görevler
    arasında kırmızı test bırakırdı; bkz. plan.md "Mimari Kararlar".)
  - Bağımlılık: T35.
- [x] **CHECKPOINT S** — `go test ./internal/cart/ ./internal/order/` yeşil (Docker açık).

### Faz 4 — Eski auth'un sökümü + kablolama
- [x] **T37 — router + main + eski auth dosyalarının silinmesi**
  - Yapılacak: `router.go` — `POST /api/register` ve `POST /api/login` rotaları **silinir**;
    imza `NewRouter(products, cartH, ordersH, requireAuth Middleware)` olur (auth handler
    parametresi gider); `POST/PUT/DELETE /api/products` + cart + orders rotalarının tümü
    `requireAuth` ile sarılır; `GET` ürün rotaları ve `/health` sarılmaz. `main.go` —
    `JWT_SECRET`/`TokenManager`/auth handler kalkar; `KEYCLOAK_ISSUER_URL` okunur (boşsa
    `log.Fatal`), `NewKeycloakVerifier` kurulur (hata → `log.Fatal`), `verifier.RequireAuth`
    router'a verilir. **Silinir:** `internal/auth/token.go`, `token_test.go`, `middleware.go`,
    `middleware_test.go`, `handler.go`, `handler_test.go`, `repository.go`, `repository_test.go`,
    `model.go`. `go mod tidy` (bcrypt düşer). `router_test.go`: register/login → `404`; token'sız
    yazma/cart/orders → `401`; geçerli token → handler'a ulaşır; token'sız `GET /api/products` →
    `200`; `/health` değişmez.
  - Kabul: `internal/auth`'ta yalnızca `keycloak.go` + testi kalır; `go.mod`'da bcrypt yok;
    tüm korumalı uçlar tek middleware'den geçer; okuma uçları public.
  - Doğrulama: `go build ./...` · `go test ./...` yeşil (Docker açık — ilk tam yeşil bu noktada).
  - Dosyalar: `internal/http/router.go`, `router_test.go`, `cmd/server/main.go`, 9 silme,
    `go.mod`/`go.sum`. **Kapsam: M** (çoğunluğu silme)
  - Bağımlılık: T36.
- [x] **T38 — Makefile + api.http**
  - Yapılacak: `Makefile` start zinciri Keycloak hazır olana dek bekler (host'tan
    `curl -sf http://localhost:8081/realms/vibe-shop` döngüsü); sunucuya `KEYCLOAK_ISSUER_URL`'in
    mevcut env sağlama yöntemiyle (DATABASE_URL nasıl veriliyorsa aynı yolla) ulaştığı doğrulanır;
    `JWT_SECRET` referansları temizlenir. `api.http`: register/login örnekleri **silinir**;
    Keycloak token istekleri eklenir (`# @name kcLogin` testuser, `# @name kcLogin2` testuser2);
    ürün yazma + cart + orders örnekleri Keycloak token'ına geçer; token'sız `401` ve izolasyon
    (kcLogin2 ile boş sepet) örnekleri eklenir.
  - Kabul: `make start` tek komutla Postgres + Keycloak + API'yi ayağa kaldırır; `api.http`
    örnekleri güncel akışla çalışır.
  - Doğrulama: `make start` sonrası SPEC §10.2 curl akışı elle koşulur.
  - Dosyalar: `Makefile`, `api.http`, (gerekirse) `.env`. **Kapsam: S**
  - Bağımlılık: T33, T37.
- [x] **CHECKPOINT T** — Uçtan uca: `docker compose down -v` + `docker compose up -d` (PG + KC) +
  migration'lar + sunucu; token'sız `POST /api/products` / `POST /api/cart` / `POST /api/orders` →
  `401`; `kcLogin` token'ıyla `POST /api/products` → `201` (listede görünür), `PUT` → `200`,
  `DELETE` → `204`; aynı token'la sepete ekle → gör → sipariş → sepet boşalır; `kcLogin2`
  token'ı ilk kullanıcının sepetini/siparişini göremez; token'sız `GET /api/products` → `200`;
  `POST /api/register` ve `POST /api/login` → `404`; `/health` → `{"status":"ok"}`.

### Faz 5 — Kalite kapısı
- [x] **T39 — Kalite doğrulaması**
  - Yapılacak: `go mod tidy` son kontrol (keyfunc kayıtlı, bcrypt yok).
  - Doğrulama: `gofmt -l .` boş · `go vet ./...` temiz · `go test ./...` yeşil (Docker açık).
- [ ] **CHECKPOINT U (final)** — İnsan onayı; dilim tamam.

---

## Dilim 6 — Frontend (SPA) ← *şu anki dilim*

Detaylar: [plan.md](./plan.md#dilim-6--frontend-spa--şu-anki-dilim) · Spec: [../SPEC.md](../SPEC.md) §11

### Faz 0 — İskelet
- [x] **T40 — frontend/ iskeleti: Vite + React + TS + Tailwind + shadcn/ui**
  - Yapılacak: `npm create vite@latest frontend -- --template react-ts`; Tailwind
    (`@tailwindcss/vite`) ve shadcn/ui (`npx shadcn@latest init`, zinc teması) kurulur;
    `vite.config.ts`'e `/api` → `http://localhost:8080` proxy'si eklenir; Vitest + Testing
    Library + MSW dev bağımlılıkları eklenir; `App.tsx`'te React Router ile boş rota iskeleti
    (`/login`, `/`, `/products/:id`, `/cart`).
  - Kabul: `npm run dev` 5173'te açılır; proxy üzerinden `/api/products` gerçek API'ye ulaşır;
    `npm run build` ve `npm run lint` temiz; Go tarafında hiçbir dosya değişmez.
  - Doğrulama: `npm run dev` + tarayıcıda `http://localhost:5173`; `curl localhost:5173/api/products`.
  - Dosyalar: `frontend/` (yeni ağaç). **Kapsam: M**
- [x] **T41 — Keycloak client'a webOrigins**
  - Yapılacak: `keycloak/vibe-shop-realm.json`'da `vibe-shop-api` client'ına
    `"webOrigins": ["http://localhost:5173"]` eklenir; Keycloak container'ı yeniden yaratılır.
  - Kabul: tarayıcıdan (origin 5173) token endpoint'ine `grant_type=password` isteği CORS
    engeline takılmaz.
  - Doğrulama: `docker compose up -d --force-recreate keycloak` sonrası tarayıcı konsolundan
    fetch denemesi (veya CHECKPOINT W'de canlı giriş).
  - Dosyalar: `keycloak/vibe-shop-realm.json`. **Kapsam: S**
- [x] **CHECKPOINT V** — `npm run dev` ayakta, proxy çalışıyor, `npm run build` temiz.

### Faz 1 — Kimlik dikey dilimi
- [x] **T42 — auth altyapısı + login sayfası + rota koruması**
  - Yapılacak: `lib/auth.tsx` — AuthContext: `login(email, password)` Keycloak token
    endpoint'ine ROPC isteği atar, `access_token`/`refresh_token`'ı localStorage'da saklar;
    `logout()` temizler; `refresh()` tek sefer yeniler. `lib/api.ts` — fetch sarmalayıcı:
    göreli `/api/...`, `Authorization: Bearer` ekler, `401` → refresh + retry, olmadı →
    oturumu temizle + `/login`. `components/require-auth.tsx` — token yoksa `state.from` ile
    `/login`'e. `pages/login.tsx` — `design/01`: ortalanmış kart, E-posta/Parola alanları,
    "Giriş Yap" butonu, `invalid_grant` → "E-posta veya parola hatalı" form hatası; girişte
    geldiği rotaya döner; token varken `/login` → `/`.
  - Kabul: token'sız her rota `/login`'e düşer; geçerli kimlikle giriş çalışır; yanlış parola
    form hatası gösterir; 401→refresh→retry akışı çalışır.
  - Doğrulama: `npm run test` — MSW ile login başarı/başarısızlık, guard yönlendirmesi,
    api 401→refresh senaryoları yeşil.
  - Dosyalar: `frontend/src/lib/auth.tsx`, `lib/api.ts`, `components/require-auth.tsx`,
    `pages/login.tsx`, testleri. **Kapsam: L**
  - Bağımlılık: T40, T41.
- [x] **CHECKPOINT W** — Gerçek stack'le (make start + npm run dev): token'sız `/` → `/login`;
  `testuser`/`test1234` girişi başarılı; yanlış parola hata mesajı; çıkış sonrası tekrar `/login`.

### Faz 2 — Ürün sayfaları
- [x] **T43 — navbar + ürün listesi sayfası**
  - Yapılacak: `components/navbar.tsx` — `design/02` üst barı: vibe-shop wordmark, "Ürünler",
    arama girdisi (görsel; filtre client-side isteğe bağlı), sepet ikonu + kalem sayısı rozeti,
    kullanıcı menüsü (Çıkış). `pages/products.tsx` — `GET /api/products`'tan 4 kolonlu kart
    grid'i: zinc-100 görsel alanı (placeholder), ad, `₺` fiyat (`Intl.NumberFormat('tr-TR')`),
    "Sepete Ekle" butonu (`POST /api/cart`, quantity 1, başarıda rozet güncellenir + toast).
    Boş liste durumu.
  - Kabul: liste gerçek API verisiyle render olur; sepete ekleme çalışır; boş durum düzgün.
  - Doğrulama: `npm run test` — MSW ile liste render, boş durum, sepete ekleme istek gövdesi.
  - Dosyalar: `frontend/src/components/navbar.tsx`, `pages/products.tsx`, testleri. **Kapsam: M**
  - Bağımlılık: T42.
- [x] **T44 — ürün detay sayfası**
  - Yapılacak: `pages/product-detail.tsx` — `design/03`: breadcrumb, sol büyük zinc-100 görsel,
    sağda ad + fiyat + açıklama metni + adet stepper'ı (- n +, min 1) + "Sepete Ekle"
    (`POST /api/cart` seçili adetle) + "Kargo ve İade"/"Malzeme" akordeonu (statik metin).
    Olmayan id → 404 durumu ("Ürün bulunamadı" + listeye dön).
  - Kabul: detay gerçek veriyle render; adet stepper'ı doğru gövde gönderir; 404 durumu düzgün.
  - Doğrulama: `npm run test` — MSW ile render, stepper, POST gövdesi, 404 senaryosu.
  - Dosyalar: `frontend/src/pages/product-detail.tsx`, testi. **Kapsam: M**
  - Bağımlılık: T43.

### Faz 3 — Sepet ve sipariş
- [x] **T45 — sepet sayfası + sipariş onayı görünümü**
  - Yapılacak: `pages/cart.tsx` — `design/04`: solda kalem tablosu (görsel placeholder, ad,
    birim fiyat, adet **salt-okunur**, satır toplamı), sağda "Sipariş Özeti" kartı (ara toplam,
    kargo "Ücretsiz", toplam, "Siparişi Tamamla"). Buton `POST /api/orders` çağırır; başarıda
    `design/05` onay görünümü (yeşil check, "Siparişin Alındı", sipariş no, kalemler snapshot
    fiyatlarıyla, toplam, "Alışverişe Devam Et" → `/`); sepet rozeti sıfırlanır. Boş sepet
    durumu: mesaj + listeye yönlendiren buton ("Siparişi Tamamla" gizli/pasif).
  - Kabul: toplamlar API'dekiyle aynı; sipariş sonrası onay görünümü + boş sepet; stepper/silme
    bilinçli olarak yok (SPEC §11.1).
  - Doğrulama: `npm run test` — MSW ile dolu/boş sepet, sipariş akışı, toplam hesapları.
  - Dosyalar: `frontend/src/pages/cart.tsx`, testi. **Kapsam: M**
  - Bağımlılık: T43.
- [x] **CHECKPOINT X** — Uçtan uca gerçek stack'le: login → ürünler listelenir → detaydan 2 adet
  sepete ekle → sepette doğru toplam → "Siparişi Tamamla" → onay görünümü (no + kalemler +
  toplam) → sepet boş; `testuser2` ile girişte sepet boş (izolasyon).

### Faz 4 — Kalite kapısı
- [x] **T46 — Kalite + tasarım karşılaştırması**
  - Yapılacak: `npm run lint`/`build`/`test` son koşu; her sayfanın `design/` mockup'ıyla yan
    yana karşılaştırılması, kalan görsel farkların düzeltilmesi veya SPEC §11.1'e bilinen sapma
    olarak eklenmesi; README/SPEC komut tablolarının güncel olduğunun kontrolü.
  - Doğrulama: `npm run build` + `npm run lint` + `npm run test` temiz · `go test ./...`
    etkilenmedi · görsel karşılaştırma tamam.
- [ ] **CHECKPOINT Y (final)** — İnsan onayı; dilim tamam.
