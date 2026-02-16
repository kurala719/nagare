import { authFetch } from '../utils/authFetch'

/**
 * Fetch alert data from backend via Vite proxy.
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

export async function fetchAlertData(params = {}) {
    const { limit = 100, offset = 0, ...rest } = params || {};
    const url = `/api/v1/alerts/${buildQuery({ ...rest, limit, offset })}`;
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
        console.error('fetchAlertData error:', err);
        throw err;
    }
}

/**
 * Get a single alert by ID.
 * @param {number} id - Alert ID
 * @returns {Promise} - Response from the backend
 */
export async function getAlertById(id) {
    const url = `/api/v1/alerts/${id}`;
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
        console.error('getAlertById error:', err);
        throw err;
    }
}

/**
 * Add a new alert to the database.
 * @param {Object} data - Alert data to add
 * @returns {Promise} - Response from the backend
 */
export async function addAlert(data) {
    const url = '/api/v1/alerts/';
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
        console.error('addAlert error:', err);
        throw err;
    }
}

/**
 * Update an existing alert.
 * @param {number} id - Alert ID
 * @param {Object} data - Alert data to update
 * @returns {Promise} - Response from the backend
 */
export async function updateAlert(id, data) {
    const url = `/api/v1/alerts/${id}`;
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
        console.error('updateAlert error:', err);
        throw err;
    }
}

/**
 * Delete an alert by ID.
 * @param {number} id - Alert ID
 * @returns {Promise} - Response from the backend
 */
export async function deleteAlert(id) {
    const url = `/api/v1/alerts/${id}`;
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
        console.error('deleteAlert error:', err);
        throw err;
    }
}

/**
 * Send an alert to AI for consultation.
 * @param {number} id - Alert ID
 * @returns {Promise} - Response from the backend
 */
export async function consultAlertAI(id) {
    const url = `/api/v1/alerts/${id}/consult`;
    try {
        const resp = await authFetch(url, {
            method: 'POST',
            headers: { 
                'Content-Type': 'application/json',
                'Accept': 'application/json' 
            },
            credentials: 'include',
        });
        if (!resp.ok) {
            throw new Error(`Request failed with status ${resp.status}`);
        }
        return await resp.json();
    } catch (err) {
        console.error('consultAlertAI error:', err);
        throw err;
    }
}
