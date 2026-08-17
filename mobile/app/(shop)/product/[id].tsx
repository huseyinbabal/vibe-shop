import { Link, useLocalSearchParams } from "expo-router"
import { useEffect, useState } from "react"
import { Pressable, ScrollView, Text, View } from "react-native"

import { api, ApiError, type Product } from "../../../lib/api"
import { refreshCartCount } from "../../../lib/cart"
import { formatPrice } from "../../../lib/format"

// Product detail — mobile counterpart of design/03: image, price, quantity
// stepper, add to cart, static info sections.
export default function ProductDetailScreen() {
  const { id } = useLocalSearchParams<{ id: string }>()
  const [product, setProduct] = useState<Product | null>(null)
  const [state, setState] = useState<"loading" | "ok" | "notfound" | "error">("loading")
  const [quantity, setQuantity] = useState(1)
  const [feedback, setFeedback] = useState("")

  useEffect(() => {
    setState("loading")
    api<Product>(`/api/products/${id}`)
      .then((p) => {
        setProduct(p)
        setState("ok")
      })
      .catch((err: unknown) => {
        setState(err instanceof ApiError && err.status === 404 ? "notfound" : "error")
      })
  }, [id])

  async function addToCart() {
    if (!product) return
    try {
      await api("/api/cart", {
        method: "POST",
        body: JSON.stringify({ product_id: product.id, quantity }),
      })
      setFeedback(`Sepete eklendi: ${product.name} × ${quantity}`)
      void refreshCartCount()
    } catch {
      setFeedback("Sepete eklenemedi")
    }
  }

  if (state === "loading") {
    return <Text className="p-6 text-zinc-500">Yükleniyor…</Text>
  }
  if (state !== "ok" || !product) {
    return (
      <View className="flex-1 items-center justify-center gap-4 p-6">
        <Text className="text-zinc-500">
          {state === "notfound" ? "Ürün bulunamadı." : "Ürün yüklenemedi."}
        </Text>
        <Link href="/(shop)" className="font-medium text-zinc-900">
          Ürünlere dön
        </Link>
      </View>
    )
  }

  return (
    <ScrollView className="flex-1 px-4" contentContainerClassName="pb-10">
      <Link href="/(shop)" className="mt-4 text-sm text-zinc-500">
        Ürünler / <Text className="text-zinc-900">{product.name}</Text>
      </Link>

      <View className="mt-4 aspect-square w-full items-center justify-center rounded-lg bg-zinc-100">
        <Text className="text-6xl">📦</Text>
      </View>

      <Text testID="detail-name" className="mt-5 text-3xl font-bold text-zinc-900">
        {product.name}
      </Text>
      <Text className="mt-1 text-2xl font-semibold text-zinc-900">{formatPrice(product.price)}</Text>
      <Text className="mt-3 text-zinc-500">
        El yapımı, minimal tasarımlı seramik ürün. Günlük kullanım için idealdir; her parça küçük
        tonal farklılıklar taşır.
      </Text>

      <View className="mt-5 flex-row items-center gap-4">
        <View className="flex-row items-center rounded-md border border-zinc-200">
          <Pressable
            testID="qty-dec"
            className="h-11 w-11 items-center justify-center"
            onPress={() => setQuantity((q) => Math.max(1, q - 1))}
          >
            <Text className="text-xl text-zinc-900">−</Text>
          </Pressable>
          <Text testID="qty-value" className="w-8 text-center text-base font-medium text-zinc-900">
            {quantity}
          </Text>
          <Pressable
            testID="qty-inc"
            className="h-11 w-11 items-center justify-center"
            onPress={() => setQuantity((q) => q + 1)}
          >
            <Text className="text-xl text-zinc-900">+</Text>
          </Pressable>
        </View>
        <Pressable
          testID="detail-add"
          className="h-11 flex-1 items-center justify-center rounded-md bg-zinc-950"
          onPress={addToCart}
        >
          <Text className="font-semibold text-white">Sepete Ekle</Text>
        </Pressable>
      </View>
      {feedback !== "" && (
        <Text testID="detail-feedback" className="mt-3 text-sm text-green-700">
          {feedback}
        </Text>
      )}

      <View className="mt-8 gap-4 border-t border-zinc-200 pt-4">
        <View>
          <Text className="font-semibold text-zinc-900">Kargo ve İade</Text>
          <Text className="mt-1 text-sm text-zinc-500">
            Siparişler 2 iş günü içinde kargoya verilir. 14 gün içinde ücretsiz iade edebilirsiniz.
          </Text>
        </View>
        <View>
          <Text className="font-semibold text-zinc-900">Malzeme</Text>
          <Text className="mt-1 text-sm text-zinc-500">
            Yüksek ısıda fırınlanmış stoneware seramik; kurşunsuz sır. Bulaşık makinesinde
            yıkanabilir.
          </Text>
        </View>
      </View>
    </ScrollView>
  )
}
