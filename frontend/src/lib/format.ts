const formatter = new Intl.NumberFormat("tr-TR", {
  style: "currency",
  currency: "TRY",
})

// formatPrice renders API float prices the way the mockups do: ₺249,90.
export function formatPrice(value: number): string {
  return formatter.format(value)
}
