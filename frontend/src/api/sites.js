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

export async function fetchSiteData(params = {}) {
  const { limit = 100, offset = 0, ...rest } = params || {};
  const url = `/api/v1/sites/${buildQuery({ ...rest, limit, offset })}`;
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

export async function fetchSiteDetail(id) {
  const url = `/api/v1/sites/${id}/detail`;
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

export async function addSite(data) {
  const url = '/api/v1/sites/';
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

export async function updateSite(id, data) {
  const url = `/api/v1/sites/${id}`;
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

export async function deleteSite(id) {
  const url = `/api/v1/sites/${id}`;
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

export async function checkSiteStatus(id) {
  const url = `/api/v1/sites/${id}/check`;
  const resp = await authFetch(url, {
    method: 'POST',
    headers: { 'Accept': 'application/json' },
    credentials: 'include',
  });
  if (!resp.ok) {
    throw new Error(`Request failed with status ${resp.status}`);
  }
  return await resp.json();
}

export async function checkAllSitesStatus() {
  const url = '/api/v1/sites/check';
  const resp = await authFetch(url, {
    method: 'POST',
    headers: { 'Accept': 'application/json' },
    credentials: 'include',
  });
  if (!resp.ok) {
    throw new Error(`Request failed with status ${resp.status}`);
  }
  return await resp.json();
}
