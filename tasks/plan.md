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

## Dilim 2 — Ürün Okuma API'si ✅ tamamlandı

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

---

## Dilim 3 — Ürün Yazma API'si ✅ tamamlandı

> Kaynak: [SPEC.md §8](../SPEC.md). Kapsam: **`POST /api/products`, `PUT /api/products/{id}`,
> `DELETE /api/products/{id}` — doğrulamayla birlikte, read API ile aynı tarzda**.
> Sepet/sipariş/PATCH/auth bu dilimde YOK.

### Mimari Kararlar

- **Girdi tipi + tek yerde doğrulama:** İstek gövdesi `Product`'a değil, ayrı bir
  `Input{Name string, Price float64}` tipine decode edilir; `Input.Validate() error`
  alan bazlı mesaj döner (örn. `"name must not be empty"`). POST ve PUT aynı tipi ve
  aynı `Validate`'i kullanır — kural kopyası yok (SPEC §8.6).
- **Doğrulama kuralları:** `name` kırpılınca boş olamaz, ≤ 200 karakter; `price > 0`,
  en fazla iki ondalık (`math.Round(price*100)` karşılaştırmasıyla), üst sınır
  `99_999_999.99` (DB sütunu `NUMERIC(10,2)` taşmasın, DB hatası yerine `400` dönsün).
- **Repository genişlemesi:** `Create(ctx, Product) (Product, error)`,
  `Update(ctx, Product) (Product, error)`, `Delete(ctx, id uint) error`. Update/Delete
  GORM `RowsAffected == 0` durumunda `ErrNotFound` döner — handler mevcut `errors.Is`
  desenini aynen kullanır; ekstra SELECT yok.
- **Handler tarzı korunur:** Yeni `Create/Update/Delete` metodları mevcut
  `writeJSON`/`writeError` yardımcılarını kullanır. Geçersiz JSON → `400 "invalid JSON body"`;
  doğrulama hatası → `400` + `Validate`'in mesajı; sayısal olmayan id → `400`;
  `ErrNotFound` → `404 "product not found"`; diğer → `500`. Gövdedeki `id` yok sayılır
  (decode hedefi `Input` olduğu için doğal olarak).
- **Test izolasyonu (kritik):** Paket, tek paylaşılan container + sabit seed (Widget/Gadget)
  kullanıyor ve `TestList` şu an **tam 2 ürün** bekliyor. Yazma testleri eklenince bu kırılgan
  olur (test sırası garanti değil). Karar: (a) yazma testleri kendi satırlarını oluşturur,
  seed satırlarına (id 1-2) asla dokunmaz, `t.Cleanup` ile kendi satırlarını siler;
  (b) `TestList_ReturnsSeededProducts` "tam 2" yerine "Widget ve Gadget listede var"
  şeklinde gevşetilir. Böylece testler sıradan bağımsız kalır.
- **Router imzası değişmiyor:** `NewRouter(products *product.Handler)` aynı kalır; yalnızca
  üç yeni rota eklenir (`"POST /api/products"`, `"PUT /api/products/{id}"`,
  `"DELETE /api/products/{id}"`).

### Bağımlılık Grafiği

```
internal/product/model.go — Input + Validate (T18)
        │
        ▼
internal/product/repository.go — Create/Update/Delete (T19)
        │
        ▼
internal/product/handler.go — Create/Update/Delete metodları (T20)
        │
        ▼
internal/product/handler_test.go — yazma senaryoları + TestList gevşetme (T21)
        │
        ▼
              CHECKPOINT J (go test ./internal/product/ yeşil, Docker gerekli)

internal/http/router.go + router_test.go — 3 yeni rota (T22)
        │
        ▼
              CHECKPOINT K (uçtan uca: docker compose + go run + curl POST/PUT/DELETE)

Kalite kapısı (T23)
        │
        ▼
              CHECKPOINT L (final, insan onayı)
```

### Dikey Dilimleme Yaklaşımı

Üç uç tek dikey yolda ilerler: **istek → router → handler → doğrulama → repository →
Postgres → JSON yanıt**. Domain katmanı (T18-T21) test-first tamamlanır; kablolama (T22)
ve kalite kapısı (T23) sona kalır. `main.go` ve `internal/db` değişmez — handler zaten
kablolu, yalnızca yeni metotlar rotalara bağlanır.

### Fazlar ve Checkpoint'ler

#### Faz 1 — Ürün domain'i: yazma yolları (test-first)
- **T18** `internal/product/model.go` — `Input` tipi + `Validate()`.
- **T19** `internal/product/repository.go` — arayüz + GORM impl. genişler.
- **T20** `internal/product/handler.go` — `Create`, `Update`, `Delete` metodları.
- **T21** `internal/product/handler_test.go` — SPEC §8.5'teki tüm senaryolar + `TestList` gevşetme.
- **CHECKPOINT J:** `go test ./internal/product/` yeşil (Docker açık).

#### Faz 2 — Kablolama
- **T22** `internal/http/router.go` + `router_test.go` — POST/PUT/DELETE rotaları.
- **CHECKPOINT K:** Uçtan uca manuel doğrulama — `docker compose up -d` + `go run ./cmd/server`;
  `curl -X POST` → `201` + listede görünür; `curl -X PUT` → `200` + değişiklik yansır;
  `curl -X DELETE` → `204` + ardından `GET` → `404`; `/health` hâlâ `{"status":"ok"}`.

#### Faz 3 — Kalite kapısı
- **T23** Tüm doğrulamalar: `gofmt -l .` boş, `go vet ./...` temiz, `go test ./...` yeşil (Docker açık).
- **CHECKPOINT L (final):** İnsan onayı → dilim tamam.

### Riskler

| Risk | Etki | Önlem |
|------|------|-------|
| Yazma testleri paylaşılan seed veriyi bozar, testler sıraya bağımlı hale gelir | Orta | Mimari karar: kendi satırını oluştur + `t.Cleanup` + `TestList` gevşetme (T21) |
| `float64` ile ondalık hassasiyeti (örn. `199.90` → `199.89999…`) | Düşük | İki-ondalık kontrolü `math.Round` toleransıyla; DB zaten `NUMERIC(10,2)`'ye yuvarlar |
| `NUMERIC(10,2)` taşması DB hatası → yanlışlıkla `500` | Düşük | Üst sınır doğrulaması `Validate` içinde → `400` (T18) |

### Kapsam Dışı (bilerek, SPEC §8.6 ile uyumlu)
- `PATCH` / kısmi güncelleme, soft-delete, optimistic locking.
- Kimlik doğrulama/yetkilendirme, bulk uçlar.
- Model/migration değişikliği, read API davranış değişikliği.

Detaylı görev listesi: [todo.md](./todo.md).

---

## Dilim 4 — Auth + Sepet + Sipariş ✅ tamamlandı

> Kaynak: [SPEC.md §9](../SPEC.md). Kapsam: **`POST /api/register`, `POST /api/login` (public) +
> `POST`/`GET /api/cart`, `POST /api/orders` (JWT korumalı)**. Her kullanıcı yalnızca kendi
> sepetini/siparişini görür; `user_id` daima JWT'den okunur. Tablolar: `users`, `cart`,
> `orders`, `order_items`. Refresh token, roller, parola sıfırlama, stok, ödeme bu dilimde YOK.

### Mimari Kararlar

- **Paylaşılan HTTP yardımcıları (önce çıkarılır):** `writeJSON`/`writeError` şu an
  `internal/product`'a **private**. Auth/cart/order aynı JSON ve `{"error":"..."}` gövdesini
  kullanacağı için bunlar yeni bir **`internal/httpx`** paketine taşınır (`httpx.WriteJSON`,
  `httpx.WriteError`); `product` bunları çağıracak şekilde güncellenir. Davranış birebir korunur
  (mevcut ürün testleri kanıt). Bu, tüm yeni paketlerin ön koşuludur.
- **Kullanıcı kimliği context'te:** `internal/auth` içinde **private** bir context key; exported
  `auth.UserIDFromContext(ctx) (uint, bool)` erişimcisi. Middleware `Authorization: Bearer <jwt>`
  okur, doğrular, `user_id`'yi context'e koyar; yoksa/geçersizse `401`. Handler'lar `user_id`'yi
  **yalnızca** buradan alır — asla istek gövdesinden/query'den.
- **JWT (golang-jwt/jwt/v5):** `token.go` içinde secret + ttl'e bağlı `Issue(userID) (string,error)`
  ve `Parse(tokenStr) (uint,error)`. HS256, claim `sub=user_id`, `exp` sonlu (varsayılan 24s).
  Secret **`JWT_SECRET`** env'inden `main.go`'da okunur ve enjekte edilir; global/hardcode yok.
- **Parola (bcrypt):** `bcrypt.GenerateFromPassword` (varsayılan cost) ile hash;
  `bcrypt.CompareHashAndPassword` ile doğrulama. Parola/hash hiçbir yanıtta/logda görünmez
  (`json:"-"`).
- **Sepet modeli:** `cart` her satır bir kalem `(user_id, product_id, quantity)`,
  `UNIQUE(user_id, product_id)`. `AddOrIncrement` upsert (`ON CONFLICT ... DO UPDATE
  quantity = quantity + EXCLUDED.quantity`); aynı ürün tekrar eklenince yeni satır açılmaz.
  `ListByUser` ürünle join'leyip ad/fiyat/satır toplamı döner; `ClearByUser` sepeti boşaltır.
- **Sipariş (transaction + snapshot):** `CreateFromCart(ctx, userID)` tek bir gorm
  transaction'ı içinde: sepeti (o anki ürün fiyatıyla) oku → boşsa `ErrCartEmpty` → `orders`
  ekle → her kalemi `order_items`'a **`unit_price` = o anki fiyat** olarak yaz → `total` = Σ
  `quantity*unit_price` → sepeti sil. Kısmi sipariş oluşmaz (hata → rollback).
- **Router imzası büyür:** `NewRouter(products, auth, cart, order, requireAuth)`. Register/login
  public; cart/orders auth middleware ile sarılır. Ürün uçları **sarılmaz** (korumasız kalır).
- **Test altyapısı:** Mevcut testcontainers deseni; her yeni test paketi ihtiyacı olan
  migration'ları **numara sırasıyla** `WithInitScripts(...)` ile uygular (cart → products+users+cart;
  order → hepsi). Testlerde sabit bir `JWT_SECRET` ile token üretilir; izolasyon testleri iki
  ayrı kullanıcı/token kullanır.
- **Yeni bağımlılıklar:** `github.com/golang-jwt/jwt/v5`, `golang.org/x/crypto/bcrypt` — ilk
  kullanan görevde `go get` + `go mod tidy` ile eklenir (yalnızca `internal/auth` içinde).

### Bağımlılık Grafiği

```
internal/httpx — WriteJSON/WriteError (T24) ── product bunu kullanır
        │
        ▼
migrations/0002_users → internal/auth: model.go + repository.go (T25)
        │
        ▼
internal/auth: token.go (JWT) + middleware.go (Bearer→context) (T26)  ← httpx
        │
        ▼
internal/auth: handler.go (register+login) + router/main wiring (T27)
        │                    CHECKPOINT M (auth uçtan uca)
        ▼
migrations/0003_cart → internal/cart: model.go + repository.go (T28)  ← users + products
        │
        ▼
internal/cart: handler.go + rotalar (auth mw arkasında) (T29)
        │                    CHECKPOINT N (sepet uçtan uca + izolasyon)
        ▼
migrations/0004_orders + 0005_order_items → internal/order: model + repository CreateFromCart (T30)
        │
        ▼
internal/order: handler.go + rota (auth mw arkasında) (T31)
        │                    CHECKPOINT O (sipariş uçtan uca + snapshot + izolasyon)
        ▼
Kalite kapısı (T32)          CHECKPOINT P (final, insan onayı)
```

### Dikey Dilimleme Yaklaşımı

Üç kullanıcı yolu ayrı dikey dilimler: **(1) kayıt+giriş**, **(2) sepete ekle+gör**,
**(3) siparişi tamamla**. Her dilim şema → repository → handler → route → test sırasıyla
tek seferde tamamlanır ve checkpoint'te uçtan uca çalışır. Ortak altyapı (`httpx`, JWT,
middleware) en başta, bir kez kurulur. `internal/db` değişmez; `router.go`/`main.go` her
dilimde artımlı büyür.

### Fazlar ve Checkpoint'ler

#### Faz 0 — Paylaşılan altyapı
- **T24** `internal/httpx` — `WriteJSON`/`WriteError`; `product` bunları kullanır (davranış korunur).

#### Faz 1 — Auth dikey dilimi
- **T25** `migrations/0002_create_users.sql` + `internal/auth/model.go` + `repository.go`.
- **T26** `internal/auth/token.go` (JWT) + `middleware.go` + `UserIDFromContext` (birim testli).
- **T27** `internal/auth/handler.go` (register/login) + router + `main.go` (`JWT_SECRET`).
- **CHECKPOINT M:** register→login→token; token'sız korumalı istek `401`.

#### Faz 2 — Sepet dikey dilimi
- **T28** `migrations/0003_create_cart.sql` + `internal/cart/model.go` + `repository.go`.
- **T29** `internal/cart/handler.go` (POST/GET) + rotalar (auth mw) + `main.go`.
- **CHECKPOINT N:** sepete ekle → tekrar ekle artırır → `GET` doğru toplam; A/B izolasyonu; token'sız `401`.

#### Faz 3 — Sipariş dikey dilimi
- **T30** `migrations/0004_create_orders.sql` + `0005_create_order_items.sql` + `internal/order/model.go` + `repository.go` (transaction + snapshot).
- **T31** `internal/order/handler.go` (POST) + rota (auth mw) + `main.go`.
- **CHECKPOINT O:** dolu sepet → sipariş `201` + sepet boşalır; fiyat snapshot; boş sepet `400`; A/B izolasyonu.

#### Faz 4 — Kalite kapısı
- **T32** `gofmt -l .` boş · `go vet ./...` temiz · `go test ./...` yeşil (Docker) · `go mod tidy`.
- **CHECKPOINT P (final):** İnsan onayı → dilim tamam.

### Riskler

| Risk | Etki | Önlem |
|------|------|-------|
| `writeJSON`/`writeError` çıkarımı ürün yanıtlarını değiştirebilir | Orta | Birebir aynı implementasyon; mevcut ürün testleri kanıt (T24) |
| Çoklu migration + FK sırası testcontainers'ta yanlış uygulanır | Orta | `WithInitScripts`'e migration yolları **numara sırasıyla** verilir |
| Testte `JWT_SECRET` yönetimi | Düşük | Token secret+ttl parametresiyle kurulur; testte sabit secret |
| `float64` fiyat → `numeric(10,2)` snapshot hassasiyeti | Düşük | Ürünle aynı desen; `total` snap'lenmiş `unit_price`'tan hesaplanır |
| Yeni modüllerin indirilmesi (ağ) gerekir | Düşük | İlk kullanan görevde `go get`/`go mod tidy`; offline ise önden çekilir |
| bcrypt cost testleri yavaşlatır | Düşük | Kabul edilebilir; gerekirse testte düşük cost, prod'da varsayılan |

### Kapsam Dışı (bilerek, SPEC §9.6 ile uyumlu)
- Refresh token, rol/yetki, parola sıfırlama, email doğrulama.
- Sepetten kalem silme/adet düşürme (`DELETE`/`PATCH /api/cart`).
- Stok/envanter, ödeme entegrasyonu, sipariş listeleme/detay uçları.
- Ürün (`/api/products`) uçlarına auth ekleme veya davranışını değiştirme.

Detaylı görev listesi: [todo.md](./todo.md).

---

## Dilim 5 — Keycloak'a Geçiş: Tek Kimlik Sağlayıcı ← *şu anki dilim*

> Kaynak: [SPEC.md §10](../SPEC.md). Kapsam: **eski auth (`/api/register`, `/api/login`,
> HS256 JWT, bcrypt, `users` tablosu) kaldırılır; tek kimlik sağlayıcı Keycloak olur (local'de
> Docker). Tüm korumalı uçlar — cart, orders ve artık ürün yazma uçları — yalnızca geçerli bir
> Keycloak (RS256) token'ı ile; `GET /api/products*` public kalır.** Rol/yetki (`403`),
> audience kontrolü, Keycloak production konfigürasyonu bu dilimde YOK.

### Mimari Kararlar

- **Keycloak, sıfır-tıklama kurulumla:** `docker-compose.yml`'e `keycloak` servisi
  (`quay.io/keycloak/keycloak:26.3`, `start-dev --import-realm`), host portu **8081**
  (API 8080'i kullanıyor). Realm, repoya eklenen `keycloak/vibe-shop-realm.json`'dan otomatik
  import edilir: realm `vibe-shop`, **public** client `vibe-shop-api`
  (`directAccessGrantsEnabled: true` — curl ile `grant_type=password` token alınabilsin),
  **iki** test kullanıcısı `testuser`/`test1234` ve `testuser2`/`test1234` (izolasyon
  senaryoları için). Admin konsolunda elle adım yok; kurulum deterministik ve tekrarlanabilir.
- **Eski auth tamamen sökülür (kullanıcı kararı):** `/api/register` ve `/api/login` rotaları
  silinir (`404` olur); `internal/auth`'tan `token.go`, `middleware.go`, `handler.go`,
  `repository.go`, `model.go` ve testleri silinir; bcrypt bağımlılığı `go mod tidy` ile düşer
  (`golang-jwt/jwt/v5` RS256 parse için kalır). Kullanıcı yönetimi Keycloak'ın işidir.
  Ölü kod bırakılmaz (SPEC §10.6 "asla yapma").
- **Kimlik = Keycloak `sub` (string):** `auth.SubjectFromContext(ctx) (string, bool)` eski
  `UserIDFromContext(uint)`'in yerini alır. cart/order modellerinde `UserID string`; repository
  metotları `userID string` alır. İzolasyon deseni (her sorgu `user_id` filtreli, kimlik yalnızca
  context'ten) **aynen korunur** — yalnızca tip ve kaynak değişir.
- **Şema geçişi (migration 0006):** `cart.user_id` ve `orders.user_id` → `TEXT`; `users`'a giden
  FK'lar düşürülür; `users` tablosu **drop** edilir. `UNIQUE(user_id, product_id)` korunur.
  Dev-only veri olduğundan mapping/backfill yapılmaz; temiz geçiş için `docker compose down -v`
  önerilir. cart ve order **aynı tabloyu paylaştığı için** (order, sepetten okur) domain geçişi
  tek görevde yapılır — ~5 dosya kuralı bilinçli olarak aşılır; bölmek, görevler arasında
  kırmızı test / yarı-geçmiş şema bırakırdı.
- **Doğrulama JWKS ile, tek yerde:** Keycloak RS256 imzalar; API imzayı
  `<issuer>/protocol/openid-connect/certs` JWKS'inden aldığı public key ile doğrular.
  Yeni bağımlılık (SPEC §10.4'te onaylı): `github.com/MicahParks/keyfunc/v3` — JWKS
  indirme/önbellekleme/anahtar rotasyonu; `golang-jwt/jwt/v5` ile entegre, yalnızca
  `internal/auth` içinde. `Verify(tokenStr) (sub, err)` = `jwt.Parse` +
  `WithValidMethods(["RS256"])` + `iss` kontrolü + boş olmayan `sub`; tüm hatalar
  `ErrInvalidToken`'a katlanır (eski `TokenManager.Parse` deseni). `alg=none`/HS256 reddi
  (alg confusion) açık test senaryosudur.
- **Fail-fast başlangıç:** `NewKeycloakVerifier(issuerURL)` başlangıçta JWKS'i çeker;
  erişilemezse `main.go` `log.Fatal` (mevcut `DATABASE_URL` tarzı). Lazy-init alternatifi
  (ilk istekte çekme) değerlendirildi ve elendi: sunucu "ayakta ama her korumalı istek 401/500"
  gibi belirsiz bir duruma düşmesin. `make start` sıralamayı üstlenir (Keycloak hazır olana
  dek bekler). `JWT_SECRET` env'i tamamen kalkar.
- **Middleware adı `RequireAuth` kalır, router sadeleşir:** `(v *KeycloakVerifier).RequireAuth`
  mevcut `apphttp.Middleware` tipiyle uyumlu; Bearer okuma deseni eskiyle birebir; doğrulanan
  `sub` context'e konur. `NewRouter(products, cartH, ordersH, requireAuth)` — auth handler
  parametresi ve register/login rotaları gider; ürün yazma + cart + orders rotalarının tümü
  aynı middleware ile sarılır. `internal/http` stdlib-only kalır.
- **Geçiş sırası build'i yeşil tutar:** yeni verifier/middleware önce **eskiyle yan yana**
  eklenir (farklı receiver, farklı erişimci adı — çakışma yok); sonra domain `sub`'a geçer;
  en son eski auth sökülür ve kablolama tamamlanır. Hiçbir görev sonunda derleme kırık kalmaz;
  `go test ./...`'in tam yeşili söküm görevinin (T37) checkpoint'inde aranır.
- **Test yaklaşımı — gerçek imza, sahte sunucu:** birim testler kendi RSA çiftini üretip
  JWKS'i `httptest` ile servis eder; doğrulama kriptografik olarak gerçektir, container
  gerekmez. cart/order testcontainers testleri aynı senaryoları korur, yalnızca token üretimi
  test RSA anahtarına geçer ve init script listesine 0006 eklenir. Gerçek Keycloak,
  CHECKPOINT T'de uçtan uca curl akışıyla kanıtlanır. (testcontainers-go'nun resmi Keycloak
  modülü yok; community bağımlılık eklemek yerine bu sınırlı istisna seçildi — DB'ye dokunan
  hiçbir test mock'a dönmüyor.)

### Bağımlılık Grafiği

```
docker-compose.yml + keycloak/vibe-shop-realm.json + .env.example (T33)
        │
        ▼
              CHECKPOINT Q (Keycloak ayakta, curl ile token alınıyor — manuel)

internal/auth/keycloak.go — KeycloakVerifier: JWKS + RS256 + sub (T34)   ← T33'ten bağımsız yürüyebilir
        │
        ▼
internal/auth/keycloak.go — RequireAuth middleware + SubjectFromContext (T35)  ← eskiyle yan yana
        │
        ▼
              CHECKPOINT R (go test ./internal/auth/ yeşil — container gerekmez)

migrations/0006 + internal/cart + internal/order — kimlik `sub` (string)'e geçer (T36)
        │
        ▼
              CHECKPOINT S (go test ./internal/cart/ ./internal/order/ yeşil — Docker gerekli)

internal/http/router.go + cmd/server/main.go + eski auth dosyalarının silinmesi + go mod tidy (T37)
        │
        ▼
Makefile (Keycloak bekleme) + api.http (Keycloak token'lı örnekler) (T38)   ← T33 + T37
        │
        ▼
              CHECKPOINT T (uçtan uca: compose + gerçek Keycloak token'ı + tüm akışlar)

Kalite kapısı (T39)
        │
        ▼
              CHECKPOINT U (final, insan onayı)
```

### Dikey Dilimleme Yaklaşımı

Tek kullanıcı yolu vardır: **Keycloak'tan token al → korumalı isteğe ekle → API doğrular →
`sub` context'ten okunur → mevcut handler çalışır**. Altyapı (T33) ve doğrulama çekirdeği
(T34-T35) birbirinden bağımsız ilerleyebilir; kimlik geçişi (T36) domain'i yeni kimliğe
taşır; söküm + kablolama (T37) eski yolu kaldırır; T38 geliştirici deneyimini tamamlar.
Ürün domain'inin iş mantığı hiç değişmez — yalnızca rotaları korumaya alınır.

### Fazlar ve Checkpoint'ler

#### Faz 1 — Keycloak local altyapısı
- **T33** `docker-compose.yml`'e keycloak servisi + `keycloak/vibe-shop-realm.json` (2 kullanıcı) + `.env.example`.
- **CHECKPOINT Q:** `docker compose up -d` sonrası curl ile parola grant'i token döner (manuel).

#### Faz 2 — Token doğrulama çekirdeği (test-first, container'sız)
- **T34** `internal/auth/keycloak.go` — `KeycloakVerifier` (`go get keyfunc/v3`, JWKS, RS256, iss/exp/sub).
- **T35** `RequireAuth` middleware + `SubjectFromContext` (eski auth koduyla yan yana; söküm T37'de).
- **CHECKPOINT R:** `go test ./internal/auth/` yeşil.

#### Faz 3 — Kimlik geçişi (cart + order → Keycloak `sub`)
- **T36** `migrations/0006` + `internal/cart` + `internal/order` — `UserID string`, handler'lar
  `SubjectFromContext` okur, testler test-RSA token'larına geçer. (Bilinçli büyük görev; gerekçe
  yukarıda.)
- **CHECKPOINT S:** `go test ./internal/cart/ ./internal/order/` yeşil (Docker açık).

#### Faz 4 — Eski auth'un sökümü + kablolama
- **T37** `router.go` (register/login silinir, tüm korumalı rotalar `requireAuth` ile) +
  `main.go` (`JWT_SECRET`/`TokenManager` kalkar; `KEYCLOAK_ISSUER_URL` + verifier) + eski auth
  dosyalarının silinmesi + `go mod tidy` (bcrypt düşer).
- **T38** `Makefile` (Keycloak hazır-bekleme) + `api.http` (register/login örnekleri silinir;
  kcLogin/kcLogin2 ile tüm akışlar).
- **CHECKPOINT T:** Uçtan uca — token'sız yazma/cart/orders `401`; Keycloak token'ıyla ürün
  yazma `201/200/204` ve cart→order akışı çalışır; `testuser2` izolasyonu; token'sız `GET` `200`;
  `/api/register`+`/api/login` `404`.

#### Faz 5 — Kalite kapısı
- **T39** `go mod tidy` doğrulaması · `gofmt -l .` boş · `go vet ./...` temiz · `go test ./...` yeşil (Docker).
- **CHECKPOINT U (final):** İnsan onayı → dilim tamam.

### Riskler

| Risk | Etki | Önlem |
|------|------|-------|
| Alg confusion: HS256/none token'ı RS256 doğrulayıcıyı geçer | Yüksek | `jwt.WithValidMethods(["RS256"])` + HS256/none token'la açık ret testi (T34) |
| Kimlik tipi geçişi (uint→string) cart/order'da gözden kaçan bir sorguyu kırar | Orta | T36 tek görevde, iki paketin tüm testcontainers testleri aynı senaryoları korur; CHECKPOINT S tam paket yeşili ister |
| Migration 0006 mevcut dev verisiyle uyumsuz (eski int user_id'ler anlamsızlaşır) | Düşük | Dev-only; `docker compose down -v` ile temiz başlangıç önerilir; backfill bilinçli olarak yok |
| Keycloak yavaş açılır; sunucu fail-fast olduğu için boot başarısız | Orta | `make start` Keycloak realm URL'i 200 dönene dek bekler (T38); hata mesajı yönlendirici |
| 8081 portu makinede dolu | Düşük | Compose'ta tek satır değişiklik; port değişikliği "önce sor" (SPEC §10.6) |
| Realm import formatı Keycloak sürümüyle uyumsuz | Düşük | İmaj `26.3`'e pinli; sürüm değişikliği "önce sor" |
| Söküm (T37) yanlışlıkla gereken kodu siler | Orta | Silinecek dosya listesi görevde açık; `go build ./...` + `go test ./...` söküm sonrası zorunlu |

### Kapsam Dışı (bilerek, SPEC §10.6 ile uyumlu)
- Rol/yetki (`403`), audience (`aud`/`azp`) kontrolü.
- `sub` dışındaki claim'lerin kullanımı/saklanması (email, username vb.).
- Ürün/sepet/sipariş iş mantığında, doğrulama kurallarında veya yanıt gövdelerinde değişiklik.
- Keycloak'ın production konfigürasyonu (TLS, gerçek kullanıcılar, `start` modu).

Detaylı görev listesi: [todo.md](./todo.md).
