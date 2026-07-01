# vibe-shop — Spec

> Küçük bir e-ticaret backend'i. Bu belge, kod yazılmadan önce **ne** inşa edeceğimizi ve
> **sınırları** tanımlar. Değişiklikler önce burada güncellenir, sonra kodda uygulanır.

## 1. Amaç (Objective)

Ürünleri listeleyen, ileride sepet ve sipariş yeteneklerini alacak sade bir HTTP API.

- **Hedef kullanıcı:** vibe-shop istemcisi (web/mobil frontend) ve API'yi tüketen geliştiriciler.
- **Bu dilimin kapsamı (İlk dilim):** Yalnızca boş iskelet + çalışan bir `GET /health` kontrolü.
  Ürün/sepet/sipariş **bu dilimde yok**; sonraki dilimlere bırakıldı.
- **Başarı ölçütü:** `go run ./cmd/server` ile sunucu 8080 portunda ayağa kalkar ve
  `GET /health` isteği `200` + `{"status":"ok"}` döner.

### Yol haritası
1. **İskelet + /health** ✅ tamamlandı
2. **Ürün okuma API'si** (`GET /api/products`, `GET /api/products/:id`) ← *şu anki dilim, bkz. §7*
3. Sepet (`POST /cart`, ...)
4. Sipariş (`POST /orders`, ...)

## 2. Komutlar (Commands)

| Komut | Amaç |
|-------|------|
| `go run ./cmd/server` | Sunucuyu 8080 portunda başlatır |
| `go build ./...` | Tüm paketleri derler |
| `go test ./...` | Tüm testleri çalıştırır |
| `go vet ./...` | Statik analiz |
| `gofmt -l .` | Biçim ihlallerini listeler (çıktı boş olmalı) |
| `curl -s localhost:8080/health` | Sağlık kontrolü — `{"status":"ok"}` beklenir |

## 3. Proje Yapısı (Project Structure)

```
vibe-shop/
  go.mod                    # module vibe-shop, sadece stdlib
  SPEC.md
  cmd/
    server/
      main.go               # giriş noktası: router'ı kur, 8080'de dinle
  internal/
    http/
      router.go             # rotaları http.ServeMux'a bağlar
    health/
      handler.go            # GET /health → {"status":"ok"}
      handler_test.go
```

- `cmd/server` yalnızca kablolama (wiring) yapar; iş mantığı içermez.
- `internal/` altındaki paketler tek sorumluluk taşır; sepet/sipariş ileride
  `internal/cart`, `internal/order` olarak eklenecek.
- Port ve benzeri ayarlar `main.go` içinde sabit (`:8080`); env okuması ileride eklenebilir.

## 4. Kod Stili (Code Style)

- **Dil:** Go 1.26 (kurulu sürüm). `go.mod` içindeki `go` direktifi buna göre.
- **Bağımlılık:** Sadece standart kütüphane. `go.mod` içinde dış `require` **olmayacak**.
- **Router:** `net/http` + `http.ServeMux`. Üçüncü parti router yok.
- **Biçim:** `gofmt` zorunlu. `go vet` temiz geçmeli.
- **Adlandırma:** İdiomatik Go (exported için `CamelCase`, dosya adları `snake` değil sade küçük harf).
- **Handler imzası:** Standart `func(http.ResponseWriter, *http.Request)`.
- **JSON:** `encoding/json` ile üretilir; elle string birleştirme yok.
- **Hata yönetimi:** Hatalar sarmalanarak (`fmt.Errorf("...: %w", err)`) yukarı taşınır; `panic` yok.

## 5. Test Stratejisi (Testing Strategy)

- **Çerçeve:** Standart `testing` paketi + `net/http/httptest`. Dış test kütüphanesi yok.
- **Bu dilimin testi:** `internal/health/handler_test.go`
  - `GET /health` → status `200`
  - `Content-Type: application/json`
  - gövde `{"status":"ok"}`
- **Kural:** Her yeni davranış, kodla birlikte gelen bir testle kanıtlanır.
- **Geçiş ölçütü:** `go test ./...` ve `go vet ./...` temiz; `gofmt -l .` boş.

## 6. Sınırlar (Boundaries)

**Her zaman yap (Always):**
- Sadece standart kütüphane kullan.
- `gofmt` + `go vet` temiz tut.
- Yeni davranışı testle kanıtla.
- Değişiklikleri önce bu SPEC'e yansıt, sonra kodla.
- Kapsamı bu dilimle sınırlı tut: iskelet + `/health`.

**Önce sor (Ask first):**
- Herhangi bir dış bağımlılık eklemeden önce.
- Proje yapısını (klasör düzeni) değiştirmeden önce.
- `/health` dışında yeni endpoint eklemeden önce.
- Port/config'i env değişkeninden okumaya geçmeden önce.

**Asla yapma (Never):**
- Bu dilimde ürün/sepet/sipariş mantığı yazma.
- Veritabanı, ORM veya harici servis entegrasyonu ekleme.
- İstenmeyen "temizlik"/refactor ile alakasız dosyalara dokunma.
- Anlamadığın kodu/yorumu silme.

> **Not:** Yukarıdaki "sadece stdlib / DB yok" kuralları **İlk Dilim**'e (health/router) özgüdür.
> Dilim 2 (§7) bu kuralları kendi kapsamında, açıkça belirtilen şekilde gevşetir; health ve
> router paketleri stdlib-only kalmaya devam eder.

---

## 7. Dilim 2 — Ürün Okuma API'si (Products Read API)

### 7.1 Amaç
Kullanıcının ürünleri okuyabildiği iki endpoint:
- `GET /api/products` → tüm ürünleri JSON dizi olarak döner.
- `GET /api/products/{id}` → id'ye göre tek ürün döner; bulunamazsa `404` + JSON hata gövdesi.

Veri kaynağı **local Postgres** (`products` tablosu); mock/hardcoded veri yok.
Local geliştirmede Postgres, repo köküne eklenecek `docker-compose.yml` ile bir
**Docker container** olarak ayağa kaldırılır (kurulum gerektirmez, `docker compose up -d` yeterli).
Testlerde ayrı olarak **testcontainers-go** kullanılır (§7.5) — ikisi birbirinden bağımsızdır.

**Başarı ölçütü:** `docker compose up -d` ile local Postgres container'ı ayaktayken
`go run ./cmd/server` çalışır; `curl localhost:8080/api/products` ve
`curl localhost:8080/api/products/1` gerçek DB satırlarından üretilmiş JSON döner;
`go test ./...` (Docker + testcontainers ile) yeşil.

### 7.2 Komutlar (ek)

| Komut | Amaç |
|-------|------|
| `docker compose up -d` | Local Postgres container'ını ayağa kaldırır (bkz. `docker-compose.yml`) |
| `docker compose down -v` | Container'ı ve verisini siler (temiz sıfırlama) |
| `psql "$DATABASE_URL" -f migrations/0001_create_products.sql` | `products` tablosunu oluşturur |
| `DATABASE_URL=postgres://vibeshop:vibeshop@localhost:5432/vibeshop?sslmode=disable go run ./cmd/server` | Sunucuyu local Postgres container'ına bağlı başlatır |
| `go test ./...` | Testleri çalıştırır (testcontainers için Docker gerektirir) |

### 7.3 Proje Yapısı (ek)

```
vibe-shop/
  docker-compose.yml           # local Postgres container (dev-only)
  .env.example                 # DATABASE_URL örneği (docker-compose ile eşleşir)
  migrations/
    0001_create_products.sql   # products tablosu şeması (id, name, price)
  internal/
    db/
      db.go                    # DATABASE_URL'den GORM bağlantısı kurar
    product/
      model.go                 # GORM Product struct (id, name, price)
      repository.go            # GORM ile DB erişimi: List(), GetByID(id)
      handler.go                # GET /api/products, GET /api/products/{id}
      handler_test.go           # testcontainers-go ile integration test
    http/
      router.go                 # /api/products, /api/products/{id} eklenir
```

### 7.4 Kod Stili (ek/değişiklik)
- **Local Postgres (Docker):** `docker-compose.yml` içinde `postgres:16-alpine` image'i,
  `5432:5432` port eşlemesi, `POSTGRES_USER=vibeshop`, `POSTGRES_PASSWORD=vibeshop`,
  `POSTGRES_DB=vibeshop`. Named volume ile veri container yeniden başlatıldığında kalıcı olur.
  Bu sadece **local geliştirme** içindir; production/deploy konfigürasyonu bu dilimin kapsamı dışında.
- Bu dilimde **GORM** (`gorm.io/gorm`, `gorm.io/driver/postgres`) kullanımı istisna olarak
  izinlidir; sadece `internal/db` ve `internal/product` içinde. `internal/health` ve
  `internal/http` stdlib-only kalmaya devam eder.
- DB bağlantısı tek yerden (`internal/db`) kurulur; handler'lar `*gorm.DB`'ye doğrudan
  erişmez, `product.Repository` arayüzü üzerinden çalışır (test edilebilirlik için).
- Route parametresi Go 1.22+ `http.ServeMux` pattern'i ile okunur: `/api/products/{id}`.
  Geçersiz/sayısal olmayan `id` → `400`.
- JSON gövdeler `encoding/json` ile üretilir; elle string birleştirme yok (mevcut kural devam eder).
- Hatalar sarmalanır (`fmt.Errorf("...: %w", err)`); `panic` yok (mevcut kural devam eder).

### 7.5 Test Stratejisi (ek)
- **testcontainers-go**: her test paketi çalışırken geçici bir Postgres container ayağa
  kaldırır, `migrations/0001_create_products.sql` uygulanır, seed veri eklenir.
- `internal/product/handler_test.go`: gerçek container'a karşı —
  - `GET /api/products` → 200, seed edilen ürünleri döner.
  - `GET /api/products/{var-olan-id}` → 200, doğru ürün.
  - `GET /api/products/{olmayan-id}` → 404 + JSON hata gövdesi.
  - `GET /api/products/{geçersiz-id}` → 400.
- Mock/sahte repository ile testleri "yeşil" gösterme; testler gerçek Postgres'e karşı çalışır.
  Docker yoksa test **başarısız olur** (sessizce skip edilmez).
- **Geçiş ölçütü:** `go test ./...` (Docker mevcutken) yeşil, `go vet ./...` temiz,
  `gofmt -l .` boş.

### 7.6 Sınırlar (ek/değişiklik)

**Her zaman yap (ek):**
- GORM kullanımını `internal/db` ve `internal/product` ile sınırlı tut.
- DB bağlantı string'ini `DATABASE_URL` env değişkeninden oku; hardcode etme.
- `404` durumunda JSON hata gövdesi dön (örn. `{"error":"product not found"}`).
- Yeni davranışı testcontainers tabanlı testle kanıtla.

**Önce sor (ek):**
- `migrations/` altına yeni migration eklemeden veya şemayı değiştirmeden önce.
- GORM dışında başka bir ORM/query builder eklemeden önce.
- `/api/products` dışında yeni endpoint eklemeden önce (sepet/sipariş hâlâ kapsam dışı).
- `docker-compose.yml` içindeki image/port/credential varsayılanlarını değiştirmeden önce.

**Asla yapma (ek):**
- Ürün verisini kod içinde hardcode etme veya mock repository ile "gerçek" gibi gösterme.
- Testcontainers gerektiren testleri sessizce skip edilebilir hale getirme.
- `price`/`id` gibi alanları migration şemasıyla tutarsız şekilde değiştirme.
- `docker-compose.yml`'i production deploy konfigürasyonu olarak kullanma/genişletme (sadece local dev).
- Gerçek kimlik bilgilerini/secret'ları `docker-compose.yml` veya repoya commit etme (dev-only
  sabit şifre bu dilimde kabul edilebilir çünkü sadece localhost'ta çalışır).
