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
3. **Ürün yazma API'si** (`POST`, `PUT`, `DELETE /api/products`) ← *şu anki dilim, bkz. §8*
4. Sepet (`POST /cart`, ...)
5. Sipariş (`POST /orders`, ...)

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
| `DATABASE_URL=postgres://vibeshop:vibeshop@localhost:5432/vibeshop?sslmode=disable go run ./cmd/server` | Sunucuyu başlatır; açılışta bekleyen migration'ları otomatik uygular (bkz. §7.7), sonra 8080'de dinler |
| `go test ./...` | Testleri çalıştırır (testcontainers için Docker gerektirir) |

### 7.3 Proje Yapısı (ek)

```
vibe-shop/
  docker-compose.yml           # local Postgres container (dev-only)
  .env.example                 # DATABASE_URL örneği (docker-compose ile eşleşir)
  migrations/
    embed.go                   # *.sql dosyalarını binary'e gömer (//go:embed)
    0001_create_products.sql   # products tablosu şeması (id, name, price)
  internal/
    db/
      db.go                    # DATABASE_URL'den GORM bağlantısı kurar
    migrate/
      migrate.go               # açılışta migration'ları uygular, schema_migrations'ta izler
      migrate_test.go          # testcontainers ile: şema oluşumu + idempotency
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
- `internal/migrate/migrate_test.go`: init script'siz temiz bir container'da runner'ı
  çalıştırır; `products` tablosunun oluştuğunu ve runner'ın iki kez çalıştırılınca
  migration'ı tekrar uygulamadığını (idempotency) kanıtlar.
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

### 7.7 Migration Mekanizması

Şema, elle `psql` çalıştırmadan yönetilir. `migrations/` altındaki `*.sql` dosyaları binary'e
gömülüdür (`migrations/embed.go`, `//go:embed`) ve `internal/migrate` sunucu açılışında bunları
uygular (`cmd/server/main.go` → `migrate.Run(db, migrations.FS)`).

**Nasıl çalışır:**
- Uygulanan migration'lar `schema_migrations(version TEXT PK, applied_at TIMESTAMPTZ)` tablosunda
  izlenir; `version`, dosya adının `.sql`'siz hâlidir (örn. `0001_create_products`).
- Dosyalar **ada göre sıralı** uygulanır → sırayı sayısal önek belirler (`0001_`, `0002_`, …).
- Her migration kendi transaction'ında uygulanır; `schema_migrations`'a kaydı olan atlanır.
  Bu yüzden runner her açılışta güvenle çalışır (idempotent) ve uygulanan migration iki kez çalışmaz.
- Dosya tek bir batch olarak yürütülür; `Exec` bind argümanı almadığından pgx **simple protocol**
  kullanır, yani bir dosya birden fazla ifade içerebilir.

**Yeni migration eklerken:**
- `migrations/` altına bir sonraki sıralı öneke sahip yeni bir `NNNN_aciklama.sql` dosyası ekle.
  Ekstra kayıt/kod gerekmez; açılışta otomatik uygulanır.
- Uygulanmış bir migration dosyasını **sonradan düzenleme**; yeni bir migration ile ileri al.
- Şemayı zaten var olan bir DB'ye runner'ı ilk kez bağlarken, mevcut şemaya karşılık gelen
  version(lar)ı `schema_migrations`'a elle ekleyerek **baseline'la** (yoksa runner var olan
  nesneleri yeniden oluşturmaya çalışıp hata verir).

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
