import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { http, HttpResponse } from "msw"
import { setupServer } from "msw/node"
import { MemoryRouter } from "react-router-dom"
import { afterAll, afterEach, beforeAll, beforeEach, describe, expect, it } from "vitest"

import App from "@/App"
import { clearTokens, setTokens } from "@/lib/auth"

const CART = {
  items: [
    { product_id: 1, name: "Seramik Fincan", price: 249.9, quantity: 2, line_total: 499.8 },
    { product_id: 2, name: "Seramik Tabak", price: 329.9, quantity: 1, line_total: 329.9 },
  ],
  total: 829.7,
}

const ORDER = {
  id: 1024,
  user_id: "u",
  total: 829.7,
  created_at: "2026-07-15T12:00:00Z",
  items: [
    { id: 1, order_id: 1024, product_id: 1, quantity: 2, unit_price: 249.9 },
    { id: 2, order_id: 1024, product_id: 2, quantity: 1, unit_price: 329.9 },
  ],
}

const server = setupServer(http.get("/api/cart", () => HttpResponse.json(CART)))

beforeAll(() => server.listen({ onUnhandledRequest: "error" }))
afterEach(() => server.resetHandlers())
afterAll(() => server.close())
beforeEach(() => {
  clearTokens()
  setTokens({ accessToken: "at", refreshToken: "rt" })
})

function renderCart() {
  return render(
    <MemoryRouter initialEntries={["/cart"]}>
      <App />
    </MemoryRouter>,
  )
}

describe("cart page", () => {
  it("lists the cart lines with line totals and the API's grand total", async () => {
    renderCart()

    expect(await screen.findByText("Seramik Fincan")).toBeInTheDocument()
    expect(screen.getByText("Seramik Tabak")).toBeInTheDocument()
    expect(screen.getByText("₺499,80")).toBeInTheDocument()
    // Grand total appears in the order summary card.
    expect(screen.getAllByText("₺829,70").length).toBeGreaterThan(0)
    expect(screen.getByRole("button", { name: "Siparişi Tamamla" })).toBeInTheDocument()
  })

  it("shows an empty state without the checkout button", async () => {
    server.use(http.get("/api/cart", () => HttpResponse.json({ items: [], total: 0 })))
    renderCart()

    expect(await screen.findByText("Sepetin boş.")).toBeInTheDocument()
    expect(screen.queryByRole("button", { name: "Siparişi Tamamla" })).not.toBeInTheDocument()
    expect(screen.getByRole("link", { name: "Alışverişe Başla" })).toBeInTheDocument()
  })

  it("places the order and shows the confirmation with snapshot prices", async () => {
    const user = userEvent.setup()
    let orderPlaced = false
    server.use(
      http.post("/api/orders", () => {
        orderPlaced = true
        return HttpResponse.json(ORDER, { status: 201 })
      }),
      http.get("/api/cart", () =>
        HttpResponse.json(orderPlaced ? { items: [], total: 0 } : CART),
      ),
    )

    renderCart()
    await user.click(await screen.findByRole("button", { name: "Siparişi Tamamla" }))

    expect(await screen.findByText("Siparişin Alındı")).toBeInTheDocument()
    expect(screen.getByText(/#1024/)).toBeInTheDocument()
    expect(screen.getByRole("link", { name: "Alışverişe Devam Et" })).toBeInTheDocument()
  })

  it("surfaces the API error when ordering an empty cart", async () => {
    const user = userEvent.setup()
    server.use(
      http.post("/api/orders", () =>
        HttpResponse.json({ error: "cart is empty" }, { status: 400 }),
      ),
    )

    renderCart()
    await user.click(await screen.findByRole("button", { name: "Siparişi Tamamla" }))

    expect(await screen.findByText("cart is empty")).toBeInTheDocument()
  })
})
