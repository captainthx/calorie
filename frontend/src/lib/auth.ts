const TOKEN_KEY = 'calorie_token'

const ROLE_MAP: Record<string, 'user' | 'admin'> = {
  'user-token-123': 'user',
  'user-token-456': 'user',
  'admin-token-789': 'admin',
}

export function getToken(): string {
  return localStorage.getItem(TOKEN_KEY) ?? ''
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY)
}

export function getRole(token: string): 'user' | 'admin' | null {
  return ROLE_MAP[token] ?? null
}
