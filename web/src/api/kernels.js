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

export async function runKernelTest(payload) {
  const data = await parseJson(
    await fetch('/api/kernel-tests/run', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    })
  );
  return data.results || [];
}

export async function listKernelTestPlans() {
  const data = await parseJson(await fetch('/api/kernel-test-plans'));
  return data.plans || [];
}

export async function getKernelTestPlan(id) {
  return parseJson(await fetch(`/api/kernel-test-plans/${encodeURIComponent(id)}`));
}

export async function createKernelTestPlan(payload) {
  return parseJson(
    await fetch('/api/kernel-test-plans', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    })
  );
}

export async function updateKernelTestPlan(id, payload) {
  return parseJson(
    await fetch(`/api/kernel-test-plans/${encodeURIComponent(id)}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    })
  );
}

export async function deleteKernelTestPlan(id) {
  return parseJson(
    await fetch(`/api/kernel-test-plans/${encodeURIComponent(id)}`, {
      method: 'DELETE'
    })
  );
}
