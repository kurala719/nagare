import request from '@/utils/request'

export function fetchGroupData(params) {
  return request({
    url: '/monitoring/groups',
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
    url: `/monitoring/groups/${id}`,
    method: 'get'
  })
}

export function fetchGroupDetail(id) {
  return request({
    url: `/monitoring/groups/${id}/detail`,
    method: 'get'
  })
}

export function addGroup(data) {
  return request({
    url: '/monitoring/groups',
    method: 'post',
    data
  })
}

export function updateGroup(id, data) {
  return request({
    url: `/monitoring/groups/${id}`,
    method: 'put',
    data
  })
}

export function deleteGroup(id, push = false) {
  return request({
    url: `/monitoring/groups/${id}${push ? '?push=true' : ''}`,
    method: 'delete'
  })
}

export function checkGroupStatus(id) {
  return request({
    url: `/monitoring/groups/${id}/check`,
    method: 'post'
  })
}

export function checkAllGroupsStatus() {
  return request({
    url: '/monitoring/groups/check',
    method: 'post'
  })
}

export function pullGroup(id) {
  return request({
    url: `/monitoring/groups/${id}/pull`,
    method: 'post'
  })
}

export function pushGroup(id) {
  return request({
    url: `/monitoring/groups/${id}/push`,
    method: 'post'
  })
}
