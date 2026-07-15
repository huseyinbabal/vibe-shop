import { useSyncExternalStore } from "react"

import { api, type CartResponse } from "@/lib/api"

// Tiny shared store for the navbar's cart badge: pages call refreshCartCount
// after mutating the cart, the navbar re-renders through useCartCount.
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
