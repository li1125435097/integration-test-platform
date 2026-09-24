const LANG_LABELS = {
  javascript: 'JavaScript',
  python: 'Python',
  shell: 'Shell',
  java: 'Java',
  go: 'Go',
  rust: 'Rust',
  ruby: 'Ruby',
  php: 'PHP',
  perl: 'Perl',
  lua: 'Lua',
  csharp: 'C# (.NET)',
  kotlin: 'Kotlin',
  scala: 'Scala',
  dart: 'Dart',
  deno: 'Deno',
  bun: 'Bun'
};

export const interpreterLanguageOptions = Object.entries(LANG_LABELS).map(([value, label]) => ({
  value,
  label
}));

export function languageLabel(lang) {
  return LANG_LABELS[lang] || lang || '—';
}

export function formatTime(iso) {
  if (!iso) return '—';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return d.toLocaleString('zh-CN', { hour12: false });
}
