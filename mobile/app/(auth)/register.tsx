import { Link } from "expo-router"
import { useState } from "react"
import { Pressable, Text, View } from "react-native"

import { Field } from "../../components/field"
import { EmailTakenError, register } from "../../lib/auth"

// Registration posts to the backend's /api/register (Keycloak Admin API
// underneath) and then signs the user straight in.
export default function RegisterScreen() {
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [error, setError] = useState("")
  const [busy, setBusy] = useState(false)

  async function onSubmit() {
    setError("")
    if (password.length < 8) {
      setError("Parola en az 8 karakter olmalı")
      return
    }
    setBusy(true)
    try {
      await register(email.trim(), password)
    } catch (err) {
      setError(
        err instanceof EmailTakenError
          ? "Bu e-posta zaten kayıtlı"
          : "Kayıt olunamadı, lütfen tekrar deneyin",
      )
    } finally {
      setBusy(false)
    }
  }

  return (
    <View className="flex-1 items-center justify-center bg-zinc-50 p-6">
      <View className="w-full max-w-md rounded-xl border border-zinc-200 bg-white p-8 shadow-sm">
        <Text className="text-center text-2xl font-bold text-zinc-900">vibe-shop</Text>
        <Text className="mt-6 text-center text-2xl font-bold text-zinc-900">Kayıt Ol</Text>
        <Text className="mt-1 text-center text-zinc-500">Yeni bir hesap oluştur</Text>

        <View className="mt-6 gap-4">
          <Field
            label="E-posta"
            testID="register-email"
            placeholder="adınız@örnek.com"
            keyboardType="email-address"
            value={email}
            onChangeText={setEmail}
          />
          <Field
            label="Parola"
            testID="register-password"
            placeholder="en az 8 karakter"
            secureTextEntry
            value={password}
            onChangeText={setPassword}
          />
          {error !== "" && (
            <Text testID="register-error" className="text-sm text-red-600">
              {error}
            </Text>
          )}
          <Pressable
            testID="register-submit"
            className="h-12 items-center justify-center rounded-md bg-zinc-950"
            disabled={busy}
            onPress={onSubmit}
          >
            <Text className="font-semibold text-white">Kayıt Ol</Text>
          </Pressable>
          <Link href="/login" testID="go-login" className="text-center text-sm text-zinc-500">
            Zaten hesabın var mı? <Text className="font-medium text-zinc-900">Giriş Yap</Text>
          </Link>
        </View>
      </View>
    </View>
  )
}
