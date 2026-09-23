import { buildPrimaryVars } from './color';

/**
 * 基于 Element Plus 的主题预设：亮色只改 --el-color-primary*，深色使用官方 dark/css-vars + class="dark"。
 * @typedef {{ id: string, name: string, hint?: string, dark?: boolean, primary: string }} ThemePreset
 */

/** @type {ThemePreset[]} */
export const themes = [
  {
    id: 'default',
    name: '默认',
    hint: 'Element Plus 标准蓝',
    primary: '#409eff'
  },
  {
    id: 'dark',
    name: '深色',
    hint: '官方暗色变量',
    dark: true,
    primary: '#409eff'
  },
  {
    id: 'green',
    name: '绿色',
    hint: 'Success 色系主色',
    primary: '#67c23a'
  },
  {
    id: 'teal',
    name: '青绿',
    hint: '偏冷色主色',
    primary: '#009688'
  },
  {
    id: 'purple',
    name: '紫色',
    hint: '偏品牌紫主色',
    primary: '#626aef'
  },
  {
    id: 'orange',
    name: '橙色',
    hint: 'Warning 色系主色',
    primary: '#e6a23c'
  }
];

const themeById = new Map(themes.map((t) => [t.id, t]));

export const DEFAULT_THEME_ID = 'default';
export const STORAGE_KEY = 'itp-theme';

export function getThemeById(id) {
  return themeById.get(id) ?? themeById.get(DEFAULT_THEME_ID);
}

/** @returns {{ id: string, name: string, hint?: string, dark: boolean, vars: Record<string, string> }} */
export function resolveTheme(id) {
  const preset = getThemeById(id);
  return {
    id: preset.id,
    name: preset.name,
    hint: preset.hint,
    dark: !!preset.dark,
    vars: buildPrimaryVars(preset.primary)
  };
}
