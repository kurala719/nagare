import request from '@/utils/request'

export function fetchMonitorData(params) {
  return request({
    url: '/monitor/monitors',
    method: 'get',
    params: {
      limit: 100,
      offset: 0,
      ...params
    }
  })
}

export function getMonitorById(id) {
  return request({
    url: `/monitor/monitors/${id}`,
    method: 'get'
  })
}

export function addMonitor(data) {
  return request({
    url: '/monitor/monitors',
    method: 'post',
    data
  })
}

export function updateMonitor(id, data) {
  return request({
    url: `/monitor/monitors/${id}`,
    method: 'put',
    data
  })
}

export function deleteMonitor(id) {
  return request({
    url: `/monitor/monitors/${id}`,
    method: 'delete'
  })
}

export function loginMonitor(id, data) {
  return request({
    url: `/monitor/monitors/${id}/login`,
    method: 'post',
    data
  })
}

export function checkMonitorStatus(id) {
  return request({
    url: `/monitor/monitors/${id}/check`,
    method: 'post'
  })
}

export function checkAllMonitorsStatus() {
  return request({
    url: '/monitor/monitors/check',
    method: 'post'
  })
}

export function regenerateMonitorEventToken(id) {
  return request({
    url: `/monitor/monitors/${id}/event-token`,
    method: 'post'
  })
}

export function syncGroupsFromMonitor(monitorId) {
  return request({
    url: `/monitor/monitors/${monitorId}/groups/pull`,
    method: 'post'
  })
}

export function refreshEventToken(id, token) {
  return request({
    url: `/monitor/monitors/${id}/event-token/refresh?token=${token}`,
    method: 'post'
  })
}
