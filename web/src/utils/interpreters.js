import { languageLabel } from '@/utils/format';

/** Unique language options derived from configured interpreters (解释器管理). */
export function languageOptionsFromInterpreters(interpreters, includeLanguage = '') {
  const set = new Set();
  for (const item of interpreters || []) {
    if (item.language) set.add(item.language);
  }
  if (includeLanguage) set.add(includeLanguage);
  return [...set]
    .sort((a, b) => languageLabel(a).localeCompare(languageLabel(b), 'zh-CN'))
    .map((value) => ({ value, label: languageLabel(value) }));
}

/** Whether interpreter management has a default runtime for this language. */
export function hasDefaultInterpreterForLanguage(interpreters, language) {
  return (interpreters || []).some((i) => i.language === language && i.isDefault);
}

/** Resolve list-cell label and availability for a script row. */
export function scriptInterpreterCell(row, interpreters) {
  const list = interpreters || [];
  if (!row?.interpreterId) {
    return {
      text: '默认',
      ok: hasDefaultInterpreterForLanguage(list, row?.language)
    };
  }
  const interp = list.find((i) => i.id === row.interpreterId);
  const ok = !!interp && interp.language === row.language;
  return {
    text: interp?.version?.trim() || '—',
    ok
  };
}

/** Interpreters for script language, sorted with default first. */
export function interpretersForLanguage(interpreters, language) {
  return (interpreters || [])
    .filter((i) => i.language === language)
    .slice()
    .sort((a, b) => Number(b.isDefault) - Number(a.isDefault));
}
