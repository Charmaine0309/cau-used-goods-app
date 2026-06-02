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
    url: '/admin/logs?page=1&pageSize=10'
  })
}
