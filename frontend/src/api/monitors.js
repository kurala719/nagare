import request from '@/utils/request'

export function fetchMonitorData(params) {
  return request({
    url: '/monitoring/monitors',
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
    url: `/monitoring/monitors/${id}`,
    method: 'get'
  })
}

export function fetchMonitorHistory(id, params) {
  return request({
    url: `/analysis/monitors/${id}/history`,
    method: 'get',
    params
  })
}

export function addMonitor(data) {
  return request({
    url: '/monitoring/monitors',
    method: 'post',
    data
  })
}

export function updateMonitor(id, data) {
  return request({
    url: `/monitoring/monitors/${id}`,
    method: 'put',
    data
  })
}

export function deleteMonitor(id) {
  return request({
    url: `/monitoring/monitors/${id}`,
    method: 'delete'
  })
}

export function loginMonitor(id, data) {
  return request({
    url: `/monitoring/monitors/${id}/sessions`,
    method: 'post',
    data
  })
}

export function checkMonitorStatus(id) {
  return request({
    url: `/monitoring/monitors/${id}/checks`,
    method: 'post'
  })
}

export function checkAllMonitorsStatus() {
  return request({
    url: '/monitoring/monitors/checks',
    method: 'post'
  })
}

export function regenerateMonitorEventToken(id) {
  return request({
    url: `/monitoring/monitors/${id}/event-tokens`,
    method: 'post'
  })
}

export function syncGroupsFromMonitor(monitorId) {
  return request({
    url: `/monitoring/monitors/${monitorId}/group-imports`,
    method: 'post'
  })
}

export function refreshEventToken(id, token) {
  return request({
    url: `/monitoring/monitors/${id}/event-token-refreshes?token=${token}`,
    method: 'post'
  })
}
