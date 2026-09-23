const LEGACY_TEMP_SUFFIX = '临时快照';

export const TEMP_SNAPSHOT_REMARK = LEGACY_TEMP_SUFFIX;

function maxVersionNum(versions, strictOnly) {
  let max = 0;
  for (const v of versions) {
    const re = strictOnly ? /^v(\d+)$/i : /^v(\d+)/i;
    const m = re.exec(v.id || '');
    if (m) max = Math.max(max, Number.parseInt(m[1], 10));
  }
  return max;
}

/** Next formal snapshot id: v1, v2, … (ignores legacy ids like v1临时快照). */
export function nextVersionId(versions) {
  return `v${maxVersionNum(versions, true) + 1}`;
}

/** Id for auto snapshot before version switch (counts legacy vN临时快照 ids). */
export function nextAutoSnapshotVersionId(versions) {
  return `v${maxVersionNum(versions, false) + 1}`;
}

/** Display id; legacy snapshots stored id as vN临时快照. */
export function versionDisplayId(version) {
  const id = version?.id || '';
  if (id.endsWith(LEGACY_TEMP_SUFFIX)) {
    return id.slice(0, -LEGACY_TEMP_SUFFIX.length);
  }
  return id;
}

export function versionRemark(version) {
  const stored = (version?.remark || '').trim();
  if (stored) return stored;
  const id = version?.id || '';
  if (id.endsWith(LEGACY_TEMP_SUFFIX)) return LEGACY_TEMP_SUFFIX;
  return '';
}
