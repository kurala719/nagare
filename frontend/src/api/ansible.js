import request from '@/utils/request'

export function fetchPlaybooks(params) {
  return request({
    url: '/ansible/playbooks',
    method: 'get',
    params
  })
}

export function getPlaybook(id) {
  return request({
    url: `/ansible/playbooks/${id}`,
    method: 'get'
  })
}

export function createPlaybook(data) {
  return request({
    url: '/ansible/playbooks',
    method: 'post',
    data
  })
}

export function updatePlaybook(id, data) {
  return request({
    url: `/ansible/playbooks/${id}`,
    method: 'put',
    data
  })
}

export function deletePlaybook(id) {
  return request({
    url: `/ansible/playbooks/${id}`,
    method: 'delete'
  })
}

export function runPlaybook(id, data) {
  return request({
    url: `/ansible/playbooks/${id}/run`,
    method: 'post',
    data
  })
}

export function fetchAnsibleJobs(params) {
  return request({
    url: '/ansible/jobs',
    method: 'get',
    params
  })
}

export function getAnsibleJob(id) {
  return request({
    url: `/ansible/jobs/${id}`,
    method: 'get'
  })
}

export function recommendPlaybook(data) {
  return request({
    url: '/ansible/playbooks/recommend',
    method: 'post',
    data
  })
}
