import { Route, Routes } from "react-router-dom"

import { Layout } from "@/components/layout"
import { RequireAuth } from "@/components/require-auth"
import { Toaster } from "@/components/ui/sonner"
import LoginPage from "@/pages/login"
import ProductsPage from "@/pages/products"

// Route skeleton for slice 6 (SPEC §11): remaining pages land in T44–T45.
function Placeholder({ name }: { name: string }) {
  return <div className="text-muted-foreground">{name}</div>
}

function App() {
  return (
    <>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route
          element={
            <RequireAuth>
              <Layout />
            </RequireAuth>
          }
        >
          <Route path="/" element={<ProductsPage />} />
          <Route path="/products/:id" element={<Placeholder name="product-detail" />} />
          <Route path="/cart" element={<Placeholder name="cart" />} />
        </Route>
      </Routes>
      <Toaster />
    </>
  )
}

export default App
