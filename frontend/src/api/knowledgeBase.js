import request from '@/utils/request'

export function fetchKnowledgeBase(params) {
  return request({
    url: '/knowledge-base',
    method: 'get',
    params
  })
}

export function getKnowledgeBaseById(id) {
  return request({
    url: `/knowledge-base/${id}`,
    method: 'get'
  })
}

export function addKnowledgeBase(data) {
  return request({
    url: '/knowledge-base',
    method: 'post',
    data
  })
}

export function updateKnowledgeBase(id, data) {
  return request({
    url: `/knowledge-base/${id}`,
    method: 'put',
    data
  })
}

export function deleteKnowledgeBase(id) {
  return request({
    url: `/knowledge-base/${id}`,
    method: 'delete'
  })
}

export function bulkDeleteKnowledgeBase(ids) {
  return Promise.all(ids.map(id => deleteKnowledgeBase(id)))
}
