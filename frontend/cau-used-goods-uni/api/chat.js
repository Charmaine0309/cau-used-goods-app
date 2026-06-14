import { request } from '../utils/request'

export const createOrGetConversation = (productId) => request({
  url: '/chat/conversations',
  method: 'POST',
  data: { productId }
})

export const listConversations = (params = {}) => request({
  url: '/chat/conversations',
  data: params
})

export const listMessages = (conversationId, params = {}) => request({
  url: `/chat/conversations/${conversationId}/messages`,
  data: params
})

export const sendMessage = (conversationId, content) => request({
  url: `/chat/conversations/${conversationId}/messages`,
  method: 'POST',
  data: { content }
})

export const markConversationRead = (conversationId) => request({
  url: `/chat/conversations/${conversationId}/read`,
  method: 'PUT'
})
