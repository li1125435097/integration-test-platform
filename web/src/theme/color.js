function parseHex(hex) {
  const h = hex.replace('#', '');
  const n = parseInt(h.length === 3 ? h.split('').map((c) => c + c).join('') : h, 16);
  return [(n >> 16) & 255, (n >> 8) & 255, n & 255];
}

function toHex(r, g, b) {
  return `#${[r, g, b]
    .map((v) => Math.round(v).toString(16).padStart(2, '0'))
    .join('')}`;
}

export function mixHex(color1, color2, weight) {
  const [r1, g1, b1] = parseHex(color1);
  const [r2, g2, b2] = parseHex(color2);
  const w = Math.min(1, Math.max(0, weight));
  return toHex(r1 + (r2 - r1) * w, g1 + (g2 - g1) * w, b1 + (b2 - b1) * w);
}

/** Element Plus 主色及衍生色（与 theme-chalk 变量命名一致）。 */
export function buildPrimaryVars(primary) {
  return {
    '--el-color-primary': primary,
    '--el-color-primary-light-3': mixHex(primary, '#ffffff', 0.3),
    '--el-color-primary-light-5': mixHex(primary, '#ffffff', 0.5),
    '--el-color-primary-light-7': mixHex(primary, '#ffffff', 0.7),
    '--el-color-primary-light-8': mixHex(primary, '#ffffff', 0.8),
    '--el-color-primary-light-9': mixHex(primary, '#ffffff', 0.9),
    '--el-color-primary-dark-2': mixHex(primary, '#000000', 0.2)
  };
}

export const PRIMARY_VAR_KEYS = Object.keys(buildPrimaryVars('#409eff'));
