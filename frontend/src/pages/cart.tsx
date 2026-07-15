import { CheckCircle2, Package } from "lucide-react"
import { useEffect, useState } from "react"
import { Link } from "react-router-dom"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { api, ApiError, type CartResponse, type Order } from "@/lib/api"
import { refreshCartCount } from "@/lib/cart"
import { formatPrice } from "@/lib/format"

// design/04-cart.png (line items + order summary) and, after checkout,
// design/05-order-confirmation.png. Quantities are read-only in this slice:
// the backend has no decrement/remove endpoints yet (SPEC §11.1).
export default function CartPage() {
  const [cart, setCart] = useState<CartResponse | null>(null)
  const [error, setError] = useState("")
  const [placed, setPlaced] = useState<Order | null>(null)
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    api<CartResponse>("/api/cart")
      .then(setCart)
      .catch(() => setError("Sepet yüklenemedi."))
  }, [])

  async function placeOrder() {
    setBusy(true)
    try {
      const order = await api<Order>("/api/orders", { method: "POST" })
      setPlaced(order)
      refreshCartCount()
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Sipariş oluşturulamadı")
    } finally {
      setBusy(false)
    }
  }

  if (placed) {
    return (
      <div className="mx-auto max-w-xl py-8 text-center">
        <CheckCircle2 className="mx-auto size-14 text-green-600" aria-hidden />
        <h1 className="mt-4 text-3xl font-bold tracking-tight">Siparişin Alındı</h1>
        <p className="mt-1 text-muted-foreground">
          Sipariş No: #{placed.id} ·{" "}
          {new Date(placed.created_at).toLocaleDateString("tr-TR", { dateStyle: "long" })}
        </p>
        <Card className="mt-8 text-left">
          <CardContent className="p-6">
            <h2 className="mb-4 font-semibold">Sipariş Detayı</h2>
            <ul className="space-y-3">
              {placed.items.map((item) => (
                <li key={item.id} className="flex items-center justify-between text-sm">
                  <span>
                    Ürün #{item.product_id} <span className="text-muted-foreground">× {item.quantity}</span>
                  </span>
                  <span className="font-medium">{formatPrice(item.unit_price * item.quantity)}</span>
                </li>
              ))}
            </ul>
            <Separator className="my-4" />
            <div className="flex items-center justify-between font-semibold">
              <span>Toplam</span>
              <span>{formatPrice(placed.total)}</span>
            </div>
          </CardContent>
        </Card>
        <Button asChild className="mt-8">
          <Link to="/">Alışverişe Devam Et</Link>
        </Button>
      </div>
    )
  }

  if (error) return <p className="text-muted-foreground">{error}</p>
  if (cart === null) return <p className="text-muted-foreground">Yükleniyor…</p>

  if (cart.items.length === 0) {
    return (
      <div className="py-16 text-center">
        <h1 className="text-3xl font-bold tracking-tight">Sepetim</h1>
        <p className="mt-4 text-muted-foreground">Sepetin boş.</p>
        <Button asChild className="mt-6">
          <Link to="/">Alışverişe Başla</Link>
        </Button>
      </div>
    )
  }

  return (
    <div>
      <h1 className="mb-8 text-3xl font-bold tracking-tight">Sepetim</h1>
      <div className="grid grid-cols-1 gap-8 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardContent className="p-2 sm:p-4">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Ürün</TableHead>
                  <TableHead className="text-right">Birim Fiyat</TableHead>
                  <TableHead className="text-center">Adet</TableHead>
                  <TableHead className="text-right">Tutar</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {cart.items.map((line) => (
                  <TableRow key={line.product_id}>
                    <TableCell>
                      <div className="flex items-center gap-3">
                        <div className="flex size-12 items-center justify-center rounded-md bg-zinc-100 dark:bg-zinc-800">
                          <Package className="size-6 text-zinc-300" aria-hidden />
                        </div>
                        <Link to={`/products/${line.product_id}`} className="font-medium hover:underline">
                          {line.name}
                        </Link>
                      </div>
                    </TableCell>
                    <TableCell className="text-right">{formatPrice(line.price)}</TableCell>
                    <TableCell className="text-center">{line.quantity}</TableCell>
                    <TableCell className="text-right font-medium">
                      {formatPrice(line.line_total)}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        <Card className="h-fit lg:sticky lg:top-8">
          <CardContent className="p-6">
            <h2 className="mb-4 font-semibold">Sipariş Özeti</h2>
            <div className="space-y-2 text-sm">
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Ara Toplam</span>
                <span>{formatPrice(cart.total)}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Kargo</span>
                <span>Ücretsiz</span>
              </div>
            </div>
            <Separator className="my-4" />
            <div className="flex items-center justify-between font-semibold">
              <span>Toplam</span>
              <span>{formatPrice(cart.total)}</span>
            </div>
            <Button className="mt-6 w-full" onClick={placeOrder} disabled={busy}>
              Siparişi Tamamla
            </Button>
            <p className="mt-3 text-center text-xs text-muted-foreground">
              Fiyatlar sipariş anında sabitlenir.
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
