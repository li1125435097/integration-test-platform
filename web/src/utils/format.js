const LANG_LABELS = {
  javascript: 'JavaScript',
  python: 'Python',
  shell: 'Shell'
};

export function languageLabel(lang) {
  return LANG_LABELS[lang] || lang || '—';
}

export function formatTime(iso) {
  if (!iso) return '—';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toLocaleString('zh-CN', { hour12: false });
}
