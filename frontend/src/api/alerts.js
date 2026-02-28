import request from '@/utils/request'

export function fetchAlertData(params) {
  return request({
    url: '/monitor/alerts',
    method: 'get',
    params: {
      limit: 100,
      offset: 0,
      ...params
    }
  })
}

export function getAlertById(id) {
  return request({
    url: `/monitor/alerts/${id}`,
    method: 'get'
  })
}

export function addAlert(data) {
  return request({
    url: '/monitor/alerts',
    method: 'post',
    data
  })
}

export function updateAlert(id, data) {
  return request({
    url: `/monitor/alerts/${id}`,
    method: 'put',
    data
  })
}

export function deleteAlert(id) {
  return request({
    url: `/monitor/alerts/${id}`,
    method: 'delete'
  })
}

export function consultAlertAI(id, params) {
  return request({
    url: `/monitor/alerts/${id}/consult`,
    method: 'post',
    params
  })
}

export function getAlertScore() {
  return request({
    url: '/monitor/alerts/score',
    method: 'get'
  })
}

export function generateTestAlerts() {
  return request({
    url: '/monitor/alerts/generate-test',
    method: 'post'
  })
}
