/**
 * gh-aw v0.87+ embeds safe-output config JSON as YAML-escaped strings
 * (`\"key\":\"value\"`). Normalize those quotes so tests can assert
 * JSON structure without depending on compiler quoting.
 */
export function normalizeCompiledLock(lock) {
  return lock.replaceAll('\\"', '"');
}
