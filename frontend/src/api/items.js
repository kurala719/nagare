/**
 * Fetch item data from backend via Vite proxy.
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

export async function fetchItemData(params = {}) {
    const { limit = 100, offset = 0, ...rest } = params || {};
    const url = `/api/v1/items/${buildQuery({ ...rest, limit, offset })}`;
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
        console.error('fetchItemData error:', err);
        throw err;
    }
}

/**
 * Fetch items by host ID from backend.
 * @param {number} hostId - Host ID to filter items by
 * @returns {Promise} - Response from the backend
 */
export async function fetchItemsByHost(hostId) {
    const url = `/api/v1/items/?hid=${hostId}`;
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
        console.error('fetchItemsByHost error:', err);
        throw err;
    }
}

/**
 * Get a single item by ID.
 * @param {number} id - Item ID
 * @returns {Promise} - Response from the backend
 */
export async function getItemById(id) {
    const url = `/api/v1/items/${id}`;
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
        console.error('getItemById error:', err);
        throw err;
    }
}

/**
 * Add a new item to the database.
 * @param {Object} data - Item data to add
 * @returns {Promise} - Response from the backend
 */
export async function addItem(data) {
    const url = '/api/v1/items/';
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
        console.error('addItem error:', err);
        throw err;
    }
}

/**
 * Update an existing item.
 * @param {number} id - Item ID
 * @param {Object} data - Item data to update
 * @returns {Promise} - Response from the backend
 */
export async function updateItem(id, data) {
    const url = `/api/v1/items/${id}`;
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
        console.error('updateItem error:', err);
        throw err;
    }
}

/**
 * Delete an item by ID.
 * @param {number} id - Item ID
 * @returns {Promise} - Response from the backend
 */
export async function deleteItem(id) {
    const url = `/api/v1/items/${id}`;
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
        console.error('deleteItem error:', err);
        throw err;
    }
}

/**
 * Send an item to AI for consultation.
 * @param {number} id - Item ID
 * @returns {Promise} - Response from the backend
 */
export async function consultItemAI(id) {
    const url = `/api/v1/items/${id}/consult`;
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
        console.error('consultItemAI error:', err);
        throw err;
    }
}

/**
 * Fetch item history data for trend charts.
 * @param {number} id - Item ID
 * @param {Object} params - Optional query params: from, to, limit
 * @returns {Promise}
 */
export async function fetchItemHistory(id, params = {}) {
    const url = `/api/v1/items/${id}/history${buildQuery(params)}`;
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
        console.error('fetchItemHistory error:', err);
        throw err;
    }
}

/**
 * Sync items from external monitor for a host.
 * @param {number} hostId - Host ID
 */
export async function syncItemsFromMonitor(monitorId, hostId) {
    const url = `/api/v1/monitors/${monitorId}/hosts/${hostId}/items/pull`;
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
        console.error('syncItemsFromMonitor error:', err);
        throw err;
    }
}

/**
 * Add items from external monitor for a host (import only).
 * @param {number} hostId - Host ID
 */
export async function addItemsByHostFromMonitor(hostId) {
    const url = `/api/v1/items/hosts/${hostId}/import`;
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
        console.error('addItemsByHostFromMonitor error:', err);
        throw err;
    }
}

/**
 * Pull all items from a monitor by monitor ID.
 * @param {number} monitorId - Monitor ID
 */
export async function pullItemsFromMonitor(monitorId) {
    const url = `/api/v1/monitors/${monitorId}/items/pull`;
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
        console.error('pullItemsFromMonitor error:', err);
        throw err;
    }
}

/**
 * Pull items from a monitor for a specific host.
 * @param {number} monitorId - Monitor ID
 * @param {number} hostId - Host ID
 */
export async function pullItemsFromHost(monitorId, hostId) {
    const url = `/api/v1/monitors/${monitorId}/hosts/${hostId}/items/pull`;
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
        console.error('pullItemsFromHost error:', err);
        throw err;
    }
}

/**
 * Push items to a monitor for a specific host.
 * @param {number} monitorId - Monitor ID
 * @param {number} hostId - Host ID
 */
export async function pushItemsToHost(monitorId, hostId) {
    const url = `/api/v1/monitors/${monitorId}/hosts/${hostId}/items/push`;
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
        console.error('pushItemsToHost error:', err);
        throw err;
    }
}
