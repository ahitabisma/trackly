import { api } from '~/utils/api'

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

  async createTransaction(req: TransactionRequest): Promise<Transaction> {
    const url = `${this.getBaseURL()}/api/transactions`
    const res = await api<{ success: boolean; data: Transaction }>(url, {
      method: 'POST', body: req,
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
    const res = await api<{ success: boolean; data: Transaction[] }>(url, {
      method: 'GET',
    })
    return res.data || []
  }

  async getPositions(): Promise<Position[]> {
    const url = `${this.getBaseURL()}/api/positions`
    const res = await api<{ success: boolean; data: Position[] }>(url, {
      method: 'GET',
    })
    return res.data || []
  }

  async getPositionAnalysis(ticker: string): Promise<PositionReviewResponse> {
    const url = `${this.getBaseURL()}/api/positions/${encodeURIComponent(ticker)}/analysis`
    const res = await api<{ success: boolean; data: PositionReviewResponse }>(url, {
      method: 'GET',
    })
    return res.data
  }

  async updateTransaction(id: number, req: UpdateTransactionRequest): Promise<Transaction> {
    const url = `${this.getBaseURL()}/api/transactions/${id}`
    const res = await api<{ success: boolean; data: Transaction }>(url, {
      method: 'PATCH', body: req,
    })
    return res.data
  }

  async deleteTransaction(id: number): Promise<void> {
    const url = `${this.getBaseURL()}/api/transactions/${id}`
    await api(url, { method: 'DELETE' })
  }

  async postAiInsight(ticker: string, dateEnd: string, indicators: any, snapshot?: any): Promise<string> {
    const url = `${this.getBaseURL()}/api/analisis/ai-insight`
    const res = await api<{ success: boolean; data: { ai_insight: string } }>(url, {
      method: 'POST', body: { ticker, date_end: dateEnd, indicators, snapshot },
    })
    return res.data.ai_insight
  }
}

export const tradingService = new TradingService()
