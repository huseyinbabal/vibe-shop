import { Link, Redirect, Stack, usePathname } from "expo-router"
import { Pressable, Text, View } from "react-native"

import { logout, useAuth } from "../../lib/auth"
import { resetCartCount, useCartCount } from "../../lib/cart"

// Shared chrome for the shop: wordmark, cart badge and logout — the mobile
// counterpart of the web navbar. Anonymous visitors bounce to /login.
function ShopHeader() {
  const cartCount = useCartCount()
  const pathname = usePathname()

  return (
    <View className="flex-row items-center justify-between border-b border-zinc-200 bg-white px-4 pb-3 pt-14">
      <Link href="/(shop)">
        <Text className="text-xl font-bold text-zinc-900">vibe-shop</Text>
      </Link>
      <View className="flex-row items-center gap-5">
        {pathname !== "/cart" && (
          <Link href="/cart" testID="nav-cart">
            <View className="flex-row items-center gap-1">
              <Text className="text-base text-zinc-900">Sepet</Text>
              {cartCount > 0 && (
                <View className="h-5 min-w-5 items-center justify-center rounded-full bg-red-500 px-1">
                  <Text testID="cart-count" className="text-xs font-medium text-white">
                    {cartCount}
                  </Text>
                </View>
              )}
            </View>
          </Link>
        )}
        <Pressable
          testID="nav-logout"
          onPress={() => {
            logout()
            resetCartCount()
          }}
        >
          <Text className="text-base text-zinc-500">Çıkış</Text>
        </Pressable>
      </View>
    </View>
  )
}

export default function ShopLayout() {
  const { isAuthenticated } = useAuth()
  if (!isAuthenticated) {
    return <Redirect href="/login" />
  }

  return (
    <View className="flex-1 bg-white">
      <ShopHeader />
      <Stack screenOptions={{ headerShown: false, contentStyle: { backgroundColor: "#fff" } }} />
    </View>
  )
}
