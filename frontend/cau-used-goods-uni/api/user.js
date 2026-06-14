import { request } from '../utils/request'

export const getPublicProfile = (id) => request({
  url: `/users/${id}/public`
})
