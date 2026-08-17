import { Link } from "expo-router"
import { useEffect, useState } from "react"
import { Pressable, ScrollView, Text, View } from "react-native"

import { api, ApiError, type CartResponse, type Order } from "../../lib/api"
import { refreshCartCount } from "../../lib/cart"
import { formatPrice } from "../../lib/format"

// Cart + order confirmation — mobile counterpart of design/04 and design/05.
// Quantities are read-only (backend has no decrement/remove yet, SPEC §11.1).
export default function CartScreen() {
  const [cart, setCart] = useState<CartResponse | null>(null)
  const [error, setError] = useState("")
  const [placed, setPlaced] = useState<Order | null>(null)
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    api<CartResponse>("/api/cart")
      .then(setCart)
      .catch(() => setError("Sepet yüklenemedi."))
  }, [])

  async function placeOrder() {
    setBusy(true)
    setError("")
    try {
      const order = await api<Order>("/api/orders", { method: "POST" })
      setPlaced(order)
      void refreshCartCount()
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Sipariş oluşturulamadı")
    } finally {
      setBusy(false)
    }
  }

  if (placed) {
    return (
      <ScrollView className="flex-1 px-4" contentContainerClassName="items-center pb-10 pt-10">
        <Text className="text-5xl">✅</Text>
        <Text testID="order-confirmed" className="mt-4 text-3xl font-bold text-zinc-900">
          Siparişin Alındı
        </Text>
        <Text className="mt-1 text-zinc-500">Sipariş No: #{placed.id}</Text>

        <View className="mt-8 w-full rounded-lg border border-zinc-200 bg-white p-5">
          <Text className="mb-3 font-semibold text-zinc-900">Sipariş Detayı</Text>
          {placed.items.map((item) => (
            <View key={item.id} className="flex-row justify-between py-1">
              <Text className="text-sm text-zinc-900">
                Ürün #{item.product_id} <Text className="text-zinc-500">× {item.quantity}</Text>
              </Text>
              <Text className="text-sm font-medium text-zinc-900">
                {formatPrice(item.unit_price * item.quantity)}
              </Text>
            </View>
          ))}
          <View className="mt-3 flex-row justify-between border-t border-zinc-200 pt-3">
            <Text className="font-semibold text-zinc-900">Toplam</Text>
            <Text testID="order-total" className="font-semibold text-zinc-900">
              {formatPrice(placed.total)}
            </Text>
          </View>
        </View>

        <Link href="/(shop)" testID="continue-shopping" className="mt-8">
          <View className="h-12 items-center justify-center rounded-md bg-zinc-950 px-6">
            <Text className="font-semibold text-white">Alışverişe Devam Et</Text>
          </View>
        </Link>
      </ScrollView>
    )
  }

  if (cart === null && error === "") {
    return <Text className="p-6 text-zinc-500">Yükleniyor…</Text>
  }

  if (cart && cart.items.length === 0) {
    return (
      <View className="flex-1 items-center justify-center gap-4 p-6">
        <Text className="text-3xl font-bold text-zinc-900">Sepetim</Text>
        <Text testID="cart-empty" className="text-zinc-500">
          Sepetin boş.
        </Text>
        <Link href="/(shop)" testID="start-shopping">
          <View className="h-12 items-center justify-center rounded-md bg-zinc-950 px-6">
            <Text className="font-semibold text-white">Alışverişe Başla</Text>
          </View>
        </Link>
      </View>
    )
  }

  return (
    <ScrollView className="flex-1 px-4" contentContainerClassName="pb-10">
      <Text className="mt-6 text-3xl font-bold text-zinc-900">Sepetim</Text>
      {error !== "" && (
        <Text testID="cart-error" className="mt-2 text-sm text-red-600">
          {error}
        </Text>
      )}

      <View className="mt-4 rounded-lg border border-zinc-200 bg-white">
        {cart?.items.map((line, i) => (
          <View
            key={line.product_id}
            className={`flex-row items-center gap-3 p-4 ${i > 0 ? "border-t border-zinc-100" : ""}`}
          >
            <View className="h-12 w-12 items-center justify-center rounded-md bg-zinc-100">
              <Text>📦</Text>
            </View>
            <View className="flex-1">
              <Text className="font-medium text-zinc-900">{line.name}</Text>
              <Text className="text-sm text-zinc-500">
                {formatPrice(line.price)} × {line.quantity}
              </Text>
            </View>
            <Text className="font-medium text-zinc-900">{formatPrice(line.line_total)}</Text>
          </View>
        ))}
      </View>

      <View className="mt-4 rounded-lg border border-zinc-200 bg-white p-5">
        <Text className="mb-3 font-semibold text-zinc-900">Sipariş Özeti</Text>
        <View className="flex-row justify-between py-1">
          <Text className="text-zinc-500">Ara Toplam</Text>
          <Text className="text-zinc-900">{formatPrice(cart?.total ?? 0)}</Text>
        </View>
        <View className="flex-row justify-between py-1">
          <Text className="text-zinc-500">Kargo</Text>
          <Text className="text-zinc-900">Ücretsiz</Text>
        </View>
        <View className="mt-2 flex-row justify-between border-t border-zinc-200 pt-3">
          <Text className="font-semibold text-zinc-900">Toplam</Text>
          <Text testID="cart-total" className="font-semibold text-zinc-900">
            {formatPrice(cart?.total ?? 0)}
          </Text>
        </View>
        <Pressable
          testID="place-order"
          className="mt-4 h-12 items-center justify-center rounded-md bg-zinc-950"
          disabled={busy}
          onPress={placeOrder}
        >
          <Text className="font-semibold text-white">Siparişi Tamamla</Text>
        </Pressable>
        <Text className="mt-2 text-center text-xs text-zinc-500">
          Fiyatlar sipariş anında sabitlenir.
        </Text>
      </View>
    </ScrollView>
  )
}
