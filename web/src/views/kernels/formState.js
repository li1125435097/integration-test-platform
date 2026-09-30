export const KERNEL_TEST_FORM_KEY = 'kernel-test-form';

export function launchArgs(form) {
  const flags = Array.isArray(form?.flags) ? form.flags : [];
  const custom = Array.isArray(form?.customFlags) ? form.customFlags : [];
  return [...flags, ...custom.map((item) => String(item).trim()).filter(Boolean)];
}

export function fingerprintPayload(form) {
  if (!form?.fingerprintDirty && form?.activeSavedName) {
    return { savedName: form.activeSavedName };
  }
  if (form?.cipherMode) {
    return { contentBase64: String(form.fingerprintText || '').trim() };
  }
  return {
    plaintext: form?.fingerprintText ?? '',
    baseUrl: form?.baseUrl ?? '',
    bearer: form?.bearer ?? ''
  };
}

export function readCachedForm() {
  try {
    const raw = localStorage.getItem(KERNEL_TEST_FORM_KEY);
    if (!raw) return null;
    const data = JSON.parse(raw);
    if (!data || typeof data !== 'object' || Array.isArray(data)) return null;
    return data;
  } catch {
    return null;
  }
}

export function writeCachedForm(form) {
  localStorage.setItem(KERNEL_TEST_FORM_KEY, JSON.stringify(form));
}

export function clearCachedForm() {
  localStorage.removeItem(KERNEL_TEST_FORM_KEY);
}
