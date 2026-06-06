import { request, uploadImage } from '../utils/request'

export const getProduct = (id) => request({ url: `/products/${id}`, auth: false })
export const createAppointment = (data) => request({
  url: '/orders',
  method: 'POST',
  data: { ...data, productId: Number(data.productId) }
})
export const getOrders = (role) => request({ url: '/orders', data: { role } })
export const getOrder = (id) => request({ url: `/orders/${id}` })
export const changeOrderStatus = (id, action, data = {}) => request({
  url: `/orders/${id}/${action}`,
  method: 'POST',
  data
})
export const getFavorites = () => request({ url: '/favorites' })
export const addFavorite = (productId) => request({
  url: '/favorites',
  method: 'POST',
  data: { productId: Number(productId) }
})
export const removeFavorite = (productId) => request({
  url: `/favorites/${productId}`,
  method: 'DELETE'
})
export const getMessages = () => request({ url: '/messages' })
export const getMessage = (id) => request({ url: `/messages/${id}` })
export const markMessageRead = (id) => request({ url: `/messages/${id}/read`, method: 'PUT' })
export const createReview = (data) => request({
  url: '/reviews',
  method: 'POST',
  data: { ...data, orderId: Number(data.orderId) }
})
export const createReport = async (data) => {
  const images = []
  for (const filePath of data.images || []) images.push(await uploadImage(filePath))
  return request({
    url: '/reports',
    method: 'POST',
    data: {
      targetType: data.targetType,
      targetId: Number(data.targetId),
      reasonType: data.reasonType || data.reason,
      description: data.detail,
      images
    }
  })
}
export const getReports = () => request({ url: '/reports/my' })
