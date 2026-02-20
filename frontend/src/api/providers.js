import request from '@/utils/request'

export function fetchProviderData(params) {
  return request({
    url: '/providers',
    method: 'get',
    params: {
      limit: 100,
      offset: 0,
      ...params
    }
  })
}

export function addProvider(data) {
  return request({
    url: '/providers',
    method: 'post',
    data
  })
}

export function updateProvider(id, data) {
  return request({
    url: `/providers/${id}`,
    method: 'put',
    data
  })
}

export function deleteProvider(id) {
  return request({
    url: `/providers/${id}`,
    method: 'delete'
  })
}

export function checkProviderStatus(id) {
  return request({
    url: `/providers/${id}/check`,
    method: 'post'
  })
}

export function checkAllProvidersStatus() {
  return request({
    url: '/providers/check',
    method: 'post'
  })
}

export function getProviderById(id) {
  return request({
    url: `/providers/${id}`,
    method: 'get'
  })
}
