declare global {
  interface Window {
    _env_?: { API_BASE_URL?: string }
  }
}

const runtimeApiBaseUrl =
  typeof window === 'undefined' ? undefined : window._env_?.API_BASE_URL

export const config = {
  API_BASE_URL:
    runtimeApiBaseUrl ||
    import.meta.env.VITE_API_BASE_URL ||
    'http://localhost:8080/api',
}
