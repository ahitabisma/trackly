import { $fetch } from 'ofetch'
import type { Response } from '~/types/response.type'

const SESSION_KEY = 'trackly_session'

function getToken(): string | null {
  try {
    const raw = localStorage.getItem(SESSION_KEY)
    if (!raw) return null
    const session = JSON.parse(raw)
    return session?.token || null
  } catch { return null }
}

function authHeaders(): Record<string, string> {
  const token = getToken()
  return token ? { Authorization: `Bearer ${token}` } : {}
}

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

export class ScreeningService {
  private getBaseURL(): string {
    const url = (typeof globalThis !== 'undefined' && (globalThis as any).__API_URL__)
      || (typeof process !== 'undefined' && process.env.NUXT_PUBLIC_API_URL)
      || 'https://api-trackly.aksanara.id'
    return url
  }

  async getLatest(): Promise<Response> {
    try {
      const baseURL = this.getBaseURL()
      const url = `${baseURL}/api/screening/latest`
      const response = await $fetch<Response>(url, { method: 'GET', headers: authHeaders() })
      return response
    } catch (error) {
      console.error('Failed to fetch latest screening:', error)
      throw error
    }
  }

  async getByDate(date: string): Promise<Response> {
    try {
      const baseURL = this.getBaseURL()
      const url = `${baseURL}/api/screening/${date}`
      const response = await $fetch<Response>(url, { method: 'GET', headers: authHeaders() })
      return response
    } catch (error) {
      console.error('Failed to fetch screening by date:', error)
      throw error
    }
  }

  async triggerScreening(): Promise<Response> {
    try {
      const baseURL = this.getBaseURL()
      const url = `${baseURL}/api/screening/trigger`
      const response = await $fetch<Response>(url, { method: 'POST', headers: authHeaders() })
      return response
    } catch (error) {
      console.error('Failed to trigger screening:', error)
      throw error
    }
  }
}

export const screeningService = new ScreeningService()
