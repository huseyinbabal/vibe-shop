import { Route, Routes } from "react-router-dom"

// Route skeleton for slice 6 (SPEC §11): pages land in T42–T45.
function Placeholder({ name }: { name: string }) {
  return <div className="p-8 text-muted-foreground">{name}</div>
}

function App() {
  return (
    <Routes>
      <Route path="/login" element={<Placeholder name="login" />} />
      <Route path="/" element={<Placeholder name="products" />} />
      <Route path="/products/:id" element={<Placeholder name="product-detail" />} />
      <Route path="/cart" element={<Placeholder name="cart" />} />
    </Routes>
  )
}

export default App
