import request from '@/utils/request'

export function fetchAlertData(params) {
  return request({
    url: '/alert/alerts',
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
    url: `/alert/alerts/${id}`,
    method: 'get'
  })
}

export function addAlert(data) {
  return request({
    url: '/alert/alerts',
    method: 'post',
    data
  })
}

export function updateAlert(id, data) {
  return request({
    url: `/alert/alerts/${id}`,
    method: 'put',
    data
  })
}

export function deleteAlert(id) {
  return request({
    url: `/alert/alerts/${id}`,
    method: 'delete'
  })
}

export function consultAlertAI(id, params) {
  return request({
    url: `/alert/alerts/${id}/consult`,
    method: 'post',
    params
  })
}

export function getAlertScore() {
  return request({
    url: '/alert/alerts/score',
    method: 'get'
  })
}

export function generateTestAlerts() {
  return request({
    url: '/alert/alerts/generate-test',
    method: 'post'
  })
}
