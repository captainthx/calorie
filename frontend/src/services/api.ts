import { getToken, clearToken } from '../lib/auth';
import type { FoodEntry, DailySummary, AdminReport, FoodEntryPayload } from '../types/api';

const BASE_URL = import.meta.env.VITE_API_BASE_URL as string

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getToken()
  const res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
      ...options.headers,
    },
  })

  if (res.status === 401) {
    clearToken()
    window.location.reload()
    return undefined as never
  }

  const body = await res.json()

  if (!res.ok) {
    const err = new Error(body.error ?? `${res.status} ${res.statusText}`) as Error & { status: number }
    err.status = res.status
    throw err
  }

  return body.data as T
}

// User food entries
export const getFoodEntries = (dateFrom: string, dateTo: string) =>
  request<FoodEntry[]>(`/food-entries?date_from=${dateFrom}&date_to=${dateTo}`)

export const createFoodEntry = (payload: FoodEntryPayload) =>
  request<FoodEntry>('/food-entries', { method: 'POST', body: JSON.stringify(payload) })

export const updateFoodEntry = (id: number, payload: FoodEntryPayload) =>
  request<FoodEntry>(`/food-entries/${id}`, { method: 'PUT', body: JSON.stringify(payload) })

export const deleteFoodEntry = (id: number) =>
  request<void>(`/food-entries/${id}`, { method: 'DELETE' })

export const getDailySummary = (date: string) =>
  request<DailySummary>(`/daily-summary?date=${date}`)

// Admin food entries
export const getAdminFoodEntries = (dateFrom: string, dateTo: string) =>
  request<FoodEntry[]>(`/admin/food-entries?date_from=${dateFrom}&date_to=${dateTo}`)

export const createAdminFoodEntry = (payload: FoodEntryPayload) =>
  request<FoodEntry>('/admin/food-entries', { method: 'POST', body: JSON.stringify(payload) })

export const updateAdminFoodEntry = (id: number, payload: FoodEntryPayload) =>
  request<FoodEntry>(`/admin/food-entries/${id}`, { method: 'PUT', body: JSON.stringify(payload) })

export const deleteAdminFoodEntry = (id: number) =>
  request<void>(`/admin/food-entries/${id}`, { method: 'DELETE' })

export const getAdminReports = () =>
  request<AdminReport>('/admin/reports')
