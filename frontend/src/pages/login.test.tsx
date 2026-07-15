import { render, screen } from "@testing-library/react"
import userEvent from "@testing-library/user-event"
import { http, HttpResponse } from "msw"
import { setupServer } from "msw/node"
import { MemoryRouter } from "react-router-dom"
import { afterAll, afterEach, beforeAll, beforeEach, describe, expect, it } from "vitest"

import App from "@/App"
import { clearTokens, getTokens, setTokens, TOKEN_ENDPOINT } from "@/lib/auth"

const server = setupServer(
  // Pages behind the guard may fetch data after login; tests here only care
  // about the auth flow, so serve harmless defaults.
  http.get("/api/products", () => HttpResponse.json([])),
  http.get("/api/cart", () => HttpResponse.json({ items: [], total: 0 })),
)

beforeAll(() => server.listen({ onUnhandledRequest: "error" }))
afterEach(() => server.resetHandlers())
afterAll(() => server.close())
beforeEach(() => clearTokens())

function renderAt(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <App />
    </MemoryRouter>,
  )
}

describe("route guard", () => {
  it("redirects every protected route to /login when there is no token", () => {
    renderAt("/cart")
    expect(screen.getByRole("heading", { name: "Giriş Yap" })).toBeInTheDocument()
  })

  it("redirects /login to the home page when already authenticated", () => {
    setTokens({ accessToken: "at", refreshToken: "rt" })
    renderAt("/login")
    expect(screen.queryByRole("heading", { name: "Giriş Yap" })).not.toBeInTheDocument()
  })
})

describe("login page", () => {
  it("stores tokens and returns to the target route after a successful login", async () => {
    const user = userEvent.setup()
    server.use(
      http.post(TOKEN_ENDPOINT, async ({ request }) => {
        const body = new URLSearchParams(await request.text())
        expect(body.get("grant_type")).toBe("password")
        expect(body.get("client_id")).toBe("vibe-shop-api")
        expect(body.get("username")).toBe("testuser@vibe.shop")
        expect(body.get("password")).toBe("test1234")
        return HttpResponse.json({ access_token: "at-login", refresh_token: "rt-login" })
      }),
    )

    renderAt("/cart")
    await user.type(screen.getByLabelText("E-posta"), "testuser@vibe.shop")
    await user.type(screen.getByLabelText("Parola"), "test1234")
    await user.click(screen.getByRole("button", { name: "Giriş Yap" }))

    // Returned to the route the user originally asked for (/cart).
    expect(await screen.findByText("Sepetin boş.")).toBeInTheDocument()
    expect(getTokens()?.accessToken).toBe("at-login")
  })

  it("shows a form error on invalid credentials and keeps the user on /login", async () => {
    const user = userEvent.setup()
    server.use(
      http.post(TOKEN_ENDPOINT, () =>
        HttpResponse.json({ error: "invalid_grant" }, { status: 401 }),
      ),
    )

    renderAt("/")
    await user.type(screen.getByLabelText("E-posta"), "testuser@vibe.shop")
    await user.type(screen.getByLabelText("Parola"), "yanlis")
    await user.click(screen.getByRole("button", { name: "Giriş Yap" }))

    expect(await screen.findByText("E-posta veya parola hatalı")).toBeInTheDocument()
    expect(getTokens()).toBeNull()
  })
})
