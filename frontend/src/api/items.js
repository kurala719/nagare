import request from '@/utils/request'

export function fetchItemData(params) {
  return request({
    url: '/items',
    method: 'get',
    params: {
      limit: 100,
      offset: 0,
      ...params
    }
  })
}

export function fetchItemsByHost(hostId) {
  return request({
    url: `/items?hid=${hostId}&limit=1000`,
    method: 'get'
  })
}

export function getItemById(id) {
  return request({
    url: `/items/${id}`,
    method: 'get'
  })
}

export function addItem(data) {
  return request({
    url: '/items',
    method: 'post',
    data
  })
}

export function updateItem(id, data) {
  return request({
    url: `/items/${id}`,
    method: 'put',
    data
  })
}

export function deleteItem(id) {
  return request({
    url: `/items/${id}`,
    method: 'delete'
  })
}

export function consultItemAI(id) {
  return request({
    url: `/items/${id}/consult`,
    method: 'post'
  })
}

export function fetchItemHistory(id, params) {
  return request({
    url: `/items/${id}/history`,
    method: 'get',
    params
  })
}

export function pullItemsFromHost(monitorId, hostId) {
  return request({
    url: `/monitors/${monitorId}/hosts/${hostId}/items/pull`,
    method: 'post'
  })
}

export function pushItemsToHost(monitorId, hostId) {
  return request({
    url: `/monitors/${monitorId}/hosts/${hostId}/items/push`,
    method: 'post'
  })
}

export function addItemsByHostIDFromMonitor(hostId) {
  return request({
    url: `/items/hosts/${hostId}/import`,
    method: 'post'
  })
}

export function pullItemsFromMonitor(monitorId) {
  return request({
    url: `/monitors/${monitorId}/items/pull`,
    method: 'post'
  })
}

export function pushItemsFromMonitor(monitorId) {
  return request({
    url: `/monitors/${monitorId}/items/push`,
    method: 'post'
  })
}
