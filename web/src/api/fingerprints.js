async function parseJson(res) {
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data.error || res.statusText || '请求失败');
  }
  return data;
}

export async function listFingerprints() {
  const data = await parseJson(await fetch('/api/fingerprints'));
  return data.files || [];
}

export async function readFingerprint(name) {
  const data = await parseJson(
    await fetch(`/api/fingerprints/${encodeURIComponent(name)}`)
  );
  return data.contentBase64 || '';
}

export async function saveFingerprint(name, contentBase64) {
  return parseJson(
    await fetch('/api/fingerprints', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name, contentBase64 })
    })
  );
}

export async function encryptFingerprint({ plaintext, baseUrl, bearer }) {
  const data = await parseJson(
    await fetch('/api/fingerprints/encrypt', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ plaintext, baseUrl, bearer })
    })
  );
  return data.contentBase64 || '';
}

export async function decryptFingerprint(contentBase64) {
  const data = await parseJson(
    await fetch('/api/fingerprints/decrypt', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ contentBase64 })
    })
  );
  return data.plaintext || '';
}
