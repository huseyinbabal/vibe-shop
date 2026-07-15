import { Minus, Package, Plus, ShoppingCart } from "lucide-react"
import { useEffect, useState } from "react"
import { Link, useParams } from "react-router-dom"
import { toast } from "sonner"

import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion"
import { Button } from "@/components/ui/button"
import { api, ApiError, type Product } from "@/lib/api"
import { refreshCartCount } from "@/lib/cart"
import { formatPrice } from "@/lib/format"

// design/03-product-detail.png: breadcrumb, large image left, details +
// quantity stepper + add-to-cart right, static info accordion below.
export default function ProductDetailPage() {
  const { id } = useParams()
  const [product, setProduct] = useState<Product | null>(null)
  const [state, setState] = useState<"loading" | "ok" | "notfound" | "error">("loading")
  const [quantity, setQuantity] = useState(1)

  useEffect(() => {
    setState("loading")
    api<Product>(`/api/products/${id}`)
      .then((p) => {
        setProduct(p)
        setState("ok")
      })
      .catch((err: unknown) => {
        setState(err instanceof ApiError && err.status === 404 ? "notfound" : "error")
      })
  }, [id])

  async function addToCart() {
    if (!product) return
    try {
      await api("/api/cart", {
        method: "POST",
        body: JSON.stringify({ product_id: product.id, quantity }),
      })
      toast.success("Sepete eklendi", { description: `${product.name} × ${quantity}` })
      refreshCartCount()
    } catch (err) {
      toast.error(err instanceof ApiError ? err.message : "Sepete eklenemedi")
    }
  }

  if (state === "loading") return <p className="text-muted-foreground">Yükleniyor…</p>
  if (state === "notfound" || state === "error" || !product) {
    return (
      <div className="py-16 text-center">
        <p className="text-muted-foreground">
          {state === "notfound" ? "Ürün bulunamadı." : "Ürün yüklenemedi."}
        </p>
        <Button asChild variant="outline" className="mt-4">
          <Link to="/">Ürünlere dön</Link>
        </Button>
      </div>
    )
  }

  return (
    <div>
      <p className="mb-6 text-sm text-muted-foreground">
        <Link to="/" className="hover:text-foreground">
          Ürünler
        </Link>{" "}
        / <span className="text-foreground">{product.name}</span>
      </p>

      <div className="grid grid-cols-1 gap-10 lg:grid-cols-2">
        <div className="flex aspect-square items-center justify-center rounded-lg bg-zinc-100 dark:bg-zinc-800">
          <Package className="size-24 text-zinc-300" aria-hidden />
        </div>

        <div>
          <h1 className="text-3xl font-bold tracking-tight">{product.name}</h1>
          <p className="mt-2 text-2xl font-semibold">{formatPrice(product.price)}</p>
          <p className="mt-4 text-muted-foreground">
            El yapımı, minimal tasarımlı seramik ürün. Günlük kullanım için idealdir; her parça
            küçük tonal farklılıklar taşır.
          </p>

          <div className="mt-6 flex items-center gap-4">
            <div className="flex items-center rounded-md border">
              <Button
                variant="ghost"
                size="icon"
                aria-label="Adet azalt"
                onClick={() => setQuantity((q) => Math.max(1, q - 1))}
              >
                <Minus className="size-4" />
              </Button>
              <span className="w-10 text-center font-medium">{quantity}</span>
              <Button
                variant="ghost"
                size="icon"
                aria-label="Adet artır"
                onClick={() => setQuantity((q) => q + 1)}
              >
                <Plus className="size-4" />
              </Button>
            </div>
            <Button className="flex-1" onClick={addToCart}>
              <ShoppingCart className="size-4" /> Sepete Ekle
            </Button>
          </div>

          <Accordion type="single" collapsible className="mt-8">
            <AccordionItem value="shipping">
              <AccordionTrigger>Kargo ve İade</AccordionTrigger>
              <AccordionContent>
                Siparişler 2 iş günü içinde kargoya verilir. 14 gün içinde ücretsiz iade
                edebilirsiniz.
              </AccordionContent>
            </AccordionItem>
            <AccordionItem value="material">
              <AccordionTrigger>Malzeme</AccordionTrigger>
              <AccordionContent>
                Yüksek ısıda fırınlanmış stoneware seramik; kurşunsuz sır. Bulaşık makinesinde
                yıkanabilir.
              </AccordionContent>
            </AccordionItem>
          </Accordion>
        </div>
      </div>
    </div>
  )
}
