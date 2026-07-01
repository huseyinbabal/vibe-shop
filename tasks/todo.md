# vibe-shop — Todo (İlk Dilim)

Detaylar: [plan.md](./plan.md) · Spec: [../SPEC.md](../SPEC.md)

## Faz 1 — Derlenen iskelet
- [ ] **T1 — go.mod oluştur**
  - Yapılacak: `module vibe-shop`, `go 1.26`, dış `require` yok.
  - Kabul: dosya var, modül yolu `vibe-shop`, dış bağımlılık yok.
  - Doğrulama: `go build ./...` hatasız.
- [ ] **CHECKPOINT A** — `go build ./...` başarılı.

## Faz 2 — Health yolu (test-önce)
- [ ] **T2 — internal/health/handler.go**
  - Yapılacak: exported `Handler(http.ResponseWriter, *http.Request)`; `Content-Type: application/json`, status 200, gövde `{"status":"ok"}` (encoding/json ile).
  - Kabul: handler stdlib imzasında, JSON elle string birleştirmeden üretiliyor.
  - Doğrulama: `go build ./internal/health/`.
- [ ] **T3 — internal/health/handler_test.go**
  - Yapılacak: `httptest` ile 3 assert — status 200, content-type application/json, gövde `{"status":"ok"}`.
  - Kabul: test health handler'ı gerçek istekle çağırır.
  - Doğrulama: `go test ./internal/health/` yeşil.
- [ ] **CHECKPOINT B** — `go test ./internal/health/` yeşil.

## Faz 3 — Kablolama ve çalıştırma
- [ ] **T4 — internal/http/router.go**
  - Yapılacak: `http.ServeMux` kuran, `/health` → `health.Handler` bağlayan fonksiyon (ör. `NewRouter() http.Handler`).
  - Kabul: router health paketine bağlanıyor; iş mantığı içermiyor.
  - Doğrulama: `go build ./internal/http/`.
- [ ] **T5 — cmd/server/main.go**
  - Yapılacak: router'ı al, `http.ListenAndServe(":8080", ...)`; başlangıç logu; hata dönerse `log.Fatal`.
  - Kabul: giriş noktası yalnızca kablolama yapıyor.
  - Doğrulama: `go build ./cmd/server`.
- [ ] **CHECKPOINT C** — `go run ./cmd/server` + `curl -s localhost:8080/health` → `{"status":"ok"}`.

## Faz 4 — Kalite kapısı
- [ ] **T6 — Kalite doğrulaması**
  - Doğrulama: `gofmt -l .` boş · `go vet ./...` temiz · `go test ./...` yeşil.
- [ ] **CHECKPOINT D (final)** — İnsan onayı; dilim tamam.
