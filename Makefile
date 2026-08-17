.PHONY: start stop restart docker-up db-up kc-up db-down server-start server-stop build test vet fmt health logs

# --- Yapılandırma ---
# .env varsa oradaki değerler (ADDR, DATABASE_URL, KEYCLOAK_ISSUER_URL,
# POSTGRES_PORT) hem make değişkeni hem de alt süreçlerin env'i olur.
-include .env
export

BIN      := bin/server
PID_FILE := .server.pid
LOG_FILE := server.log

# API portu ADDR'den türetilir (örn. ADDR=:8090 → PORT=8090); yoksa 8080.
ADDR ?= :8080
PORT := $(patsubst :%,%,$(ADDR))

# --- Ana hedefler ---

## start: Postgres + Keycloak'ı (docker) ve API sunucusunu başlatır
start: db-up kc-up server-start
	@echo "✅ vibe-shop çalışıyor → http://localhost:$(PORT) · Keycloak → http://localhost:8081"

## stop: API sunucusunu, Postgres'i ve Keycloak'ı durdurur
stop: server-stop db-down
	@echo "🛑 vibe-shop durduruldu"

## restart: Durdurup yeniden başlatır
restart: stop start

# --- Veritabanı (docker compose) ---

docker-up:
	@if ! docker info >/dev/null 2>&1; then \
		echo "🐳 Docker engine kapalı, başlatılıyor..."; \
		open -a Docker 2>/dev/null || (echo "❌ Docker Desktop bulunamadı, elle başlatın" && exit 1); \
		printf "   Docker'ın hazır olması bekleniyor"; \
		until docker info >/dev/null 2>&1; do printf "."; sleep 2; done; \
		echo " hazır."; \
	fi

db-up: docker-up
	@echo "🐘 Postgres başlatılıyor..."
	@docker compose up -d
	@docker compose exec -T postgres sh -c 'until pg_isready -U vibeshop -d vibeshop; do sleep 1; done' >/dev/null 2>&1
	@echo "   Postgres hazır."

kc-up: docker-up
	@printf "🔑 Keycloak'ın hazır olması bekleniyor"
	@until curl -sf -o /dev/null http://localhost:8081/realms/vibe-shop; do printf "."; sleep 2; done
	@echo " hazır."

db-down:
	@echo "🐘 Postgres ve Keycloak durduruluyor..."
	@docker compose down

# --- Sunucu (arka planda, PID dosyası ile) ---

server-start: build
	@if [ -f $(PID_FILE) ] && kill -0 $$(cat $(PID_FILE)) 2>/dev/null; then \
		echo "⚠️  Sunucu zaten çalışıyor (PID $$(cat $(PID_FILE)))"; \
	else \
		echo "🚀 Sunucu başlatılıyor (http://localhost:$(PORT))..."; \
		$(BIN) > $(LOG_FILE) 2>&1 & echo $$! > $(PID_FILE); \
		sleep 1; \
		if kill -0 $$(cat $(PID_FILE)) 2>/dev/null; then \
			echo "   PID $$(cat $(PID_FILE)) · loglar: $(LOG_FILE)"; \
		else \
			echo "❌ Sunucu başlayamadı, son loglar:"; tail -3 $(LOG_FILE); rm -f $(PID_FILE); exit 1; \
		fi; \
	fi

server-stop:
	@if [ -f $(PID_FILE) ]; then \
		echo "🛑 Sunucu durduruluyor (PID $$(cat $(PID_FILE)))..."; \
		kill $$(cat $(PID_FILE)) 2>/dev/null || true; \
		rm -f $(PID_FILE); \
	else \
		echo "ℹ️  Çalışan sunucu yok"; \
	fi

# --- Yardımcılar ---

build:
	@go build -o $(BIN) ./cmd/server

test:
	@go test ./...

vet:
	@go vet ./...

fmt:
	@gofmt -l .

health:
	@curl -s localhost:$(PORT)/health && echo

logs:
	@tail -f $(LOG_FILE)
