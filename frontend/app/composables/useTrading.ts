import { ref } from 'vue'
import { tradingService, type Transaction, type TransactionRequest, type Position, type PositionReviewResponse } from '~/services/trading.service'

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

  const loadAnalysis = async (ticker: string) => {
    positionLoading.value = true; error.value = null; positionAnalysis.value = null; aiInsight.value = null
    try { positionAnalysis.value = await tradingService.getPositionAnalysis(ticker) }
    catch (e: any) { error.value = e?.message || 'Gagal muat analisis posisi' }
    finally { positionLoading.value = false }
  }

  const requestAiInsight = async (ticker: string) : Promise<void> => {
    aiLoading.value = true; aiError.value = null; aiInsight.value = null
    const end = new Date().toISOString().split('T')[0]
    const start = new Date(Date.now() - 365 * 86400000).toISOString().split('T')[0]
    try {
      aiInsight.value = await tradingService.postAiInsight(ticker, start, end)
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

  return { transactions, positions, positionAnalysis, loading, positionLoading, error, aiInsight, aiLoading, aiError, add, load, loadPositions, loadAnalysis, requestAiInsight, updateTransaction, removeTransaction }
}