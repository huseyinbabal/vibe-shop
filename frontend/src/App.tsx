import { Route, Routes } from "react-router-dom"

import { RequireAuth } from "@/components/require-auth"
import LoginPage from "@/pages/login"

// Route skeleton for slice 6 (SPEC §11): remaining pages land in T43–T45.
function Placeholder({ name }: { name: string }) {
  return <div className="p-8 text-muted-foreground">{name}</div>
}

function guarded(element: React.ReactNode) {
  return <RequireAuth>{element}</RequireAuth>
}

function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/" element={guarded(<Placeholder name="products" />)} />
      <Route path="/products/:id" element={guarded(<Placeholder name="product-detail" />)} />
      <Route path="/cart" element={guarded(<Placeholder name="cart" />)} />
    </Routes>
  )
}

export default App
