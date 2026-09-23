import { PRIMARY_VAR_KEYS } from './color';
import { DEFAULT_THEME_ID, STORAGE_KEY, getThemeById, resolveTheme } from './themes';

function clearPrimaryOverrides(root) {
  for (const key of PRIMARY_VAR_KEYS) {
    root.style.removeProperty(key);
  }
}

export function applyTheme(themeId) {
  const resolved = resolveTheme(themeId);
  const root = document.documentElement;

  clearPrimaryOverrides(root);
  root.classList.toggle('dark', resolved.dark);
  root.setAttribute('data-theme', resolved.id);

  for (const [key, value] of Object.entries(resolved.vars)) {
    root.style.setProperty(key, value);
  }

  try {
    localStorage.setItem(STORAGE_KEY, resolved.id);
  } catch {
    /* ignore */
  }
  return resolved.id;
}

export function loadStoredThemeId() {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored && getThemeById(stored)) {
      return stored;
    }
  } catch {
    /* ignore */
  }
  return DEFAULT_THEME_ID;
}

export function initTheme() {
  return applyTheme(loadStoredThemeId());
}
