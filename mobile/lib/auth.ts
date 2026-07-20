import * as SecureStore from "expo-secure-store"
import { useSyncExternalStore } from "react"

// Identity lives in Keycloak (SPEC §12): the login form posts a Direct Access
// Grant straight to the realm's token endpoint; registration goes through the
// backend's /api/register (Keycloak Admin API). Tokens live in SecureStore.
const KEYCLOAK_URL = process.env.EXPO_PUBLIC_KEYCLOAK_URL ?? "http://localhost:8081"
export const API_URL = process.env.EXPO_PUBLIC_API_URL ?? "http://localhost:8090"
export const TOKEN_ENDPOINT = `${KEYCLOAK_URL}/realms/vibe-shop/protocol/openid-connect/token`
const CLIENT_ID = "vibe-shop-api"
const STORE_KEY = "vibe-shop.tokens"

export type Tokens = { accessToken: string; refreshToken: string }

export class InvalidCredentialsError extends Error {
  constructor() {
    super("invalid credentials")
  }
}

export class EmailTakenError extends Error {
  constructor() {
    super("email already registered")
  }
}

// Hermes'in URLSearchParams'ı eksik (toString desteklenmiyor); form gövdesini
// elle kodluyoruz.
function formBody(fields: Record<string, string>): string {
  return Object.entries(fields)
    .map(([k, v]) => `${encodeURIComponent(k)}=${encodeURIComponent(v)}`)
    .join("&")
}

let tokens: Tokens | null = null
let ready = false
const listeners = new Set<() => void>()

function notify() {
  for (const l of listeners) l()
}

// initAuth loads persisted tokens once at app start; the root layout waits
// for it before rendering guarded routes.
export async function initAuth(): Promise<void> {
  if (ready) return
  const raw = await SecureStore.getItemAsync(STORE_KEY).catch(() => null)
  if (raw) {
    try {
      tokens = JSON.parse(raw) as Tokens
    } catch {
      tokens = null
    }
  }
  ready = true
  notify()
}

export function getTokens(): Tokens | null {
  return tokens
}

function persist() {
  if (tokens) {
    void SecureStore.setItemAsync(STORE_KEY, JSON.stringify(tokens))
  } else {
    void SecureStore.deleteItemAsync(STORE_KEY)
  }
}

export function setTokens(next: Tokens) {
  tokens = next
  persist()
  notify()
}

export function clearTokens() {
  tokens = null
  persist()
  notify()
}

export async function login(email: string, password: string): Promise<void> {
  const res = await fetch(TOKEN_ENDPOINT, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: formBody({
      grant_type: "password",
      client_id: CLIENT_ID,
      username: email,
      password,
    }),
  })
  if (res.status === 400 || res.status === 401) throw new InvalidCredentialsError()
  if (!res.ok) throw new Error(`login failed: ${res.status}`)
  const data = (await res.json()) as { access_token: string; refresh_token: string }
  setTokens({ accessToken: data.access_token, refreshToken: data.refresh_token })
}

// register creates the user through the backend, then signs in with the same
// credentials so the user lands in the shop without a second form.
export async function register(email: string, password: string): Promise<void> {
  const res = await fetch(`${API_URL}/api/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  })
  if (res.status === 409) throw new EmailTakenError()
  if (!res.ok) {
    const body = (await res.json().catch(() => null)) as { error?: string } | null
    throw new Error(body?.error ?? `register failed: ${res.status}`)
  }
  await login(email, password)
}

export function logout() {
  clearTokens()
}

export async function tryRefresh(): Promise<boolean> {
  if (!tokens) return false
  const res = await fetch(TOKEN_ENDPOINT, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: formBody({
      grant_type: "refresh_token",
      client_id: CLIENT_ID,
      refresh_token: tokens.refreshToken,
    }),
  }).catch(() => null)
  if (!res || !res.ok) {
    clearTokens()
    return false
  }
  const data = (await res.json()) as { access_token: string; refresh_token: string }
  setTokens({ accessToken: data.access_token, refreshToken: data.refresh_token })
  return true
}

function subscribe(listener: () => void): () => void {
  listeners.add(listener)
  return () => listeners.delete(listener)
}

export function useAuth() {
  const snapshot = useSyncExternalStore(subscribe, () => (ready ? (tokens ? "in" : "out") : "loading"))
  return {
    isReady: snapshot !== "loading",
    isAuthenticated: snapshot === "in",
    login,
    register,
    logout,
  }
}
