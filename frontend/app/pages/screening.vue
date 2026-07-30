<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useAuthSession } from "~/composables/useAuth";
import { screeningService } from "~/services/screening.service";
import type { ScreeningResult } from "~/services/screening.service";

definePageMeta({ layout: "main", middleware: "auth" });

const { isAdmin } = useAuthSession();

const results = ref<ScreeningResult[]>([]);
const loading = ref(true);
const error = ref("");
const triggering = ref(false);
const expandedTicker = ref<string | null>(null);

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

const formatMd = (s: string) =>
  s.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
   .replace(/\n\n/g, '</p><p>')
   .replace(/\n/g, '<br>');

const formatPrice = (v: any) => {
  if (v === null || v === undefined) return "N/A";
  const n = Number(v);
  return isNaN(n) ? "N/A" : n.toLocaleString("id-ID");
};

const NAFmt = (v: any) => {
  if (v === null || v === undefined) return "N/A";
  if (typeof v === "number" && v === 0) return "N/A";
  if (typeof v === "boolean") return v ? "Ya" : "Tidak";
  return v;
};

const signalBadgeClass = (overall: string) => {
  if (overall === "bullish") return "bg-green-100 text-green-800 border-green-300";
  if (overall === "bearish") return "bg-red-100 text-red-800 border-red-300";
  return "bg-gray-100 text-gray-600 border-gray-300";
};

const scoreColor = (score: number) => {
  if (score > 0.3) return "bg-green-500";
  if (score > 0) return "bg-yellow-500";
  return "bg-red-500";
};

const scanDate = computed(() => {
  if (!results.value.length) return "";
  return results.value[0].scan_date;
});

const latestDateLabel = computed(() => {
  if (!results.value.length) return "";
  const d = new Date(results.value[0].created_at);
  return d.toLocaleDateString("id-ID", {
    day: "numeric", month: "long", year: "numeric",
    hour: "2-digit", minute: "2-digit",
  });
});

const isToday = computed(() => {
  if (!results.value.length) return false;
  const today = new Date().toISOString().split("T")[0];
  return results.value[0].scan_date === today;
});

async function fetchData() {
  loading.value = true;
  error.value = "";
  try {
    const res = await screeningService.getLatest();
    results.value = res.data || [];
  } catch (e: any) {
    error.value = e?.data?.message || e?.message || "Gagal memuat screening";
  } finally {
    loading.value = false;
  }
}

async function handleTrigger() {
  if (triggering.value) return;
  triggering.value = true;
  try {
    await screeningService.triggerScreening();
    await new Promise(r => setTimeout(r, 2000));
    await fetchData();
  } catch (e: any) {
    error.value = e?.data?.message || e?.message || "Gagal trigger screening";
  } finally {
    triggering.value = false;
  }
}

const toggleExpand = (ticker: string) => {
  expandedTicker.value = expandedTicker.value === ticker ? null : ticker;
};

onMounted(fetchData);
</script>

<template>
  <main class="min-h-screen bg-cream text-ink font-sans antialiased overflow-x-hidden pt-20 pb-24">
    <div class="max-w-5xl mx-auto px-5 sm:px-8">
      <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-8">
        <div>
          <h1 class="font-serif text-4xl text-ink" style="letter-spacing: -1px">Screening</h1>
          <p class="font-mono text-xs text-muted mt-1 uppercase tracking-wider">
            Top 10 Saham — Nightly Screening
          </p>
        </div>
        <div v-if="results.length" class="flex items-center gap-3">
          <span class="font-mono text-xs text-muted">
            {{ isToday ? "Hari ini" : "Data: " + scanDate }}
          </span>
          <span class="font-mono text-[10px] text-muted">{{ latestDateLabel }}</span>
          <button
            v-if="isAdmin"
            class="neo-btn-sm"
            :disabled="triggering"
            @click="handleTrigger"
          >
            <span v-if="triggering" class="inline-block w-3 h-3 border-2 border-ink border-t-transparent rounded-full animate-spin mr-1.5"></span>
            {{ triggering ? "Memproses..." : "Trigger Screening" }}
          </button>
        </div>
      </div>

      <div v-if="loading" class="fu2 text-center py-14">
        <div class="inline-block w-8 h-8 border-2 border-ink border-t-transparent rounded-full animate-spin mb-3"></div>
        <p class="font-mono text-sm text-muted">Memuat hasil screening...</p>
      </div>

      <div v-else-if="error" class="fu2 bg-red-50 border-2 border-red-300 rounded-xl px-5 py-4 mb-8" style="box-shadow: 3px 3px 0 #e25757">
        <p class="font-mono text-sm text-red-700">{{ error }}</p>
      </div>

      <div v-else-if="!results.length" class="fu2 text-center py-14 bg-card border-2 border-ink rounded-2xl" style="box-shadow: 4px 4px 0 #1a1612">
        <p class="font-mono text-sm text-muted mb-4">Belum ada data screening.</p>
        <p v-if="isAdmin" class="font-mono text-xs text-muted mb-4">Trigger screening manual atau tunggu jadwal malam (21:00 WIB).</p>
        <button v-if="isAdmin" class="neo-btn-primary" :disabled="triggering" @click="handleTrigger">
          <span v-if="triggering" class="inline-block w-4 h-4 border-2 border-cream border-t-transparent rounded-full animate-spin mr-2"></span>
          {{ triggering ? "Memproses..." : "Trigger Screening Sekarang" }}
        </button>
      </div>

      <div v-else class="space-y-4">
        <div class="fu3 bg-card border-2 border-ink rounded-2xl overflow-hidden" style="box-shadow: 4px 4px 0 #1a1612">
          <table class="w-full text-left font-mono text-xs">
            <thead>
              <tr class="bg-gray-100 text-muted uppercase tracking-wider">
                <th class="py-3 px-4 w-12">Rank</th>
                <th class="py-3 px-4">Ticker</th>
                <th class="py-3 px-4">Score</th>
                <th class="py-3 px-4 hidden sm:table-cell">Signal</th>
                <th class="py-3 px-4 hidden md:table-cell">Confidence</th>
                <th class="py-3 px-4 hidden sm:table-cell">Avg Volume</th>
                <th class="py-3 px-4 w-10"></th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(r, i) in results"
                :key="r.ticker"
                class="border-t border-bdr/50 cursor-pointer hover:bg-cream2 transition-colors"
                :class="{ 'bg-bluebg/30': expandedTicker === r.ticker }"
                @click="toggleExpand(r.ticker)"
              >
                <td class="py-3 px-4 font-bold text-base">{{ r.rank }}</td>
                <td class="py-3 px-4">
                  <span class="font-mono text-sm font-bold text-bluedk bg-bluebg px-2 py-0.5 rounded-md">{{ r.ticker }}</span>
                </td>
                <td class="py-3 px-4">
                  <div class="flex items-center gap-2">
                    <div class="w-16 h-2 bg-gray-200 rounded-full overflow-hidden">
                      <div
                        class="h-full rounded-full transition-all"
                        :class="scoreColor(r.score)"
                        :style="{ width: Math.abs(r.score) * 100 + '%' }"
                      ></div>
                    </div>
                    <span class="font-bold text-sm" :class="r.score > 0 ? 'text-green-700' : 'text-red-700'">
                      {{ r.score.toFixed(2) }}
                    </span>
                  </div>
                </td>
                <td class="py-3 px-4 hidden sm:table-cell">
                  <span class="text-[10px] font-bold px-2 py-0.5 rounded-md border uppercase" :class="signalBadgeClass(r.overall)">
                    {{ r.overall }}
                  </span>
                </td>
                <td class="py-3 px-4 hidden md:table-cell">
                  <span class="text-[10px] px-2 py-0.5 rounded border font-semibold"
                    :class="r.confidence === 'high' ? 'bg-green-50 text-green-700 border-green-200' : r.confidence === 'medium' ? 'bg-yellow-50 text-yellow-700 border-yellow-200' : 'bg-gray-50 text-gray-500 border-gray-200'"
                  >{{ r.confidence.toUpperCase() }}</span>
                </td>
                <td class="py-3 px-4 hidden sm:table-cell font-semibold">{{ formatCompact(r.avg_volume) }}</td>
                <td class="py-3 px-4">
                  <svg v-if="r.trading_plan || r.ai_insight" class="w-3 h-3 text-muted transition-transform" :class="{ 'rotate-180': expandedTicker === r.ticker }" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
                    <path d="m6 9 6 6 6-6" />
                  </svg>
                </td>
              </tr>

              <!-- Expanded Row -->
              <tr v-for="r in results.filter(x => x.trading_plan || x.ai_insight && expandedTicker === x.ticker)" :key="'detail-'+r.ticker">
                <td colspan="7" class="p-0">
                  <div class="bg-cream2 border-t border-bdr/50 p-5 space-y-5">
                    <!-- Trading Plan -->
                    <div v-if="r.trading_plan" class="bg-white border border-bdr rounded-xl p-4">
                      <h3 class="font-mono text-xs text-muted uppercase tracking-wider mb-3">Trading Plan</h3>
                      <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-3">
                        <div>
                          <div class="font-mono text-[10px] text-muted uppercase">Bias</div>
                          <div class="font-mono text-sm font-bold mt-0.5 uppercase"
                            :class="r.trading_plan.bias === 'buy' ? 'text-green-700' : r.trading_plan.bias === 'sell' ? 'text-red-700' : 'text-gray-500'"
                          >{{ r.trading_plan.bias }}</div>
                        </div>
                        <div>
                          <div class="font-mono text-[10px] text-muted uppercase">Entry</div>
                          <div class="font-mono text-sm font-bold mt-0.5">{{ formatPrice(r.trading_plan.entry_price) }}</div>
                        </div>
                        <div>
                          <div class="font-mono text-[10px] text-muted uppercase">Stop Loss</div>
                          <div class="font-mono text-sm font-bold mt-0.5 text-red-600">{{ formatPrice(r.trading_plan.stop_loss) }}</div>
                        </div>
                        <div>
                          <div class="font-mono text-[10px] text-muted uppercase">Entry Zone</div>
                          <div class="font-mono text-sm font-bold mt-0.5" v-if="r.trading_plan.entry_zone">
                            {{ formatPrice(r.trading_plan.entry_zone.low) }} — {{ formatPrice(r.trading_plan.entry_zone.high) }}
                          </div>
                          <div v-else class="font-mono text-sm mt-0.5 text-muted">N/A</div>
                        </div>
                      </div>

                      <!-- TP Targets -->
                      <div v-if="r.trading_plan.targets?.length">
                        <div class="font-mono text-[10px] text-muted uppercase tracking-wider mb-1.5">Target Take Profit</div>
                        <table class="w-full text-left font-mono text-xs data-table border border-bdr rounded-lg overflow-hidden">
                          <thead>
                            <tr class="bg-gray-100 text-muted uppercase tracking-wider">
                              <th class="py-1.5 px-3">Level</th>
                              <th class="py-1.5 px-3">Harga</th>
                              <th class="py-1.5 px-3">R:R Ratio</th>
                            </tr>
                          </thead>
                          <tbody>
                            <tr v-for="t in r.trading_plan.targets" :key="t.level" class="border-t border-bdr/50">
                              <td class="py-1.5 px-3 font-bold">TP{{ t.level }}</td>
                              <td class="py-1.5 px-3">{{ formatPrice(t.price) }}</td>
                              <td class="py-1.5 px-3 text-green-700 font-bold">{{ Number(t.rr_ratio).toFixed(2) }}R</td>
                            </tr>
                          </tbody>
                        </table>
                      </div>

                      <!-- Invalidation & Disclaimer -->
                      <div v-if="r.trading_plan.invalidation_note" class="mt-3 p-2.5 bg-yellow-50 border border-yellow-200 rounded-lg">
                        <div class="font-mono text-[10px] text-muted uppercase tracking-wider mb-0.5">Invalidation</div>
                        <p class="font-mono text-xs text-yellow-800">{{ r.trading_plan.invalidation_note }}</p>
                      </div>
                      <div v-if="r.trading_plan.disclaimer" class="mt-2 p-2.5 bg-gray-100 border border-gray-200 rounded-lg">
                        <div class="font-mono text-[10px] text-muted uppercase tracking-wider mb-0.5">Disclaimer</div>
                        <p class="font-mono text-xs text-gray-600">{{ r.trading_plan.disclaimer }}</p>
                      </div>
                    </div>

                    <!-- AI Insight -->
                    <div v-if="r.ai_insight" class="bg-white border border-bdr rounded-xl p-4">
                      <h3 class="font-mono text-xs text-muted uppercase tracking-wider mb-2">AI Analisis</h3>
                      <div v-html="'<p>' + formatMd(r.ai_insight) + '</p>'" class="font-sans text-sm text-ink leading-relaxed [&_strong]:font-bold [&_p]:mt-2 [&_p:first-child]:mt-0"></div>
                    </div>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </main>
</template>
