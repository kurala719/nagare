import request from '../utils/request';
import { authFetch } from '../utils/authFetch';
 
export function getMonitors() {
    return request({
        method:'GET',
        url:'/monitors/',
    })
}

export function addMonitor(data) {
    return request({
        method:'POST',
        url:'/monitors/',
        data,
    })
}

export function updateMonitor(id, data) {
    return request({
        method:'PUT',
        url:`/monitors/${id}`,
        data,
    })
}

export function deleteMonitor(id) {
    return request({
        method:'DELETE',
        url:`/monitors/${id}`,
    })
}

export function loginMonitor(id) {
    return request({
        method:'POST',
        url:`/monitors/${id}/login`,
    })
}

export function regenerateMonitorEventToken(id) {
    return request({
        method: 'POST',
        url: `/monitors/${id}/event-token`,
    })
}

export function checkMonitorStatus(id) {
    return request({
        method: 'POST',
        url: `/monitors/${id}/check`,
    })
}

export function checkAllMonitorsStatus() {
    return request({
        method: 'POST',
        url: '/monitors/check',
    })
}

export function syncGroupsFromMonitor(id) {
    return request({
        method: 'POST',
        url: `/monitors/${id}/groups/pull`,
    })
}

/**
 * Fetch monitor data from backend via Vite proxy.
 * Uses /api prefix which gets rewritten by proxy to backend.
 * Returns parsed JSON or throws an error.
 */
function buildQuery(params = {}) {
    const qs = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
        if (value === undefined || value === null || value === '') return;
        qs.set(key, String(value));
    });
    const query = qs.toString();
    return query ? `?${query}` : '';
}

export async function fetchMonitorData(params = {}) {
    const { limit = 100, offset = 0, ...rest } = params || {};
    const url = `/api/v1/monitors/${buildQuery({ ...rest, limit, offset })}`;
    try {
        const resp = await authFetch(url, {
            method: 'GET',
            headers: { 'Accept': 'application/json' },
            credentials: 'include',
        });
        if (!resp.ok) {
            throw new Error(`Request failed with status ${resp.status}`);
        }
        return await resp.json();
    } catch (err) {
        console.error('fetchMonitorData error:', err);
        throw err;
    }
}