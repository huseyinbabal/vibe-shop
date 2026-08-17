import { ChevronDown, Package } from "lucide-react"
import { useEffect, useState } from "react"
import { Link } from "react-router-dom"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { api, ApiError, type Product } from "@/lib/api"
import { refreshCartCount } from "@/lib/cart"
import { formatPrice } from "@/lib/format"

type Sort = "default" | "asc" | "desc"

const sortLabels: Record<Sort, string> = {
  default: "Sırala",
  asc: "Fiyat (artan)",
  desc: "Fiyat (azalan)",
}

// design/02-product-list.png: heading + sort on the right, then a grid of
// product cards with a zinc image placeholder, name, price and add-to-cart.
export default function ProductsPage() {
  const [products, setProducts] = useState<Product[] | null>(null)
  const [error, setError] = useState("")
  const [sort, setSort] = useState<Sort>("default")

  useEffect(() => {
    api<Product[]>("/api/products")
      .then((list) => setProducts(list ?? []))
      .catch(() => setError("Ürünler yüklenemedi."))
  }, [])

  async function addToCart(product: Product) {
    try {
      await api("/api/cart", {
        method: "POST",
        body: JSON.stringify({ product_id: product.id, quantity: 1 }),
      })
      toast.success("Sepete eklendi", { description: product.name })
      refreshCartCount()
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Sepete eklenemedi")
    }
  }

  if (error) return <p className="text-muted-foreground">{error}</p>
  if (products === null) return <p className="text-muted-foreground">Yükleniyor…</p>

  const sorted =
    sort === "default" ? products : [...products].sort((a, b) => (sort === "asc" ? a.price - b.price : b.price - a.price))

  return (
    <div>
      <div className="mb-8 flex items-end justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Ürünler</h1>
          <p className="mt-1 text-muted-foreground">En yeni ve trend ürünlerimize göz atın.</p>
        </div>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline">
              {sortLabels[sort]} <ChevronDown className="size-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => setSort("default")}>Varsayılan</DropdownMenuItem>
            <DropdownMenuItem onClick={() => setSort("asc")}>Fiyat (artan)</DropdownMenuItem>
            <DropdownMenuItem onClick={() => setSort("desc")}>Fiyat (azalan)</DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      {sorted.length === 0 ? (
        <p className="text-muted-foreground">Henüz ürün yok.</p>
      ) : (
        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
          {sorted.map((product) => (
            <Card key={product.id} className="rounded-lg">
              <CardContent className="flex h-full flex-col gap-3 p-4">
                <Link
                  to={`/products/${product.id}`}
                  className="flex aspect-square items-center justify-center rounded-md bg-zinc-100 dark:bg-zinc-800"
                >
                  <Package className="size-12 text-zinc-300" aria-hidden />
                </Link>
                <div className="flex-1">
                  <Link to={`/products/${product.id}`} className="font-medium hover:underline">
                    {product.name}
                  </Link>
                  <p className="mt-1 font-semibold">{formatPrice(product.price)}</p>
                </div>
                <Button className="w-full" onClick={() => addToCart(product)}>
                  Sepete Ekle
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
