import request from '@/utils/request'

export function fetchSystemStatus() {
  return request({
    url: '/system/status',
    method: 'get'
  })
}

export function fetchHealthScore() {
  return request({
    url: '/system/health',
    method: 'get'
  })
}

export function fetchNetworkStatusHistory(params) {
  return request({
    url: '/system/health/history',
    method: 'get',
    params
  })
}

export function fetchNetworkMetrics(params) {
  return request({
    url: '/system/metrics',
    method: 'get',
    params
  })
}
