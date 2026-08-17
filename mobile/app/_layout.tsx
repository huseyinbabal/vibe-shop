import { Stack } from "expo-router"
import { useEffect } from "react"
import { View } from "react-native"

import { initAuth, useAuth } from "../lib/auth"
import "../global.css"

export default function RootLayout() {
  const { isReady } = useAuth()

  useEffect(() => {
    void initAuth()
  }, [])

  // Hold rendering until persisted tokens are loaded so guards don't flash
  // the login screen for an already signed-in user.
  if (!isReady) {
    return <View className="flex-1 bg-white" />
  }

  return (
    <Stack screenOptions={{ headerShown: false }}>
      <Stack.Screen name="(auth)" />
      <Stack.Screen name="(shop)" />
    </Stack>
  )
}
