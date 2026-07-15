import { render, screen, within } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { http, HttpResponse } from "msw"
import { setupServer } from "msw/node"
import { MemoryRouter } from "react-router-dom"
import { afterAll, afterEach, beforeAll, beforeEach, describe, expect, it } from "vitest"

import App from "@/App"
import { clearTokens, setTokens } from "@/lib/auth"

const PRODUCTS = [
  { id: 1, name: "Seramik Fincan", price: 249.9 },
  { id: 2, name: "Seramik Tabak", price: 329.9 },
]

const server = setupServer(
  http.get("/api/products", () => HttpResponse.json(PRODUCTS)),
  http.get("/api/cart", () => HttpResponse.json({ items: [], total: 0 })),
)

beforeAll(() => server.listen({ onUnhandledRequest: "error" }))
afterEach(() => server.resetHandlers())
afterAll(() => server.close())
beforeEach(() => {
  clearTokens()
  setTokens({ accessToken: "at", refreshToken: "rt" })
})

function renderHome() {
  return render(
    <MemoryRouter initialEntries={["/"]}>
      <App />
    </MemoryRouter>,
  )
}

describe("products page", () => {
  it("renders API products as cards with formatted prices", async () => {
    renderHome()

    expect(await screen.findByText("Seramik Fincan")).toBeInTheDocument()
    expect(screen.getByText("Seramik Tabak")).toBeInTheDocument()
    expect(screen.getByText("₺249,90")).toBeInTheDocument()
    expect(screen.getAllByRole("button", { name: "Sepete Ekle" })).toHaveLength(2)
  })

  it("shows an empty state when the API returns no products", async () => {
    server.use(http.get("/api/products", () => HttpResponse.json([])))
    renderHome()

    expect(await screen.findByText("Henüz ürün yok.")).toBeInTheDocument()
  })

  it("adds a product to the cart with quantity 1 and updates the navbar badge", async () => {
    const user = userEvent.setup()
    let cartBody: unknown = null
    let cartCount = 0
    server.use(
      http.post("/api/cart", async ({ request }) => {
        cartBody = await request.json()
        cartCount = 1
        return HttpResponse.json(
          { id: 1, user_id: "u", product_id: 1, quantity: 1 },
          { status: 201 },
        )
      }),
      http.get("/api/cart", () =>
        HttpResponse.json({
          items:
            cartCount === 0
              ? []
              : [{ product_id: 1, name: "Seramik Fincan", price: 249.9, quantity: 1, line_total: 249.9 }],
          total: cartCount === 0 ? 0 : 249.9,
        }),
      ),
    )

    renderHome()
    const card = (await screen.findByText("Seramik Fincan")).closest("[data-slot=card]")!
    await user.click(within(card as HTMLElement).getByRole("button", { name: "Sepete Ekle" }))

    expect(cartBody).toEqual({ product_id: 1, quantity: 1 })
    const badge = await screen.findByTestId("cart-count")
    expect(badge).toHaveTextContent("1")
  })
})
