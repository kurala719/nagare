import request from '@/utils/request'

export function fetchSiteMessages(params) {
  return request({
    url: '/sys/site-messages',
    method: 'get',
    params
  })
}

export function getUnreadCount() {
  return request({
    url: '/sys/site-messages/unread-count',
    method: 'get'
  })
}

export function markAsRead(id) {
  return request({
    url: `/sys/site-messages/${id}/read`,
    method: 'put'
  })
}

export function markAllAsRead() {
  return request({
    url: '/sys/site-messages/read-all',
    method: 'put'
  })
}

export function deleteSiteMessage(id) {
  return request({
    url: `/sys/site-messages/${id}`,
    method: 'delete'
  })
}
