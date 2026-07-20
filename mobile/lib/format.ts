// formatPrice renders API float prices the way the web app does: ₺249,90.
// Hermes ships full Intl, so this matches the web output exactly.
const formatter = new Intl.NumberFormat("tr-TR", {
  style: "currency",
  currency: "TRY",
})

export function formatPrice(value: number): string {
  return formatter.format(value)
}
