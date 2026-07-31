import { $fetch } from 'ofetch'
import type { Response } from '~/types/response.type'

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

      const response = await $fetch<Response>(url, { method: 'GET' })

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

      const response = await $fetch<Response>(url, { method: 'GET' })

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

      const response = await $fetch<Response>(url, { method: 'POST' })

      return response
    } catch (error) {
      console.error('Failed to trigger screening:', error)
      throw error
    }
  }
}

export const screeningService = new ScreeningService()
