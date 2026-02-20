import request from '@/utils/request'

export function fetchMediaTypeData(params) {
  return request({
    url: '/media-types',
    method: 'get',
    params: {
      limit: 100,
      offset: 0,
      ...params
    }
  })
}

export function addMediaType(data) {
  return request({
    url: '/media-types',
    method: 'post',
    data
  })
}

export function updateMediaType(id, data) {
  return request({
    url: `/media-types/${id}`,
    method: 'put',
    data
  })
}

export function deleteMediaType(id) {
  return request({
    url: `/media-types/${id}`,
    method: 'delete'
  })
}

export function getMediaTypeById(id) {
  return request({
    url: `/media-types/${id}`,
    method: 'get'
  })
}
