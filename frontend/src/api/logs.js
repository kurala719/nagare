import { authFetch } from '../utils/authFetch'

function buildQuery(params = {}) {
  const qs = new URLSearchParams()
  Object.entries(params).forEach(([key, value]) => {
    if (value === undefined || value === null || value === '') return
    qs.set(key, String(value))
  })
  const query = qs.toString()
  return query ? `?${query}` : ''
}

async function fetchLogs(url, params = {}) {
  const query = buildQuery(params)
  const resp = await authFetch(`${url}${query}`, {
    method: 'GET',
    headers: { 'Accept': 'application/json' },
    credentials: 'include',
  })
  if (!resp.ok) {
    throw new Error(`Request failed with status ${resp.status}`)
  }
  return await resp.json()
}

export async function fetchSystemLogs(params = {}) {
  return fetchLogs('/api/v1/logs/system', params)
}

export async function fetchServiceLogs(params = {}) {
  return fetchLogs('/api/v1/logs/service', params)
}
