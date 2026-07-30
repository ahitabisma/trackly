import { ref } from 'vue'
import { tradingService, type Transaction, type TransactionRequest, type Position, type PositionReviewResponse } from '~/services/trading.service'

const CACHE_PREFIX = 'trackly_analysis_'
const TS_KEY = 'trackly_analysis_ts'

function loadCached<T>(key: string): T | null {
  try {
    const raw = localStorage.getItem(key)
    return raw ? JSON.parse(raw) : null
  } catch { return null }
}

function saveCache<T>(key: string, data: T) {
  try { localStorage.setItem(key, JSON.stringify(data)) } catch { }
}

function removeCache(key: string) {
  try { localStorage.removeItem(key) } catch { }
}

const _timestamps = ref<Record<string, string>>(loadCached<Record<string, string>>(TS_KEY) || {})

function _persistTimestamps() {
  saveCache(TS_KEY, _timestamps.value)
}

export function useTrading() {
  const transactions = ref<Transaction[]>([])
  const positions = ref<Position[]>([])
  const positionAnalysis = ref<PositionReviewResponse | null>(null)
  const loading = ref(false)
  const positionLoading = ref(false)
  const error = ref<string | null>(null)
  const aiInsight = ref<string | null>(null)
  const aiLoading = ref(false)
  const aiError = ref<string | null>(null)

  const add = async (req: TransactionRequest) => {
    loading.value = true; error.value = null
    try {
      const t = await tradingService.createTransaction(req)
      transactions.value.unshift(t)
      return t
    } catch (e: any) { error.value = e?.message || 'Gagal simpan transaksi'; throw e }
    finally { loading.value = false }
  }

  const load = async (params?: { ticker?: string; dateStart?: string; dateEnd?: string; type?: string }) => {
    loading.value = true; error.value = null
    try { transactions.value = await tradingService.getTransactions(params) }
    catch (e: any) { error.value = e?.message || 'Gagal muat transaksi' }
    finally { loading.value = false }
  }

  const loadPositions = async () => {
    loading.value = true; error.value = null
    try { positions.value = await tradingService.getPositions() }
    catch (e: any) { error.value = e?.message || 'Gagal muat posisi' }
    finally { loading.value = false }
  }

  const loadAnalysis = async (ticker: string, forceRefresh = false) => {
    positionLoading.value = true; error.value = null; positionAnalysis.value = null; aiInsight.value = null
    try {
      if (!forceRefresh) {
        const cached = loadCached<PositionReviewResponse>(CACHE_PREFIX + ticker)
        if (cached) {
          positionAnalysis.value = cached
          positionLoading.value = false
          return
        }
      }
      const data = await tradingService.getPositionAnalysis(ticker)
      positionAnalysis.value = data
      saveCache(CACHE_PREFIX + ticker, data)
      _timestamps.value[ticker] = new Date().toISOString()
      _persistTimestamps()
    } catch (e: any) { error.value = e?.message || 'Gagal muat analisis posisi' }
    finally { positionLoading.value = false }
  }

  const refreshAnalysis = async (ticker: string) => {
    removeCache(CACHE_PREFIX + ticker)
    delete _timestamps.value[ticker]
    _persistTimestamps()
    await loadAnalysis(ticker, true)
  }

  const getAnalysisTimestamp = (ticker: string): string => {
    return _timestamps.value[ticker] || ''
  }

  const hasCachedAnalysis = (ticker: string): boolean => {
    return !!_timestamps.value[ticker]
  }

  const requestAiInsight = async (ticker: string): Promise<void> => {
    aiLoading.value = true; aiError.value = null; aiInsight.value = null
    const pa = positionAnalysis.value
    if (!pa?.indicators) {
      aiError.value = 'Data analisis belum tersedia'
      aiLoading.value = false
      return
    }
    try {
      aiInsight.value = await tradingService.postAiInsight(ticker,
        new Date().toISOString().split('T')[0],
        pa.indicators, pa.snapshot || null,
        pa.position, pa.position_review, pa.signal,
      )
    } catch (e: any) { aiError.value = e?.message || 'AI insight failed' }
    finally { aiLoading.value = false }
  }

  const updateTransaction = async (id: number, req: any) => {
    error.value = null
    try {
      const t = await tradingService.updateTransaction(id, req)
      const idx = transactions.value.findIndex(x => x.id === id)
      if (idx >= 0) transactions.value[idx] = t
      return t
    } catch (e: any) { error.value = e?.message || 'Gagal update'; throw e }
  }

  const removeTransaction = async (id: number) => {
    error.value = null
    try {
      await tradingService.deleteTransaction(id)
      transactions.value = transactions.value.filter(x => x.id !== id)
    } catch (e: any) { error.value = e?.message || 'Gagal hapus'; throw e }
  }

  return { transactions, positions, positionAnalysis, loading, positionLoading, error, aiInsight, aiLoading, aiError, add, load, loadPositions, loadAnalysis, refreshAnalysis, getAnalysisTimestamp, hasCachedAnalysis, requestAiInsight, updateTransaction, removeTransaction }
}