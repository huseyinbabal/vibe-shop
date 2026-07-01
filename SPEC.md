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

### Yol haritası (bilgi amaçlı, bu dilimde uygulanmayacak)
1. **İskelet + /health** ← *şu anki dilim*
2. Ürün listeleme (`GET /products`)
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
