import { normalizeErrorMessage } from './request'

export function navigate(path, params = {}) {
  const query = Object.entries(params)
    .filter(([, value]) => value !== undefined && value !== null && value !== '')
    .map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(value)}`)
    .join('&')
  uni.navigateTo({ url: query ? `${path}?${query}` : path })
}

export function showSuccess(title) {
  uni.showToast({ title, icon: 'success' })
}

export function showError(error) {
  uni.showToast({
    title: normalizeErrorMessage(error?.message, '操作失败，请稍后重试'),
    icon: 'none',
    duration: 2400
  })
}
