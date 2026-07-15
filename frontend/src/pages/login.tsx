import { useState } from "react"
import { Navigate, useLocation, useNavigate } from "react-router-dom"

import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { InvalidCredentialsError, useAuth } from "@/lib/auth"

// design/01-login.png: centered card on a dotted zinc backdrop. The form posts
// straight to Keycloak (ROPC) — see SPEC §11.1 for the known deviations.
export default function LoginPage() {
  const { isAuthenticated, login } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [error, setError] = useState("")
  const [busy, setBusy] = useState(false)

  if (isAuthenticated) {
    return <Navigate to="/" replace />
  }

  const from = (location.state as { from?: { pathname: string } } | null)?.from?.pathname ?? "/"

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError("")
    setBusy(true)
    try {
      await login(email, password)
      navigate(from, { replace: true })
    } catch (err) {
      setError(
        err instanceof InvalidCredentialsError
          ? "E-posta veya parola hatalı"
          : "Giriş yapılamadı, lütfen tekrar deneyin",
      )
    } finally {
      setBusy(false)
    }
  }

  return (
    <div
      className="flex min-h-svh items-center justify-center bg-zinc-50 p-6"
      style={{
        backgroundImage: "radial-gradient(circle, var(--color-zinc-300) 1px, transparent 1px)",
        backgroundSize: "24px 24px",
      }}
    >
      <div className="w-full max-w-md rounded-xl border bg-zinc-50 p-2 shadow-sm">
        <div className="py-5 text-center text-2xl font-bold tracking-tight">vibe-shop</div>
        <Card className="rounded-lg">
          <CardContent className="px-8 py-6">
            <div className="mb-6 text-center">
              <h1 className="text-2xl font-bold tracking-tight">Giriş Yap</h1>
              <p className="mt-1 text-muted-foreground">Hesabınla devam et</p>
            </div>
            <form onSubmit={onSubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="email">E-posta</Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="adınız@örnek.com"
                  autoComplete="username"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="password">Parola</Label>
                <Input
                  id="password"
                  type="password"
                  autoComplete="current-password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                />
              </div>
              {error && (
                <p role="alert" className="text-sm text-destructive">
                  {error}
                </p>
              )}
              <Button type="submit" className="w-full" disabled={busy}>
                Giriş Yap
              </Button>
            </form>
          </CardContent>
        </Card>
        <p className="py-4 text-center text-sm text-muted-foreground">
          Hesap yönetimi Keycloak üzerinden yapılır.
        </p>
      </div>
    </div>
  )
}
