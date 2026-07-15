import "@testing-library/jest-dom/vitest"

// Node 22+ exposes a global localStorage stub (usable only with
// --localstorage-file) that shadows jsdom's implementation under vitest.
// Replace it with a real in-memory Storage so app code works unchanged.
const store = new Map<string, string>()
const memoryStorage: Storage = {
  getItem: (key) => store.get(key) ?? null,
  setItem: (key, value) => void store.set(key, String(value)),
  removeItem: (key) => void store.delete(key),
  clear: () => store.clear(),
  key: (index) => [...store.keys()][index] ?? null,
  get length() {
    return store.size
  },
}
Object.defineProperty(globalThis, "localStorage", {
  value: memoryStorage,
  configurable: true,
})
