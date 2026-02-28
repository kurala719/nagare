import request from '@/utils/request'

export function fetchRetentionPolicies() {
  return request({
    url: '/system/retention/policies',
    method: 'get'
  })
}

export function updateRetentionPolicy(data) {
  return request({
    url: '/system/retention/policies',
    method: 'post',
    data
  })
}

export function performRetentionCleanup() {
  return request({
    url: '/system/retention/cleanup',
    method: 'post'
  })
}
