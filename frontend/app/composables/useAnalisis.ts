import { ref, computed } from 'vue'
import { analisisService, type TickerSearchResult, type AnalisisResult } from '~/services/analisis.service'

export function useTickerSearch() {
    const query = ref('')
    const allTickers = ref<TickerSearchResult[]>([])
    const loading = ref(false)
    const error = ref<string | null>(null)
    const loaded = ref(false)

    const fetchAll = async () => {
        if (loaded.value) return
        loading.value = true
        error.value = null
        try {
            allTickers.value = await analisisService.searchTickers()
            loaded.value = true
        } catch (e: any) {
            error.value = e?.message || 'Failed to load tickers'
        } finally {
            loading.value = false
        }
    }

    const suggestions = computed(() => {
        if (!query.value.trim()) return []
        const q = query.value.toLowerCase()
        return allTickers.value.filter(t =>
            t.kode.toLowerCase().includes(q) ||
            t.nama_perusahaan.toLowerCase().includes(q)
        ).slice(0, 20)
    })

    const clear = () => {
        query.value = ''
    }

    return { query, suggestions, loading, error, fetchAll, clear }
}

export function useAnalisis() {
    const result = ref<AnalisisResult | null>(null)
    const loading = ref(false)
    const error = ref<string | null>(null)

    const run = async (ticker: string, dateStart: string, dateEnd: string) => {
        loading.value = true
        error.value = null
        result.value = null
        try {
            result.value = await analisisService.postAnalisis(ticker, dateStart, dateEnd)
        } catch (e: any) {
            error.value = e?.message || 'Analisis failed'
        } finally {
            loading.value = false
        }
    }

    const clear = () => {
        result.value = null
        error.value = null
    }

    return { result, loading, error, run, clear }
}
