async function parseJson(res) {
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data.error || res.statusText || '请求失败');
  }
  return data;
}

function fileNameFromDisposition(header) {
  const match = /filename="([^"]+)"/.exec(header || '');
  return match?.[1] || 'itp-data.zip';
}

export async function exportData() {
  const res = await fetch('/api/data/export');
  if (!res.ok) {
    await parseJson(res);
    return;
  }
  const blob = await res.blob();
  const name = fileNameFromDisposition(res.headers.get('Content-Disposition'));
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = name;
  link.click();
  URL.revokeObjectURL(url);
}

export async function importData(file) {
  const body = new FormData();
  body.append('file', file);
  return parseJson(
    await fetch('/api/data/import', {
      method: 'POST',
      body
    })
  );
}
