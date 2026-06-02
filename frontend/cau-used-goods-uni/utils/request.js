import { getToken, clearAuth } from './auth'

export const BASE_URL = 'http://127.0.0.1:8080'

export const request = ({
  url,
  method = 'GET',
  data = {},
  header = {},
  auth = true
}) => {
  const token = getToken()
  const requestHeader = {
    'Content-Type': 'application/json',
    ...header
  }

  if (auth && token) {
    requestHeader.Authorization = `Bearer ${token}`
  }

  return new Promise((resolve, reject) => {
    uni.request({
      url: `${BASE_URL}${url}`,
      method,
      data,
      header: requestHeader,
      success: (res) => {
        const body = res.data || {}

        if (res.statusCode === 401) {
          clearAuth()
          reject(new Error(body.message || 'Login expired, please login again'))
          return
        }

        if (res.statusCode < 200 || res.statusCode >= 300 || body.code !== 0) {
          reject(new Error(body.message || 'Request failed'))
          return
        }

        resolve(body.data)
      },
      fail: () => {
        reject(new Error('Cannot connect to server. Please make sure backend is running.'))
      }
    })
  })
}

export const uploadImage = (filePath) => {
  const token = getToken()
  return new Promise((resolve, reject) => {
    uni.uploadFile({
      url: `${BASE_URL}/upload/image`,
      filePath,
      name: 'file',
      header: {
        Authorization: token ? `Bearer ${token}` : ''
      },
      success: (res) => {
        let body = {}
        try {
          body = JSON.parse(res.data || '{}')
        } catch (error) {
          reject(new Error('图片上传响应解析失败'))
          return
        }
        if (res.statusCode < 200 || res.statusCode >= 300 || body.code !== 0) {
          reject(new Error(body.message || '图片上传失败'))
          return
        }
        resolve(body.data.url)
      },
      fail: () => reject(new Error('图片上传失败，请稍后重试'))
    })
  })
}
