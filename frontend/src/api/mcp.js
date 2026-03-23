import request from '@/utils/request'

export function getMCPServers() {
  return request({
    url: '/ai/mcp-servers',
    method: 'get'
  })
}

export function saveMCPServers(data) {
  return request({
    url: '/ai/mcp-servers',
    method: 'post',
    data
  })
}

export function reloadMCPServers() {
  return request({
    url: '/ai/mcp-servers/reload',
    method: 'post'
  })
}

export function getMCPClientStatus() {
  return request({
    url: '/ai/mcp-servers/status',
    method: 'get'
  })
}

export function testMCPServer(data) {
  return request({
    url: '/ai/mcp-servers/test',
    method: 'post',
    data
  })
}
