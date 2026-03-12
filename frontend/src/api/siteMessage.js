import request from '@/utils/request'

export function fetchSiteMessages(params) {
  return request({
    url: '/delivery/site-messages',
    method: 'get',
    params
  })
}

export function getUnreadCount() {
  return request({
    url: '/delivery/site-messages/unread/count',
    method: 'get'
  })
}

export function markAsRead(id) {
  return request({
    url: `/delivery/site-messages/${id}/read-status`,
    method: 'put'
  })
}

export function markAllAsRead() {
  return request({
    url: '/delivery/site-messages/read-status',
    method: 'put'
  })
}

export function deleteSiteMessage(id) {
  return request({
    url: `/delivery/site-messages/${id}`,
    method: 'delete'
  })
}
