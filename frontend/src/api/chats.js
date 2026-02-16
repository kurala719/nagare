import request from '../utils/request';
import { authFetch } from '../utils/authFetch';
 
export function getMessages() {
    return request({
        method:'GET',
        url:'/messages/',
    })
}
 
export function getMessageByID(id) {
    return request({
        method:'GET',
        url:`/messages/${id}`,
    })
}
export function updateMessage(id,data) {
    return request({
        method:'PUT',
        url:`/messages/${id}`,
        data,
    })
}

export function addMessage(data) {
    return request({
        method:'POST',
        url:'/messages/',
        data,
    })
}

/**
 * Send a chat message to the backend.
 * @param {Object} data - Message data { message: string }
 * @returns {Promise} - Response from the backend
 */
export async function sendChatMessage(data) {
    const url = '/api/v1/chats/';
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
        console.error('sendChatMessage error:', err);
        throw err;
    }
}

/**
 * Fetch chat history from the backend.
 * @returns {Promise} - Response from the backend
 */
export async function fetchChatHistory() {
    const url = '/api/v1/chats/';
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
        console.error('fetchChatHistory error:', err);
        throw err;
    }
}