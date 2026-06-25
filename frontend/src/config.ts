declare global {
  interface Window {
    _env_?: { API_BASE_URL?: string }
  }
}

export const config = {
  API_BASE_URL:
    window._env_?.API_BASE_URL ||
    import.meta.env.VITE_API_BASE_URL ||
    'http://localhost:8080/api',
}
