import request from '@/utils/request'

export function fetchChatHistory(params) {
  return request({
    url: '/chats',
    method: 'get',
    params
  })
}

export function sendChatMessage(data) {
  return request({
    url: '/chats',
    method: 'post',
    data,
    timeout: 120000 // 120 seconds timeout for AI processing
  })
}

export function getMessages() {
  return request({
    url: '/messages',
    method: 'get'
  })
}

export function getMessageByID(id) {
  return request({
    url: `/messages/${id}`,
    method: 'get'
  })
}

export function updateMessage(id, data) {
  return request({
    url: `/messages/${id}`,
    method: 'put',
    data
  })
}

export function addMessage(data) {
  return request({
    url: '/messages',
    method: 'post',
    data
  })
}
