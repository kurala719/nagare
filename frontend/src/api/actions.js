import request from '@/utils/request'

export function fetchActionData(params) {
  return request({
    url: '/delivery/actions',
    method: 'get',
    params: {
      limit: 100,
      offset: 0,
      ...params
    }
  })
}

export function addAction(data) {
  return request({
    url: '/delivery/actions',
    method: 'post',
    data
  })
}

export function updateAction(id, data) {
  return request({
    url: `/delivery/actions/${id}`,
    method: 'put',
    data
  })
}

export function deleteAction(id) {
  return request({
    url: `/delivery/actions/${id}`,
    method: 'delete'
  })
}

export function getActionById(id) {
  return request({
    url: `/delivery/actions/${id}`,
    method: 'get'
  })
}

export function testAction(id) {
  return request({
    url: `/delivery/actions/${id}/test`,
    method: 'post'
  })
}
