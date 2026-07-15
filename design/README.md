# vibe-shop — UI Tasarım Mockup'ları

shadcn/ui görünümünde konsept çizimleri (Nano Banana Pro ile üretildi, 2026-07-15).
Kod değildir; frontend dilimine girdi olacak görsel yön önerisidir.

| Dosya | Sayfa | Not |
|---|---|---|
| `01-login.png` | Giriş | E-posta/parola formunun gerçek evi Keycloak login theme'idir (keycloakify ile shadcn görünümü verilebilir); uygulamadaki ana yol "Keycloak ile devam et" yönlendirmesidir (Authorization Code + PKCE). |
| `02-product-list.png` | Ürün listesi | `GET /api/products`; kartlar zinc-100 zeminli görsel + ad + fiyat + "Sepete Ekle". |
| `03-product-detail.png` | Ürün detay | `GET /api/products/{id}`; adet seçici + sepete ekle. |
| `04-cart.png` | Sepet | `GET/POST /api/cart`. Dikkat: adet azaltma ve satır silme backend'de henüz yok (SPEC "önce sor") — bu tasarım uygulanacaksa ayrı dilim gerekir. |
| `05-order-confirmation.png` | Sipariş onayı | `POST /api/orders`; kalem fiyatları sipariş anındaki `unit_price` snapshot'ıdır. |
| `06-admin-products.png` | Admin ürün yönetimi | Korumalı yazma uçları (`POST/PUT/DELETE /api/products`, Keycloak token). Dialog doğrulaması API kurallarıyla aynı olmalı: ad ≤ 200 karakter, fiyat > 0, en fazla iki ondalık. |

## Tasarım dili

- **Palet:** zinc nötrler, tek koyu birincil buton (zinc-950); ürün görselleri `zinc-100` zeminde.
- **Bileşen karşılıkları (shadcn/ui):** Card, Button, Badge, Table, Dialog, Accordion, Separator, Sonner (toast).
- **Tipografi:** Inter; başlıklar semibold, ikincil metin zinc-500.

Bilinen görsel kusurlar (AI üretimi): ürün listesinde birkaç bozuk ürün adı,
admin sayfasında logo yanında gereksiz "zinc-50" rozeti. Konsept amaçlı önemsizdir.
