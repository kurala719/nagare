import request from '@/utils/request'

export function fetchProviderData(params) {
  return request({
    url: '/sys/providers',
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
    url: '/sys/providers',
    method: 'post',
    data
  })
}

export function updateProvider(id, data) {
  return request({
    url: `/sys/providers/${id}`,
    method: 'put',
    data
  })
}

export function deleteProvider(id) {
  return request({
    url: `/sys/providers/${id}`,
    method: 'delete'
  })
}

export function checkProviderStatus(id) {
  return request({
    url: `/sys/providers/${id}/check`,
    method: 'post'
  })
}

export function checkAllProvidersStatus() {
  return request({
    url: '/sys/providers/check',
    method: 'post'
  })
}

export function fetchProviderModels(id) {
  return request({
    url: `/sys/providers/${id}/fetch-models`,
    method: 'post'
  })
}

export function fetchModelsDirect(data) {
  return request({
    url: '/sys/providers/fetch-models-direct',
    method: 'post',
    data
  })
}

export function getProviderById(id) {
  return request({
    url: `/sys/providers/${id}`,
    method: 'get'
  })
}
