import { clearTokens, getTokens, tryRefresh } from "@/lib/auth"

// Shapes returned by the Go API (SPEC §7–§9).
export type Product = { id: number; name: string; price: number }
export type CartLine = {
  product_id: number
  name: string
  price: number
  quantity: number
  line_total: number
}
export type CartResponse = { items: CartLine[]; total: number }
export type OrderItem = {
  id: number
  order_id: number
  product_id: number
  quantity: number
  unit_price: number
}
export type Order = {
  id: number
  user_id: string
  total: number
  created_at: string
  items: OrderItem[]
}

// ApiError carries the HTTP status and the API's {"error":"..."} message so
// pages can branch on status (404 vs 400) and show the server's wording.
export class ApiError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

function doFetch(path: string, init?: RequestInit): Promise<Response> {
  const headers = new Headers(init?.headers)
  const tokens = getTokens()
  if (tokens) headers.set("Authorization", `Bearer ${tokens.accessToken}`)
  if (init?.body) headers.set("Content-Type", "application/json")
  return fetch(path, { ...init, headers })
}

// api is the single door to the Go API: relative /api/... paths (Vite proxies
// them in dev), Bearer attached, and a one-shot refresh+retry on 401.
export async function api<T = unknown>(path: string, init?: RequestInit): Promise<T> {
  let res = await doFetch(path, init)
  if (res.status === 401 && (await tryRefresh())) {
    res = await doFetch(path, init)
  }
  if (res.status === 401) {
    clearTokens()
    throw new ApiError(401, "oturum süresi doldu")
  }
  if (res.status === 204) return undefined as T
  const body = (await res.json().catch(() => null)) as { error?: string } | null
  if (!res.ok) throw new ApiError(res.status, body?.error ?? `request failed: ${res.status}`)
  return body as T
}
