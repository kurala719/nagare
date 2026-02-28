import request from '@/utils/request'

export function fetchProviderData(params) {
  return request({
    url: '/ai/providers',
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
    url: '/ai/providers',
    method: 'post',
    data
  })
}

export function updateProvider(id, data) {
  return request({
    url: `/ai/providers/${id}`,
    method: 'put',
    data
  })
}

export function deleteProvider(id) {
  return request({
    url: `/ai/providers/${id}`,
    method: 'delete'
  })
}

export function checkProviderStatus(id) {
  return request({
    url: `/ai/providers/${id}/check`,
    method: 'post'
  })
}

export function checkAllProvidersStatus() {
  return request({
    url: '/ai/providers/check',
    method: 'post'
  })
}

export function fetchProviderModels(id) {
  return request({
    url: `/ai/providers/${id}/fetch-models`,
    method: 'post'
  })
}

export function fetchModelsDirect(data) {
  return request({
    url: '/ai/providers/fetch-models-direct',
    method: 'post',
    data
  })
}

export function getProviderById(id) {
  return request({
    url: `/ai/providers/${id}`,
    method: 'get'
  })
}
