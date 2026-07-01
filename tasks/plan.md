# vibe-shop — İlk Dilim Planı

> Kaynak: [SPEC.md](../SPEC.md). Kapsam: **boş iskelet + çalışan `GET /health`**.
> Ürün/sepet/sipariş bu dilimde YOK.

## Bağımlılık Grafiği

```
go.mod  (T1)
  └──> internal/health/handler.go        (T2)  ── testi: handler_test.go (T3, aynı dilim)
  └──> internal/http/router.go           (T4)  ← health.Handler'a bağlanır
          └──> cmd/server/main.go         (T5)  ← router'ı 8080'de dinler
                  └──> uçtan uca doğrulama (T6)
```

- `go.mod` her şeyin ön koşulu (modül yolu `vibe-shop`).
- `router.go`, `health` paketine bağımlı → health handler'ı önce gelir.
- `main.go`, `router`'a bağımlı → en sonda kablolanır.
- Dış bağımlılık yok; grafik tamamen stdlib içinde.

## Dikey Dilimleme Yaklaşımı

Tek bir çalışır uçtan-uca yol hedefliyoruz: **istek → router → health handler → JSON yanıt**.
Katman katman (önce tüm handler'lar, sonra tüm router) DEĞİL; en küçük tam yol.
İlk dilim tek bir vertical slice olduğundan tasks küçük ve sıralı tutuldu.

## Fazlar ve Checkpoint'ler

### Faz 1 — Derlenen iskelet
- **T1** `go.mod` (modül `vibe-shop`, go 1.26, dış require yok)
- **CHECKPOINT A:** `go build ./...` başarılı (henüz kod yok, boş modül derlenir).

### Faz 2 — Health yolu (test-önce)
- **T2** `internal/health/handler.go` — `GET /health` → 200 + `application/json` + `{"status":"ok"}`
- **T3** `internal/health/handler_test.go` — httptest ile status/content-type/gövde doğrular
- **CHECKPOINT B:** `go test ./internal/health/` yeşil.

### Faz 3 — Kablolama ve çalıştırma
- **T4** `internal/http/router.go` — `http.ServeMux`, `/health` → `health.Handler`
- **T5** `cmd/server/main.go` — router'ı kur, `:8080` dinle
- **CHECKPOINT C:** `go run ./cmd/server` ayakta + `curl localhost:8080/health` → `{"status":"ok"}`

### Faz 4 — Kalite kapısı
- **T6** Tüm doğrulamalar: `gofmt -l .` boş, `go vet ./...` temiz, `go test ./...` yeşil
- **CHECKPOINT D (final):** İnsan onayı → dilim tamam.

## Kapsam Dışı (bilerek)
- Ürün/sepet/sipariş endpoint'leri
- Veritabanı / config / env okuması
- Dış bağımlılıklar (router kütüphaneleri dahil)
- Structured logging, middleware, graceful shutdown (sonraki dilimlere aday)
