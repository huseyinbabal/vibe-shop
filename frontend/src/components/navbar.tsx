import { Search, ShoppingCart, User } from "lucide-react"
import { Link, useNavigate } from "react-router-dom"

import { Button } from "@/components/ui/button"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Input } from "@/components/ui/input"
import { useAuth } from "@/lib/auth"
import { resetCartCount, useCartCount } from "@/lib/cart"

// design/02-product-list.png top bar: wordmark, Ürünler, search, cart with a
// count badge, user menu. The search box is visual-only in this slice.
export function Navbar() {
  const cartCount = useCartCount()
  const { logout } = useAuth()
  const navigate = useNavigate()

  function onLogout() {
    logout()
    resetCartCount()
    navigate("/login", { replace: true })
  }

  return (
    <header className="border-b bg-background">
      <div className="mx-auto flex h-16 max-w-6xl items-center gap-8 px-6">
        <Link to="/" className="text-xl font-bold tracking-tight">
          vibe-shop
        </Link>
        <nav>
          <Link to="/" className="text-sm text-muted-foreground hover:text-foreground">
            Ürünler
          </Link>
        </nav>
        <div className="ml-auto flex items-center gap-4">
          <div className="relative hidden sm:block">
            <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input placeholder="Ara..." className="w-56 rounded-full pl-9" />
          </div>
          <Button asChild variant="ghost" size="icon" className="relative">
            <Link to="/cart" aria-label="Sepet">
              <ShoppingCart className="size-5" />
              {cartCount > 0 && (
                <span
                  data-testid="cart-count"
                  className="absolute -right-1 -top-1 flex size-5 items-center justify-center rounded-full bg-red-500 text-xs font-medium text-white"
                >
                  {cartCount}
                </span>
              )}
            </Link>
          </Button>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="rounded-full bg-muted" aria-label="Hesap">
                <User className="size-5" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={onLogout}>Çıkış</DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </header>
  )
}
