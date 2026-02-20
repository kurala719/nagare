import request from '@/utils/request'

export function fetchAlarmData(params) {
  return request({
    url: '/alarms',
    method: 'get',
    params: {
      limit: 100,
      offset: 0,
      ...params
    }
  })
}

export function getAlarmById(id) {
  return request({
    url: `/alarms/${id}`,
    method: 'get'
  })
}

export function addAlarm(data) {
  return request({
    url: '/alarms',
    method: 'post',
    data
  })
}

export function updateAlarm(id, data) {
  return request({
    url: `/alarms/${id}`,
    method: 'put',
    data
  })
}

export function deleteAlarm(id) {
  return request({
    url: `/alarms/${id}`,
    method: 'delete'
  })
}

export function loginAlarm(id, data) {
  return request({
    url: `/alarms/${id}/login`,
    method: 'post',
    data
  })
}

export function regenerateAlarmEventToken(id) {
  return request({
    url: `/alarms/${id}/event-token`,
    method: 'post'
  })
}

export function refreshAlarmEventToken(id, token) {
  return request({
    url: `/alarms/${id}/event-token/refresh?token=${token}`,
    method: 'post'
  })
}

export function setupAlarmMedia(id, data) {
  return request({
    url: `/alarms/${id}/setup-media`,
    method: 'post',
    data
  })
}
