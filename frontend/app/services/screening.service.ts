import { api } from '~/utils/api'

export interface ScreeningResult {
  id: string
  scan_date: string
  ticker: string
  rank: number
  score: number
  confidence: string
  overall: string
  avg_volume: number | null
  trading_plan: any | null
  ai_insight: string | null
  created_at: string
}

export const screeningService = {
  getLatest: () => api<{ data: ScreeningResult[] }>('/api/screening/latest'),
  getByDate: (date: string) => api<{ data: ScreeningResult[] }>(`/api/screening/${date}`),
  triggerScreening: () => api<{ data: { status: string } }>('/api/screening/trigger', { method: 'POST' }),
}
