import request from '@/utils/request'

export function fetchHostData(params) {
  return request({
    url: '/monitoring/hosts',
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
    url: '/monitoring/hosts',
    method: 'post',
    data
  })
}

export function updateHost(id, data) {
  return request({
    url: `/monitoring/hosts/${id}`,
    method: 'put',
    data
  })
}

export function deleteHost(id, push = false) {
  return request({
    url: `/monitoring/hosts/${id}${push ? '?push=true' : ''}`,
    method: 'delete'
  })
}

export function getHostById(id) {
  return request({
    url: `/monitoring/hosts/${id}`,
    method: 'get'
  })
}

export function consultHostAI(id, params) {
  return request({
    url: `/monitoring/hosts/${id}/consult`,
    method: 'post',
    params
  })
}

export function fetchHostHistory(id, params) {
  return request({
    url: `/monitoring/hosts/${id}/history`,
    method: 'get',
    params
  })
}

export function pullHostFromMonitor(monitorId, hostId) {
  return request({
    url: `/monitoring/monitors/${monitorId}/hosts/${hostId}/pull`,
    method: 'post'
  })
}

export function pushHostToMonitor(monitorId, hostId) {
  return request({
    url: `/monitoring/monitors/${monitorId}/hosts/${hostId}/push`,
    method: 'post'
  })
}

export function syncHostsFromMonitor(monitorId) {
  return request({
    url: `/monitoring/monitors/${monitorId}/hosts/pull`,
    method: 'post'
  })
}

export function pushHostsToMonitor(monitorId) {
  return request({
    url: `/monitoring/monitors/${monitorId}/hosts/push`,
    method: 'post'
  })
}

export function testSNMP(id) {
  return request({
    url: `/monitoring/snmp-poll-direct/${id}`,
    method: 'post',
    timeout: 120000
  })
}

export function probeSnmpOid(id, oid) {
  const encodedOid = encodeURIComponent(oid || '')
  return request({
    url: `/monitoring/hosts/${id}/snmp/probe${encodedOid ? `?oid=${encodedOid}` : ''}`,
    method: 'post',
    timeout: 120000
  })
}
