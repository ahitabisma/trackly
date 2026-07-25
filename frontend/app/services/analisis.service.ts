import { $fetch } from 'ofetch'

export interface TickerSearchResult {
    kode: string
    nama_perusahaan: string
    papan_pencatatan: string
}

export interface Snapshot {
    kode: string
    company_name: string
    price: number
    high52: number
    low52: number
    volume: number
    marketcap: number
    pe: number | null
    eps: number
    currency: string
}

export interface OHLCVRow {
    date: string
    open: number
    high: number
    low: number
    close: number
    volume: number
}

export interface Indicators {
    sma20: number
    sma50: number
    sma200: number | null
    ema20: number
    ema50: number | null
    adx: number | null
    di_plus: number | null
    di_minus: number | null
    rsi: number | null
    macd: number | null
    macd_signal: number | null
    macd_histogram: number | null
    stoch_k: number | null
    stoch_d: number | null
    bb_upper: number | null
    bb_middle: number | null
    bb_lower: number | null
    bb_width: number | null
    atr: number | null
    obv: number | null
    volume_ma20: number | null
    volume_spike: boolean
    support: number
    resistance: number
    fib_23_6: number | null
    fib_38_2: number | null
    fib_50_0: number | null
    fib_61_8: number | null
}

export interface SignalBreakdown {
    indicator: string
    signal: string
    note: string
    score: number
}

export interface SignalResult {
    overall: string
    score: number
    breakdown: SignalBreakdown[]
    ticker: string
}

export interface TPTarget {
    level: number
    price: number
    rr_ratio: number
}

export interface TradingPlan {
    bias: string
    entry_zone: number | null
    entry_price: number | null
    stop_loss: number | null
    targets: TPTarget[]
    suggested_position_size_pct: number
    invalidation_note: string
    disclaimer: string
}

export interface AnalisisResult {
    snapshot: Snapshot
    ohlcv: OHLCVRow[]
    indicators: Indicators
    signal: SignalResult
    trading_plan: TradingPlan
    chart_image: string
}

export class AnalisisService {
    private getBaseURL(): string {
        return (typeof globalThis !== 'undefined' && (globalThis as any).__API_URL__)
            || (typeof process !== 'undefined' && process.env.NUXT_PUBLIC_API_URL)
            || 'https://api-trackly.aksanara.id'
    }

    async searchTickers(): Promise<TickerSearchResult[]> {
        try {
            const baseURL = this.getBaseURL()
            const url = `${baseURL}/api/tickers`
            const response = await $fetch<{ success: boolean; data: TickerSearchResult[] }>(url, { method: 'GET' })
            return response.data || []
        } catch (error) {
            console.error('Failed to load tickers:', error)
            throw error
        }
    }

    async getTicker(kode: string): Promise<Snapshot> {
        try {
            const baseURL = this.getBaseURL()
            const url = `${baseURL}/api/ticker/${encodeURIComponent(kode)}`
            const response = await $fetch<{ success: boolean; data: Snapshot }>(url, { method: 'GET' })
            return response.data
        } catch (error) {
            console.error('Failed to get ticker:', error)
            throw error
        }
    }

    async postAnalisis(ticker: string, dateStart: string, dateEnd: string): Promise<AnalisisResult> {
        try {
            const baseURL = this.getBaseURL()
            const url = `${baseURL}/api/analisis`
            const response = await $fetch<{ success: boolean; data: AnalisisResult }>(url, {
                method: 'POST',
                body: { ticker, date_start: dateStart, date_end: dateEnd },
            })
            return response.data
        } catch (error) {
            console.error('Analisis failed:', error)
            throw error
        }
    }
}

export const analisisService = new AnalisisService()
