export type Result<T> = { ok: true; value: T } | { ok: false; error: Error };

export function Success<T>(v: T): Result<T> {
  return {
    ok: true,
    value: v,
  };
}

export function Fail<T>(e: Error): Result<T> {
  return {
    ok: false,
    error: e,
  };
}
