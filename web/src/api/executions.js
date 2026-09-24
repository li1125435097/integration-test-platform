async function parseJson(res) {
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data.error || res.statusText || '请求失败');
  }
  return data;
}

export async function listExecutionRecords() {
  const res = await fetch('/api/execution-records');
  const data = await parseJson(res);
  return data.records || [];
}

export async function getExecutionRecord(id) {
  return parseJson(await fetch(`/api/execution-records/${encodeURIComponent(id)}`));
}
