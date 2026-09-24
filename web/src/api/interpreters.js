async function parseJson(res) {
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data.error || res.statusText || '请求失败');
  }
  return data;
}

export async function listInterpreters() {
  const res = await fetch('/api/interpreters');
  const data = await parseJson(res);
  return data.interpreters || [];
}

export async function createInterpreter(payload) {
  return parseJson(
    await fetch('/api/interpreters', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    })
  );
}

export async function updateInterpreter(id, payload) {
  return parseJson(
    await fetch(`/api/interpreters/${encodeURIComponent(id)}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    })
  );
}

export async function setInterpreterDefault(id, isDefault) {
  return parseJson(
    await fetch(`/api/interpreters/${encodeURIComponent(id)}/default`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ isDefault: !!isDefault })
    })
  );
}

export async function deleteInterpreter(id) {
  return parseJson(
    await fetch(`/api/interpreters/${encodeURIComponent(id)}`, { method: 'DELETE' })
  );
}

export async function updateInterpreterPath(id, path, version = '') {
  return parseJson(
    await fetch(`/api/interpreters/${encodeURIComponent(id)}/path`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ path, version: version || '' })
    })
  );
}

export async function discoverInterpreters(language = '') {
  const qs = language ? `?language=${encodeURIComponent(language)}` : '';
  const res = await fetch(`/api/interpreters/discover${qs}`);
  const data = await parseJson(res);
  return data.candidates || [];
}

export async function batchCreateInterpreters(items) {
  return parseJson(
    await fetch('/api/interpreters/batch', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ items })
    })
  );
}
