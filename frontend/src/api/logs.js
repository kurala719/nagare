import request from '@/utils/request'

export function fetchSystemLogs(params) {
  return request({
    url: '/logs/system',
    method: 'get',
    params
  })
}

export function fetchServiceLogs(params) {
  return request({
    url: '/logs/service',
    method: 'get',
    params
  })
}
