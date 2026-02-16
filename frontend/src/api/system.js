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

/**
 * Fetch network status history (health score) for trend charts.
 * @param {Object} params - Optional query params: from, to, limit
 * @returns {Promise}
 */
export async function fetchNetworkStatusHistory(params = {}) {
    const url = `/api/v1/system/health/history${buildQuery(params)}`
    try {
        const resp = await authFetch(url, {
            method: 'GET',
            headers: { 'Accept': 'application/json' },
            credentials: 'include',
        })
        if (!resp.ok) {
            throw new Error(`Request failed with status ${resp.status}`)
        }
        return await resp.json()
    } catch (err) {
        console.error('fetchNetworkStatusHistory error:', err)
        throw err
    }
}
