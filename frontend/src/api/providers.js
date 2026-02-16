/**
 * Fetch provider data from backend via Vite proxy.
 * Uses /api prefix which gets rewritten by proxy to backend.
 * Returns parsed JSON or throws an error.
 */
import { authFetch } from '../utils/authFetch'

function buildQuery(params = {}) {
    const qs = new URLSearchParams();
    Object.entries(params).forEach(([key, value]) => {
        if (value === undefined || value === null || value === '') return;
        qs.set(key, String(value));
    });
    const query = qs.toString();
    return query ? `?${query}` : '';
}

export async function fetchProviderData(params = {}) {
    const { limit = 100, offset = 0, ...rest } = params || {};
    const url = `/api/v1/providers/${buildQuery({ ...rest, limit, offset })}`;
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
        console.error('fetchProviderData error:', err);
        throw err;
    }
}

/**
 * Add a new provider to the database.
 * @param {Object} data - Provider data to add
 * @returns {Promise} - Response from the backend
 */
export async function addProvider(data) {
    const url = '/api/v1/providers/';
    try {
        const resp = await authFetch(url, {
            method: 'POST',
            headers: { 
                'Content-Type': 'application/json',
                'Accept': 'application/json' 
            },
            credentials: 'include',
            body: JSON.stringify(data),
        });
        if (!resp.ok) {
            throw new Error(`Request failed with status ${resp.status}`);
        }
        return await resp.json();
    } catch (err) {
        console.error('addProvider error:', err);
        throw err;
    }
}

/**
 * Update an existing provider in the database.
 * @param {number} id - Provider ID
 * @param {Object} data - Provider data to update
 * @returns {Promise} - Response from the backend
 */
export async function updateProvider(id, data) {
    const url = `/api/v1/providers/${id}`;
    try {
        const resp = await authFetch(url, {
            method: 'PUT',
            headers: { 
                'Content-Type': 'application/json',
                'Accept': 'application/json' 
            },
            credentials: 'include',
            body: JSON.stringify(data),
        });
        if (!resp.ok) {
            throw new Error(`Request failed with status ${resp.status}`);
        }
        return await resp.json();
    } catch (err) {
        console.error('updateProvider error:', err);
        throw err;
    }
}

/**
 * Delete a provider from the database.
 * @param {number} id - Provider ID
 * @returns {Promise} - Response from the backend
 */
export async function deleteProvider(id) {
    const url = `/api/v1/providers/${id}`;
    try {
        const resp = await authFetch(url, {
            method: 'DELETE',
            headers: { 'Accept': 'application/json' },
            credentials: 'include',
        });
        if (!resp.ok) {
            throw new Error(`Request failed with status ${resp.status}`);
        }
        return await resp.json();
    } catch (err) {
        console.error('deleteProvider error:', err);
        throw err;
    }
}

export async function checkProviderStatus(id) {
    const url = `/api/v1/providers/${id}/check`;
    try {
        const resp = await authFetch(url, {
            method: 'POST',
            headers: { 'Accept': 'application/json' },
            credentials: 'include',
        });
        if (!resp.ok) {
            throw new Error(`Request failed with status ${resp.status}`);
        }
        return await resp.json();
    } catch (err) {
        console.error('checkProviderStatus error:', err);
        throw err;
    }
}

export async function checkAllProvidersStatus() {
    const url = '/api/v1/providers/check';
    try {
        const resp = await authFetch(url, {
            method: 'POST',
            headers: { 'Accept': 'application/json' },
            credentials: 'include',
        });
        if (!resp.ok) {
            throw new Error(`Request failed with status ${resp.status}`);
        }
        return await resp.json();
    } catch (err) {
        console.error('checkAllProvidersStatus error:', err);
        throw err;
    }
}
