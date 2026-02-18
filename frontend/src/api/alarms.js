import request from '../utils/request';
import { authFetch } from '../utils/authFetch';

export function getAlarms() {
    return request({
        method: 'GET',
        url: '/alarms/',
    });
}

export function addAlarm(data) {
    return request({
        method: 'POST',
        url: '/alarms/',
        data,
    });
}

export function updateAlarm(id, data) {
    return request({
        method: 'PUT',
        url: `/alarms/${id}`,
        data,
    });
}

export function deleteAlarm(id) {
    return request({
        method: 'DELETE',
        url: `/alarms/${id}`,
    });
}

export function loginAlarm(id) {
    return request({
        method: 'POST',
        url: `/alarms/${id}/login`,
    });
}

export function regenerateAlarmEventToken(id) {
    return request({
        method: 'POST',
        url: `/alarms/${id}/event-token`,
    });
}

function buildQuery(params = {}) {
    const qs = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
        if (value === undefined || value === null || value === '') return;
        qs.set(key, String(value));
    });
    const query = qs.toString();
    return query ? `?${query}` : '';
}

export async function fetchAlarmData(params = {}) {
    const { limit = 100, offset = 0, ...rest } = params || {};
    const url = `/api/v1/alarms/${buildQuery({ ...rest, limit, offset })}`;
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
        console.error('fetchAlarmData error:', err);
        throw err;
    }
}
