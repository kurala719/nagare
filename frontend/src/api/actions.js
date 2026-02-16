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

export async function fetchActionData(params = {}) {
  const { limit = 100, offset = 0, ...rest } = params || {};
  const url = `/api/v1/actions/${buildQuery({ ...rest, limit, offset })}`;
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

export async function addAction(data) {
  const url = '/api/v1/actions/';
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

export async function updateAction(id, data) {
  const url = `/api/v1/actions/${id}`;
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

export async function deleteAction(id) {
  const url = `/api/v1/actions/${id}`;
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
