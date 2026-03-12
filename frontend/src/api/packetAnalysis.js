import request from '@/utils/request'

export function fetchPacketAnalyses(params) {
  return request({
    url: '/ai/packet-analyses',
    method: 'get',
    params
  })
}

export function uploadPacket(formData) {
  return request({
    url: '/ai/packet-analyses',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

export function deletePacketAnalysis(id) {
  return request({
    url: `/ai/packet-analyses/${id}`,
    method: 'delete'
  })
}

export function startPacketAnalysis(id) {
  return request({
    url: `/ai/packet-analyses/${id}/runs`,
    method: 'post'
  })
}
