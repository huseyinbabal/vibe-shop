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
