import { Route, Routes } from "react-router-dom"

import { Layout } from "@/components/layout"
import { RequireAuth } from "@/components/require-auth"
import { Toaster } from "@/components/ui/sonner"
import CartPage from "@/pages/cart"
import LoginPage from "@/pages/login"
import ProductDetailPage from "@/pages/product-detail"
import ProductsPage from "@/pages/products"

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
          <Route path="/products/:id" element={<ProductDetailPage />} />
          <Route path="/cart" element={<CartPage />} />
        </Route>
      </Routes>
      <Toaster />
    </>
  )
}

export default App
