export interface FoodEntry {
  id: number
  user_id: number
  user_name?: string
  food_name: string
  calories: number
  price: number
  entry_date: string
  created_at: string
}

export interface DailySummary {
  date: string
  total_calories: number
  calorie_limit: number
  calorie_exceeded: boolean
  total_price: number
  price_limit: number
  price_exceeded: boolean
}

export interface AdminReport {
  entries_last_7_days: number
  entries_previous_7_days: number
  average_calories_per_user_last_7_days: number
  users_count: number
  entries_comparison: {
    current_week: number
    previous_week: number
    difference: number
  }
}

export interface FoodEntryPayload {
  food_name: string
  calories: number
  price: number
  entry_date: string
  user_id?: number
}
