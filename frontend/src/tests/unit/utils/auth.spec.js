import { describe, it, expect, beforeEach, vi } from 'vitest'
import { getToken, setToken, clearToken, getUserClaims, getUserPrivileges } from '../../../utils/auth'

describe('Auth Utils', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.restoreAllMocks()
  })

  it('should set and get token', () => {
    setToken('test-token')
    expect(getToken()).toBe('test-token')
  })

  it('should clear token', () => {
    setToken('test-token')
    clearToken()
    expect(getToken()).toBeNull()
  })

  it('should parse user claims', () => {
    // mock jwt token: header.payload.signature
    // payload: {"privileges": 2, "sub": "user"}
    const payload = btoa(JSON.stringify({ privileges: 2, sub: 'user' }))
    const token = `header.${payload}.signature`
    setToken(token)

    const claims = getUserClaims()
    expect(claims).toEqual({ privileges: 2, sub: 'user' })
  })

  it('should return null for invalid token', () => {
    setToken('invalid-token')
    expect(getUserClaims()).toBeNull()
  })

  it('should get user privileges', () => {
    const payload = btoa(JSON.stringify({ privileges: 3 }))
    const token = `header.${payload}.signature`
    setToken(token)

    expect(getUserPrivileges()).toBe(3)
  })

  it('should return 0 for missing privileges', () => {
    const payload = btoa(JSON.stringify({ sub: 'user' }))
    const token = `header.${payload}.signature`
    setToken(token)

    expect(getUserPrivileges()).toBe(0)
  })
})
