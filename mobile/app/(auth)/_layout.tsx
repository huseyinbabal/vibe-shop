import { Redirect, Stack } from "expo-router"

import { useAuth } from "../../lib/auth"

export default function AuthLayout() {
  const { isAuthenticated } = useAuth()
  if (isAuthenticated) {
    return <Redirect href="/(shop)" />
  }
  return <Stack screenOptions={{ headerShown: false }} />
}
