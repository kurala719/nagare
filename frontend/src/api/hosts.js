import request from '@/utils/request'

export function fetchHostData(params) {
  return request({
    url: '/hosts',
    method: 'get',
    params: {
      limit: 100,
      offset: 0,
      ...params
    }
  })
}

export function addHost(data) {
  return request({
    url: '/hosts',
    method: 'post',
    data
  })
}

export function updateHost(id, data) {
  return request({
    url: `/hosts/${id}`,
    method: 'put',
    data
  })
}

export function deleteHost(id) {
  return request({
    url: `/hosts/${id}`,
    method: 'delete'
  })
}

export function getHostById(id) {
  return request({
    url: `/hosts/${id}`,
    method: 'get'
  })
}

export function consultHostAI(id) {
  return request({
    url: `/hosts/${id}/consult`,
    method: 'post'
  })
}

export function fetchHostHistory(id, params) {
  return request({
    url: `/hosts/${id}/history`,
    method: 'get',
    params
  })
}

export function pullHostFromMonitor(monitorId, hostId) {
  return request({
    url: `/monitors/${monitorId}/hosts/${hostId}/pull`,
    method: 'post'
  })
}

export function pushHostToMonitor(monitorId, hostId) {
  return request({
    url: `/monitors/${monitorId}/hosts/${hostId}/push`,
    method: 'post'
  })
}

export function syncHostsFromMonitor(monitorId) {
  return request({
    url: `/monitors/${monitorId}/hosts/pull`,
    method: 'post'
  })
}

export function pushHostsToMonitor(monitorId) {
  return request({
    url: `/monitors/${monitorId}/hosts/push`,
    method: 'post'
  })
}

export function testSNMP(id) {
  return request({
    url: `/snmp-poll-direct/${id}`,
    method: 'post',
    timeout: 120000
  })
}
