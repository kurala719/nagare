const TOKEN_KEY = 'nagare_token'

function notifyAuthChanged() {
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent('auth-changed'))
  }
}

export function getToken() {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token) {
  localStorage.setItem(TOKEN_KEY, token)
  notifyAuthChanged()
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY)
  notifyAuthChanged()
}

export function getUserClaims() {
  const token = getToken()
  if (!token) return null
  const parts = token.split('.')
  if (parts.length < 2) return null
  try {
    let payload = parts[1].replace(/-/g, '+').replace(/_/g, '/')
    const pad = payload.length % 4
    if (pad) {
      payload += '='.repeat(4 - pad)
    }
    const decoded = JSON.parse(atob(payload))
    return decoded
  } catch {
    return null
  }
}

export function getUserPrivileges() {
  const claims = getUserClaims()
  if (typeof claims?.privileges === 'number') return claims.privileges
  if (typeof claims?.privileges === 'string') {
    const parsed = Number.parseInt(claims.privileges, 10)
    return Number.isNaN(parsed) ? 0 : parsed
  }
  return 0
}
