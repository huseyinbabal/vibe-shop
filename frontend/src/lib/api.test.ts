import { http, HttpResponse } from "msw"
import { setupServer } from "msw/node"
import { afterAll, afterEach, beforeAll, beforeEach, describe, expect, it } from "vitest"

import { api, ApiError } from "@/lib/api"
import { clearTokens, setTokens, getTokens, TOKEN_ENDPOINT } from "@/lib/auth"

const server = setupServer()

beforeAll(() => server.listen({ onUnhandledRequest: "error" }))
afterEach(() => server.resetHandlers())
afterAll(() => server.close())
beforeEach(() => clearTokens())

describe("api", () => {
  it("attaches the access token as a Bearer header", async () => {
    setTokens({ accessToken: "at-1", refreshToken: "rt-1" })
    let seenAuth = ""
    server.use(
      http.get("/api/products", ({ request }) => {
        seenAuth = request.headers.get("Authorization") ?? ""
        return HttpResponse.json([])
      }),
    )

    await api("/api/products")

    expect(seenAuth).toBe("Bearer at-1")
  })

  it("refreshes once on 401 and retries with the new token", async () => {
    setTokens({ accessToken: "expired", refreshToken: "rt-1" })
    const authHeaders: string[] = []
    server.use(
      http.get("/api/cart", ({ request }) => {
        const auth = request.headers.get("Authorization") ?? ""
        authHeaders.push(auth)
        if (auth === "Bearer fresh") {
          return HttpResponse.json({ items: [], total: 0 })
        }
        return HttpResponse.json({ error: "invalid or expired token" }, { status: 401 })
      }),
      http.post(TOKEN_ENDPOINT, () =>
        HttpResponse.json({ access_token: "fresh", refresh_token: "rt-2" }),
      ),
    )

    const body = await api<{ total: number }>("/api/cart")

    expect(body.total).toBe(0)
    expect(authHeaders).toEqual(["Bearer expired", "Bearer fresh"])
    expect(getTokens()?.accessToken).toBe("fresh")
  })

  it("clears the session and throws 401 when refresh fails", async () => {
    setTokens({ accessToken: "expired", refreshToken: "dead" })
    server.use(
      http.get("/api/cart", () =>
        HttpResponse.json({ error: "invalid or expired token" }, { status: 401 }),
      ),
      http.post(TOKEN_ENDPOINT, () =>
        HttpResponse.json({ error: "invalid_grant" }, { status: 400 }),
      ),
    )

    await expect(api("/api/cart")).rejects.toMatchObject({ status: 401 })
    expect(getTokens()).toBeNull()
  })

  it("surfaces the API's error body as an ApiError", async () => {
    setTokens({ accessToken: "at", refreshToken: "rt" })
    server.use(
      http.post("/api/orders", () =>
        HttpResponse.json({ error: "cart is empty" }, { status: 400 }),
      ),
    )

    const err = await api("/api/orders", { method: "POST" }).catch((e: unknown) => e)
    expect(err).toBeInstanceOf(ApiError)
    expect((err as ApiError).status).toBe(400)
    expect((err as ApiError).message).toBe("cart is empty")
  })
})
