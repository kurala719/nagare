import request from '@/utils/request'

export function fetchTriggerData(params) {
  return request({
    url: '/triggers',
    method: 'get',
    params: {
      limit: 100,
      offset: 0,
      ...params
    }
  })
}

export function addTrigger(data) {
  return request({
    url: '/triggers',
    method: 'post',
    data
  })
}

export function updateTrigger(id, data) {
  return request({
    url: `/triggers/${id}`,
    method: 'put',
    data
  })
}

export function deleteTrigger(id) {
  return request({
    url: `/triggers/${id}`,
    method: 'delete'
  })
}
