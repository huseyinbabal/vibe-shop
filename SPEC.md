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
2. **Ürün okuma API'si** (`GET /api/products`, `GET /api/products/:id`) ✅ tamamlandı, bkz. §7
3. **Ürün yazma API'si** (`POST`, `PUT`, `DELETE /api/products`) ✅ tamamlandı, bkz. §8
4. **Auth + Sepet + Sipariş** (`POST /api/register`, `POST /api/login`, `POST`/`GET /api/cart`,
   `POST /api/orders`) ✅ tamamlandı, bkz. §9. Not: bu dilim, önceki yol haritasındaki
   ayrı "Sepet" ve "Sipariş" adımlarını **kimlik doğrulama** ile birlikte tek dilimde birleştirir,
   çünkü "her kullanıcı yalnızca kendi sepetini/siparişlerini görür" kuralı bir kullanıcı kimliği gerektirir.
5. **Keycloak'a geçiş — tek kimlik sağlayıcı** (eski `/api/register` + `/api/login` (HS256 JWT)
   kaldırılır; tüm korumalı uçlar — sepet, sipariş ve artık ürün yazma uçları — yalnızca geçerli
   bir Keycloak token'ı ile çalışır; `GET /api/products*` herkese açık kalır) ✅ tamamlandı, bkz. §10.
6. **Frontend (SPA)** — ürün listesi, ürün detay, sepet ve Keycloak girişli login sayfaları;
   giriş yapmamış herkes login'e yönlendirilir; veri Go API'den (:8080); `design/` mockup'larına
   sadık, shadcn/ui ile ← *şu anki dilim*, bkz. §11.

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

---

## 8. Dilim 3 — Ürün Yazma API'si (Products Write API)

### 8.1 Amaç

Ürünler artık sadece okunmakla kalmaz; eklenir, güncellenir ve silinir. Tüm uçlar
Dilim 2'deki read API ile **aynı tarzda** çalışır: aynı `Handler`/`Repository` deseni,
aynı `writeJSON`/`writeError` yardımcıları, aynı `{"error":"..."}` hata gövdesi,
aynı GORM + Postgres altyapısı, aynı `http.ServeMux` desen rotaları.

| Metot & Yol | Başarı | Gövde (istek → yanıt) |
|---|---|---|
| `POST /api/products` | `201` | `{"name","price"}` → oluşturulan ürün (`id` dahil) |
| `PUT /api/products/{id}` | `200` | `{"name","price"}` → güncellenen ürünün son hali (tam güncelleme) |
| `DELETE /api/products/{id}` | `204` | — → boş gövde |

**Doğrulama (POST ve PUT'ta birebir aynı):**
- `name`: zorunlu; boşluklar kırpıldıktan sonra boş olamaz; en fazla 200 karakter.
- `price`: zorunlu; `> 0`; en fazla iki ondalık basamak (DB sütunu `numeric(10,2)` ile uyum).
- Gövdedeki `id` alanı yok sayılır; `id` yalnızca path'ten ve DB'den gelir.
- Geçersiz JSON gövde → `400`. Doğrulama hatası → `400` + hangi alanın neden geçersiz
  olduğunu söyleyen `{"error":"..."}` mesajı.
- Sayısal olmayan `{id}` → `400`; var olmayan `{id}` (PUT/DELETE) → `404` + `{"error":"product not found"}`
  (read API'deki davranışla aynı).

**Başarı ölçütü:** Local Postgres ayaktayken (`docker compose up -d`) sunucu çalışır;
`curl -X POST` ile eklenen ürün `GET /api/products`'ta görünür, `PUT` ile güncellenen
alanlar `GET /api/products/{id}`'de yansır, `DELETE` sonrası aynı id `404` döner;
`go test ./...` (Docker + testcontainers ile) yeşil.

### 8.2 Komutlar (ek)

| Komut | Amaç |
|-------|------|
| `curl -s -X POST localhost:8080/api/products -d '{"name":"Fincan","price":249.90}'` | Ürün ekler → `201` |
| `curl -s -X PUT localhost:8080/api/products/1 -d '{"name":"Fincan","price":199.90}'` | Ürünü günceller → `200` |
| `curl -s -X DELETE localhost:8080/api/products/1 -i` | Ürünü siler → `204` |

### 8.3 Proje Yapısı (ek)

Yeni dosya gerekmez; mevcut `internal/product` paketi genişletilir:

```
internal/product/
  model.go          # Product struct (değişmez) + istek DTO'su ve doğrulama kuralları
  repository.go     # Repository arayüzüne Create/Update/Delete eklenir (GORM impl. dahil)
  handler.go        # Create/Update/Delete handler metotları eklenir
  handler_test.go   # yeni uçların testcontainers tabanlı testleri eklenir
internal/http/
  router.go         # POST/PUT/DELETE rotaları eklenir
```

### 8.4 Kod Stili (ek/değişiklik)

- `Repository` arayüzü genişler: `Create(ctx, Product) (Product, error)`,
  `Update(ctx, Product) (Product, error)`, `Delete(ctx, id uint) error`.
  Update/Delete, satır yoksa `ErrNotFound` döner (read API'deki desenle aynı).
- İstek gövdesi ayrı bir girdi tipine decode edilir (örn. `productInput{Name, Price}`);
  doğrulama bu tip üzerinde tek bir yerde yapılır ki POST ve PUT aynı kuralları paylaşsın.
- Handler'lar `writeJSON`/`writeError` yardımcılarını kullanmaya devam eder; yeni yanıt
  üretme yolu eklenmez.
- Rotalar `http.ServeMux` desenleriyle bağlanır: `"POST /api/products"`,
  `"PUT /api/products/{id}"`, `"DELETE /api/products/{id}"`.
- GORM kullanımı `internal/db` ve `internal/product` ile sınırlı kalmaya devam eder;
  `internal/health` ve `internal/http` stdlib-only kalır.

### 8.5 Test Stratejisi (ek)

- Mevcut testcontainers düzeni aynen kullanılır; mock repository ile "yeşil" gösterme yok.
- `internal/product/handler_test.go`'ya eklenecek senaryolar:
  - `POST` geçerli gövde → `201`, yanıtta `id` var; ardından `GET /api/products/{id}` aynı ürünü döner.
  - `POST` geçersiz JSON → `400`; boş/boşluk `name` → `400`; `price <= 0` → `400`;
    200 karakterden uzun `name` → `400`.
  - `PUT` var olan id + geçerli gövde → `200`, yanıt ve DB güncel; doğrulama hataları POST ile aynı → `400`.
  - `PUT`/`DELETE` olmayan id → `404` + JSON hata gövdesi; sayısal olmayan id → `400`.
  - `DELETE` var olan id → `204` + boş gövde; ardından `GET` aynı id → `404`.
- **Geçiş ölçütü:** `go test ./...` (Docker mevcutken) yeşil, `go vet ./...` temiz, `gofmt -l .` boş.

### 8.6 Sınırlar (ek/değişiklik)

**Her zaman yap (ek):**
- Doğrulama kurallarını POST ve PUT'ta tek bir ortak yerden uygula; iki kopya kural tutma.
- Yazma uçlarında da hata gövdesini `{"error":"..."}` formatında dön.
- İstemciden gelen `id`'yi asla yazma işleminde kullanma (path hariç).

**Önce sor (ek):**
- `PATCH` (kısmi güncelleme) eklemeden önce — bu dilim yalnızca tam `PUT` içerir.
- Ürün modeline/migration'a yeni alan eklemeden önce.
- Soft-delete, versiyonlama veya optimistic locking gibi davranışlar eklemeden önce.

**Asla yapma (ek):**
- Kimlik doğrulama/yetkilendirme ekleme (ileriki dilim; uçlar şimdilik korumasız ve bu bilinçli).
- Toplu (bulk) ekleme/silme uçları ekleme.
- Read API'nin mevcut davranışını (`GET` uçları, hata formatı, model alanları) değiştirme.

> **Not:** Dilim 3'teki *"kimlik doğrulama ekleme"* yasağı **o dilime** özgüydü. Dilim 4 (§9) bu
> kuralı bilinçli olarak kaldırır: JWT tabanlı auth ekler ve sepet/sipariş uçlarını korur.
> Ürün (`/api/products`) uçları bu dilimde **korumasız kalmaya devam eder** (değiştirilmez).

---

## 9. Dilim 4 — Auth + Sepet + Sipariş (Auth + Cart + Orders)

### 9.1 Amaç

Kullanıcılar kayıt olup giriş yapar, sepete ürün ekler, sepetini görür ve siparişini tamamlar.
**Her kullanıcı yalnızca kendi sepetini ve siparişlerini görür/değiştirir** — bu izolasyon,
JWT'den okunan `user_id` ile sağlanır; istemcinin gönderdiği hiçbir `user_id`'ye güvenilmez.

| Metot & Yol | Koruma | Başarı | Gövde (istek → yanıt) |
|---|---|---|---|
| `POST /api/register` | public | `201` | `{"email","password"}` → `{"id","email"}` (parola dönmez) |
| `POST /api/login` | public | `200` | `{"email","password"}` → `{"token"}` (JWT) |
| `POST /api/cart` | JWT | `201` | `{"product_id","quantity"}` → eklenen/güncellenen sepet kalemi |
| `GET /api/cart` | JWT | `200` | — → kullanıcının sepet kalemleri + satır/genel toplam |
| `POST /api/orders` | JWT | `201` | — → sepetten oluşturulan sipariş (kalemler + toplam) |

**Kimlik (Auth):**
- Login, `HS256` ile imzalanmış bir JWT döner; `sub = user_id`, `exp` sonlu (varsayılan 24 saat).
  İmza secret'ı `JWT_SECRET` env değişkeninden okunur (hardcode edilmez).
- Korumalı uçlar `Authorization: Bearer <jwt>` bekler. Token yoksa/geçersizse/süresi geçmişse
  → `401` + `{"error":"..."}`. Doğrulanan `user_id` request context'ine konur; handler'lar
  yalnızca oradan okur.
- Parolalar **bcrypt** ile hash'lenir; düz metin parola asla saklanmaz/loglanmaz.

**Sepet (Cart):**
- `cart` tablosunun her satırı bir sepet kalemidir: `(user_id, product_id, quantity)`.
- `POST /api/cart`: `product_id` var olan bir ürün olmalı (yoksa `404`); `quantity` tamsayı ve
  `> 0` (değilse `400`). Aynı ürün tekrar eklenirse yeni satır açılmaz, mevcut satırın
  `quantity`'si artırılır (`UNIQUE(user_id, product_id)`).
- `GET /api/cart`: yalnızca o kullanıcının kalemlerini, ürün adı/fiyatı, satır toplamı ve
  genel toplamla birlikte döner. Boş sepet → `200` + boş liste (hata değil).

**Sipariş (Order):**
- `POST /api/orders`: kullanıcının mevcut sepetini siparişe dönüştürür:
  1. `orders` satırı oluşturulur (`user_id`, `total`).
  2. Her sepet kalemi bir `order_items` satırına kopyalanır; **o anki ürün fiyatı
     `unit_price` olarak snapshot'lanır** (ürün fiyatı sonradan değişse bile sipariş sabit kalır).
  3. `total`, kalemlerin `quantity * unit_price` toplamıdır.
  4. Kullanıcının sepeti **boşaltılır**.
  Boş sepette çağrılırsa → `400` + `{"error":"cart is empty"}`.
  Tüm bu adımlar **tek bir DB transaction**'ı içinde yapılır (kısmi sipariş oluşmaz).

**Başarı ölçütü:** Local Postgres ayaktayken sunucu çalışır; `register` → `login` ile token
alınır; token ile `POST /api/cart` sepete ürün ekler, `GET /api/cart` doğru toplamı döner,
`POST /api/orders` siparişi oluşturup sepeti boşaltır; başka bir kullanıcının token'ıyla aynı
sepet/sipariş **görünmez**; token'sız korumalı istek `401` döner; `go test ./...` (Docker +
testcontainers ile) yeşil.

### 9.2 Komutlar (ek)

| Komut | Amaç |
|-------|------|
| `psql "$DATABASE_URL" -f migrations/0002_create_users.sql` | `users` tablosunu oluşturur |
| `psql "$DATABASE_URL" -f migrations/0003_create_cart.sql` | `cart` tablosunu oluşturur |
| `psql "$DATABASE_URL" -f migrations/0004_create_orders.sql` | `orders` tablosunu oluşturur |
| `psql "$DATABASE_URL" -f migrations/0005_create_order_items.sql` | `order_items` tablosunu oluşturur |
| `curl -s -X POST localhost:8080/api/register -d '{"email":"a@b.com","password":"parola123"}'` | Kayıt → `201` |
| `curl -s -X POST localhost:8080/api/login -d '{"email":"a@b.com","password":"parola123"}'` | Giriş → `{"token":"..."}` |
| `curl -s -X POST localhost:8080/api/cart -H "Authorization: Bearer $TOKEN" -d '{"product_id":5,"quantity":2}'` | Sepete ekle → `201` |
| `curl -s localhost:8080/api/cart -H "Authorization: Bearer $TOKEN"` | Sepeti gör → `200` |
| `curl -s -X POST localhost:8080/api/orders -H "Authorization: Bearer $TOKEN"` | Siparişi tamamla → `201` |

> `JWT_SECRET` env değişkeni sunucu başlatılırken verilmelidir (örn.
> `JWT_SECRET=dev-secret DATABASE_URL=... go run ./cmd/server`).

### 9.3 Proje Yapısı (ek)

```
vibe-shop/
  migrations/
    0002_create_users.sql        # users(id, email UNIQUE, password_hash, created_at)
    0003_create_cart.sql         # cart(id, user_id, product_id, quantity) + UNIQUE(user_id,product_id)
    0004_create_orders.sql       # orders(id, user_id, total numeric(10,2), created_at)
    0005_create_order_items.sql  # order_items(id, order_id, product_id, quantity, unit_price numeric(10,2))
  internal/
    auth/
      model.go         # User struct (GORM) — email, password_hash
      repository.go    # user erişimi: Create(user), GetByEmail(email)
      token.go         # golang-jwt ile JWT üret/doğrula (HS256, JWT_SECRET)
      middleware.go    # Bearer token → user_id'yi context'e koyar; yoksa 401
      handler.go       # POST /api/register, POST /api/login
      handler_test.go  # testcontainers tabanlı auth testleri
    cart/
      model.go         # CartItem struct (user_id, product_id, quantity)
      repository.go    # AddOrIncrement(userID,...), ListByUser(userID), ClearByUser(userID)
      handler.go       # POST /api/cart, GET /api/cart
      handler_test.go
    order/
      model.go         # Order + OrderItem struct'ları
      repository.go    # CreateFromCart(ctx, userID) — transaction; snapshot fiyat
      handler.go       # POST /api/orders
      handler_test.go
    http/
      router.go        # public register/login + JWT-korumalı cart/orders rotaları
```

- `cmd/server/main.go`: `JWT_SECRET` okunur ve auth katmanına verilir (mevcut wiring tarzıyla).
- Auth middleware, korumalı rota grubunu (cart, orders) sarar; ürün uçları sarılmaz.

### 9.4 Kod Stili (ek/değişiklik)

- **Yeni dış bağımlılıklar (onaylı):** `github.com/golang-jwt/jwt/v5` (yalnızca `internal/auth`)
  ve `golang.org/x/crypto/bcrypt` (yalnızca `internal/auth`). GORM kullanımı `internal/cart`,
  `internal/order`, `internal/auth` paketlerine genişler. `internal/health` ve `internal/http`
  stdlib-only kalır (router yalnızca kablolama yapar).
- **İzolasyon deseni:** her cart/order repository metodu `userID` parametresi alır ve sorguyu
  buna göre filtreler. Handler'lar `userID`'yi **yalnızca** `auth.UserIDFromContext(ctx)`'ten
  okur; istek gövdesinden/query'den asla değil.
- **JWT:** HS256, secret `JWT_SECRET`'ten. Claim `sub=user_id`, `exp` sonlu. Doğrulama
  `token.go`'da tek yerde; middleware bunu çağırır.
- **Parola:** `bcrypt.GenerateFromPassword` (varsayılan cost) ile hash; doğrulamada
  `bcrypt.CompareHashAndPassword`. Parola hiçbir yanıtta/loglamada görünmez.
- **Doğrulama:** istek gövdeleri ayrı input tiplerine (`registerInput`, `loginInput`,
  `cartInput`) decode edilir; doğrulama tek yerde. Mevcut `writeJSON`/`writeError` yardımcıları
  ve `{"error":"..."}` hata gövdesi kullanılmaya devam eder.
- **Transaction:** sipariş oluşturma `gorm` transaction (`db.Transaction(func(tx)...)`) içinde;
  order + order_items + sepet temizliği atomik.
- Rotalar `http.ServeMux` desenleriyle bağlanır: `"POST /api/register"`, `"POST /api/login"`,
  `"POST /api/cart"`, `"GET /api/cart"`, `"POST /api/orders"`.

### 9.5 Test Stratejisi (ek)

Mevcut testcontainers düzeni aynen kullanılır; mock repository ile "yeşil" gösterme yok.
Her test paketi geçici Postgres container'ı ayağa kaldırır, ilgili migration'ları uygular.

- **auth (`internal/auth/handler_test.go`):**
  - `register` geçerli → `201`; yanıt parola/hash içermez.
  - `register` var olan email → `409` (veya `400`) + JSON hata.
  - `register` geçersiz email / kısa parola → `400`.
  - `login` doğru kimlik → `200` + geçerli JWT; yanlış parola / olmayan email → `401`.
  - middleware: token'sız / bozuk / süresi geçmiş token → `401`; geçerli token → context'te doğru `user_id`.
- **cart (`internal/cart/handler_test.go`):**
  - `POST /api/cart` geçerli → `201`; aynı ürünü tekrar ekleyince `quantity` artar (yeni satır yok).
  - olmayan `product_id` → `404`; `quantity <= 0` veya tamsayı değil → `400`; token'sız → `401`.
  - `GET /api/cart` yalnızca o kullanıcının kalemlerini + doğru toplamı döner.
  - **İzolasyon:** kullanıcı A'nın eklediği ürün, kullanıcı B'nin `GET /api/cart`'ında görünmez.
- **order (`internal/order/handler_test.go`):**
  - dolu sepette `POST /api/orders` → `201`; `order_items` fiyatı sipariş anındaki ürün fiyatı;
    ardından ürün fiyatı değişse bile sipariş `total` değişmez; sepet boşalır.
  - boş sepette `POST /api/orders` → `400` + `{"error":"cart is empty"}`.
  - **İzolasyon:** A'nın siparişi B tarafından oluşturulamaz/görülemez; `user_id` daima token'dan.
- **Geçiş ölçütü:** `go test ./...` (Docker mevcutken) yeşil, `go vet ./...` temiz, `gofmt -l .` boş.

### 9.6 Sınırlar (ek/değişiklik)

**Her zaman yap (ek):**
- `user_id`'yi **yalnızca** doğrulanmış JWT'den (context) al; istek gövdesi/query'deki `user_id`'ye asla güvenme.
- Her cart/order DB sorgusunu `user_id`'ye göre filtrele (sahiplik zorunlu).
- `JWT_SECRET`'i env'den oku; parolaları bcrypt ile hash'le.
- Sipariş oluşturmayı tek transaction içinde yap; fiyatı `order_items`'a snapshot'la.
- Doğrulama kurallarını tek ortak yerde uygula; hata gövdesini `{"error":"..."}` formatında dön.

**Önce sor (ek):**
- Refresh token, rol/yetki (admin vb.), parola sıfırlama, email doğrulama eklemeden önce.
- Token süresi (`exp`) varsayılanını değiştirmeden önce.
- Sepetten kalem silme / adet düşürme (`DELETE /api/cart/...`, `PATCH`) uçları eklemeden önce.
- Stok/envanter kontrolü veya ödeme entegrasyonu eklemeden önce.
- `migrations/` şemasını (yeni alan/tablo) değiştirmeden önce.

**Asla yapma (ek):**
- Düz metin parola saklama/dönme/loglama; token'ları loglama.
- İstemciden gelen `user_id` ile bir kullanıcının verisine erişme.
- Bir kullanıcının sepet/siparişini başka kullanıcıya döndürme.
- `JWT_SECRET`'i koda hardcode etme veya repoya commit etme.
- Ürün (`/api/products`) uçlarının mevcut davranışını değiştirme (bu dilimde korumasız kalırlar).
- Siparişi transaction dışında, kısmi (order var ama items yok gibi) oluşturma.

> **Not:** Dilim 4'teki *"ürün uçlarının davranışını değiştirme"* yasağı ve HS256 JWT tabanlı
> kimlik tasarımı **o dilime** özgüdür. Dilim 5 (§10) bilinçli olarak: (a) ürün **yazma**
> uçlarını korumaya alır (okuma uçları ve tüm iş mantığı/yanıt gövdeleri değişmez), (b) local
> auth'u (`/api/register`, `/api/login`, HS256 token, bcrypt, `users` tablosu) **kaldırıp**
> tek kimlik sağlayıcı olarak **Keycloak**'a geçer.

---

## 10. Dilim 5 — Keycloak'a Geçiş: Tek Kimlik Sağlayıcı (Keycloak-Only Auth)

### 10.1 Amaç

Uygulamanın kendi kimlik sistemi (Dilim 4: `/api/register`, `/api/login`, HS256 JWT, bcrypt,
`users` tablosu) **kaldırılır**; tek kimlik sağlayıcı **Keycloak** olur. Keycloak local'de
Docker ile çalışır; kullanıcı token'ını Keycloak'tan alır, API her korumalı istekte bu token'ı
**doğrular** (imza + issuer + süre). Ürün **yazma** uçları da artık korumalıdır; **okuma**
uçları herkese açık kalır. API kendi login/register ucunu **açmaz**.

| Metot & Yol | Koruma | Değişiklik |
|---|---|---|
| `GET /api/products`, `GET /api/products/{id}` | public | değişmez |
| `POST /api/products` | Keycloak Bearer | **yeni:** korumaya alınır; iş mantığı/`201` değişmez |
| `PUT /api/products/{id}` | Keycloak Bearer | **yeni:** korumaya alınır; `200` değişmez |
| `DELETE /api/products/{id}` | Keycloak Bearer | **yeni:** korumaya alınır; `204` değişmez |
| `POST /api/cart`, `GET /api/cart` | Keycloak Bearer | koruma HS256 → Keycloak; iş mantığı değişmez |
| `POST /api/orders` | Keycloak Bearer | koruma HS256 → Keycloak; iş mantığı değişmez |
| `POST /api/register`, `POST /api/login` | — | **kaldırılır** (`404`); kullanıcı yönetimi Keycloak'ta |

- Token yok / `Bearer` değil / bozuk / süresi geçmiş / imzası geçersiz / yanlış issuer →
  `401` + `{"error":"..."}` (mevcut hata gövdesi formatı).
- Bu dilimde **rol/yetki (authorization) yok**: geçerli token'ı olan her kullanıcı yazabilir;
  `403` senaryosu yok. Rol bazlı kısıt ileriki dilim ("önce sor").

**Kullanıcı kimliği ve izolasyon:**
- Kimlik, Keycloak token'ının **`sub`** claim'idir (Keycloak kullanıcı UUID'si, **string**).
- "Her kullanıcı yalnızca kendi sepetini/siparişini görür" kuralı aynen sürer; `user_id` artık
  local `users.id` değil Keycloak `sub`'ıdır. Middleware doğrulanan `sub`'ı request context'ine
  koyar; handler'lar kimliği **yalnızca** oradan okur (Dilim 4 deseni korunur).
- **Şema değişikliği (migration 0006):** `cart.user_id` ve `orders.user_id` `TEXT`'e çevrilir,
  `users` tablosuna giden FK'lar kaldırılır ve `users` tablosu **düşürülür**.
  `UNIQUE(user_id, product_id)` korunur. Dev-only veri olduğundan temiz geçiş için
  `docker compose down -v` ile sıfırdan başlamak önerilir.

**Keycloak (local, Docker):**
- `docker-compose.yml`'e `keycloak` servisi eklenir: `quay.io/keycloak/keycloak:26.3`,
  `start-dev --import-realm`, host portu **8081** (API 8080'de olduğu için).
- Realm **dosyadan otomatik import** edilir (`keycloak/vibe-shop-realm.json`): realm `vibe-shop`,
  public client `vibe-shop-api` (Direct Access Grants açık — curl ile parola grant'i alınabilsin),
  **iki** test kullanıcısı: `testuser`/`test1234` ve `testuser2`/`test1234` (izolasyon
  senaryoları için). Elle admin konsolu tıklaması **gerekmez**; `docker compose up -d` yeterli.
- Kayıt/parola yönetimi artık API'nin işi değildir: kullanıcılar Keycloak'ta yaratılır
  (dev'de realm import; gerekirse admin konsolu `http://localhost:8081`, `admin`/`admin` —
  yalnızca local dev, §7'deki dev-only credential istisnasıyla aynı).
- Docker'dan gelen istekler Keycloak'a dış IP gibi göründüğünden `sslRequired` varsayılanı
  HTTP'yi engeller: `vibe-shop` realm'i bunu import dosyasındaki `"sslRequired": "none"` ile,
  **master** realm (admin konsolu) ise compose'taki tek seferlik `keycloak-init` servisiyle
  (`kcadm update realms/master -s sslRequired=NONE`) çözer — ikisi de yalnızca local dev içindir.

**Token doğrulama (API tarafı):**
- Keycloak **RS256** imzalı JWT verir. API, imzayı Keycloak'ın **JWKS** ucundan
  (`<issuer>/protocol/openid-connect/certs`) aldığı public key ile doğrular; `iss` claim'i
  `KEYCLOAK_ISSUER_URL` ile birebir eşleşmeli ve `exp` geçerli olmalı; `sub` boş olamaz.
- Yalnızca `RS256` kabul edilir; `alg=none` veya HS256 imzalı token'lar ("alg confusion"
  saldırısı, örn. eski `/api/login` token'ları) **reddedilir**.
- Issuer `KEYCLOAK_ISSUER_URL` env değişkeninden okunur
  (örn. `http://localhost:8081/realms/vibe-shop`); hardcode edilmez. `JWT_SECRET` artık **yoktur**.

**Başarı ölçütü:** `docker compose up -d` ile Postgres + Keycloak ayakta; Keycloak'tan parola
grant'iyle token alınır; token'sız yazma/sepet/sipariş istekleri → `401`; token ile ürün yazma
`201`/`200`/`204` döner ve sepet/sipariş akışı (ekle → gör → sipariş → sepet boşalır) çalışır;
`testuser2` token'ı `testuser`'ın sepetini/siparişini **göremez**; `GET /api/products*` token'sız
`200`; `/api/register` ve `/api/login` → `404`; `go test ./...` yeşil.

### 10.2 Komutlar (ek/değişiklik)

| Komut | Amaç |
|-------|------|
| `docker compose up -d` | Postgres **ve** Keycloak'ı başlatır (Keycloak: `http://localhost:8081`) |
| `psql "$DATABASE_URL" -f migrations/0006_switch_to_keycloak_identity.sql` | `cart`/`orders.user_id` → `TEXT`; `users` tablosunu düşürür |
| `curl -s -X POST http://localhost:8081/realms/vibe-shop/protocol/openid-connect/token -d grant_type=password -d client_id=vibe-shop-api -d username=testuser -d password=test1234` | Keycloak'tan token alır → `{"access_token":"...",...}` |
| `curl -s -X POST localhost:8080/api/products -H "Authorization: Bearer $KC_TOKEN" -d '{"name":"Fincan","price":249.90}'` | Token'la ürün ekler → `201` |
| `curl -s -X POST localhost:8080/api/products -d '{"name":"Fincan","price":249.90}' -i` | Token'sız yazma → `401` |
| `curl -s -X POST localhost:8080/api/cart -H "Authorization: Bearer $KC_TOKEN" -d '{"product_id":1,"quantity":2}'` | Keycloak token'ıyla sepete ekler → `201` |
| `KEYCLOAK_ISSUER_URL=http://localhost:8081/realms/vibe-shop DATABASE_URL=... go run ./cmd/server` | Sunucuyu başlatır (`JWT_SECRET` artık kullanılmaz) |

### 10.3 Proje Yapısı (ek/değişiklik)

```
vibe-shop/
  docker-compose.yml            # + keycloak servisi (8081, --import-realm)
  keycloak/
    vibe-shop-realm.json        # realm + public client + 2 test kullanıcısı (otomatik import)
  .env.example                  # + KEYCLOAK_ISSUER_URL (JWT_SECRET kalkar)
  migrations/
    0006_switch_to_keycloak_identity.sql  # cart/orders.user_id → TEXT; FK'lar + users tablosu düşer
  internal/
    auth/
      keycloak.go               # KeycloakVerifier (JWKS, RS256) + RequireAuth middleware + SubjectFromContext
      keycloak_test.go          # httptest JWKS ile birim testler (container gerekmez)
      # SİLİNİR: token.go, middleware.go, handler.go, repository.go, model.go (+ testleri)
    cart/                       # UserID string (Keycloak sub) — model/repository/handler/testler güncellenir
    order/                      # UserID string (Keycloak sub) — model/repository/handler/testler güncellenir
    http/
      router.go                 # register/login rotaları silinir; tüm korumalı rotalar tek middleware'le sarılır
  cmd/server/main.go            # JWT_SECRET/TokenManager kalkar; KEYCLOAK_ISSUER_URL + verifier kurulur
  Makefile                      # start, Keycloak hazır olana dek bekler
  api.http                      # register/login örnekleri silinir; tüm korumalı örnekler Keycloak token'ına geçer
```

### 10.4 Kod Stili (ek/değişiklik)

- **Yeni dış bağımlılık (onaylı):** `github.com/MicahParks/keyfunc/v3` — JWKS indirme,
  önbellekleme ve anahtar rotasyonunu üstlenir; mevcut `golang-jwt/jwt/v5` ile doğrudan entegre
  çalışır. Yalnızca `internal/auth` içinde kullanılır. (Alternatif olan elle JWK çözümleme,
  anahtar rotasyonu/önbellekleme kodunu bize yıkardı; keyfunc küçük ve tek amaçlıdır.)
- **Kaldırılan bağımlılıklar/kod:** `golang.org/x/crypto/bcrypt` (kullanan kod silinince
  `go mod tidy` ile düşer); `internal/auth`'tan `token.go`, `middleware.go`, `handler.go`,
  `repository.go`, `model.go` ve testleri. `golang-jwt/jwt/v5` **kalır** (RS256 parse için).
  Ölü kod bırakılmaz.
- **`KeycloakVerifier`:** `NewKeycloakVerifier(issuerURL string) (*KeycloakVerifier, error)` —
  JWKS URL'i issuer'dan türetilir (`<issuer>/protocol/openid-connect/certs`). Başlangıçta JWKS
  erişilemezse hata döner; `main.go` bunu `log.Fatal` yapar (`DATABASE_URL` ile aynı fail-fast
  tarzı — Keycloak ayakta değilse sunucu hiç başlamaz, sessiz kırıklık olmaz).
- **Doğrulama tek yerde:** `Verify(tokenStr) (sub string, err error)` — `jwt.Parse` +
  `jwt.WithValidMethods([]string{"RS256"})` + `iss` eşitlik kontrolü + boş olmayan `sub`;
  `exp` kütüphanece zorlanır. Tüm hatalar `ErrInvalidToken`'a katlanır → middleware tek bir
  `401` döner (eski `TokenManager.Parse` deseniyle aynı).
- **Middleware:** `(v *KeycloakVerifier) RequireAuth(next http.HandlerFunc) http.HandlerFunc` —
  mevcut `apphttp.Middleware` tipiyle uyumlu; Bearer okuma deseni eski middleware ile birebir.
  Doğrulanan `sub` context'e konur; erişimci **`auth.SubjectFromContext(ctx) (string, bool)`**
  (eski `UserIDFromContext(uint)`'in yerini alır). Handler'lar kimliği yalnızca buradan okur.
- **cart/order kimlik tipi:** modellerde `UserID string`; repository metotları `userID string`
  alır; tüm sorgular yine `user_id` filtrelidir (izolasyon deseni değişmez, yalnızca tip değişir).
  İş mantığı, doğrulama kuralları ve yanıt gövdeleri **değişmez**.
- **Router sadeleşir:** `NewRouter(products, cartH, ordersH, requireAuth Middleware)` —
  register/login rotaları silinir; ürün yazma + cart + orders rotalarının tümü aynı
  `requireAuth` ile sarılır. `internal/http` stdlib-only kalır.
- Ürün handler/repository/model kodu **değişmez** (yalnızca rotaları korumaya alınır).

### 10.5 Test Stratejisi (ek/değişiklik)

- **`internal/auth/keycloak_test.go` (container gerekmez):** test kendi RSA anahtar çiftini
  üretir, public key'i JWKS JSON'u olarak bir `httptest` sunucusundan servis eder; verifier bu
  JWKS'e karşı **gerçek kriptografik imza doğrulaması** yapar (doğrulamanın kendisi mock'lanmaz).
  Senaryolar:
  - geçerli RS256 token → geçer, doğru `sub` döner;
  - süresi geçmiş → hata; yanlış `iss` → hata; farklı anahtarla imzalanmış → hata; `sub` boş → hata;
  - HS256 imzalı veya `alg=none` token → hata (alg confusion);
  - bilinmeyen `kid` → hata; bozuk/boş string → hata.
  - middleware: header yok / `Bearer` değil / geçersiz token → `401` + `{"error":"..."}`;
    geçerli token → `next` çağrılır ve context'te doğru `sub` bulunur.
- **cart/order testleri (testcontainers, mevcut düzen):** senaryolar Dilim 4'tekiyle aynı kalır;
  yalnızca token üretimi değişir — testler httptest JWKS'li gerçek verifier + test RSA anahtarıyla
  imzalanmış token'lar kullanır. İzolasyon senaryoları iki farklı `sub` ile koşulur. Init script
  listesine `0006` eklenir (migration'lar **numara sırasıyla**: 0001…0006 — 0002 `users`'ı kurar,
  0006 düşürür; sıra bozulmadığı sürece sorun değildir).
- **`internal/http/router_test.go`:** httptest JWKS'li gerçek verifier ile — token'sız
  `POST/PUT/DELETE /api/products` ve cart/orders istekleri → `401`; geçerli token → handler'a
  ulaşır (sahte repository); token'sız `GET /api/products` → `200`; `POST /api/register` ve
  `POST /api/login` → `404`; `/health` değişmez.
- **Gerçek Keycloak** manuel uçtan uca checkpoint'te doğrulanır (docker compose + curl akışı).
  Not: testcontainers-go'nun resmi bir Keycloak modülü yok; community bağımlılığı eklemek
  yerine gerçek Keycloak entegrasyonu checkpoint'te kanıtlanır — imza doğrulaması birim
  testlerde zaten gerçektir. (Bu, "testcontainers'sız iş görme" kuralının bilinçli ve sınırlı
  bir istisnasıdır; DB'ye dokunan hiçbir test mock'a dönmez.)
- **Geçiş ölçütü:** `go test ./...` (Docker mevcutken) yeşil, `go vet ./...` temiz, `gofmt -l .` boş.

### 10.6 Sınırlar (ek/değişiklik)

**Her zaman yap (ek):**
- Yalnızca `RS256` kabul et; `alg=none`/HS256 token'ları reddet.
- İmza + `iss` + `exp` üçünü birden doğrula; doğrulamayı tek yerde (`keycloak.go`) tut.
- Kimliği **yalnızca** doğrulanmış token'ın `sub`'ından (context) al; istek gövdesi/query'deki
  hiçbir kimlik bilgisine güvenme (Dilim 4 kuralının devamı).
- Her cart/order sorgusunu `user_id` (`sub`) ile filtrele (sahiplik zorunlu, değişmedi).
- `GET /api/products*` ve `/health` uçlarını public bırak.
- `KEYCLOAK_ISSUER_URL`'i env'den oku; issuer/JWKS URL'lerini hardcode etme.
- Realm import dosyasında yalnızca dev-only test kullanıcıları/parolaları tut.

**Önce sor (ek):**
- Rol/yetki (`403`) veya audience (`aud`/`azp`) kontrolü eklemeden önce.
- `sub` dışında bir claim'i (email, preferred_username vb.) kullanmadan/saklamadan önce.
- Keycloak imaj sürümünü, portu (8081), realm/client/test-kullanıcı adlarını değiştirmeden önce.
- `keyfunc` dışında yeni bağımlılık eklemeden önce.
- `migrations/` şemasını 0006'nın ötesinde değiştirmeden önce.

**Asla yapma (ek):**
- Token'ı imza doğrulamadan decode edip claim'lerine güvenme.
- Access token'ları loglama/yanıtta yansıtma.
- Okuma uçlarına veya `/health`'e auth ekleme.
- Ürün/sepet/sipariş iş mantığını, doğrulama kurallarını veya yanıt gövdelerini değiştirme
  (yalnızca kimlik katmanı değişir).
- Eski auth'tan ölü kod bırakma (`users` tablosu, bcrypt, HS256 token kodu "belki lazım olur"
  diye tutulmaz).
- `keycloak/vibe-shop-realm.json`'daki dev credential'ları production için kullanma/önerme.

---

## 11. Dilim 6 — Frontend (SPA): Ürünler, Sepet ve Keycloak Girişi

### 11.1 Amaç

vibe-shop'a, Go API'sini (:8080) tüketen bir web arayüzü eklenir. Görsel yön `design/`
klasöründeki mockup'lardır (shadcn/ui görünümü); sayfalar bu çizimlere **mümkün olan en yakın**
şekilde, shadcn/ui bileşenleriyle inşa edilir.

| Sayfa | Rota | Kaynak tasarım | Veri |
|---|---|---|---|
| Login | `/login` | `design/01-login.png` | Keycloak token endpoint'i |
| Ürün listesi | `/` | `design/02-product-list.png` | `GET /api/products` |
| Ürün detay | `/products/:id` | `design/03-product-detail.png` | `GET /api/products/{id}` + `POST /api/cart` |
| Sepet | `/cart` | `design/04-cart.png` | `GET /api/cart` + `POST /api/orders` |
| Sipariş onayı | `/cart` akışının sonucu | `design/05-order-confirmation.png` | `POST /api/orders` yanıtı |

- **Koruma kuralı (kullanıcı kararı):** giriş yapmamış kullanıcı **hangi rotaya girerse girsin**
  `/login`'e yönlendirilir — ürün listesi API'de public olsa bile SPA'da giriş ister.
  Login olan kullanıcı `/login`'e giderse `/`'a yönlendirilir.
- **Giriş (Keycloak, ROPC):** login sayfasındaki e-posta/parola formu, Keycloak'ın token
  endpoint'ine **Direct Access Grant** (`grant_type=password`, client `vibe-shop-api` — realm'de
  zaten açık) ile doğrudan istek atar; dönen `access_token`/`refresh_token` saklanır. Bu bilinçli
  bir sadelik tercihidir (dev/öğrenme projesi): Authorization Code + PKCE yönlendirme akışına
  geçiş ayrı bir dilim adayıdır ("önce sor"). Çıkış = token'ları silip `/login`'e dönmek.
- **Token yenileme:** Keycloak access token'ı kısa ömürlüdür (~5 dk). API çağrısı `401` dönerse
  istemci **bir kez** `refresh_token` ile yeniler ve isteği tekrarlar; yenileme de başarısızsa
  token'lar silinir ve `/login`'e yönlendirilir. Realm'in token ömürlerini değiştirmek "önce sor".
- **CORS/ağ:** Go API'ye CORS eklenmez; dev'de Vite proxy'si `/api` → `http://localhost:8080`'e
  iletir (SPA istekleri hep göreli `/api/...`). Keycloak'a tarayıcıdan doğrudan istek için
  client'a `webOrigins: ["http://localhost:5173"]` eklenir (realm import dosyasında).
- **Bilinen tasarım sapmaları (bilinçli):**
  - `design/04-cart.png`'deki adet artır/azalt ve satır silme kontrolleri **bu dilimde yok** —
    backend desteklemiyor (SPEC §9.6 "önce sor"); sepet satırları adetleriyle salt-okunur gösterilir.
  - `design/01-login.png`'deki "Keycloak ile devam et" butonu yoktur; form zaten Keycloak'a
    gittiği için ikinci bir yol koymak kafa karıştırır. Form alanları + hata durumu birebir uygulanır.
  - Admin sayfası (`design/06`) bu dilimin kapsamı dışındadır.

**Başarı ölçütü:** Backend + Keycloak ayaktayken `npm run dev` ile SPA açılır; token'sız her
rota `/login`'e düşer; `testuser`/`test1234` ile giriş çalışır; ürünler API'den listelenir,
detay sayfasından sepete ürün eklenir, sepette doğru toplam görünür, "Siparişi Tamamla" siparişi
oluşturup onay görünümünü (sipariş no + kalemler + toplam) gösterir ve sepet boşalır; çıkış
sonrası korumalı rotalar yine `/login`'e düşer; `npm run build`, `npm run lint` ve
`npm run test` temiz.

### 11.2 Komutlar (ek)

| Komut | Amaç |
|-------|------|
| `cd frontend && npm install` | Frontend bağımlılıklarını kurar |
| `cd frontend && npm run dev` | Vite dev sunucusu — `http://localhost:5173` (proxy: `/api` → `:8080`) |
| `cd frontend && npm run build` | Production build (`tsc` + `vite build`) |
| `cd frontend && npm run test` | Vitest ile birim/bileşen testleri |
| `cd frontend && npm run lint` | ESLint |

> Tam akış için üç süreç gerekir: `make start` (Postgres + Keycloak + API) ve `npm run dev`.

### 11.3 Proje Yapısı (ek)

```
vibe-shop/
  frontend/                    # SPA — Go modülünden tamamen ayrı
    package.json               # React 19 + Vite + TypeScript
    vite.config.ts             # @tailwindcss/vite + /api proxy → :8080
    components.json            # shadcn/ui yapılandırması
    src/
      main.tsx                 # router + AuthProvider kablolaması
      lib/
        api.ts                 # fetch sarmalayıcı: Bearer ekler, 401'de bir kez refresh
        auth.tsx               # AuthContext: login (ROPC), logout, token saklama
      components/
        ui/                    # shadcn/ui bileşenleri (üretilmiş)
        navbar.tsx             # wordmark + Ürünler + sepet rozeti + kullanıcı/çıkış
        require-auth.tsx       # korumalı rota sarmalayıcı → /login
      pages/
        login.tsx              # design/01
        products.tsx           # design/02
        product-detail.tsx     # design/03
        cart.tsx               # design/04 + onay görünümü design/05
      pages/*.test.tsx         # Vitest + Testing Library + MSW
  keycloak/vibe-shop-realm.json  # + webOrigins (5173)
```

### 11.4 Kod Stili (ek)

- **Stack:** Vite + React 19 + TypeScript (strict) + Tailwind CSS + **shadcn/ui** (zinc teması).
  Next.js/SSR yok — API ayrı olduğundan SPA yeterli, proje sadeliği korunur.
- **Bileşenler:** shadcn/ui üretilen bileşenler `components/ui/` altında kalır ve **elle
  değiştirilmez**; özelleştirme kompozisyonla yapılır. Sayfalar `pages/`, paylaşılanlar
  `components/` altında; dosya adları `kebab-case.tsx`.
- **Veri erişimi tek yerden:** tüm HTTP istekleri `lib/api.ts` üzerinden geçer (göreli `/api/...`);
  sayfalar `fetch`'i doğrudan çağırmaz. Token yalnızca `lib/auth.tsx` tarafından yönetilir.
- **Görünüm dili (mockup'larla eşleşen):** zinc nötr palet, tek koyu birincil buton,
  `zinc-100` zeminli ürün görselleri, ince `zinc-200` border'lı `rounded-lg` kartlar, Inter.
  Fiyat biçimi `₺249,90` (`Intl.NumberFormat('tr-TR')`).
- Durum yönetimi için ek kütüphane yok (React state + context yeterli); yeni bağımlılık "önce sor".

### 11.5 Test Stratejisi (ek)

- **Çerçeve:** Vitest + React Testing Library + **MSW** (ağ katmanı testte taklit edilir —
  DB'siz frontend testinde endüstri standardı; gerçek API/Keycloak entegrasyonu manuel
  checkpoint'te kanıtlanır, dilim 5'teki desenle aynı).
- Kapsanacak davranışlar:
  - token'sız korumalı rota → `/login`'e yönlendirme; login sonrası hedefe dönüş.
  - login: geçersiz kimlik → form hatası (`invalid_grant` → "E-posta veya parola hatalı");
    başarı → token saklanır ve yönlendirilir.
  - `lib/api.ts`: `401` → bir kez refresh + retry; refresh başarısız → token'lar silinir, `/login`.
  - ürün listesi: API'den gelen ürünler kart olarak render edilir; boş liste durumu.
  - ürün detay: adet seçimi + "Sepete Ekle" → doğru gövde ve `Authorization` header'ı ile `POST /api/cart`.
  - sepet: kalemler + satır/genel toplam; "Siparişi Tamamla" → onay görünümü; boş sepet durumu.
- **Geçiş ölçütü:** `npm run test` yeşil · `npm run build` temiz · `npm run lint` temiz ·
  Go tarafında `go test ./...` etkilenmeden yeşil kalır.

### 11.6 Sınırlar (ek/değişiklik)

**Her zaman yap (ek):**
- Veriyi yalnızca Go API'den al; SPA içinde ürün/sepet verisi hardcode etme.
- Token'ı yalnızca `lib/auth.tsx` yönetsin; `Authorization` header'ı tek yerden eklensin.
- Tasarım kararlarında `design/` mockup'larını referans al; bilinçli sapmaları SPEC'e yaz.
- Backend'e ve `migrations/`'a dokunma (yalnızca `keycloak/vibe-shop-realm.json`'a `webOrigins` eklenir).

**Önce sor (ek):**
- Yeni npm bağımlılığı eklemeden önce (kurulum iskeletinin getirdikleri + MSW hariç).
- ROPC yerine Authorization Code + PKCE'ye geçmeden önce; realm token ömürlerini değiştirmeden önce.
- Sepette adet azaltma/satır silme için backend ucu eklemeden önce (ayrı dilim).
- Admin sayfasını (design/06) inşa etmeden önce; Go API'ye CORS eklemeden önce.
- Makefile'a frontend hedefleri eklemeden önce.

**Asla yapma (ek):**
- Token'ları URL'de taşıma veya loglama; parolayı Keycloak dışına gönderme.
- `components/ui/` altındaki üretilmiş shadcn dosyalarını elle yamalama.
- API'nin davranışına SPA tarafında güvenmemek gerekeni: istemci doğrulaması API doğrulamasının
  yerine geçmez (400'ler yine de düzgün gösterilir).
- Backend Go kodunda bu dilim kapsamında değişiklik yapma.
