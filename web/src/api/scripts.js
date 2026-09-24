async function parseJson(res) {
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data.error || res.statusText || '请求失败');
  }
  return data;
}

export async function listScripts() {
  const res = await fetch('/api/scripts');
  const data = await parseJson(res);
  return data.scripts || [];
}

export async function getScript(id) {
  return parseJson(await fetch(`/api/scripts/${encodeURIComponent(id)}`));
}

export async function createScript(payload) {
  return parseJson(
    await fetch('/api/scripts', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    })
  );
}

export async function updateScript(id, payload) {
  return parseJson(
    await fetch(`/api/scripts/${encodeURIComponent(id)}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    })
  );
}

export async function deleteScript(id) {
  return parseJson(
    await fetch(`/api/scripts/${encodeURIComponent(id)}`, { method: 'DELETE' })
  );
}

export async function listVersions(id) {
  const res = await fetch(`/api/scripts/${encodeURIComponent(id)}/versions`);
  const data = await parseJson(res);
  return data.versions || [];
}

export async function addVersion(id, name, remark = '') {
  return parseJson(
    await fetch(`/api/scripts/${encodeURIComponent(id)}/versions`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: name ?? '', remark: remark ?? '' })
    })
  );
}

export async function restoreVersion(id, versionId) {
  return parseJson(
    await fetch(`/api/scripts/${encodeURIComponent(id)}/restore`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ versionId })
    })
  );
}

export async function updateVersionRemark(id, versionId, remark = '') {
  return parseJson(
    await fetch(
      `/api/scripts/${encodeURIComponent(id)}/versions/${encodeURIComponent(versionId)}`,
      {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ remark: remark ?? '' })
      }
    )
  );
}

export async function deleteVersion(id, versionId) {
  return parseJson(
    await fetch(
      `/api/scripts/${encodeURIComponent(id)}/versions/${encodeURIComponent(versionId)}`,
      { method: 'DELETE' }
    )
  );
}
