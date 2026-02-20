import request from '@/utils/request'

export function fetchGroupData(params) {
  return request({
    url: '/groups',
    method: 'get',
    params: {
      limit: 100,
      offset: 0,
      ...params
    }
  })
}

export function getGroupById(id) {
  return request({
    url: `/groups/${id}`,
    method: 'get'
  })
}

export function fetchGroupDetail(id) {
  return request({
    url: `/groups/${id}/detail`,
    method: 'get'
  })
}

export function addGroup(data) {
  return request({
    url: '/groups',
    method: 'post',
    data
  })
}

export function updateGroup(id, data) {
  return request({
    url: `/groups/${id}`,
    method: 'put',
    data
  })
}

export function deleteGroup(id) {
  return request({
    url: `/groups/${id}`,
    method: 'delete'
  })
}

export function checkGroupStatus(id) {
  return request({
    url: `/groups/${id}/check`,
    method: 'post'
  })
}

export function checkAllGroupsStatus() {
  return request({
    url: '/groups/check',
    method: 'post'
  })
}

export function pullGroup(id) {
  return request({
    url: `/groups/${id}/pull`,
    method: 'post'
  })
}

export function pushGroup(id) {
  return request({
    url: `/groups/${id}/push`,
    method: 'post'
  })
}
