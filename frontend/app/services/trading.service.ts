import { $fetch } from 'ofetch'

export interface TransactionRequest {
  ticker: string
  transaction_type: 'buy' | 'sell'
  lot: number
  price: number
  transaction_date: string
  notes?: string
}

export interface UpdateTransactionRequest {
  ticker?: string
  transaction_type?: 'buy' | 'sell'
  lot?: number
  price?: number
  transaction_date?: string
  notes?: string
}

export interface Transaction {
  id: number
  ticker: string
  transaction_type: string
  lot: number
  price: number
  transaction_date: string
  notes: string | null
  created_at: string
}

export interface Position {
  ticker: string
  total_lot: number
  average_buy_price: number
  status: string
}

export interface PositionReviewResponse {
  ticker: string
  position: Position
  indicators: any
  signal: any
  position_review: Record<string, any>
  ai_insight: string
}

export class TradingService {
  private getBaseURL(): string {
    return (typeof globalThis !== 'undefined' && (globalThis as any).__API_URL__)
      || (typeof process !== 'undefined' && process.env.NUXT_PUBLIC_API_URL)
      || 'https://api-trackly.aksanara.id'
  }

  private getToken(): string | null {
    try {
      const raw = localStorage.getItem('trackly_session')
      if (!raw) return null
      const session = JSON.parse(raw)
      return session?.token || null
    } catch { return null }
  }

  private headers(): Record<string, string> {
    const token = this.getToken()
    const h: Record<string, string> = {}
    if (token) h['Authorization'] = `Bearer ${token}`
    return h
  }

  async createTransaction(req: TransactionRequest): Promise<Transaction> {
    const url = `${this.getBaseURL()}/api/transactions`
    const res = await $fetch<{ success: boolean; data: Transaction }>(url, {
      method: 'POST', body: req, headers: this.headers(),
    })
    return res.data
  }

  async getTransactions(params?: { ticker?: string; dateStart?: string; dateEnd?: string; type?: string }): Promise<Transaction[]> {
    const q = new URLSearchParams()
    if (params?.ticker) q.set('ticker', params.ticker)
    if (params?.dateStart) q.set('date_start', params.dateStart)
    if (params?.dateEnd) q.set('date_end', params.dateEnd)
    if (params?.type) q.set('type', params.type)
    const qs = q.toString()
    const url = `${this.getBaseURL()}/api/transactions${qs ? `?${qs}` : ''}`
    const res = await $fetch<{ success: boolean; data: Transaction[] }>(url, {
      method: 'GET', headers: this.headers(),
    })
    return res.data || []
  }

  async getPositions(): Promise<Position[]> {
    const url = `${this.getBaseURL()}/api/positions`
    const res = await $fetch<{ success: boolean; data: Position[] }>(url, {
      method: 'GET', headers: this.headers(),
    })
    return res.data || []
  }

  async getPositionAnalysis(ticker: string): Promise<PositionReviewResponse> {
    const url = `${this.getBaseURL()}/api/positions/${encodeURIComponent(ticker)}/analysis`
    const res = await $fetch<{ success: boolean; data: PositionReviewResponse }>(url, {
      method: 'GET', headers: this.headers(),
    })
    return res.data
  }

  async updateTransaction(id: number, req: UpdateTransactionRequest): Promise<Transaction> {
    const url = `${this.getBaseURL()}/api/transactions/${id}`
    const res = await $fetch<{ success: boolean; data: Transaction }>(url, {
      method: 'PATCH', body: req, headers: this.headers(),
    })
    return res.data
  }

  async deleteTransaction(id: number): Promise<void> {
    const url = `${this.getBaseURL()}/api/transactions/${id}`
    await $fetch(url, { method: 'DELETE', headers: this.headers() })
  }

  async postAiInsight(ticker: string, dateStart: string, dateEnd: string): Promise<string> {
    const url = `${this.getBaseURL()}/api/analisis/ai-insight`
    const res = await $fetch<{ success: boolean; data: { ai_insight: string } }>(url, {
      method: 'POST', body: { ticker, date_start: dateStart, date_end: dateEnd }, headers: this.headers(),
    })
    return res.data.ai_insight
  }
}

export const tradingService = new TradingService()