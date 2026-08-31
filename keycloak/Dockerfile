# İki aşamalı prod imajı: kc.sh build ile Postgres'e göre optimize edilir,
# böylece "start --optimized" runtime'da yeniden build gerektirmez.
# Dev compose'daki "start-dev" ile de uyumludur (dev modu kendi ayarını kurar).
# Realm import imaja gömülüdür: compose dosyası bind mount istemez (Dokploy
# repoyu kendi container'ına klonladığı için host yolları çözülmez).
FROM quay.io/keycloak/keycloak:26.3 AS builder
ENV KC_DB=postgres
ENV KC_HEALTH_ENABLED=true
RUN /opt/keycloak/bin/kc.sh build

FROM quay.io/keycloak/keycloak:26.3
COPY --from=builder /opt/keycloak/ /opt/keycloak/
COPY vibe-shop-realm.json /opt/keycloak/data/import/vibe-shop-realm.json
