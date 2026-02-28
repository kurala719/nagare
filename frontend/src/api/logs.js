import request from '@/utils/request'

export function fetchSystemLogs(params) {
  return request({
    url: '/system/logs/system',
    method: 'get',
    params
  })
}

