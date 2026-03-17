const trimTrailingSlash = (value) => String(value || '').replace(/\/+$/, '')

const toWebSocketOrigin = (value) => {
  const normalized = trimTrailingSlash(value)
  if (!normalized) return ''
  if (normalized.startsWith('wss://') || normalized.startsWith('ws://')) return normalized
  if (normalized.startsWith('https://')) return `wss://${normalized.slice('https://'.length)}`
  if (normalized.startsWith('http://')) return `ws://${normalized.slice('http://'.length)}`
  return normalized
}

export const getApiBaseURL = () => {
  const configuredBase = trimTrailingSlash(import.meta.env.VITE_API_BASE_URL)
  if (import.meta.env.DEV) {
    // Keep axios on same-origin in dev so Vite proxy can handle /api requests.
    return ''
  }
  if (configuredBase) return configuredBase
  if (typeof window !== 'undefined') return window.location.origin
  return 'http://localhost:8080'
}

export const getWebSocketBaseURL = () => {
  const configuredWs = toWebSocketOrigin(import.meta.env.VITE_WS_BASE_URL)
  if (configuredWs) return configuredWs

  if (import.meta.env.DEV) {
    // In dev, keep websocket same-origin so Vite proxy can forward /api WS.
    if (typeof window !== 'undefined') {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      return `${protocol}//${window.location.host}`
    }
    return 'ws://localhost:5173'
  }

  const configuredApi = toWebSocketOrigin(import.meta.env.VITE_API_BASE_URL)
  if (configuredApi) return configuredApi

  if (typeof window !== 'undefined') {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    return `${protocol}//${window.location.host}`
  }

  return 'ws://localhost:8080'
}