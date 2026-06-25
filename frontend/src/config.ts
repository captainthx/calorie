declare global {
  // eslint-disable-next-line no-var
  var _env_: { API_BASE_URL?: string } | undefined
}

export const config = {
  API_BASE_URL:
    globalThis._env_?.API_BASE_URL ||
    import.meta.env.VITE_API_BASE_URL ||
    'http://localhost:8080/api',
}
