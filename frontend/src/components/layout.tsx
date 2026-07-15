import { useEffect } from "react"
import { Outlet } from "react-router-dom"

import { Navbar } from "@/components/navbar"
import { refreshCartCount } from "@/lib/cart"

// Shared chrome for every authenticated page: navbar on top, page content in
// a centered column, cart badge primed once on mount.
export function Layout() {
  useEffect(() => {
    refreshCartCount()
  }, [])

  return (
    <div className="min-h-svh bg-background">
      <Navbar />
      <main className="mx-auto max-w-6xl px-6 py-8">
        <Outlet />
      </main>
    </div>
  )
}
