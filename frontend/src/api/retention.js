import request from '@/utils/request'

export function fetchRetentionPolicies() {
  return request({
    url: '/retention/policies',
    method: 'get'
  })
}

export function updateRetentionPolicy(data) {
  return request({
    url: '/retention/policies',
    method: 'post',
    data
  })
}

export function performRetentionCleanup() {
  return request({
    url: '/retention/cleanup',
    method: 'post'
  })
}
