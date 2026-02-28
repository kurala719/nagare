import request from '@/utils/request'

export function fetchChatHistory(params) {
  return request({
    url: '/ai/chats',
    method: 'get',
    params
  })
}

export function sendChatMessage(data) {
  return request({
    url: '/ai/chats',
    method: 'post',
    data,
    timeout: 120000 // 120 seconds timeout for AI processing
  })
}
