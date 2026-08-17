import { Link } from "expo-router"
import { useCallback, useEffect, useState } from "react"
import { FlatList, Pressable, Text, View } from "react-native"

import { api, type Product } from "../../lib/api"
import { refreshCartCount } from "../../lib/cart"
import { formatPrice } from "../../lib/format"

// Product grid — mobile counterpart of the web products page (design/02):
// zinc image placeholder, name, price, add-to-cart per card.
export default function ProductsScreen() {
  const [products, setProducts] = useState<Product[] | null>(null)
  const [error, setError] = useState("")
  const [added, setAdded] = useState("")

  const load = useCallback(() => {
    api<Product[]>("/api/products")
      .then((list) => setProducts(list ?? []))
      .catch(() => setError("Ürünler yüklenemedi."))
  }, [])

  useEffect(() => {
    load()
    void refreshCartCount()
  }, [load])

  async function addToCart(product: Product) {
    try {
      await api("/api/cart", {
        method: "POST",
        body: JSON.stringify({ product_id: product.id, quantity: 1 }),
      })
      setAdded(`${product.name} sepete eklendi`)
      void refreshCartCount()
    } catch {
      setAdded("Sepete eklenemedi")
    }
  }

  if (error !== "") {
    return <Text className="p-6 text-zinc-500">{error}</Text>
  }
  if (products === null) {
    return <Text className="p-6 text-zinc-500">Yükleniyor…</Text>
  }

  return (
    <View className="flex-1 px-4">
      <Text className="mt-6 text-3xl font-bold text-zinc-900">Ürünler</Text>
      <Text className="mb-4 mt-1 text-zinc-500">En yeni ve trend ürünlerimize göz atın.</Text>
      {added !== "" && (
        <Text testID="add-feedback" className="mb-2 text-sm text-green-700">
          {added}
        </Text>
      )}
      {products.length === 0 ? (
        <Text className="text-zinc-500">Henüz ürün yok.</Text>
      ) : (
        <FlatList
          testID="product-list"
          data={products}
          keyExtractor={(p) => String(p.id)}
          numColumns={2}
          columnWrapperClassName="gap-3"
          contentContainerClassName="gap-3 pb-8"
          renderItem={({ item }) => (
            <View className="flex-1 rounded-lg border border-zinc-200 bg-white p-3">
              <Link href={{ pathname: "/product/[id]", params: { id: String(item.id) } }} testID={`product-${item.id}`}>
                <View className="w-full">
                  <View className="aspect-square w-full items-center justify-center rounded-md bg-zinc-100">
                    <Text className="text-3xl">📦</Text>
                  </View>
                  <Text className="mt-2 font-medium text-zinc-900" numberOfLines={1}>
                    {item.name}
                  </Text>
                  <Text className="mt-1 font-semibold text-zinc-900">{formatPrice(item.price)}</Text>
                </View>
              </Link>
              <Pressable
                testID={`add-${item.id}`}
                className="mt-2 h-10 items-center justify-center rounded-md bg-zinc-950"
                onPress={() => addToCart(item)}
              >
                <Text className="text-sm font-semibold text-white">Sepete Ekle</Text>
              </Pressable>
            </View>
          )}
        />
      )}
    </View>
  )
}
