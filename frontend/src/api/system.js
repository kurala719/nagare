import request from '@/utils/request'

export function fetchSystemStatus() {
  return request({
    url: '/sys/system/status',
    method: 'get'
  })
}

export function fetchHealthScore() {
  return request({
    url: '/sys/system/health',
    method: 'get'
  })
}

export function fetchNetworkStatusHistory(params) {
  return request({
    url: '/sys/system/health/history',
    method: 'get',
    params
  })
}

export function fetchNetworkMetrics(params) {
  return request({
    url: '/sys/system/metrics',
    method: 'get',
    params
  })
}
