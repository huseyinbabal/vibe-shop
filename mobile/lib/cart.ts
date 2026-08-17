import { useSyncExternalStore } from "react"

import { api, type CartResponse } from "./api"

// Shared store for the cart badge (mirror of the web frontend's lib/cart.ts).
let count = 0
const listeners = new Set<() => void>()

function notify() {
  for (const l of listeners) l()
}

function subscribe(listener: () => void): () => void {
  listeners.add(listener)
  return () => listeners.delete(listener)
}

export async function refreshCartCount() {
  try {
    const cart = await api<CartResponse>("/api/cart")
    count = cart.items.length
  } catch {
    count = 0
  }
  notify()
}

export function resetCartCount() {
  count = 0
  notify()
}

export function useCartCount(): number {
  return useSyncExternalStore(subscribe, () => count)
}
