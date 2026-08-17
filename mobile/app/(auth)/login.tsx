import { Link } from "expo-router"
import { useState } from "react"
import { KeyboardAvoidingView, Platform, Pressable, ScrollView, Text, View } from "react-native"

import { Field } from "../../components/field"
import { InvalidCredentialsError, login } from "../../lib/auth"

// Mobile counterpart of the web login page (design/01): centered zinc card,
// email + password, error line. Successful login flips the auth store and the
// (auth) layout redirects into the shop.
export default function LoginScreen() {
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [error, setError] = useState("")
  const [busy, setBusy] = useState(false)

  async function onSubmit() {
    setError("")
    setBusy(true)
    try {
      await login(email.trim(), password)
    } catch (err) {
      setError(
        err instanceof InvalidCredentialsError
          ? "E-posta veya parola hatalı"
          : "Giriş yapılamadı, lütfen tekrar deneyin",
      )
    } finally {
      setBusy(false)
    }
  }

  return (
    <KeyboardAvoidingView
      className="flex-1"
      behavior={Platform.OS === "ios" ? "padding" : undefined}
    >
      <ScrollView
        className="bg-zinc-50"
        contentContainerClassName="flex-grow items-center justify-center p-6"
        keyboardShouldPersistTaps="handled"
      >
      <View className="w-full max-w-md rounded-xl border border-zinc-200 bg-white p-8 shadow-sm">
        <Text className="text-center text-2xl font-bold text-zinc-900">vibe-shop</Text>
        <Text className="mt-6 text-center text-2xl font-bold text-zinc-900">Giriş Yap</Text>
        <Text className="mt-1 text-center text-zinc-500">Hesabınla devam et</Text>

        <View className="mt-6 gap-4">
          <Field
            label="E-posta"
            testID="login-email"
            placeholder="adınız@örnek.com"
            keyboardType="email-address"
            value={email}
            onChangeText={setEmail}
          />
          <Field
            label="Parola"
            testID="login-password"
            secureTextEntry
            value={password}
            onChangeText={setPassword}
          />
          {error !== "" && (
            <Text testID="login-error" className="text-sm text-red-600">
              {error}
            </Text>
          )}
          <Pressable
            testID="login-submit"
            className="h-12 items-center justify-center rounded-md bg-zinc-950"
            disabled={busy}
            onPress={onSubmit}
          >
            <Text className="font-semibold text-white">Giriş Yap</Text>
          </Pressable>
          <Link href="/register" testID="go-register" className="text-center text-sm text-zinc-500">
            Hesabın yok mu? <Text className="font-medium text-zinc-900">Kayıt Ol</Text>
          </Link>
        </View>
      </View>
        <Text className="mt-4 text-sm text-zinc-500">
          Hesap yönetimi Keycloak üzerinden yapılır.
        </Text>
      </ScrollView>
    </KeyboardAvoidingView>
  )
}
