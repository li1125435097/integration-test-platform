async function parseJson(res) {
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data.error || res.statusText || '请求失败');
  }
  return data;
}

export async function listKernels() {
  const res = await fetch('/api/kernels');
  const data = await parseJson(res);
  return data.kernels || [];
}
