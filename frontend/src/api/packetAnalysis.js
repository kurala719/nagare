import request from '@/utils/request'

export function fetchPacketAnalyses(params) {
  return request({
    url: '/tooling/packets',
    method: 'get',
    params
  })
}

export function uploadPacket(formData) {
  return request({
    url: '/tooling/packets/upload',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export function deletePacketAnalysis(id) {
  return request({
    url: `/tooling/packets/${id}`,
    method: 'delete'
  })
}

export function startPacketAnalysis(id) {
  return request({
    url: `/tooling/packets/${id}/analyze`,
    method: 'post'
  })
}
