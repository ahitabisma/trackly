<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useTrading } from '~/composables/useTrading'

definePageMeta({ layout: 'main', middleware: 'auth' })

const { positions, positionAnalysis, loading, positionLoading, error, aiInsight, aiLoading, aiError, loadPositions, loadAnalysis, requestAiInsight } = useTrading()

const formatNum = (v: number) => new Intl.NumberFormat('id-ID', { minimumFractionDigits: 0, maximumFractionDigits: 2 }).format(v)

const selectPosition = async (ticker: string) => {
  await loadAnalysis(ticker)
}

const pnl = computed(() => {
  if (!positionAnalysis.value) return null
  const review = positionAnalysis.value.position_review
  if (!review || review.error) return null
  return {
    unrealized_pnl: review.unrealized_pnl,
    unrealized_pnl_pct: review.unrealized_pnl_pct,
    holding_days: review.holding_days,
    recommendation: review.recommendation,
    suggested_exit_price: review.suggested_exit_price,
    suggested_stop_price: review.suggested_stop_price,
    reason: review.reason,
  }
})

onMounted(() => { loadPositions() })
</script>

<template>
  <main class="min-h-screen bg-cream text-ink font-sans antialiased overflow-x-hidden pt-20 pb-24">
    <div class="max-w-5xl mx-auto px-5 sm:px-8">
      <div class="fu mb-8">
        <h1 class="font-serif text-4xl text-ink" style="letter-spacing:-1px">Posisi Terbuka</h1>
        <p class="font-mono text-xs text-muted mt-1 uppercase tracking-wider">Analisis posisi &amp; review</p>
      </div>

      <div v-if="loading && !positions.length" class="text-center py-20">
        <div class="inline-block w-6 h-6 border-2 border-ink border-t-transparent rounded-full animate-spin mb-2"></div>
        <p class="font-mono text-xs text-muted">Memuat...</p>
      </div>

      <div v-else-if="!positions.length" class="text-center py-20">
        <p class="font-mono text-sm text-muted">Tidak ada posisi terbuka</p>
      </div>

      <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-8">
        <div v-for="p in positions" :key="p.ticker"
          class="border-2 border-ink rounded-xl p-4 cursor-pointer transition-all hover:translate-y-[-2px] hover:shadow-[4px_4px_0_#1a1612] relative"
          :class="[positionAnalysis?.ticker === p.ticker ? 'bg-bluebg border-bluedk shadow-[4px_4px_0_#2d7ab5]' : 'bg-card', positionLoading && positionAnalysis?.ticker !== p.ticker ? 'opacity-50 pointer-events-none' : '', positionLoading && positionAnalysis?.ticker === p.ticker ? 'animate-pulse border-ink/50' : '']"
          @click="selectPosition(p.ticker)">
          <div class="flex items-center justify-between mb-2">
            <span class="font-mono font-bold text-lg">{{ p.ticker }}</span>
            <span class="font-mono text-xs px-2 py-0.5 rounded-md bg-green-100 text-green-800 border border-green-300">OPEN</span>
          </div>
          <div class="grid grid-cols-2 gap-2 font-mono text-xs">
            <div><span class="text-muted">Lot:</span> {{ formatNum(p.total_lot) }}</div>
            <div><span class="text-muted">Avg Price:</span> {{ formatNum(p.average_buy_price) }}</div>
          </div>
          <div v-if="positionLoading && positionAnalysis?.ticker === p.ticker" class="absolute inset-0 flex items-center justify-center bg-black/10 rounded-xl backdrop-blur-[1px]">
            <div class="flex items-center gap-2 bg-card px-4 py-2 rounded-lg border-2 border-ink shadow-[2px_2px_0_#1a1612]">
              <div class="w-4 h-4 border-2 border-ink border-t-transparent rounded-full animate-spin"></div>
              <span class="font-mono text-xs font-semibold">Menganalisis...</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Position Analysis -->
      <div v-if="positionAnalysis" class="fu3 space-y-4">
        <div class="bg-card border-2 border-ink rounded-2xl p-5 sm:p-6" style="box-shadow:4px 4px 0 #1a1612">
          <div class="flex items-center justify-between mb-4">
            <h2 class="font-mono text-xs text-muted uppercase tracking-wider">Review Posisi {{ positionAnalysis.ticker }}</h2>
            <button class="font-mono text-xs text-muted hover:text-ink underline" @click="positionAnalysis = null">Tutup</button>
          </div>

          <div v-if="positionAnalysis.position_review?.error" class="font-mono text-xs text-red-600">
            {{ positionAnalysis.position_review.error }}
          </div>

          <div v-else-if="pnl" class="space-y-4">
            <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
              <div>
                <div class="font-mono text-xs text-muted uppercase">Unrealized P&L</div>
                <div class="font-mono text-lg font-bold mt-0.5" :class="pnl.unrealized_pnl >= 0 ? 'text-green-700' : 'text-red-700'">
                  {{ formatNum(pnl.unrealized_pnl) }}
                </div>
              </div>
              <div>
                <div class="font-mono text-xs text-muted uppercase">P&L %</div>
                <div class="font-mono text-lg font-bold mt-0.5" :class="pnl.unrealized_pnl_pct >= 0 ? 'text-green-700' : 'text-red-700'">
                  {{ pnl.unrealized_pnl_pct?.toFixed(2) }}%
                </div>
              </div>
              <div>
                <div class="font-mono text-xs text-muted uppercase">Holding</div>
                <div class="font-mono text-lg font-bold mt-0.5">{{ pnl.holding_days }} hari</div>
              </div>
              <div>
                <div class="font-mono text-xs text-muted uppercase">Rekomendasi</div>
                <div class="font-mono text-lg font-bold mt-0.5 uppercase"
                  :class="pnl.recommendation === 'sell' ? 'text-red-700' : 'text-green-700'">
                  {{ pnl.recommendation }}
                </div>
              </div>
            </div>

            <div v-if="pnl.suggested_exit_price || pnl.suggested_stop_price" class="grid grid-cols-2 gap-4">
              <div v-if="pnl.suggested_exit_price">
                <div class="font-mono text-xs text-muted uppercase">Suggested Exit</div>
                <div class="font-mono text-sm font-bold mt-0.5">{{ formatNum(pnl.suggested_exit_price) }}</div>
              </div>
              <div v-if="pnl.suggested_stop_price">
                <div class="font-mono text-xs text-muted uppercase">Suggested Stop</div>
                <div class="font-mono text-sm font-bold mt-0.5 text-red-600">{{ formatNum(pnl.suggested_stop_price) }}</div>
              </div>
            </div>

            <div v-if="pnl.reason" class="p-3 bg-yellow-50 border border-yellow-200 rounded-lg">
              <div class="font-mono text-xs text-muted uppercase tracking-wider mb-1">Alasan</div>
              <p class="font-mono text-xs text-yellow-800">{{ pnl.reason }}</p>
            </div>
          </div>
        </div>

        <div v-if="positionAnalysis.signal" class="bg-card border-2 border-ink rounded-2xl p-5 sm:p-6" style="box-shadow:4px 4px 0 #1a1612">
          <h2 class="font-mono text-xs text-muted uppercase tracking-wider mb-3">Sinyal Teknikal</h2>
          <div class="flex items-center gap-3">
            <span class="font-mono text-xs font-bold px-3 py-1 rounded-md border-2"
              :class="positionAnalysis.signal?.overall === 'bullish' ? 'bg-green-100 text-green-800 border-green-300' : positionAnalysis.signal?.overall === 'bearish' ? 'bg-red-100 text-red-800 border-red-300' : 'bg-gray-100 text-gray-600 border-gray-300'">
              {{ positionAnalysis.signal?.overall?.toUpperCase() || 'NETRAL' }}
            </span>
            <span class="font-mono text-xs text-muted">Skor: {{ positionAnalysis.signal?.score?.toFixed(2) }}</span>
            <span v-if="positionAnalysis.signal?.confidence" class="font-mono text-xs px-2 py-0.5 rounded border"
              :class="positionAnalysis.signal.confidence === 'high' ? 'bg-green-50 text-green-700 border-green-200' : positionAnalysis.signal.confidence === 'medium' ? 'bg-yellow-50 text-yellow-700 border-yellow-200' : 'bg-gray-50 text-gray-500 border-gray-200'">
              {{ positionAnalysis.signal.confidence.toUpperCase() }}
            </span>
          </div>
        </div>

        <!-- AI Insight -->
        <div class="bg-card border-2 border-ink rounded-2xl p-5 sm:p-6" style="box-shadow:4px 4px 0 #1a1612">
          <button v-if="!aiInsight && !aiLoading" class="neo-btn-primary max-w-xs" @click="requestAiInsight(positionAnalysis.ticker)" :disabled="aiLoading">
            <span v-if="aiLoading" class="inline-block w-4 h-4 border-2 border-cream border-t-transparent rounded-full animate-spin mr-2"></span>
            {{ aiLoading ? 'Memproses...' : 'Analisis AI' }}
          </button>

          <div v-if="aiLoading" class="mt-3 flex items-center gap-3">
            <div class="w-5 h-5 border-2 border-ink border-t-transparent rounded-full animate-spin"></div>
            <span class="font-mono text-sm text-muted">AI sedang menganalisis...</span>
          </div>

          <div v-if="aiInsight" class="mt-3">
            <div class="flex items-center justify-between mb-3">
              <h2 class="font-mono text-xs text-muted uppercase tracking-wider">Analisis AI</h2>
              <button class="font-mono text-xs text-muted hover:text-ink underline" @click="aiInsight = null">Tutup</button>
            </div>
            <p class="font-sans text-sm text-ink leading-relaxed">{{ aiInsight }}</p>
          </div>

          <div v-if="aiError" class="mt-3 font-mono text-xs text-red-600">{{ aiError }}</div>
        </div>
      </div>
    </div>
  </main>
</template>