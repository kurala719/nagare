import request from '@/utils/request'

export function fetchPacketAnalyses(params) {
  return request({
    url: '/packets',
    method: 'get',
    params
  })
}

export function uploadPacket(formData) {
  return request({
    url: '/packets/upload',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export function deletePacketAnalysis(id) {
  return request({
    url: `/packets/${id}`,
    method: 'delete'
  })
}

export function startPacketAnalysis(id) {
  return request({
    url: `/packets/${id}/analyze`,
    method: 'post'
  })
}
