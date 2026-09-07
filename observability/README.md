# Gözlemlenebilirlik yığını

Prometheus (metrik) · Loki (log) · Alloy (log toplayıcı) · Grafana (panel).

## Nasıl çalışır

- Ana uygulama (`docker-compose.prod.yml`) varsayılan ağını `vibe-shop` olarak
  adlandırır.
- Bu yığın aynı `vibe-shop` ağına **external** olarak bağlanır. Böylece:
  - Prometheus `api:8080/metrics`'i 15 sn'de bir toplar (retention 15 gün).
  - Alloy, Docker soketinden container'ları keşfeder, loglarını Loki'ye yollar
    ve container adını `container` label'ı olarak ekler.
  - Grafana, Prometheus ve Loki'yi otomatik veri kaynağı olarak alır.
- Tüm veriler named volume'larda kalıcıdır
  (`prometheus-data`, `loki-data`, `alloy-data`, `grafana-data`).

## Dokploy'da kurulum (Hetzner)

1. Önce `docker-compose.prod.yml` yığınını (yeniden) deploy et; `vibe-shop`
   ağı oluşur.
2. Yeni bir **Compose** servisi ekle, compose yolu:
   `observability/docker-compose.observability.yml`.
3. Environment ekranına `GRAFANA_PASSWORD` gir.
4. Grafana için bir domain tanımla → hedef port `3000`.

## Yerel test

```sh
# vibe-shop ağı ayakta olmalı (prod compose çalışıyor olmalı)
GRAFANA_PASSWORD=admin docker compose \
  -f observability/docker-compose.observability.yml up -d
```

## Not

API'de `/metrics` endpoint'i henüz yoksa Prometheus hedefi `DOWN` görünür;
uygulamaya bir Prometheus handler'ı eklendiğinde otomatik toplanmaya başlar.
