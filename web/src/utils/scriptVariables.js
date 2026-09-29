const placeholderRe = /\{\{\s*([A-Za-z_][A-Za-z0-9_]*)(?:\s*=\s*([^}]*?))?\s*\}\}/g;

/**
 * First-seen placeholders across file contents.
 * inlineDefault is null when the placeholder has no "=", otherwise the trimmed {{name=value}} text.
 * The first "=" for a name wins.
 */
export function extractVariables(contents) {
  const byName = new Map();
  const names = [];
  for (const text of contents || []) {
    placeholderRe.lastIndex = 0;
    const source = text || '';
    let match = placeholderRe.exec(source);
    while (match) {
      const name = match[1];
      const hasInline = match[0].includes('=');
      const inlineDefault = hasInline ? (match[2] ?? '').trim() : null;
      const existing = byName.get(name);
      if (!existing) {
        names.push(name);
        byName.set(name, { name, inlineDefault: hasInline ? inlineDefault : null });
      } else if (hasInline && existing.inlineDefault === null) {
        byName.set(name, { name, inlineDefault });
      }
      match = placeholderRe.exec(source);
    }
  }
  return names.map((name) => byName.get(name));
}

/**
 * Variables present in the script that will actually run.
 * Names come from the current file contents; saved defaults fill values when the name still exists.
 */
export function variablesForCurrentScript(detail) {
  const files = Array.isArray(detail?.files) && detail.files.length
    ? detail.files
    : [{ content: detail?.content || '' }];
  const found = extractVariables(files.map((file) => file?.content || ''));
  const saved = new Map((detail?.variables || []).map((item) => [item.name, item.value ?? '']));
  return found.map((item) => ({
    name: item.name,
    value: saved.has(item.name) ? saved.get(item.name) : (item.inlineDefault ?? '')
  }));
}

/** Keys whose draft value differs from the saved default, including empty string. */
export function overridesFromDefaults(defaults, drafts) {
  const base = new Map((defaults || []).map((v) => [v.name, v.value ?? '']));
  const overrides = {};
  for (const draft of drafts || []) {
    const previous = base.has(draft.name) ? base.get(draft.name) : '';
    const next = draft.value ?? '';
    if (next !== previous) {
      overrides[draft.name] = next;
    }
  }
  return overrides;
}
