import request from '@/utils/request'

export function fetchSiteMessages(params) {
  return request({
    url: '/site-messages',
    method: 'get',
    params
  })
}

export function getUnreadCount() {
  return request({
    url: '/site-messages/unread-count',
    method: 'get'
  })
}

export function markAsRead(id) {
  return request({
    url: `/site-messages/${id}/read`,
    method: 'put'
  })
}

export function markAllAsRead() {
  return request({
    url: '/site-messages/read-all',
    method: 'put'
  })
}

export function deleteSiteMessage(id) {
  return request({
    url: `/site-messages/${id}`,
    method: 'delete'
  })
}
