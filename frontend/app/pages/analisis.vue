<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from "vue";
import * as d3 from "d3";
import { useTickerSearch, useAnalisis } from "~/composables/useAnalisis";

definePageMeta({ layout: "main", middleware: "auth" });

const DATE_FORMATTER = (d: string) => {
  const dt = new Date(d);
  return dt.toLocaleDateString("id-ID", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  });
};

const NA = (v: any) =>
  v === null || v === undefined
    ? "N/A"
    : typeof v === "number"
      ? v.toLocaleString("id-ID")
      : v;
const NAFmt = (v: any) => {
  if (v === null || v === undefined) return "N/A";
  if (typeof v === "number" && v === 0) return "N/A";
  if (typeof v === "boolean") return v ? "Ya" : "Tidak";
  return v;
};

const today = new Date();
const oneYearAgo = new Date(
  today.getFullYear() - 1,
  today.getMonth(),
  today.getDate(),
);

const fmtDate = (d: Date) => d.toISOString().split("T")[0];
const dateStart = ref(fmtDate(oneYearAgo));
const dateEnd = ref(fmtDate(today));

const selectedTicker = ref<{ kode: string; nama: string } | null>(null);

const {
  query,
  suggestions,
  loading: tickerLoading,
  fetchAll,
} = useTickerSearch();
const {
  result,
  loading: analisisLoading,
  error: analisisError,
  run: runAnalisis,
  clear: clearResult,
  aiInsight,
  aiLoading,
  aiError,
  requestAiInsight,
} = useAnalisis();

const activeTab = ref<"review" | "indicators" | "snapshot" | "ai" | "chart">("review");

const handleAiTab = () => {
  activeTab.value = "ai";
  if (!aiInsight.value && !aiLoading.value && selectedTicker.value) {
    requestAiInsight(
      selectedTicker.value.kode,
      dateEnd.value,
      result.value!.indicators,
      result.value!.snapshot,
    );
  }
};

const chartSvg = ref<SVGSVGElement | null>(null);
const formatMd = (s: string) =>
  s.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
   .replace(/\n\n/g, '</p><p>')
   .replace(/\n/g, '<br>');
let d3Cleanup: (() => void) | null = null;

onMounted(() => fetchAll());

const showDropdown = computed(() => query.value && !selectedTicker.value);

const selectTicker = (s: { kode: string; nama_perusahaan?: string }) => {
  selectedTicker.value = { kode: s.kode, nama: s.nama_perusahaan || s.kode };
  query.value = `${s.kode} — ${s.nama_perusahaan || ""}`;
};

const clearTicker = () => {
  selectedTicker.value = null;
  query.value = "";
  clearResult();
};

const submit = () => {
  if (!selectedTicker.value || !dateStart.value || !dateEnd.value) return;
  runAnalisis(selectedTicker.value.kode, dateStart.value, dateEnd.value);
};

const formatPrice = (v: number | null | undefined) => {
  if (v === null || v === undefined) return "N/A";
  return new Intl.NumberFormat("id-ID", {
    minimumFractionDigits: 0,
    maximumFractionDigits: 2,
  }).format(v);
};

const formatCompact = (v: number) => {
  if (!v) return "0";
  if (v >= 1e12) return (v / 1e12).toFixed(2) + " T";
  if (v >= 1e9) return (v / 1e9).toFixed(2) + " M";
  if (v >= 1e6) return (v / 1e6).toFixed(2) + " Jt";
  return v.toLocaleString("id-ID");
};

const signalBadge = computed(() => {
  const o = result.value?.signal?.overall;
  if (o === "bullish")
    return {
      label: "BULLISH",
      class: "bg-green-100 text-green-800 border-green-300",
    };
  if (o === "bearish")
    return {
      label: "BEARISH",
      class: "bg-red-100 text-red-800 border-red-300",
    };
  return {
    label: "NETRAL",
    class: "bg-gray-100 text-gray-600 border-gray-300",
  };
});

function drawChart(
  ohlcv: {
    date: string;
    open: number;
    high: number;
    low: number;
    close: number;
    volume: number;
  }[],
) {
  if (!chartSvg.value || !ohlcv.length) return;
  if (d3Cleanup) d3Cleanup();

  const margin = { top: 20, right: 20, bottom: 40, left: 50 };
  const width = chartSvg.value.clientWidth - margin.left - margin.right;
  const height = 400 - margin.top - margin.bottom;

  const svg = d3.select(chartSvg.value);
  svg.selectAll("*").remove();

  const g = svg
    .append("g")
    .attr("transform", `translate(${margin.left},${margin.top})`);

  const parseDate = d3.timeParse("%Y-%m-%d");
  const data = ohlcv
    .map((d) => ({ ...d, parsed: parseDate(d.date)! }))
    .filter((d) => d.parsed);

  const xScale = d3
    .scaleTime()
    .domain(d3.extent(data, (d) => d.parsed) as [Date, Date])
    .range([0, width]);

  const yScale = d3
    .scaleLinear()
    .domain([
      d3.min(data, (d) => d.low)! * 0.98,
      d3.max(data, (d) => d.high)! * 1.02,
    ])
    .range([height, 0]);

  const volumeScale = d3
    .scaleLinear()
    .domain([0, d3.max(data, (d) => d.volume)!])
    .range([height, height - 80]);

  g.append("g")
    .attr("transform", `translate(0,${height})`)
    .call(
      d3
        .axisBottom(xScale)
        .ticks(6)
        .tickFormat(d3.timeFormat("%b %Y") as any),
    )
    .selectAll("text")
    .attr("font-family", "Share Tech Mono")
    .attr("font-size", "10px");

  g.append("g")
    .call(
      d3
        .axisLeft(yScale)
        .ticks(6)
        .tickFormat((d) => formatPrice(d as number) as any),
    )
    .selectAll("text")
    .attr("font-family", "Share Tech Mono")
    .attr("font-size", "10px");

  g.selectAll(".volume-bar")
    .data(data)
    .enter()
    .append("rect")
    .attr("class", "volume-bar")
    .attr("x", (d) => xScale(d.parsed)! - Math.max(1, width / data.length / 4))
    .attr("y", (d) => volumeScale(d.volume))
    .attr("width", Math.max(2, width / data.length / 2))
    .attr("height", (d) => height - volumeScale(d.volume))
    .attr("fill", (d) => (d.close >= d.open ? "#26a69a" : "#ef5350"))
    .attr("opacity", 0.3);

  g.selectAll(".candle")
    .data(data)
    .enter()
    .append("line")
    .attr("class", "candle-wick")
    .attr("x1", (d) => xScale(d.parsed)!)
    .attr("x2", (d) => xScale(d.parsed)!)
    .attr("y1", (d) => yScale(d.high))
    .attr("y2", (d) => yScale(d.low))
    .attr("stroke", (d) => (d.close >= d.open ? "#26a69a" : "#ef5350"))
    .attr("stroke-width", 1.5);

  g.selectAll(".candle-body")
    .data(data)
    .enter()
    .append("rect")
    .attr("class", "candle-body")
    .attr("x", (d) => xScale(d.parsed)! - Math.max(1, width / data.length / 3))
    .attr("y", (d) => yScale(Math.max(d.open, d.close)))
    .attr("width", Math.max(2, width / data.length / 1.5))
    .attr("height", (d) =>
      Math.max(1, Math.abs(yScale(d.open) - yScale(d.close))),
    )
    .attr("fill", (d) => (d.close >= d.open ? "#26a69a" : "#ef5350"));

  const tooltip = d3
    .select("body")
    .append("div")
    .attr("class", "candle-tooltip")
    .style("position", "fixed")
    .style("pointer-events", "none")
    .style("display", "none")
    .style("background", "#fff")
    .style("border", "2.5px solid #1a1612")
    .style("border-radius", "12px")
    .style("padding", "10px 14px")
    .style("box-shadow", "4px 4px 0 #1a1612")
    .style("z-index", "999")
    .style("font-family", "DM Sans, sans-serif")
    .style("font-size", "12px")
    .style("color", "#1a1612");

  const bisect = d3.bisector((d: (typeof data)[0]) => d.parsed).left;
  const overlay = g
    .append("rect")
    .attr("width", width)
    .attr("height", height)
    .attr("fill", "none")
    .attr("pointer-events", "all");

  overlay.on("mousemove", (event: MouseEvent) => {
    const mx = d3.pointer(event)[0];
    const x0 = xScale.invert(mx);
    const i = bisect(data, x0, 1);
    const d0 = data[i - 1];
    const d1 = data[i];
    if (!d0 || !d1) return;
    const d =
      x0.getTime() - d0.parsed.getTime() > d1.parsed.getTime() - x0.getTime()
        ? d1
        : d0;
    tooltip
      .style("display", "block")
      .html(
        `
                <div style="font-family:Share Tech Mono;font-size:11px;margin-bottom:4px;color:#3d3730">${DATE_FORMATTER(d.date)}</div>
                <div><b>O:</b> ${formatPrice(d.open)}</div>
                <div><b>H:</b> ${formatPrice(d.high)}</div>
                <div><b>L:</b> ${formatPrice(d.low)}</div>
                <div><b>C:</b> ${formatPrice(d.close)}</div>
                <div><b>Vol:</b> ${formatCompact(d.volume)}</div>
            `,
      )
      .style("left", event.clientX + 16 + "px")
      .style("top", event.clientY - 80 + "px");
  });

  overlay.on("mouseleave", () => tooltip.style("display", "none"));
  svg.on("mouseleave", () => tooltip.style("display", "none"));

  d3Cleanup = () => {
    tooltip.remove();
  };
}

watch(
  () => result.value?.ohlcv,
  (ohlcv) => {
    if (ohlcv?.length) {
      nextTick(() => drawChart(ohlcv));
    }
  },
  { deep: true },
);

watch(activeTab, (tab) => {
  if (tab === 'chart' && result.value?.ohlcv?.length) {
    nextTick(() => drawChart(result.value!.ohlcv));
  }
});

onUnmounted(() => {
  if (d3Cleanup) d3Cleanup();
});
</script>

<template>
  <main
    class="min-h-screen bg-cream text-ink font-sans antialiased overflow-x-hidden pt-20 pb-24"
  >
    <div class="max-w-5xl mx-auto px-5 sm:px-8">
      <div class="fu mb-8">
        <h1 class="font-serif text-4xl text-ink" style="letter-spacing: -1px">
          Analisis
        </h1>
        <p class="font-mono text-xs text-muted mt-1 uppercase tracking-wider">
          Cari dan analisis pergerakan harga saham
        </p>
      </div>

      <!-- Search + Date + Submit -->
      <div
        class="fu1 bg-card border-2 border-ink rounded-2xl p-5 sm:p-6 mb-8"
        style="box-shadow: 4px 4px 0 #1a1612"
      >
        <div class="flex flex-col sm:flex-row gap-4 items-end">
          <div class="flex-1 w-full relative">
            <label
              class="font-mono text-xs text-muted uppercase tracking-wider block mb-1.5"
              >Ticker</label
            >
            <div
              v-if="selectedTicker"
              class="search-input flex items-center justify-between cursor-default"
            >
              <span>
                <span
                  class="font-mono text-sm font-bold text-bluedk bg-bluebg px-2 py-0.5 rounded-md mr-2"
                  >{{ selectedTicker.kode }}</span
                >
                <span class="text-sm text-ink2">{{ selectedTicker.nama }}</span>
              </span>
              <button
                class="font-mono text-xs text-muted hover:text-ink ml-2"
                @click="clearTicker"
              >
                ✕
              </button>
            </div>
            <input
              v-else
              v-model="query"
              class="search-input"
              type="text"
              placeholder="Cari kode atau nama perusahaan..."
            />

            <div
              v-if="showDropdown"
              class="absolute top-full mt-2 left-0 right-0 bg-white border-2 border-ink rounded-xl max-h-64 min-h-fit overflow-y-auto z-50"
              style="box-shadow: 4px 4px 0 #1a1612"
            >
              <div v-if="tickerLoading" class="px-4 py-4 text-center">
                <span class="font-mono text-xs text-muted"
                  >Memuat daftar ticker...</span
                >
              </div>
              <div v-else-if="suggestions.length">
                <div
                  v-for="s in suggestions"
                  :key="s.kode"
                  class="flex items-center gap-3 px-4 py-3 cursor-pointer hover:bg-bluebg border-b border-bdr last:border-0 transition-colors"
                  @mousedown.prevent="selectTicker(s)"
                >
                  <span
                    class="font-mono text-xs font-bold text-bluedk bg-bluebg px-2 py-0.5 rounded-md min-w-[46px] text-center"
                  >
                    {{ s.kode }}
                  </span>
                  <span class="text-sm text-ink2">{{ s.nama_perusahaan }}</span>
                </div>
              </div>
              <div v-else-if="query" class="px-4 py-4 text-center">
                <p class="text-sm text-muted font-mono">
                  Ticker tidak ditemukan
                </p>
              </div>
            </div>
          </div>

          <div class="w-full sm:w-40">
            <label
              class="font-mono text-xs text-muted uppercase tracking-wider block mb-1.5"
              >Dari</label
            >
            <input v-model="dateStart" type="date" class="neo-input" />
          </div>

          <div class="w-full sm:w-40">
            <label
              class="font-mono text-xs text-muted uppercase tracking-wider block mb-1.5"
              >Sampai</label
            >
            <input v-model="dateEnd" type="date" class="neo-input" />
          </div>

          <div class="w-full sm:w-auto">
            <button
              class="neo-btn-primary whitespace-nowrap"
              :disabled="!selectedTicker || analisisLoading"
              @click="submit"
            >
              <span
                v-if="analisisLoading"
                class="inline-block w-4 h-4 border-2 border-cream border-t-transparent rounded-full animate-spin"
              ></span>
              {{ analisisLoading ? "Menganalisis..." : "Analisis" }}
            </button>
          </div>
        </div>
      </div>

      <!-- Error -->
      <div
        v-if="analisisError"
        class="fu2 bg-red-50 border-2 border-red-300 rounded-xl px-5 py-4 mb-8"
        style="box-shadow: 3px 3px 0 #e25757"
      >
        <p class="font-mono text-sm text-red-700">Gagal: {{ analisisError }}</p>
      </div>

      <!-- Loading State -->
      <div v-if="analisisLoading && !result" class="fu2 text-center py-14">
        <div
          class="inline-block w-8 h-8 border-2 border-ink border-t-transparent rounded-full animate-spin mb-3"
        ></div>
        <p class="font-mono text-sm text-muted">
          Mengambil data dan menghitung indikator...
        </p>
      </div>

      <!-- Results -->
      <div v-if="result" class="space-y-6">
        <!-- Tab Bar -->
        <div class="flex gap-2 bg-card border-2 border-ink rounded-2xl p-1.5" style="box-shadow: 4px 4px 0 #1a1612">
          <button
            v-for="tab in [
              { key: 'review', label: 'Review' },
              { key: 'indicators', label: 'Indikator' },
              { key: 'snapshot', label: 'Snapshot' },
              { key: 'ai', label: 'AI Analisis' },
              { key: 'chart', label: 'Chart' },
            ]"
            :key="tab.key"
            class="flex-1 font-mono text-xs font-bold py-2.5 px-3 rounded-xl border-2 transition-all uppercase tracking-wider"
            :class="activeTab === tab.key
              ? 'bg-ink text-cream border-ink shadow-[2px_2px_0_#1a1612]'
              : 'bg-card text-ink border-transparent hover:bg-gray-100'"
            @click="tab.key === 'ai' ? handleAiTab() : (activeTab = tab.key as any)"
          >{{ tab.label }}</button>
        </div>

        <!-- Tab: Review -->
        <div v-show="activeTab === 'review'" class="space-y-6">
        <!-- Signal Badge -->
        <div
          v-if="result.signal"
          class="fu3 bg-card border-2 border-ink rounded-2xl p-5 sm:p-6"
          style="box-shadow: 4px 4px 0 #1a1612"
        >
          <div
            class="flex flex-col sm:flex-row items-start sm:items-center gap-3 sm:gap-6"
          >
            <div class="flex items-center gap-2 shrink-0">
              <h2 class="font-mono text-xs text-muted uppercase tracking-wider">
                Sinyal
              </h2>
              <span
                class="font-mono text-[10px] bg-purple-100 text-purple-800 border border-purple-300 px-2 py-0.5 rounded-md"
                >Swing 1-4 Minggu</span
              >
            </div>
            <span
              class="font-mono text-xs font-bold px-3 py-1 rounded-md border-2"
              :class="signalBadge.class"
            >
              {{ signalBadge.label }}
            </span>
            <span class="font-mono text-xs text-muted"
              >Skor: {{ result.signal.score.toFixed(2) }}</span
            >
            <span
              v-if="result.signal.confidence"
              class="font-mono text-xs px-2 py-0.5 rounded border"
              :class="
                result.signal.confidence === 'high'
                  ? 'bg-green-50 text-green-700 border-green-200'
                  : result.signal.confidence === 'medium'
                    ? 'bg-yellow-50 text-yellow-700 border-yellow-200'
                    : 'bg-gray-50 text-gray-500 border-gray-200'
              "
            >
              {{ result.signal.confidence.toUpperCase() }}
            </span>
            <span
              v-if="result.signal.trend_filter_passed === true"
              class="font-mono text-[10px] text-green-600"
              >Trend Filter ✓</span
            >
            <span
              v-else-if="result.signal.trend_filter_passed === false"
              class="font-mono text-[10px] text-red-500"
              >Trend Filter ✗</span
            >
            <button
              class="font-mono text-xs text-bluedk underline ml-auto"
              @click="signalExpanded = !signalExpanded"
            >
              {{ signalExpanded ? "Sembunyikan" : "Detail" }}
            </button>
          </div>
          <div v-if="signalExpanded" class="mt-4 space-y-1">
            <div
              v-for="b in result.signal.breakdown"
              :key="b.indicator"
              class="flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-mono"
              :class="
                b.signal === 'bullish'
                  ? 'bg-green-50'
                  : b.signal === 'bearish'
                    ? 'bg-red-50'
                    : 'bg-gray-50'
              "
            >
              <span class="font-bold w-28 shrink-0">{{ b.indicator }}</span>
              <span
                class="text-xs uppercase font-bold"
                :class="
                  b.signal === 'bullish'
                    ? 'text-green-700'
                    : b.signal === 'bearish'
                      ? 'text-red-700'
                      : 'text-gray-500'
                "
              >
                {{ b.signal }}
              </span>
              <span class="text-xs text-muted">{{ b.note }}</span>
            </div>
          </div>
        </div>

        <!-- Trading Plan -->
        <div
          v-if="result.trading_plan"
          class="fu3 bg-card border-2 border-ink rounded-2xl p-5 sm:p-6"
          style="box-shadow: 4px 4px 0 #1a1612"
        >
          <h2
            class="font-mono text-xs text-muted uppercase tracking-wider mb-4"
          >
            Trading Plan
          </h2>

          <div v-if="result.trading_plan.time_stop_days" class="mb-3">
            <span
              class="font-mono text-[10px] bg-purple-50 text-purple-700 border border-purple-200 px-2 py-0.5 rounded-md"
              >Time Stop: {{ result.trading_plan.time_stop_days }} hari</span
            >
          </div>
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-2">
            <div>
              <div class="font-mono text-xs text-muted uppercase">Bias</div>
              <div
                class="font-mono text-lg font-bold mt-0.5 uppercase"
                :class="
                  result.trading_plan.bias === 'buy'
                    ? 'text-green-700'
                    : result.trading_plan.bias === 'sell'
                      ? 'text-red-700'
                      : 'text-gray-500'
                "
              >
                {{ result.trading_plan.bias }}
              </div>
            </div>
            <!-- <div>
              <div class="font-mono text-xs text-muted uppercase">
                Harga Saat Ini
              </div>
              <div class="font-mono text-lg font-bold mt-0.5">
                {{
                  result.trading_plan.current_price !== null &&
                  result.trading_plan.current_price !== undefined
                    ? formatPrice(result.trading_plan.current_price)
                    : "N/A"
                }}
              </div>
            </div> -->
            <div>
              <div class="font-mono text-xs text-muted uppercase">
                Entry Price
              </div>
              <div class="font-mono text-lg font-bold mt-0.5">
                {{
                  result.trading_plan.entry_price !== null &&
                  result.trading_plan.entry_price !== undefined
                    ? formatPrice(result.trading_plan.entry_price)
                    : "N/A"
                }}
              </div>
            </div>
            <div v-if="result.trading_plan.entry_zone">
              <div class="font-mono text-xs text-muted uppercase">
                Entry Zone
              </div>
              <div class="font-mono text-lg font-bold mt-0.5">
                {{ formatPrice(result.trading_plan.entry_zone.low) }} —
                {{ formatPrice(result.trading_plan.entry_zone.high) }}
              </div>
            </div>
            <div>
              <div class="font-mono text-xs text-muted uppercase">
                Stop Loss
              </div>
              <div class="font-mono text-lg font-bold mt-0.5 text-red-600">
                {{
                  result.trading_plan.stop_loss !== null &&
                  result.trading_plan.stop_loss !== undefined
                    ? formatPrice(result.trading_plan.stop_loss)
                    : "N/A"
                }}
              </div>
            </div>
          </div>
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-5">
            <!-- <div>
              <div class="font-mono text-xs text-muted uppercase">
                Tipe Entry
              </div>
              <div class="font-mono text-sm font-bold mt-0.5 uppercase">
                {{ result.trading_plan.entry_type || "—" }}
              </div>
            </div>
            <div>
              <div class="font-mono text-xs text-muted uppercase">
                Stop Loss Basis
              </div>
              <div class="font-mono text-sm font-bold mt-0.5">
                {{ result.trading_plan.stop_loss_basis || "—" }}
              </div>
            </div> -->
            <!-- <div>
              <div class="font-mono text-xs text-muted uppercase">
                Ukuran Posisi
              </div>
              <div class="font-mono text-lg font-bold mt-0.5">
                {{
                  result.trading_plan.suggested_position_size_pct.toFixed(2)
                }}%
              </div>
            </div> -->
          </div>
          <div
            v-if="result.trading_plan.entry_note"
            class="mb-4 p-3 bg-blue-50 border border-blue-200 rounded-lg"
          >
            <div
              class="font-mono text-xs text-muted uppercase tracking-wider mb-1"
            >
              Catatan Entry
            </div>
            <p class="font-mono text-xs text-blue-800">
              {{ result.trading_plan.entry_note }}
            </p>
          </div>

          <!-- TP Targets -->
          <div v-if="result.trading_plan.targets?.length" class="mb-4">
            <div
              class="font-mono text-xs text-muted uppercase tracking-wider mb-2"
            >
              Target Take Profit
            </div>
            <table
              class="w-full text-left font-mono text-xs data-table border-2 border-bdr rounded-xl overflow-hidden"
            >
              <thead>
                <tr class="bg-gray-100 text-muted uppercase tracking-wider">
                  <th class="py-2 px-3">Level</th>
                  <th class="py-2 px-3">Harga</th>
                  <th class="py-2 px-3">R:R Ratio</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="t in result.trading_plan.targets"
                  :key="t.level"
                  class="border-t border-bdr/50"
                >
                  <td class="py-2 px-3 font-bold">TP{{ t.level }}</td>
                  <td class="py-2 px-3">{{ formatPrice(t.price) }}</td>
                  <td class="py-2 px-3 text-green-700 font-bold">
                    {{ t.rr_ratio.toFixed(2) }}R
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <!-- Invalidation Note -->
          <div class="mb-4 p-3 bg-yellow-50 border border-yellow-200 rounded-lg">
            <div
              class="font-mono text-xs text-muted uppercase tracking-wider mb-1"
            >
              Invalidation
            </div>
            <p class="font-mono text-xs text-yellow-800">
              {{ result.trading_plan.invalidation_note }}
            </p>
          </div>

          <!-- Disclaimer — permanent, non-dismissible -->
          <div class="p-3 bg-gray-100 border border-gray-200 rounded-lg">
            <div
              class="font-mono text-xs text-muted uppercase tracking-wider mb-1"
            >
              Disclaimer
            </div>
            <p class="font-mono text-xs text-gray-600 leading-relaxed">
              {{ result.trading_plan.disclaimer }}
            </p>
          </div>
        </div>
      </div>

        <!-- Tab: AI -->
        <div v-show="activeTab === 'ai'" class="space-y-6">
        <!-- AI Insight -->
        <div class="fu3">
          <button
            v-if="!aiInsight && !aiLoading"
            class="neo-btn-primary max-w-full"
            @click="handleAiTab"
            :disabled="aiLoading"
          >
            <span
              v-if="aiLoading"
              class="inline-block w-4 h-4 border-2 border-cream border-t-transparent rounded-full animate-spin mr-2"
            ></span>
            {{ aiLoading ? "Memproses..." : "Analisis AI" }}
          </button>

          <div
            v-if="aiLoading"
            class="mt-4 p-5 bg-card border-2 border-ink rounded-2xl"
            style="box-shadow: 4px 4px 0 #1a1612"
          >
            <div class="flex items-center gap-3 justify-center">
              <div
                class="w-5 h-5 border-2 border-ink border-t-transparent rounded-full animate-spin"
              ></div>
              <span class="font-mono text-sm text-muted text-center"
                >AI sedang menganalisis...</span
              >
            </div>
          </div>

          <div
            v-if="aiInsight"
            class="bg-card border-2 border-ink rounded-2xl p-5 sm:p-6"
            style="box-shadow: 4px 4px 0 #1a1612"
          >
            <div class="flex items-center justify-between mb-3">
              <h2 class="font-mono text-xs text-muted uppercase tracking-wider">
                Analisis AI
              </h2>
              <button
                class="font-mono text-xs text-muted hover:text-ink underline"
                @click="aiInsight = null"
              >
                Tutup
              </button>
            </div>
            <div v-html="'<p>' + formatMd(aiInsight) + '</p>'" class="font-sans text-sm text-ink leading-relaxed [&_strong]:font-bold [&_p]:mt-2 [&_p:first-child]:mt-0"></div>
          </div>

          <div
            v-if="aiError"
            class="mt-4 p-4 bg-yellow-50 border-2 border-yellow-400 rounded-xl"
          >
            <p class="font-mono text-sm text-yellow-800 font-semibold">
              AI Insight tidak tersedia
            </p>
            <p class="font-mono text-xs text-yellow-700 mt-1">
              Layanan AI sedang sibuk. Coba lagi nanti.
            </p>
          </div>
        </div>
      </div>

        <!-- Tab: Snapshot -->
        <div v-show="activeTab === 'snapshot'" class="space-y-6">
        <!-- Snapshot Card -->
        <div
          v-if="result.snapshot"
          class="fu2 bg-card border-2 border-ink rounded-2xl p-5 sm:p-6"
          style="box-shadow: 4px 4px 0 #1a1612"
        >
          <h2
            class="font-mono text-xs text-muted uppercase tracking-wider mb-4"
          >
            Snapshot Fundamental
          </h2>
          <div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
            <div>
              <div class="font-mono text-xs text-muted uppercase">Harga</div>
              <div class="font-serif text-2xl mt-0.5">
                {{ formatPrice(result.snapshot.price) }}
              </div>
            </div>
            <div>
              <div class="font-mono text-xs text-muted uppercase">52w High</div>
              <div class="font-serif text-2xl mt-0.5">
                {{ formatPrice(result.snapshot.high52) }}
              </div>
            </div>
            <div>
              <div class="font-mono text-xs text-muted uppercase">52w Low</div>
              <div class="font-serif text-2xl mt-0.5">
                {{ formatPrice(result.snapshot.low52) }}
              </div>
            </div>
            <div>
              <div class="font-mono text-xs text-muted uppercase">Volume</div>
              <div class="font-serif text-2xl mt-0.5">
                {{ formatCompact(result.snapshot.volume) }}
              </div>
            </div>
            <div>
              <div class="font-mono text-xs text-muted uppercase">
                Market Cap
              </div>
              <div class="font-serif text-2xl mt-0.5">
                {{ formatCompact(result.snapshot.marketcap) }}
              </div>
            </div>
            <div>
              <div class="font-mono text-xs text-muted uppercase">PE</div>
              <div class="font-serif text-2xl mt-0.5">
                {{ NA(result.snapshot.pe) }}
              </div>
            </div>
            <div>
              <div class="font-mono text-xs text-muted uppercase">EPS</div>
              <div class="font-serif text-2xl mt-0.5">
                {{ formatPrice(result.snapshot.eps) }}
              </div>
            </div>
            <div>
              <div class="font-mono text-xs text-muted uppercase">Currency</div>
              <div class="font-serif text-2xl mt-0.5">
                {{ result.snapshot.currency || "N/A" }}
              </div>
            </div>
          </div>
        </div>
      </div>

        <!-- Tab: Indikator -->
        <div v-show="activeTab === 'indicators'" class="space-y-6">
        <!-- Indicators Grid -->
        <div
          v-if="result.indicators"
          class="fu3 bg-card border-2 border-ink rounded-2xl p-5 sm:p-6"
          style="box-shadow: 4px 4px 0 #1a1612"
        >
          <h2
            class="font-mono text-xs text-muted uppercase tracking-wider mb-4"
          >
            Indikator Teknikal
          </h2>
          <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-6 gap-4">
            <div>
              <div class="font-mono text-xs text-muted uppercase">SMA20</div>
              <div class="font-mono text-lg font-bold mt-0.5">
                {{ formatPrice(result.indicators.sma20) }}
              </div>
            </div>
            <div>
              <div class="font-mono text-xs text-muted uppercase">SMA50</div>
              <div class="font-mono text-lg font-bold mt-0.5">
                {{ formatPrice(result.indicators.sma50) }}
              </div>
            </div>
            <div>
              <div class="font-mono text-xs text-muted uppercase">SMA200</div>
              <div class="font-mono text-lg font-bold mt-0.5">
                {{ NAFmt(result.indicators.sma200) }}
              </div>
            </div>
            <div>
              <div class="font-mono text-xs text-muted uppercase">EMA20</div>
              <div class="font-mono text-lg font-bold mt-0.5">
                {{ formatPrice(result.indicators.ema20) }}
              </div>
            </div>
            <div>
              <div class="font-mono text-xs text-muted uppercase">RSI</div>
              <div class="font-mono text-lg font-bold mt-0.5">
                {{ NAFmt(result.indicators.rsi?.toFixed(2)) }}
              </div>
            </div>
            <div>
              <div class="font-mono text-xs text-muted uppercase">MACD</div>
              <div class="font-mono text-lg font-bold mt-0.5">
                {{ NAFmt(result.indicators.macd?.toFixed(2)) }}
              </div>
            </div>
            <div>
              <div class="font-mono text-xs text-muted uppercase">ADX</div>
              <div class="font-mono text-lg font-bold mt-0.5">
                {{ NAFmt(result.indicators.adx?.toFixed(2)) }}
              </div>
            </div>
            <div>
              <div class="font-mono text-xs text-muted uppercase">ATR</div>
              <div class="font-mono text-lg font-bold mt-0.5">
                {{ NAFmt(result.indicators.atr) }}
              </div>
            </div>
            <div>
              <div class="font-mono text-xs text-muted uppercase">Support</div>
              <div class="font-mono text-lg font-bold mt-0.5">
                {{ formatPrice(result.indicators.support) }}
              </div>
            </div>
            <div>
              <div class="font-mono text-xs text-muted uppercase">
                Resistance
              </div>
              <div class="font-mono text-lg font-bold mt-0.5">
                {{ formatPrice(result.indicators.resistance) }}
              </div>
            </div>
            <div>
              <div class="font-mono text-xs text-muted uppercase">BB Width</div>
              <div class="font-mono text-lg font-bold mt-0.5">
                {{ NAFmt(result.indicators.bb_width?.toFixed(4)) }}
              </div>
            </div>
            <div>
              <div class="font-mono text-xs text-muted uppercase">
                Volume Spike
              </div>
              <div class="font-mono text-lg font-bold mt-0.5">
                {{ NAFmt(result.indicators.volume_spike) }}
              </div>
            </div>
          </div>
        </div>
      </div>

        <!-- Tab: Chart -->
        <div v-show="activeTab === 'chart'" class="space-y-6">
        <!-- D3 Candlestick Chart -->
        <div
          v-if="result.ohlcv?.length"
          class="fu2 bg-card border-2 border-ink rounded-2xl p-5 sm:p-6"
          style="box-shadow: 4px 4px 0 #1a1612"
        >
          <h2
            class="font-mono text-xs text-muted uppercase tracking-wider mb-4"
          >
            Grafik Harga Historis
          </h2>
          <svg
            ref="chartSvg"
            class="w-full"
            :style="{ height: '400px' }"
            style="overflow: visible"
          ></svg>
        </div>

        <!-- Python Multi-panel Chart -->
        <div
          v-if="result.chart_image"
          class="fu2 bg-card border-2 border-ink rounded-2xl p-5 sm:p-6"
          style="box-shadow: 4px 4px 0 #1a1612"
        >
          <h2
            class="font-mono text-xs text-muted uppercase tracking-wider mb-4"
          >
            Chart Teknikal Lengkap
          </h2>
          <img
            :src="'data:image/png;base64,' + result.chart_image"
            class="w-full h-auto"
            alt="Technical Analysis Chart"
          />
        </div>
      </div>

        <!-- OHLCV Table -->
        <!-- <details class="fu3 bg-card border-2 border-ink rounded-2xl p-5 sm:p-6"
                    style="box-shadow:4px 4px 0 #1a1612">
                    <summary class="font-mono text-xs text-muted uppercase tracking-wider cursor-pointer select-none">Lihat Data Lengkap</summary>
                    <div class="mt-4 overflow-x-auto">
                        <table class="w-full text-left font-mono text-xs data-table">
                            <thead>
                                <tr class="text-muted uppercase tracking-wider border-b border-bdr">
                                    <th class="py-2 pr-4">Tanggal</th>
                                    <th class="py-2 pr-4">Open</th>
                                    <th class="py-2 pr-4">High</th>
                                    <th class="py-2 pr-4">Low</th>
                                    <th class="py-2 pr-4">Close</th>
                                    <th class="py-2 pr-4">Volume</th>
                                </tr>
                            </thead>
                            <tbody>
                                <tr v-for="row in result.ohlcv" :key="row.date" class="border-b border-bdr/50">
                                    <td class="py-2 pr-4">{{ DATE_FORMATTER(row.date) }}</td>
                                    <td class="py-2 pr-4">{{ formatPrice(row.open) }}</td>
                                    <td class="py-2 pr-4">{{ formatPrice(row.high) }}</td>
                                    <td class="py-2 pr-4">{{ formatPrice(row.low) }}</td>
                                    <td class="py-2 pr-4">{{ formatPrice(row.close) }}</td>
                                    <td class="py-2 pr-4">{{ formatCompact(row.volume) }}</td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </details> -->
      </div>

      <!-- Empty State -->
      <div
        v-if="!result && !analisisLoading && !analisisError"
        class="text-center py-20"
      >
        <p class="font-mono text-sm text-muted">
          Pilih ticker dan rentang tanggal, lalu klik Analisis untuk memulai.
        </p>
      </div>
    </div>
  </main>
</template>
