/**
 * Fetch host data from backend via Vite proxy.
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

export async function fetchHostData(params = {}) {
    const { limit = 100, offset = 0, ...rest } = params || {};
    const url = `/api/v1/hosts/${buildQuery({ ...rest, limit, offset })}`;
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
        console.error('fetchHostData error:', err);
        throw err;
    }
}

/**
 * Add a new host to the database.
 * @param {Object} data - Host data to add
 * @returns {Promise} - Response from the backend
 */
export async function addHost(data) {
    const url = '/api/v1/hosts/';
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
        console.error('addHost error:', err);
        throw err;
    }
}

/**
 * Update an existing host in the database.
 * @param {number} id - Host ID
 * @param {Object} data - Host data to update
 * @returns {Promise} - Response from the backend
 */
export async function updateHost(id, data) {
    const url = `/api/v1/hosts/${id}`;
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
        console.error('updateHost error:', err);
        throw err;
    }
}

/**
 * Delete a host from the database.
 * @param {number} id - Host ID
 * @returns {Promise} - Response from the backend
 */
export async function deleteHost(id) {
    const url = `/api/v1/hosts/${id}`;
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
        console.error('deleteHost error:', err);
        throw err;
    }
}

/**
 * Get a single host by ID.
 * @param {number} id - Host ID
 * @returns {Promise} - Response from the backend
 */
export async function getHostById(id) {
    const url = `/api/v1/hosts/${id}`;
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
        console.error('getHostById error:', err);
        throw err;
    }
}

/**
 * Send a host to AI for consultation.
 * @param {number} id - Host ID
 * @returns {Promise} - Response from the backend
 */
export async function consultHostAI(id) {
    const url = `/api/v1/hosts/${id}/consult`;
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
        console.error('consultHostAI error:', err);
        throw err;
    }
}

/**
 * Fetch host history data for trend charts.
 * @param {number} id - Host ID
 * @param {Object} params - Optional query params: from, to, limit
 * @returns {Promise}
 */
export async function fetchHostHistory(id, params = {}) {
    const url = `/api/v1/hosts/${id}/history${buildQuery(params)}`;
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
        console.error('fetchHostHistory error:', err);
        throw err;
    }
}

/**
 * Sync hosts from a monitor into the database.
 * @param {number} monitorId - Monitor ID
 */
export async function syncHostsFromMonitor(monitorId) {
    const url = `/api/v1/monitors/${monitorId}/hosts/pull`;
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
        console.error('syncHostsFromMonitor error:', err);
        throw err;
    }
}

/**
 * Sync both hosts and items from a monitor into the database.
 * @param {number} monitorId - Monitor ID
 */
export async function syncAllFromMonitor(monitorId) {
    const url = `/api/v1/monitors/${monitorId}/hosts/pull`;
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
        console.error('syncAllFromMonitor error:', err);
        throw err;
    }
}

/**
 * Import hosts from a monitor (add only).
 * @param {number} monitorId - Monitor ID
 */
export async function addHostsFromMonitor(monitorId) {
    const url = `/api/v1/monitors/${monitorId}/hosts/pull`;
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
        console.error('addHostsFromMonitor error:', err);
        throw err;
    }
}

/**
 * Pull a specific host from a monitor into the database.
 * @param {number} monitorId - Monitor ID
 * @param {number} hostId - Host ID
 */
export async function pullHostFromMonitor(monitorId, hostId) {
    const url = `/api/v1/monitors/${monitorId}/hosts/${hostId}/pull`;
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
        console.error('pullHostFromMonitor error:', err);
        throw err;
    }
}

/**
 * Push all hosts from local database to a monitor.
 * @param {number} monitorId - Monitor ID
 */
export async function pushHostsToMonitor(monitorId) {
    const url = `/api/v1/monitors/${monitorId}/hosts/push`;
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
        console.error('pushHostsToMonitor error:', err);
        throw err;
    }
}

/**
 * Push a specific host from local database to a monitor.
 * @param {number} monitorId - Monitor ID
 * @param {number} hostId - Host ID
 */
export async function pushHostToMonitor(monitorId, hostId) {
    const url = `/api/v1/monitors/${monitorId}/hosts/${hostId}/push`;
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
        console.error('pushHostToMonitor error:', err);
        throw err;
    }
}
