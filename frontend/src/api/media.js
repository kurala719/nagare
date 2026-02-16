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

export async function fetchMediaData(params = {}) {
  const { limit = 100, offset = 0, ...rest } = params || {};
  const url = `/api/v1/media/${buildQuery({ ...rest, limit, offset })}`;
  const resp = await authFetch(url, {
    method: 'GET',
    headers: { 'Accept': 'application/json' },
    credentials: 'include',
  });
  if (!resp.ok) {
    throw new Error(`Request failed with status ${resp.status}`);
  }
  return await resp.json();
}

export async function addMedia(data) {
  const url = '/api/v1/media/';
  const resp = await authFetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'Accept': 'application/json' },
    credentials: 'include',
    body: JSON.stringify(data),
  });
  if (!resp.ok) {
    throw new Error(`Request failed with status ${resp.status}`);
  }
  return await resp.json();
}

export async function updateMedia(id, data) {
  const url = `/api/v1/media/${id}`;
  const resp = await authFetch(url, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', 'Accept': 'application/json' },
    credentials: 'include',
    body: JSON.stringify(data),
  });
  if (!resp.ok) {
    throw new Error(`Request failed with status ${resp.status}`);
  }
  return await resp.json();
}

export async function deleteMedia(id) {
  const url = `/api/v1/media/${id}`;
  const resp = await authFetch(url, {
    method: 'DELETE',
    headers: { 'Accept': 'application/json' },
    credentials: 'include',
  });
  if (!resp.ok) {
    throw new Error(`Request failed with status ${resp.status}`);
  }
  return await resp.json();
}
