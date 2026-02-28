import request from '@/utils/request'

export function fetchRetentionPolicies() {
  return request({
    url: '/sys/retention/policies',
    method: 'get'
  })
}

export function updateRetentionPolicy(data) {
  return request({
    url: '/sys/retention/policies',
    method: 'post',
    data
  })
}

export function performRetentionCleanup() {
  return request({
    url: '/sys/retention/cleanup',
    method: 'post'
  })
}
