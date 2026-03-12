import request from '@/utils/request'

export function fetchMediaData(params) {
  return request({
    url: '/delivery/media',
    method: 'get',
    params: {
      limit: 100,
      offset: 0,
      ...params
    }
  })
}

export function addMedia(data) {
  return request({
    url: '/delivery/media',
    method: 'post',
    data
  })
}

export function updateMedia(id, data) {
  return request({
    url: `/delivery/media/${id}`,
    method: 'put',
    data
  })
}

export function deleteMedia(id) {
  return request({
    url: `/delivery/media/${id}`,
    method: 'delete'
  })
}

export function testMedia(id) {
  return request({
    url: `/delivery/media/${id}/test-runs`,
    method: 'post'
  })
}

export function getMediaById(id) {
  return request({
    url: `/delivery/media/${id}`,
    method: 'get'
  })
}
