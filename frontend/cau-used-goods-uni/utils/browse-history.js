const STORAGE_KEY = 'browse-history'
const MAX_ITEMS = 100

function normalizeItem(product = {}) {
  const id = product.id || product.productId
  if (!id) return null

  return {
    id: `${id}-${Date.now()}`,
    productId: id,
    title: product.title || '\u5546\u54c1',
    price: product.price || product.productPrice || 0,
    status: product.status || product.productStatus || '',
    image: product.coverImage || product.image || product.productImage || product.images?.[0] || '',
    sellerName: product.sellerName || product.seller?.nickname || '',
    viewedAt: Date.now()
  }
}

export function getBrowseHistory() {
  const items = uni.getStorageSync(STORAGE_KEY)
  if (!Array.isArray(items)) return []
  return items.sort((a, b) => Number(b.viewedAt || 0) - Number(a.viewedAt || 0))
}

export function addBrowseHistory(product) {
  const item = normalizeItem(product)
  if (!item) return
  const items = getBrowseHistory().filter((record) => String(record.productId) !== String(item.productId))
  uni.setStorageSync(STORAGE_KEY, [item, ...items].slice(0, MAX_ITEMS))
}

export function clearBrowseHistory() {
  uni.removeStorageSync(STORAGE_KEY)
}
