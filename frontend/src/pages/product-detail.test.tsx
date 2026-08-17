import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { http, HttpResponse } from "msw"
import { setupServer } from "msw/node"
import { MemoryRouter } from "react-router-dom"
import { afterAll, afterEach, beforeAll, beforeEach, describe, expect, it } from "vitest"

import App from "@/App"
import { clearTokens, setTokens } from "@/lib/auth"

const server = setupServer(
  http.get("/api/products/7", () =>
    HttpResponse.json({ id: 7, name: "Seramik Fincan", price: 249.9 }),
  ),
  http.get("/api/cart", () => HttpResponse.json({ items: [], total: 0 })),
)

beforeAll(() => server.listen({ onUnhandledRequest: "error" }))
afterEach(() => server.resetHandlers())
afterAll(() => server.close())
beforeEach(() => {
  clearTokens()
  setTokens({ accessToken: "at", refreshToken: "rt" })
})

function renderDetail(id: number) {
  return render(
    <MemoryRouter initialEntries={[`/products/${id}`]}>
      <App />
    </MemoryRouter>,
  )
}

describe("product detail page", () => {
  it("renders the product with its formatted price", async () => {
    renderDetail(7)

    expect(await screen.findByRole("heading", { name: "Seramik Fincan" })).toBeInTheDocument()
    expect(screen.getByText("₺249,90")).toBeInTheDocument()
    expect(screen.getByRole("button", { name: "Sepete Ekle" })).toBeInTheDocument()
  })

  it("sends the chosen quantity when adding to the cart", async () => {
    const user = userEvent.setup()
    let cartBody: unknown = null
    server.use(
      http.post("/api/cart", async ({ request }) => {
        cartBody = await request.json()
        return HttpResponse.json(
          { id: 1, user_id: "u", product_id: 7, quantity: 3 },
          { status: 201 },
        )
      }),
    )

    renderDetail(7)
    await screen.findByRole("heading", { name: "Seramik Fincan" })

    const increment = screen.getByRole("button", { name: "Adet artır" })
    await user.click(increment)
    await user.click(increment)
    expect(screen.getByText("3")).toBeInTheDocument()

    await user.click(screen.getByRole("button", { name: "Sepete Ekle" }))
    expect(cartBody).toEqual({ product_id: 7, quantity: 3 })
  })

  it("does not decrement the quantity below 1", async () => {
    const user = userEvent.setup()
    renderDetail(7)
    await screen.findByRole("heading", { name: "Seramik Fincan" })

    await user.click(screen.getByRole("button", { name: "Adet azalt" }))
    expect(screen.getByText("1")).toBeInTheDocument()
  })

  it("shows a not-found state for a missing product", async () => {
    server.use(
      http.get("/api/products/999", () =>
        HttpResponse.json({ error: "product not found" }, { status: 404 }),
      ),
    )
    renderDetail(999)

    expect(await screen.findByText("Ürün bulunamadı.")).toBeInTheDocument()
    expect(screen.getByRole("link", { name: "Ürünlere dön" })).toBeInTheDocument()
  })
})
