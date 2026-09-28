const FILE_NAME_RE = /^[\w.-]+$/;

export function extForLanguage(lang) {
  switch (lang) {
    case 'python':
      return 'py';
    case 'shell':
      return 'sh';
    default:
      return 'js';
  }
}

export function mainFileName(lang) {
  return `main.${extForLanguage(lang)}`;
}

export function withLanguageExt(name, lang) {
  const ext = extForLanguage(lang);
  const trimmed = (name || '').trim();
  if (!trimmed) return '';
  if (trimmed.toLowerCase().endsWith(`.${ext}`)) return trimmed;
  const bare = trimmed.replace(/\.[^.]+$/, '');
  return `${bare}.${ext}`;
}

export function isValidFileName(name) {
  return FILE_NAME_RE.test(name) && !name.includes('..');
}

export function emptyMainFile(lang) {
  return {
    name: mainFileName(lang),
    kind: 'main',
    content: '',
    sourceScriptId: '',
    sourceFileName: '',
    sourceScriptName: '',
    missing: false
  };
}

export function filesFromDetail(data) {
  if (Array.isArray(data?.files) && data.files.length) {
    return data.files.map((f) => ({
      name: f.name,
      kind: f.kind || 'local',
      content: f.content || '',
      sourceScriptId: f.sourceScriptId || '',
      sourceFileName: f.sourceFileName || '',
      sourceScriptName: f.sourceScriptName || '',
      missing: !!f.missing
    }));
  }
  return [
    {
      ...emptyMainFile(data?.language || 'javascript'),
      content: data?.content || ''
    }
  ];
}

export function renameOwnedFilesForLanguage(fileList, lang) {
  const ext = extForLanguage(lang);
  const used = new Set();
  return fileList.map((f) => {
    if (f.kind === 'main') {
      const name = `main.${ext}`;
      used.add(name);
      return { ...f, name };
    }
    const base = String(f.name || 'file').replace(/\.[^.]+$/, '') || 'file';
    let name = `${base}.${ext}`;
    let n = 1;
    while (used.has(name)) {
      name = `${base}${n}.${ext}`;
      n += 1;
    }
    used.add(name);
    return { ...f, name };
  });
}

export function requireHint(lang) {
  switch (lang) {
    case 'python':
      return '子文件可通过 import 模块名 调用（不含扩展名）';
    case 'shell':
      return '子文件可通过 source ./文件名 调用';
    default:
      return "子文件可通过 require('./文件名') 调用（不含扩展名）";
  }
}
