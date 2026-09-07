# Gözlemlenebilirlik yığını

Prometheus (metrik) · Alertmanager (alarm) · Loki (log) · Alloy (log toplayıcı) ·
Grafana (panel).

## Nasıl çalışır

- Ana uygulama (`docker-compose.prod.yml`) varsayılan ağını `vibe-shop` olarak
  adlandırır.
- Bu yığın aynı `vibe-shop` ağına **external** olarak bağlanır. Böylece:
  - Prometheus `api:8080/metrics`'i 15 sn'de bir toplar (retention 15 gün) ve
    kuralları değerlendirip Alertmanager'a gönderir.
  - Alloy, Docker soketinden container'ları keşfeder, loglarını Loki'ye yollar;
    container adını `container`, compose servis adını `compose_service` label'ı
    olarak ekler.
  - Loki ruler, log tabanlı alarmları değerlendirip Alertmanager'a gönderir.
  - Grafana, Prometheus ve Loki'yi otomatik veri kaynağı olarak alır ve
    dashboard'ları provision eder.
- Tüm veriler named volume'larda kalıcıdır
  (`prometheus-data`, `alertmanager-data`, `loki-data`, `alloy-data`,
  `grafana-data`).

## Dosyalar

| Dosya | İş |
|---|---|
| `prometheus.yml` | scrape hedefleri + alerting/rule_files |
| `prometheus-rules.yml` | metrik tabanlı alarmlar (TargetDown, APIDown, Prometheus sağlığı) |
| `alertmanager.yml` | yönlendirme + Slack alıcıları + inhibit kuralları |
| `loki-config.yml` | Loki (filesystem) + ruler |
| `loki-rules.yml` | log tabanlı alarmlar (APIPanic, APIHighErrorLogRate, …) |
| `config.alloy` | Docker log keşfi → Loki |
| `grafana/provisioning/` | datasource + dashboard sağlayıcıları |
| `grafana/dashboards/` | `vibe-shop / Overview`, `vibe-shop / Logs` |

## Dashboard'lar

Grafana'da **vibe-shop** klasörü altında otomatik gelir:

- **vibe-shop / Overview** — hedeflerin up/down durumu, scrape istatistikleri,
  container başına log hacmi ve hata/panik oranı, canlı log akışı.
- **vibe-shop / Logs** — `container` değişkeni ve regex arama kutusu ile
  filtrelenebilir log oranı, dağılım pastası, hata oranı ve ham loglar.

## Alarmlar

Metrik (Prometheus):

| Alarm | Koşul | Şiddet |
|---|---|---|
| `TargetDown` | `up == 0` 5 dk | critical |
| `APIDown` | `up{job="vibe-shop-api"} == 0` 2 dk | critical |
| `PrometheusConfigReloadFailed` | reload başarısız 5 dk | warning |
| `PrometheusTSDBCompactionFailing` | compaction hataları | warning |
| `PrometheusAlertmanagerNotificationsFailing` | bildirim gönderilemiyor | warning |

Log (Loki ruler):

| Alarm | Koşul | Şiddet |
|---|---|---|
| `APIPanic` | API loglarında `panic:` | critical |
| `APIHighErrorLogRate` | 5 dk'da >20 error/fatal, 10 dk sürüyor | warning |
| `KeycloakErrorSpike` | 5 dk'da >15 ERROR/SEVERE | warning |
| `ContainerLogsStopped` | api/web/keycloak 10 dk log üretmiyor | warning |

### Bildirim kanalı

Webhook URL'i repoda tutulmaz (GitHub push protection). `alertmanager.yml` onu
`/etc/alertmanager/secrets/slack_api_url` dosyasından `api_url_file` ile okur.
Aktifleştirmek için:

1. Slack'te bir incoming webhook oluştur (kanallar: `#vibe-shop-alerts`,
   `#vibe-shop-alerts-critical`).
2. Dosyayı yerleştir:
   - **Yerel:** `mkdir -p observability/secrets && echo 'https://hooks.slack.com/services/...' > observability/secrets/slack_api_url` (gitignore'da) ve
     `docker-compose.observability.yml`'deki yorumlu mount satırını aç.
   - **Dokploy:** servise bir mount ekle → içeriği webhook URL'i olan tek satır
     dosyayı `/etc/alertmanager/secrets/slack_api_url` yoluna bağla.
3. Yığını yeniden deploy et.

Dosya yoksa Alertmanager yine başlar; yalnızca gönderim başarısız olur ve
`PrometheusAlertmanagerNotificationsFailing` tetiklenir. E-posta örneği
`alertmanager.yml` sonunda yorumlu haldedir.

## Dokploy'da kurulum (Hetzner)

1. Önce `docker-compose.prod.yml` yığınını (yeniden) deploy et; `vibe-shop`
   ağı oluşur.
2. Yeni bir **Compose** servisi ekle, compose yolu:
   `docker-compose.observability.yml` (repo kökünde — Dokploy `${VAR}`
   interpolasyonu için `.env`'i `--project-directory`'ye göre arıyor, ki bu
   her zaman repo kökü; dosya alt klasörde olsaydı `GRAFANA_PASSWORD` boş
   string'e düşerdi).
3. Environment ekranına `GRAFANA_PASSWORD` gir.
4. Grafana için bir domain tanımla → hedef port `3000`.
5. (İsteğe bağlı) Alertmanager UI için domain → port `9093`.

## Yerel test

```sh
# vibe-shop ağı ayakta olmalı (prod compose çalışıyor ya da: docker network create vibe-shop)
GRAFANA_PASSWORD=admin docker compose -f docker-compose.observability.yml up -d
```

Config doğrulama:

```sh
docker run --rm --entrypoint promtool -v "$PWD/observability:/o" prom/prometheus:v3.1.0 check rules /o/prometheus-rules.yml
docker run --rm --entrypoint amtool -v "$PWD/observability:/o" prom/alertmanager:v0.28.1 check-config /o/alertmanager.yml
docker run --rm -v "$PWD/observability:/o" grafana/loki:3.3.2 -config.file=/o/loki-config.yml -verify-config
```

## Not

API'de henüz `/metrics` endpoint'i yok (`internal/http/router.go`). Bu yüzden
`vibe-shop-api` hedefi `DOWN` ve **APIDown alarmı sürekli firing** kalır — bu
doğru sinyaldir. API'ye bir Prometheus handler'ı eklenince otomatik düzelir.
O zamana kadar ana sinyal **log tabanlı alarmlar**dır.
