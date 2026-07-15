import { useSyncExternalStore } from "react"

// Identity lives in Keycloak (SPEC §11): the login form posts a Direct Access
// Grant straight to the realm's token endpoint and this module is the only
// keeper of the resulting tokens.
const KEYCLOAK_URL = import.meta.env.VITE_KEYCLOAK_URL ?? "http://localhost:8081"
const CLIENT_ID = "vibe-shop-api"
export const TOKEN_ENDPOINT = `${KEYCLOAK_URL}/realms/vibe-shop/protocol/openid-connect/token`

const STORAGE_KEY = "vibe-shop.tokens"

export type Tokens = { accessToken: string; refreshToken: string }

// InvalidCredentialsError separates "wrong email/password" from network or
// server failures so the form can show a precise message.
export class InvalidCredentialsError extends Error {
  constructor() {
    super("invalid credentials")
  }
}

let tokens: Tokens | null = loadTokens()
const listeners = new Set<() => void>()

function loadTokens(): Tokens | null {
  const raw = localStorage.getItem(STORAGE_KEY)
  if (!raw) return null
  try {
    return JSON.parse(raw) as Tokens
  } catch {
    return null
  }
}

function notify() {
  for (const l of listeners) l()
}

export function getTokens(): Tokens | null {
  return tokens
}

export function setTokens(next: Tokens) {
  tokens = next
  localStorage.setItem(STORAGE_KEY, JSON.stringify(next))
  notify()
}

export function clearTokens() {
  tokens = null
  localStorage.removeItem(STORAGE_KEY)
  notify()
}

async function requestTokens(body: URLSearchParams): Promise<Tokens | null> {
  const res = await fetch(TOKEN_ENDPOINT, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body,
  })
  if (!res.ok) return null
  const data = (await res.json()) as { access_token: string; refresh_token: string }
  return { accessToken: data.access_token, refreshToken: data.refresh_token }
}

// login exchanges the user's credentials for tokens (ROPC). Wrong credentials
// throw InvalidCredentialsError; anything else bubbles as a generic Error.
export async function login(email: string, password: string): Promise<void> {
  const res = await fetch(TOKEN_ENDPOINT, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({
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

export function logout() {
  clearTokens()
}

// tryRefresh renews the session once with the refresh token. On failure the
// session is cleared so guards send the user back to /login.
export async function tryRefresh(): Promise<boolean> {
  if (!tokens) return false
  const next = await requestTokens(
    new URLSearchParams({
      grant_type: "refresh_token",
      client_id: CLIENT_ID,
      refresh_token: tokens.refreshToken,
    }),
  ).catch(() => null)
  if (!next) {
    clearTokens()
    return false
  }
  setTokens(next)
  return true
}

function subscribe(listener: () => void): () => void {
  listeners.add(listener)
  return () => listeners.delete(listener)
}

// useAuth exposes the reactive auth state; components re-render when tokens
// change anywhere (including api.ts clearing them after a failed refresh).
export function useAuth() {
  const current = useSyncExternalStore(subscribe, getTokens)
  return { isAuthenticated: current !== null, login, logout }
}
