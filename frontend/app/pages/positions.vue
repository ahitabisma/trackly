<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useTrading } from "~/composables/useTrading";

definePageMeta({ layout: "main", middleware: "auth" });

const {
  positions,
  positionAnalysis,
  loading,
  positionLoading,
  error,
  aiInsight,
  aiLoading,
  aiError,
  loadPositions,
  loadAnalysis,
  refreshAnalysis,
  getAnalysisTimestamp,
  hasCachedAnalysis,
  requestAiInsight,
} = useTrading();

const activeTab = ref<"review" | "indicators" | "snapshot" | "ai">("review");
const signalExpanded = ref(false);

const formatNum = (v: number | null | undefined) =>
  v === null || v === undefined ? "N/A"
    : new Intl.NumberFormat("id-ID", {
        minimumFractionDigits: 0,
        maximumFractionDigits: 2,
      }).format(v);

const formatCompact = (v: number | null | undefined) => {
  if (!v) return "0";
  if (v >= 1e12) return (v / 1e12).toFixed(2) + " T";
  if (v >= 1e9) return (v / 1e9).toFixed(2) + " M";
  if (v >= 1e6) return (v / 1e6).toFixed(2) + " Jt";
  return v.toLocaleString("id-ID");
};

const NA = (v: any) =>
  v === null || v === undefined ? "N/A"
    : typeof v === "number" ? v.toLocaleString("id-ID") : v;

const NAFmt = (v: any) => {
  if (v === null || v === undefined) return "N/A";
  if (typeof v === "number" && v === 0) return "N/A";
  if (typeof v === "boolean") return v ? "Ya" : "Tidak";
  return v;
};

const formatMd = (s: string) =>
  s.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
   .replace(/\n\n/g, '</p><p>')
   .replace(/\n/g, '<br>');

const formatDate = (s: string) => {
  if (!s) return ''
  const d = new Date(s)
  return d.toLocaleDateString('id-ID', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })
}

const selectPosition = async (ticker: string) => {
  activeTab.value = "review";
  signalExpanded.value = false;
  await loadAnalysis(ticker);
};

const pnl = computed(() => {
  if (!positionAnalysis.value) return null;
  const review = positionAnalysis.value.position_review;
  if (!review || review.error) return null;
  return {
    unrealized_pnl: review.unrealized_pnl,
    unrealized_pnl_pct: review.unrealized_pnl_pct,
    holding_days: review.holding_days,
    recommendation: review.recommendation,
    suggested_exit_price: review.suggested_exit_price,
    suggested_stop: review.suggested_stop,
    reason: review.reason,
  };
});

const tradingPlan = computed(() =>
  positionAnalysis.value?.position_review?.trading_plan || null
);

const snapshot = computed(() => positionAnalysis.value?.snapshot || null);

const currentPrice = computed(() =>
  positionAnalysis.value?.position_review?.current_price || 0
);

const signalBadge = computed(() => {
  const o = positionAnalysis.value?.signal?.overall;
  if (o === "bullish")
    return { label: "BULLISH", class: "bg-green-100 text-green-800 border-green-300" };
  if (o === "bearish")
    return { label: "BEARISH", class: "bg-red-100 text-red-800 border-red-300" };
  return { label: "NETRAL", class: "bg-gray-100 text-gray-600 border-gray-300" };
});

const handleAiTab = () => {
  activeTab.value = "ai";
  if (!aiInsight.value && !aiLoading.value && positionAnalysis.value?.ticker) {
    requestAiInsight(positionAnalysis.value.ticker);
  }
};

onMounted(() => { loadPositions(); });
</script>

<template>
  <main
    class="min-h-screen bg-cream text-ink font-sans antialiased overflow-x-hidden pt-20 pb-24"
  >
    <div class="max-w-5xl mx-auto px-5 sm:px-8">
      <div class="fu mb-8">
        <h1 class="font-serif text-4xl text-ink" style="letter-spacing: -1px">
          Posisi Terbuka
        </h1>
        <p class="font-mono text-xs text-muted mt-1 uppercase tracking-wider">
          Analisis posisi &amp; review
        </p>
      </div>

      <div v-if="loading && !positions.length" class="text-center py-20">
        <div
          class="inline-block w-6 h-6 border-2 border-ink border-t-transparent rounded-full animate-spin mb-2"
        ></div>
        <p class="font-mono text-xs text-muted">Memuat...</p>
      </div>

      <div v-else-if="!positions.length" class="text-center py-20">
        <p class="font-mono text-sm text-muted">Tidak ada posisi terbuka</p>
      </div>

      <div v-else class="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-8">
        <div
          v-for="p in positions"
          :key="p.ticker"
          class="border-2 border-ink rounded-xl p-4 cursor-pointer transition-all hover:translate-y-[-2px] hover:shadow-[4px_4px_0_#1a1612] relative"
          :class="[
            positionAnalysis?.ticker === p.ticker
              ? 'bg-bluebg border-bluedk shadow-[4px_4px_0_#2d7ab5]'
              : 'bg-card',
            positionLoading && positionAnalysis?.ticker !== p.ticker
              ? 'opacity-50 pointer-events-none'
              : '',
            positionLoading && positionAnalysis?.ticker === p.ticker
              ? 'animate-pulse border-ink/50'
              : '',
          ]"
          @click="selectPosition(p.ticker)"
        >
          <div class="flex items-center justify-between mb-2">
            <span class="font-mono font-bold text-lg">{{ p.ticker }}</span>
            <span
              class="font-mono text-xs px-2 py-0.5 rounded-md bg-green-100 text-green-800 border border-green-300"
            >OPEN</span>
          </div>
          <div class="grid grid-cols-2 gap-2 font-mono text-xs">
            <div>
              <span class="text-muted">Lot:</span> {{ formatNum(p.total_lot) }}
            </div>
            <div>
              <span class="text-muted">Avg Price:</span>
              {{ formatNum(p.average_buy_price) }}
            </div>
          </div>
          <div
            v-if="positionLoading && positionAnalysis?.ticker === p.ticker"
            class="absolute inset-0 flex items-center justify-center bg-black/10 rounded-xl backdrop-blur-[1px]"
          >
            <div
              class="flex items-center gap-2 bg-card px-4 py-2 rounded-lg border-2 border-ink shadow-[2px_2px_0_#1a1612]"
            >
              <div
                class="w-4 h-4 border-2 border-ink border-t-transparent rounded-full animate-spin"
              ></div>
              <span class="font-mono text-xs font-semibold">Menganalisis...</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Analysis Section -->
      <div v-if="positionAnalysis" class="space-y-4">
        <!-- Tab Bar -->
        <div class="flex gap-2 bg-card border-2 border-ink rounded-2xl p-1.5" style="box-shadow: 4px 4px 0 #1a1612">
          <button
            v-for="tab in [
              { key: 'review', label: 'Review' },
              { key: 'indicators', label: 'Indikator' },
              { key: 'snapshot', label: 'Snapshot' },
              { key: 'ai', label: 'AI Analisis' },
            ]"
            :key="tab.key"
            class="flex-1 font-mono text-xs font-bold py-2.5 px-3 rounded-xl border-2 transition-all uppercase tracking-wider"
            :class="activeTab === tab.key
              ? 'bg-ink text-cream border-ink shadow-[2px_2px_0_#1a1612]'
              : 'bg-card text-ink border-transparent hover:bg-gray-100'"
            @click="tab.key === 'ai' ? handleAiTab() : (activeTab = tab.key as any)"
          >{{ tab.label }}</button>
        </div>

        <!-- Tab: Position Review -->
        <div v-show="activeTab === 'review'" class="fu space-y-4">
          <!-- Position Review Card -->
          <div
            class="bg-card border-2 border-ink rounded-2xl p-5 sm:p-6"
            style="box-shadow: 4px 4px 0 #1a1612"
          >
            <div class="flex items-center justify-between mb-4">
              <h2 class="font-mono text-xs text-muted uppercase tracking-wider">
                Review Posisi {{ positionAnalysis.ticker }}
              </h2>
              <div class="flex items-center gap-2">
                <span
                  v-if="getAnalysisTimestamp(positionAnalysis.ticker)"
                  class="font-mono text-[10px] text-muted hidden sm:inline"
                >{{ formatDate(getAnalysisTimestamp(positionAnalysis.ticker)) }}</span>
                <button
                  v-if="!positionLoading"
                  class="font-mono text-xs text-muted hover:text-ink underline"
                  @click="refreshAnalysis(positionAnalysis.ticker)"
                >Analisis Ulang</button>
                <button
                  class="font-mono text-xs text-muted hover:text-ink underline"
                  @click="positionAnalysis = null; aiInsight = null"
                >Tutup</button>
              </div>
            </div>

            <div
              v-if="positionAnalysis.position_review?.error"
              class="font-mono text-xs text-red-600"
            >{{ positionAnalysis.position_review.error }}</div>

            <div v-else-if="pnl" class="space-y-4">
              <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
                <div>
                  <div class="font-mono text-xs text-muted uppercase">Unrealized P&L</div>
                  <div
                    class="font-mono text-lg font-bold mt-0.5"
                    :class="pnl.unrealized_pnl >= 0 ? 'text-green-700' : 'text-red-700'"
                  >{{ formatNum(pnl.unrealized_pnl) }}</div>
                </div>
                <div>
                  <div class="font-mono text-xs text-muted uppercase">P&L %</div>
                  <div
                    class="font-mono text-lg font-bold mt-0.5"
                    :class="pnl.unrealized_pnl_pct >= 0 ? 'text-green-700' : 'text-red-700'"
                  >{{ pnl.unrealized_pnl_pct?.toFixed(2) }}%</div>
                </div>
                <div>
                  <div class="font-mono text-xs text-muted uppercase">Holding</div>
                  <div class="font-mono text-lg font-bold mt-0.5">{{ pnl.holding_days }} hari</div>
                </div>
                <div>
                  <div class="font-mono text-xs text-muted uppercase">Rekomendasi</div>
                  <div
                    class="font-mono text-lg font-bold mt-0.5 uppercase"
                    :class="pnl.recommendation === 'sell' ? 'text-red-700' : 'text-green-700'"
                  >{{ pnl.recommendation || '—' }}</div>
                </div>
              </div>

              <div
                v-if="pnl.suggested_exit_price || pnl.suggested_stop"
                class="grid grid-cols-2 gap-4"
              >
                <div v-if="pnl.suggested_exit_price">
                  <div class="font-mono text-xs text-muted uppercase">Suggested Exit</div>
                  <div class="font-mono text-sm font-bold mt-0.5">{{ formatNum(pnl.suggested_exit_price) }}</div>
                </div>
                <div v-if="pnl.suggested_stop">
                  <div class="font-mono text-xs text-muted uppercase">Suggested Stop</div>
                  <div class="font-mono text-sm font-bold mt-0.5 text-red-600">{{ formatNum(pnl.suggested_stop) }}</div>
                </div>
              </div>

              <div
                v-if="pnl.reason"
                class="p-3 bg-yellow-50 border border-yellow-200 rounded-lg"
              >
                <div class="font-mono text-xs text-muted uppercase tracking-wider mb-1">Alasan</div>
                <p class="font-mono text-xs text-yellow-800">{{ pnl.reason }}</p>
              </div>
            </div>
          </div>

          <!-- Trading Plan -->
          <div
            v-if="tradingPlan"
            class="bg-card border-2 border-ink rounded-2xl p-5 sm:p-6"
            style="box-shadow: 4px 4px 0 #1a1612"
          >
            <h2 class="font-mono text-xs text-muted uppercase tracking-wider mb-4">Trading Plan</h2>

            <div class="flex flex-wrap gap-2 mb-3">
              <span
                v-if="tradingPlan.time_stop_days"
                class="font-mono text-[10px] bg-purple-50 text-purple-700 border border-purple-200 px-2 py-0.5 rounded-md"
              >Time Stop: {{ tradingPlan.time_stop_days }} hari</span>
              <span
                v-if="tradingPlan.avg_price"
                class="font-mono text-[10px] bg-orange-50 text-orange-700 border border-orange-200 px-2 py-0.5 rounded-md"
              >Avg: {{ formatNum(tradingPlan.avg_price) }}</span>
              <span
                v-if="tradingPlan.current_vs_avg_pct != null"
                class="font-mono text-[10px] px-2 py-0.5 rounded-md border"
                :class="tradingPlan.current_vs_avg_pct >= 0 ? 'bg-green-50 text-green-700 border-green-200' : 'bg-red-50 text-red-700 border-red-200'"
              >{{ tradingPlan.current_vs_avg_pct >= 0 ? '+' : '' }}{{ tradingPlan.current_vs_avg_pct?.toFixed(2) }}%</span>
            </div>

            <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-2">
              <div>
                <div class="font-mono text-xs text-muted uppercase">Bias</div>
                <div
                  class="font-mono text-lg font-bold mt-0.5 uppercase"
                  :class="tradingPlan.bias === 'buy' ? 'text-green-700' : tradingPlan.bias === 'avoid' ? 'text-red-700' : 'text-gray-500'"
                >{{ tradingPlan.bias || '—' }}</div>
              </div>
              <div>
                <div class="font-mono text-xs text-muted uppercase">Entry Price</div>
                <div class="font-mono text-lg font-bold mt-0.5">
                  {{ tradingPlan.entry_price ? formatNum(tradingPlan.entry_price) : 'N/A' }}
                </div>
              </div>
              <div v-if="tradingPlan.entry_zone">
                <div class="font-mono text-xs text-muted uppercase">Entry Zone</div>
                <div class="font-mono text-lg font-bold mt-0.5">
                  {{ formatNum(tradingPlan.entry_zone.low) }} — {{ formatNum(tradingPlan.entry_zone.high) }}
                </div>
              </div>
              <div>
                <div class="font-mono text-xs text-muted uppercase">Stop Loss</div>
                <div class="font-mono text-lg font-bold mt-0.5 text-red-600">
                  {{ tradingPlan.stop_loss ? formatNum(tradingPlan.stop_loss) : 'N/A' }}
                </div>
              </div>
            </div>

            <div
              v-if="tradingPlan.entry_note"
              class="mb-4 p-3 bg-blue-50 border border-blue-200 rounded-lg"
            >
              <div class="font-mono text-xs text-muted uppercase tracking-wider mb-1">Catatan Entry</div>
              <p class="font-mono text-xs text-blue-800">{{ tradingPlan.entry_note }}</p>
            </div>

            <!-- TP Targets -->
            <div v-if="tradingPlan.targets?.length" class="mb-4">
              <div class="font-mono text-xs text-muted uppercase tracking-wider mb-2">Target Take Profit</div>
              <table class="w-full text-left font-mono text-xs data-table border-2 border-bdr rounded-xl overflow-hidden">
                <thead>
                  <tr class="bg-gray-100 text-muted uppercase tracking-wider">
                    <th class="py-2 px-3">Level</th>
                    <th class="py-2 px-3">Harga</th>
                    <th class="py-2 px-3">R:R Ratio</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="t in tradingPlan.targets"
                    :key="t.level"
                    class="border-t border-bdr/50"
                    :class="t.level === 0 ? 'bg-orange-50/50' : ''"
                  >
                    <td class="py-2 px-3 font-bold">{{ t.level === 0 ? 'BE' : 'TP' + t.level }}</td>
                    <td class="py-2 px-3">{{ formatNum(t.price) }}</td>
                    <td class="py-2 px-3 font-bold" :class="t.level === 0 ? 'text-muted' : 'text-green-700'">{{ t.rr_ratio?.toFixed(2) }}R</td>
                  </tr>
                </tbody>
              </table>
            </div>

            <!-- Invalidation -->
            <div
              v-if="tradingPlan.invalidation_note"
              class="mb-4 p-3 bg-yellow-50 border border-yellow-200 rounded-lg"
            >
              <div class="font-mono text-xs text-muted uppercase tracking-wider mb-1">Invalidation</div>
              <p class="font-mono text-xs text-yellow-800">{{ tradingPlan.invalidation_note }}</p>
            </div>

            <!-- Disclaimer -->
            <div class="p-3 bg-gray-100 border border-gray-200 rounded-lg">
              <div class="font-mono text-xs text-muted uppercase tracking-wider mb-1">Disclaimer</div>
              <p class="font-mono text-xs text-gray-600 leading-relaxed">{{ tradingPlan.disclaimer || positionAnalysis.position_review?.disclaimer }}</p>
            </div>
          </div>

          <!-- Signal -->
          <div
            v-if="positionAnalysis.signal"
            class="bg-card border-2 border-ink rounded-2xl p-5 sm:p-6"
            style="box-shadow: 4px 4px 0 #1a1612"
          >
            <h2 class="font-mono text-xs text-muted uppercase tracking-wider mb-3">Sinyal Teknikal</h2>
            <div class="flex items-center gap-3 mb-3">
              <span
                class="font-mono text-xs font-bold px-3 py-1 rounded-md border-2"
                :class="signalBadge.class"
              >{{ signalBadge.label }}</span>
              <span class="font-mono text-xs text-muted">Skor: {{ positionAnalysis.signal.score?.toFixed(2) }}</span>
              <span
                v-if="positionAnalysis.signal.confidence"
                class="font-mono text-xs px-2 py-0.5 rounded border"
                :class="positionAnalysis.signal.confidence === 'high' ? 'bg-green-50 text-green-700 border-green-200' : positionAnalysis.signal.confidence === 'medium' ? 'bg-yellow-50 text-yellow-700 border-yellow-200' : 'bg-gray-50 text-gray-500 border-gray-200'"
              >{{ positionAnalysis.signal.confidence.toUpperCase() }}</span>
            </div>

            <details class="group" v-if="positionAnalysis.signal.breakdown?.length">
              <summary class="font-mono text-xs text-muted cursor-pointer hover:text-ink select-none list-none flex items-center gap-2">
                <span class="text-xs transition-transform group-open:rotate-90">▶</span>
                Detail sinyal ({{ positionAnalysis.signal.breakdown.length }} indikator)
              </summary>
              <div class="mt-3 space-y-2">
                <div
                  v-for="b in positionAnalysis.signal.breakdown"
                  :key="b.indicator"
                  class="flex items-center justify-between p-2 rounded-lg border border-bdr/50"
                >
                  <div>
                    <div class="font-mono text-xs font-semibold">{{ b.indicator }}</div>
                    <div class="font-mono text-[10px] text-muted">{{ b.note }}</div>
                  </div>
                  <span
                    class="font-mono text-[10px] font-bold px-2 py-0.5 rounded uppercase"
                    :class="b.signal === 'bullish' ? 'bg-green-50 text-green-700' : b.signal === 'bearish' ? 'bg-red-50 text-red-700' : 'bg-gray-50 text-gray-500'"
                  >{{ b.signal }}</span>
                </div>
              </div>
            </details>
          </div>
        </div>

        <!-- Tab: Indicators -->
        <div v-show="activeTab === 'indicators'" class="fu">
          <div
            v-if="positionAnalysis.indicators"
            class="bg-card border-2 border-ink rounded-2xl p-5 sm:p-6"
            style="box-shadow: 4px 4px 0 #1a1612"
          >
            <h2 class="font-mono text-xs text-muted uppercase tracking-wider mb-4">Indikator Teknikal</h2>
            <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-6 gap-4">
              <div>
                <div class="font-mono text-xs text-muted uppercase">Harga</div>
                <div class="font-mono text-lg font-bold mt-0.5">{{ formatNum(currentPrice) }}</div>
              </div>
              <div>
                <div class="font-mono text-xs text-muted uppercase">Avg Buy</div>
                <div class="font-mono text-lg font-bold mt-0.5">{{ formatNum(positionAnalysis.position.average_buy_price) }}</div>
              </div>
              <div>
                <div class="font-mono text-xs text-muted uppercase">SMA20</div>
                <div class="font-mono text-lg font-bold mt-0.5">{{ formatNum(positionAnalysis.indicators.sma20) }}</div>
              </div>
              <div>
                <div class="font-mono text-xs text-muted uppercase">SMA50</div>
                <div class="font-mono text-lg font-bold mt-0.5">{{ formatNum(positionAnalysis.indicators.sma50) }}</div>
              </div>
              <div>
                <div class="font-mono text-xs text-muted uppercase">SMA200</div>
                <div class="font-mono text-lg font-bold mt-0.5">{{ NAFmt(positionAnalysis.indicators.sma200) }}</div>
              </div>
              <div>
                <div class="font-mono text-xs text-muted uppercase">EMA20</div>
                <div class="font-mono text-lg font-bold mt-0.5">{{ formatNum(positionAnalysis.indicators.ema20) }}</div>
              </div>
              <div>
                <div class="font-mono text-xs text-muted uppercase">RSI</div>
                <div class="font-mono text-lg font-bold mt-0.5">{{ NAFmt(positionAnalysis.indicators.rsi?.toFixed(2)) }}</div>
              </div>
              <div>
                <div class="font-mono text-xs text-muted uppercase">MACD</div>
                <div class="font-mono text-lg font-bold mt-0.5">{{ NAFmt(positionAnalysis.indicators.macd?.toFixed(2)) }}</div>
              </div>
              <div>
                <div class="font-mono text-xs text-muted uppercase">ADX</div>
                <div class="font-mono text-lg font-bold mt-0.5">{{ NAFmt(positionAnalysis.indicators.adx?.toFixed(2)) }}</div>
              </div>
              <div>
                <div class="font-mono text-xs text-muted uppercase">ATR</div>
                <div class="font-mono text-lg font-bold mt-0.5">{{ NAFmt(positionAnalysis.indicators.atr) }}</div>
              </div>
              <div>
                <div class="font-mono text-xs text-muted uppercase">Support</div>
                <div class="font-mono text-lg font-bold mt-0.5">{{ formatNum(positionAnalysis.indicators.support) }}</div>
              </div>
              <div>
                <div class="font-mono text-xs text-muted uppercase">Resistance</div>
                <div class="font-mono text-lg font-bold mt-0.5">{{ formatNum(positionAnalysis.indicators.resistance) }}</div>
              </div>
              <div>
                <div class="font-mono text-xs text-muted uppercase">BB Width</div>
                <div class="font-mono text-lg font-bold mt-0.5">{{ NAFmt(positionAnalysis.indicators.bb_width?.toFixed(4)) }}</div>
              </div>
              <div>
                <div class="font-mono text-xs text-muted uppercase">Volume Spike</div>
                <div class="font-mono text-lg font-bold mt-0.5">{{ NAFmt(positionAnalysis.indicators.volume_spike) }}</div>
              </div>
            </div>
          </div>
        </div>

        <!-- Tab: Snapshot -->
        <div v-show="activeTab === 'snapshot'" class="fu">
          <div
            v-if="snapshot"
            class="bg-card border-2 border-ink rounded-2xl p-5 sm:p-6"
            style="box-shadow: 4px 4px 0 #1a1612"
          >
            <h2 class="font-mono text-xs text-muted uppercase tracking-wider mb-4">Snapshot Fundamental</h2>
            <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
              <div>
                <div class="font-mono text-xs text-muted uppercase">Harga</div>
                <div class="font-serif text-2xl mt-0.5">{{ formatNum(snapshot.price) }}</div>
              </div>
              <div>
                <div class="font-mono text-xs text-muted uppercase">52w High</div>
                <div class="font-serif text-2xl mt-0.5">{{ formatNum(snapshot.high52) }}</div>
              </div>
              <div>
                <div class="font-mono text-xs text-muted uppercase">52w Low</div>
                <div class="font-serif text-2xl mt-0.5">{{ formatNum(snapshot.low52) }}</div>
              </div>
              <div>
                <div class="font-mono text-xs text-muted uppercase">Volume</div>
                <div class="font-serif text-2xl mt-0.5">{{ formatCompact(snapshot.volume) }}</div>
              </div>
              <div>
                <div class="font-mono text-xs text-muted uppercase">Market Cap</div>
                <div class="font-serif text-2xl mt-0.5">{{ formatCompact(snapshot.marketcap) }}</div>
              </div>
              <div>
                <div class="font-mono text-xs text-muted uppercase">PE</div>
                <div class="font-serif text-2xl mt-0.5">{{ NA(snapshot.pe) }}</div>
              </div>
              <div>
                <div class="font-mono text-xs text-muted uppercase">EPS</div>
                <div class="font-serif text-2xl mt-0.5">{{ formatNum(snapshot.eps) }}</div>
              </div>
              <div>
                <div class="font-mono text-xs text-muted uppercase">Currency</div>
                <div class="font-serif text-2xl mt-0.5">{{ snapshot.currency || 'N/A' }}</div>
              </div>
            </div>
          </div>
          <div
            v-else
            class="bg-card border-2 border-ink rounded-2xl p-5 sm:p-6 text-center"
            style="box-shadow: 4px 4px 0 #1a1612"
          >
            <p class="font-mono text-sm text-muted">Data snapshot tidak tersedia</p>
          </div>
        </div>

        <!-- Tab: AI Analisis -->
        <div v-show="activeTab === 'ai'" class="fu">
          <div
            v-if="aiInsight"
            class="bg-card border-2 border-ink rounded-2xl p-5 sm:p-6"
            style="box-shadow: 4px 4px 0 #1a1612"
          >
            <div class="flex items-center justify-between mb-3">
              <h2 class="font-mono text-xs text-muted uppercase tracking-wider">Analisis AI</h2>
              <button
                class="font-mono text-xs text-muted hover:text-ink underline"
                @click="aiInsight = null"
              >Tutup</button>
            </div>
            <div
              v-html="'<p>' + formatMd(aiInsight || '') + '</p>'"
              class="font-sans text-sm text-ink leading-relaxed [&_strong]:font-bold [&_p]:mt-2 [&_p:first-child]:mt-0"
            ></div>
          </div>

          <div
            v-else-if="aiLoading"
            class="bg-card border-2 border-ink rounded-2xl p-5 sm:p-6"
            style="box-shadow: 4px 4px 0 #1a1612"
          >
            <div class="flex items-center gap-3 justify-center">
              <div class="w-5 h-5 border-2 border-ink border-t-transparent rounded-full animate-spin"></div>
              <span class="font-mono text-sm text-muted">AI sedang menganalisis...</span>
            </div>
          </div>

          <div
            v-else
            class="bg-card border-2 border-ink rounded-2xl p-5 sm:p-6"
            style="box-shadow: 4px 4px 0 #1a1612"
          >
            <button
              class="neo-btn-primary max-w-full"
              @click="handleAiTab"
              :disabled="aiLoading"
            >
              <span
                v-if="aiLoading"
                class="inline-block w-4 h-4 border-2 border-cream border-t-transparent rounded-full animate-spin mr-2"
              ></span>
              {{ aiLoading ? "Memproses..." : "Generate Analisis AI" }}
            </button>
          </div>

          <div v-if="aiError" class="mt-3 font-mono text-xs text-red-600">{{ aiError }}</div>
        </div>
      </div>
    </div>


  </main>
</template>
