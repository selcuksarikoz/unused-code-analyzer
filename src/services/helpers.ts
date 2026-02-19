export function generateUUID(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0;
    const v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

export function isGlobalIdentifier(name: string): boolean {
  const globals = new Set([
    'window', 'document', 'console', 'process', 'global', 'require',
    'module', 'exports', 'Buffer', 'setTimeout', 'setInterval',
    'clearTimeout', 'clearInterval', 'setImmediate', 'clearImmediate',
    '__dirname', '__filename', 'undefined', 'null', 'true', 'false',
    'NaN', 'Infinity', 'eval', 'parseInt', 'parseFloat', 'isNaN', 'isFinite',
    'decodeURI', 'decodeURIComponent', 'encodeURI', 'encodeURIComponent',
    'escape', 'unescape',
  ]);
  return globals.has(name);
}
