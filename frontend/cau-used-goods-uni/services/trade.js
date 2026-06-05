import { BASE_URL } from '../utils/request'
import * as api from '../api/trade'

function absoluteImage(url) {
  if (!url || /^https?:\/\//.test(url)) return url
  return `${BASE_URL}${url}`
}

function normalizeProduct(item = {}) {
  return {
    ...item,
    image: absoluteImage(item.image || item.productImage || item.images?.[0]),
    meetLocation: item.meetLocation || '预约后协商'
  }
}

function normalizeOrder(item = {}) {
  return {
    ...item,
    id: String(item.id),
    sellerName: item.sellerName || item.sellerNickname || '卖家',
    buyerName: item.buyerName || item.buyerNickname || '买家',
    createdAt: item.createdAt || item.createTime,
    product: normalizeProduct(item.product || {
      id: item.productId,
      title: item.productTitleSnapshot,
      price: item.productPriceSnapshot,
      image: item.productImage,
      meetLocation: item.meetLocation
    })
  }
}

function normalizeMessage(item = {}) {
  return {
    ...item,
    id: String(item.id),
    type: item.type || item.messageType,
    targetType: item.targetType || item.relatedType,
    targetId: item.targetId || item.relatedId,
    createdAt: item.createdAt || item.createTime,
    read: item.read ?? item.readStatus === 'READ'
  }
}

export const tradeService = {
  getProduct: async (id) => normalizeProduct(await api.getProduct(id)),
  createAppointment: async (data) => normalizeOrder(await api.createAppointment(data)),
  getOrders: async (role) => (await api.getOrders(role)).items.map(normalizeOrder),
  getOrder: async (id) => normalizeOrder(await api.getOrder(id)),
  changeOrderStatus: async (id, action, data) => normalizeOrder(await api.changeOrderStatus(id, action, data)),
  getFavorites: async () => (await api.getFavorites()).items.map((item) => normalizeProduct({
    id: item.productId,
    title: item.productTitle,
    price: item.productPrice,
    status: item.productStatus,
    image: item.productImage,
    sellerName: item.sellerNickname
  })),
  addFavorite: api.addFavorite,
  removeFavorite: api.removeFavorite,
  getMessages: async () => (await api.getMessages()).items.map(normalizeMessage),
  getMessage: async (id) => normalizeMessage(await api.getMessage(id)),
  markMessageRead: api.markMessageRead,
  createReview: api.createReview,
  createReport: api.createReport,
  getReports: async () => (await api.getReports()).items.map((item) => ({
    ...item,
    id: String(item.id),
    reason: item.reasonType,
    detail: item.description,
    result: item.handleResult,
    createdAt: item.createTime
  }))
}
