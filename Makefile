.PHONY: start stop restart docker-up db-up kc-up db-down server-start server-stop build test vet fmt health logs

# --- Yapılandırma ---
BIN      := bin/server
PID_FILE := .server.pid
LOG_FILE := server.log
PORT     := 8080

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
		echo "🚀 Sunucu başlatılıyor..."; \
		{ [ -f .env ] && set -a && . ./.env && set +a; \
		  $(BIN) > $(LOG_FILE) 2>&1 & echo $$! > $(PID_FILE); }; \
		sleep 1; \
		echo "   PID $$(cat $(PID_FILE)) · loglar: $(LOG_FILE)"; \
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
