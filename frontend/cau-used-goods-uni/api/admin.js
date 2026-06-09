import { request } from '../utils/request'

export const getProductOverview = () => {
  return request({ url: '/stats/products/overview' })
}

export const getUserOverview = () => {
  return request({ url: '/stats/users/overview' })
}

export const getOrderOverview = () => {
  return request({ url: '/stats/orders/overview' })
}

export const getReportOverview = () => {
  return request({ url: '/stats/reports/overview' })
}

export const getAppealOverview = () => {
  return request({ url: '/stats/appeals/overview' })
}

export const getCategoryDistribution = () => {
  return request({ url: '/stats/products/category-distribution' })
}

export const getAdminCategories = () => {
  return request({ url: '/admin/categories' })
}

export const createAdminCategory = (payload) => {
  return request({
    url: '/admin/categories',
    method: 'POST',
    data: payload
  })
}

export const updateAdminCategoryStatus = (categoryId, status) => {
  return request({
    url: `/admin/categories/${categoryId}/status`,
    method: 'PUT',
    data: { status }
  })
}

export const getProductStatusDistribution = () => {
  return request({ url: '/stats/products/status-distribution' })
}

export const getProductTrend = (days = 7) => {
  return request({ url: `/stats/products/trend?days=${days}` })
}

export const getAdminUsers = () => {
  return request({ url: '/admin/users' })
}

export const getAdminProducts = () => {
  return request({
    url: '/products?status=ALL&page=1&pageSize=50&sort=newest',
    auth: false
  })
}

export const updateAdminProductStatus = (productId, status) => {
  return request({
    url: `/products/${productId}/status`,
    method: 'PUT',
    data: {
      status,
      reason: status === 'ON_SALE' ? '管理员上架商品' : '管理员下架商品'
    }
  })
}

export const getPendingStudentVerifications = () => {
  return request({
    url: '/admin/users/student-verifications?authStatus=PENDING'
  })
}

export const reviewStudentVerification = (userId, payload) => {
  return request({
    url: `/admin/users/${userId}/student-verify`,
    method: 'PUT',
    data: payload
  })
}

export const getAdminLogs = () => {
  return request({
    url: '/admin/logs?page=1&pageSize=50'
  })
}

export const getAdminReports = () => {
  return request({
    url: '/admin/reports?page=1&pageSize=20'
  })
}

export const markAdminReportProcessing = (reportId) => {
  return request({
    url: `/admin/reports/${reportId}/processing`,
    method: 'POST'
  })
}

export const handleAdminReport = (reportId, status, handleResult) => {
  return request({
    url: `/admin/reports/${reportId}/handle`,
    method: 'POST',
    data: {
      status,
      handleResult: handleResult || (status === 'PROCESSING' ? '举报处理中' : status === 'APPROVED' ? '举报已处理' : '举报已驳回')
    }
  })
}

export const getAdminAppeals = () => {
  return request({
    url: '/admin/appeals?page=1&pageSize=20'
  })
}

export const markAdminAppealProcessing = (appealId) => {
  return request({
    url: `/admin/appeals/${appealId}/processing`,
    method: 'POST'
  })
}

export const handleAdminAppeal = (appealId, status, handleResult) => {
  return request({
    url: `/admin/appeals/${appealId}/handle`,
    method: 'POST',
    data: {
      status,
      handleResult: handleResult || (status === 'PROCESSING' ? '申诉处理中' : status === 'APPROVED' ? '申诉已通过' : '申诉已驳回')
    }
  })
}

export const getSensitiveWords = () => {
  return request({ url: '/admin/sensitive-words?page=1&pageSize=50' })
}

export const createSensitiveWord = (payload) => {
  return request({
    url: '/admin/sensitive-words',
    method: 'POST',
    data: payload
  })
}

export const updateSensitiveWord = (id, payload) => {
  return request({
    url: `/admin/sensitive-words/${id}`,
    method: 'PUT',
    data: payload
  })
}

export const deleteSensitiveWord = (id) => {
  return request({
    url: `/admin/sensitive-words/${id}`,
    method: 'DELETE'
  })
}

export const getAnnouncements = () => {
  return request({ url: '/admin/announcements?page=1&pageSize=50' })
}

export const createAnnouncement = (payload) => {
  return request({
    url: '/admin/announcements',
    method: 'POST',
    data: payload
  })
}

export const updateAnnouncementStatus = (id, status) => {
  return request({
    url: `/admin/announcements/${id}/status`,
    method: 'PUT',
    data: { status }
  })
}

export const getAdminOrders = (status = 'ALL') => {
  const query = status && status !== 'ALL' ? `?status=${status}&page=1&pageSize=50` : '?page=1&pageSize=50'
  return request({ url: `/admin/orders${query}` })
}

export const updateAdminOrderStatus = (orderId, status, reason = '') => {
  return request({
    url: `/admin/orders/${orderId}/status`,
    method: 'PUT',
    data: { status, reason }
  })
}
