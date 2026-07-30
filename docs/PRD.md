# PRD: Trackly — Analisis Teknikal, Trading Plan & Position Management

## 1. Latar Belakang

Aplikasi memiliki frontend (Nuxt/Vue) + backend (Go) yang terhubung ke Google Sheets (via Apps Script API) sebagai sumber data Google Finance, dengan 4 sheet utama:

- `master` — daftar seluruh emiten (Kode, Nama Perusahaan, Tanggal Pencatatan, Saham, Papan Pencatatan)
- `config` — key-value control cell: `selected_ticker`, `date_start`, `date_end`
- `google_finance` — snapshot harga & fundamental per ticker
- `chart` — data historis OHLCV mengikuti ticker & rentang tanggal aktif di `config`

Modul **Analisis** untuk meneliti saham secara teknikal: user mencari ticker, memilih rentang tanggal, melihat grafik historis lengkap dengan indikator teknikal, sinyal (bullish/bearish/netral), dan **trading plan** (entry, stop loss, take profit) untuk **swing trading 1–4 minggu**.

Modul **Positions** untuk memantau posisi terbuka dari data transaksi riil (buy/sell) di database lokal: menampilkan P&L, posisi direview dengan analisis teknikal + trading plan yang **position-aware** (entry average, stop loss disesuaikan, target take profit berdasarkan avg price).

Modul **Screening** untuk **nightly screening otomatis** seluruh ticker: menjalankan analisis teknikal ringan (indikator + sinyal saja) ke semua emiten di `master` secara paralel, mem-filter yang liquid + non-bearish, mengurutkan berdasarkan skor, lalu melakukan **deep pass** ke 10 besar (indikator lengkap + trading plan + AI insight). Hasil disimpan di database lokal dan bisa diakses via API. Screening berjalan otomatis setiap hari jam 21:00 WIB via in-process goroutine scheduler (tanpa cron dependency eksternal).

> Output analisis & trading plan bersifat sinyal otomatis berbasis data historis, **bukan rekomendasi investasi**. Aplikasi wajib menampilkan disclaimer di UI.

## 2. Tujuan

- User bisa mencari ticker cepat (autocomplete dari `master`)
- User bisa melihat snapshot fundamental (price, 52w high/low, market cap, PE, EPS)
- User bisa melihat chart multi-panel: candlestick + overlay indikator + panel volume/RSI/MACD/Stochastic
- Sistem menghasilkan **skor/sinyal teknikal** dari confluence beberapa indikator dengan bobot berbeda
- Sistem menghasilkan **trading plan** otomatis: entry zone, stop loss, multi-level take profit, risk/reward, position sizing, time-stop
- User bisa mencatat transaksi jual/beli dan melihat posisi terbuka dengan P&L real-time
- Tiap posisi bisa direview dengan analisis teknikal + trading plan yang disesuaikan dengan harga rata-rata posisi
- User bisa mendapatkan **AI insight** terpisah untuk analisis fundamental + sentimen naratif
- Proses analisis end-to-end target < 10 detik
- Data analisis di-cache sementara (sessionStorage 5 menit) untuk menghindari request ulang
- **Nightly screening** otomatis seluruh ticker setiap 21:00 WIB — 10 saham terbaik berdasarkan confluence teknikal siap dilihat setiap pagi

## 3. Non-Goals

- Tidak membangun data warehouse/cache historis sendiri di luar Google Sheets
- Tidak mencakup fitur watchlist/eksekusi order otomatis
- Bukan pengganti nasihat keuangan profesional
- Tidak mengganti sumber data (tetap Google Sheets + GOOGLEFINANCE)
- Tidak membangun backtesting engine
- Tidak ada sistem multi-user terisolasi — tooling internal single-user

## 4. User Flow

### 4.1 Analisis Page (`/analisis`)

1. User membuka menu **Analisis**
2. `onMounted` → fetch semua ticker dari `GET /api/tickers`, simpan di memori
3. User mengetik nama/kode ticker → filter lokal via `computed` (instant)
4. User memilih ticker → memilih rentang tanggal (default: 1 tahun terakhir)
5. User klik "Analisis" → loading state
6. Backend orkestrasi: update config sheet → poll chart data → jalankan Python (indikator, sinyal, trading plan, render PNG)
7. Hasil ditampilkan dalam **tab system**: Review (sinyal + trading plan), Indikator (grid 14 indikator), Snapshot (fundamental), AI Analisis (insight dari LLM), Chart (D3 candlestick + PNG multi-panel)
8. Disclaimer permanen di bagian bawah

### 4.2 Positions Page (`/positions`)

1. User membuka menu **Positions**
2. `onMounted` → `GET /api/positions` — agregasi transaksi buy/sell per ticker dari database
3. Daftar posisi terbuka tampil sebagai cards (ticker, lot, avg price)
4. User klik card → `loadAnalysis(ticker)` → `GET /api/positions/{ticker}/analysis`
5. Hasil analisis ditampilkan di bawah cards dalam **tab system**: Review (P&L + trading plan position-aware + sinyal breakdown), Indikator (14 indikator), Snapshot (fundamental), AI Analisis (insight LLM dengan konteks posisi)
6. Analisis di-cache di sessionStorage (5 menit). User bisa klik "Re-Analyze" untuk refresh.
7. AI insight dipanggil terpisah (client-side) hanya saat tab AI diklik

### 4.3 Screening Page (`/screening`)

1. User membuka menu **Screening** (atau halaman yang menampilkan hasil screening)
2. `onMounted` → `GET /api/screening/latest` — mengambil hasil screening terbaru dari database
3. Tampilkan **Top 10 Saham** dalam tabel: rank, ticker, nama, skor, sinyal (bullish/bearish/netral), confidence, avg volume
4. Tiap baris bisa diklik → expand detail deep pass: indikator lengkap, trading plan, AI insight
5. Jika belum ada screening hari ini → tampilkan data screening terakhir + label "Data kemarin"
6. Admin bisa trigger ulang screening via tombol → `POST /api/screening/trigger`

## 5. Functional Requirements

### 5.1 Backend (Go)

#### Endpoints

| Endpoint | Method | Deskripsi |
|----------|--------|-----------|
| `/api/tickers` | GET | Seluruh daftar ticker dari sheet `master` (cache TTL 5 menit) |
| `/api/ticker/{kode}` | GET | Snapshot fundamental satu ticker dari sheet `google_finance` |
| `/api/analisis` | POST | Body `{ticker, date_start, date_end}` → orkestrasi update config, poll chart, jalankan Python, return hasil lengkap |
| `/api/analisis/ai-insight` | POST | Body `{ticker, date_end, indicators, snapshot, position?, position_review?, signal?}` → LLM (NVIDIA fallback Gemini) → return markdown insight |
| `/api/positions` | GET | Agregasi seluruh posisi terbuka dari tabel `trade_transactions` (buy/sell per ticker) |
| `/api/positions/{ticker}/analysis` | GET | Analisis teknikal + position review untuk satu ticker yang dimiliki |
| `/api/screening/latest` | GET | Hasil screening terbaru (Top 10 + full results) |
| `/api/screening/{date}` | GET | Hasil screening per tanggal tertentu (YYYY-MM-DD) |
| `/api/screening/trigger` | POST | Trigger screening manual (admin) — menjalankan ulang screening+deep pass+AI async |

#### Response `/api/analisis`

```json
{
  "snapshot": { "kode": "ANTM", "company_name": "...", "price": 3000, "high52": 3250, "low52": 2770, "volume": 61092600, "marketcap": 0, "pe": null, "eps": 0, "currency": "IDR" },
  "ohlcv": [{ "date": "2024-01-01", "open": 2900, "high": 3050, "low": 2880, "close": 3000, "volume": 50000000 }],
  "chart_image": "base64 PNG",
  "indicators": { "sma20": 2950, "rsi": 58.5, ... },
  "signal": { "overall": "bullish", "score": 0.45, "confidence": "medium", "trend_filter_passed": true, "breakdown": [...] },
  "trading_plan": { "bias": "buy", "entry_price": 9150, "stop_loss": 8830, "targets": [...], ... },
  "ai_insight": "markdown string",
  "error": ""
}
```

#### Response `/api/positions/{ticker}/analysis`

```json
{
  "ticker": "ANTM",
  "position": { "ticker": "ANTM", "total_lot": 5, "average_buy_price": 2800, "status": "open" },
  "indicators": { ... },
  "signal": { "overall": "bullish", ... },
  "snapshot": { ... },
  "position_review": {
    "buy_price": 2800, "current_price": 3000, "lot": 5,
    "unrealized_pnl": 1000000, "unrealized_pnl_pct": 7.14,
    "holding_days": 45, "recommendation": "hold",
    "suggested_exit_price": 3200, "suggested_stop": 2650,
    "reason": "...",
    "trading_plan": {
      "bias": "buy", "avg_price": 2800, "current_vs_avg_pct": 7.14,
      "entry_zone": { "low": 2750, "high": 2850 },
      "entry_note": "Posisi sudah masuk. Avg 2800, current 3000 (+7.14%)",
      "stop_loss": 2650, "stop_loss_basis": "structure_swing_low",
      "targets": [
        { "level": "BE", "price": 2800, "rr_ratio": 0 },
        { "level": "TP1", "price": 3100, "rr_ratio": 1.0 },
        { "level": "TP2", "price": 3400, "rr_ratio": 2.0 },
        { "level": "TP3", "price": 3700, "rr_ratio": 3.0 }
      ],
      "disclaimer": "..."
    },
    "disclaimer": "..."
  },
  "ai_insight": "markdown string"
}
```

#### Konfigurasi Service

| Field | Default | Keterangan |
|-------|---------|------------|
| `pythonBin` | auto-detect (python3/py/python) | Binary Python |
| `pythonScriptPath` | `./scripts/generate_chart.py` | Path absolut script Python |
| `pollIntervalMs` | 500 | Interval polling sheet chart (ms) |
| `pollMaxRetries` | 20 | Maksimal retry polling |
| `subprocessTimeout` | 120s | Timeout eksekusi Python |

### 5.2 Python — Indikator, Sinyal, Trading Plan & Chart

**Input:** OHLCV dari sheet `chart` via temp JSON file. Kolom wajib: `date, open, high, low, close, volume`. Minimal 30 bar data, urutan ascending.

**CLI (generate_chart.py):** `python generate_chart.py --input <temp.json> --ticker <KODE> --out <chart.png> [--equity <modal>]`

**CLI (run_position_review.py):** `python run_position_review.py --input <temp.json> --ticker <KODE> --buy-price <harga> --lot <lot> --buy-date <YYYY-MM-DD>`

**CLI (screening_pass.py):** `python screening_pass.py --input <temp.json>` — **lightweight, <1s/ticker**. Hanya indikator + sinyal (tanpa trading plan, tanpa chart). Output JSON: `{indicators, signal, avg_volume}`.

**CLI (deep_pass.py):** `python deep_pass.py --input <temp.json> --ticker <KODE>` — indikator + sinyal + **trading plan** (tanpa chart). Output JSON: `{indicators, signal, trading_plan}`.

**Stdout:** JSON `{ indicators: {...}, signal: {...}, trading_plan: {...} }` untuk generate_chart. `{ position_review: {...} }` untuk run_position_review.

#### Indikator

| Grup | Indikator |
|------|-----------|
| Trend | SMA20/50/200, EMA20/50, ADX + DI+/DI- |
| Momentum | RSI(14), MACD (12,26,9) + signal + histogram, Stochastic %K/%D |
| Volatilitas | Bollinger Bands (20,2), ATR(14) |
| Volume | OBV, Volume MA20, volume spike flag |
| Price Levels | Support/Resistance (swing high/low), Fibonacci retracement (23.6/38.2/50/61.8) |

#### Sinyal (Confluence-based)

- Tiap indikator memberi vote: bullish (+1) / bearish (-1) / netral (0)
- **Bobot berbeda**: indikator tren > oscillator jangka pendek
- Filter tren jangka panjang: `trend_filter_passed = bool(close > SMA200)`
- Skor total dinormalisasi → `bullish`/`bearish`/`netral`
- `confidence`: `high`/`medium`/`low` berdasarkan kontribusi + konsistensi

#### Trading Plan

**Mode Fresh Entry (generate_chart.py):**
- **Bias**: `buy`/`sell`/`hold` dari signal overall
- **Entry price**: area support terdekat / pullback ke SMA20/EMA20
- **Stop loss**: bawah support kunci / swing low, buffer ATR × 2.0–2.5
- **Take profit**: multi-level (1.5R, 2.5R, 3.5R)
- **Time stop**: 20 hari bursa
- **Position sizing**: `suggested_position_size_pct` (1-2% risk), `suggested_lots` (nullable)

**Mode Position-Aware (run_position_review.py):**
Semua field di atas **ditambah** konteks posisi:
- **avg_price**, **current_vs_avg_pct** — status posisi vs harga rata-rata
- **Entry zone** dihitung dari avg price, bukan current price
- **Targets**: BE (avg price) + TP1/TP2/TP3 dari range avg→SL
- **Entry note** menyebut status posisi (untung/rugi, avg price, current price)

#### Chart Render (multi-panel matplotlib PNG)

1. Panel utama: candlestick + SMA20/50/200 + EMA20/50 + BB + S/R + Fibonacci + entry/SL/TP
2. Volume bar + Volume MA20 + volume spike highlight
3. RSI + 30/70
4. MACD + signal + histogram
5. Stochastic + 20/80

### 5.3 Frontend — Analisis Page (`/analisis`)

- **Search box** autocomplete — fetch semua ticker sekali (`onMounted`), filter lokal
- **Date range picker** (`<input type="date">`, default 1 tahun)
- **Tombol "Analisis"** + loading state
- **Tab system** (5 tabs di atas area hasil):
  1. **Review** — signal badge + score + confidence + breakdown + trading plan (bias, entry, SL, multi-level TP table, position size, time stop, invalidation note) + disclaimer
  2. **Indikator** — grid 14 indikator: SMA20/50/200, EMA20, RSI, MACD + signal, ADX + DI+/DI-, Stochastic %K/%D, BB Width, ATR, volume spike, S/R, Fibonacci levels
  3. **Snapshot** — price, 52w high/low, market cap, PE, EPS, volume, currency
  4. **AI Analisis** — button "Generate Analisis AI" (kalo belum ada) / rendered markdown insight
  5. **Chart** — D3 interactive candlestick SVG (pan, zoom, crosshair tooltip) + Python multi-panel chart PNG
- **Nilai `#N/A` / null** → "N/A"
- **Disclaimer** permanen di bagian bawah

### 5.4 Frontend — Positions Page (`/positions`)

- **Daftar posisi** — cards grid (ticker, lot, avg price, P&L badge)
- **Klik card** → `loadAnalysis(ticker)` → analisis + position review tampil di bawah
- **Cache sessionStorage** — data analisis disimpan 5 menit, indicator "cached" di UI
- **Tombol "Re-Analyze"** — hapus cache, fetch ulang
- **Tab system** (4 tabs di atas area hasil):
  1. **Review** — P&L cards (unrealized P&L, %, holding, recommendation) + trading plan position-aware (avg_price, entry_zone, SL, multi-level TP table dengan BE) + sinyal breakdown expandable
  2. **Indikator** — grid 14 indikator (sama dengan halaman analisis)
  3. **Snapshot** — fundamental data (price, 52w, market cap, PE, EPS, volume)
  4. **AI Analisis** — jika `ai_insight` sudah ada di response → rendered markdown. Jika tidak → button "Generate Analisis AI" yang memanggil `/api/analisis/ai-insight` dengan konteks posisi (position, position_review, signal)

### 5.5 Nightly Screening Service

**Tujuan:** Setiap malam jam 21:00 WIB, sistem melakukan screening ke **seluruh ticker** di `master` untuk menemukan 10 saham terbaik secara teknikal.

**Alur:**

1. **Screening Pass** (seluruh ticker, paralel via goroutines):
   - Fetch OHLCV 1 tahun via `FetchOHLCV()` (reuse worker pool dari analisis service)
   - Jalankan `screening_pass.py` (indikator + sinyal saja — ringan, <1s/ticker)
   - Kumpulkan hasil: `{ticker, score, overall, confidence, avg_volume}`
   
2. **Filter & Sort:**
   - Filter: `avg_volume >= minAvgVolume` (1,000,000) — hanya saham liquid
   - Filter: exclude `overall === 'bearish'` — hanya bullish/netral
   - Sort by `score` descending, ambil **Top 10**

3. **Deep Pass** (Top 10, sequential — karena mahal):
   - Jalankan `deep_pass.py` (indikator + sinyal + **trading plan**)
   - Panggil AI insight (`GenerateInsight`) untuk tiap ticker
   - Kumpulkan: `{indicators, signal, trading_plan, ai_insight}`

4. **Simpan:**
   - `daily_screening_results` table: satu row per ticker per tanggal
   - Field: `ticker`, `score`, `overall`, `confidence`, `avg_volume`, `rank` (NULL untuk non-top-10), `indicators_json`, `signal_json`, `trading_plan_json`, `ai_insight`
   - **Upsert** (ON CONFLICT DO UPDATE) — aman di-trigger ulang

**Worker pool:** Screening pass menggunakan worker pool yang sama dengan analisis service (10 slot). Jika pool penuh, screening pass menunggu.

**Scheduler:** In-process goroutine tanpa cron dependency:
- Hitung waktu ke 21:00 WIB berikutnya → `time.After()`
- Trigger screening
- Loop ke hari berikutnya

**Manual trigger:** `POST /api/screening/trigger` menjalankan ulang screening kapan saja.

#### Database Model: `daily_screening_results`

| Column | Type | Description |
|--------|------|-------------|
| id | BIGSERIAL | Primary key |
| ticker | VARCHAR(10) | Kode saham |
| screening_date | DATE | Tanggal screening |
| score | FLOAT | Normalized confluence score (-1 to 1) |
| overall | VARCHAR(10) | bullish/bearish/netral |
| confidence | VARCHAR(10) | high/medium/low |
| avg_volume | BIGINT | Rata-rata volume harian |
| rank | INT | Peringkat (1-10 untuk top 10, NULL untuk non-top-10) |
| indicators_json | JSONB | Output lengkap indicators (null untuk screening pass) |
| signal_json | JSONB | Output lengkap signal |
| trading_plan_json | JSONB | Trading plan (null untuk non-top-10) |
| ai_insight | TEXT | Markdown AI insight (null untuk non-top-10) |
| is_deep_pass | BOOLEAN | Apakah ticker masuk deep pass |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

**Index:** UNIQUE(ticker, screening_date), INDEX(screening_date, rank)

### 5.6 AI Insight

- **Endpoint**: `POST /api/analisis/ai-insight`
- **Backend**: panggil NVIDIA DeepSeek V4 Flash → fallback Gemini 3.6 Flash
- **Input opsional konteks posisi**: `position`, `position_review`, `signal` — untuk insight yang relevan dengan posisi user
- **Dua mode prompt**:
  - **Form A** (data punya signal + trading plan): AI narrate hasil kalkulasi sistem
  - **Form B** (raw indicators only): AI analisis mandiri + bangun trading plan
- **Cache in-memory**: per `ticker:date` key, expire end-of-day
- **Frontend**: insight dirender sebagai HTML (bold, paragraph, line breaks). Auto-load jika sudah ada di response analisis. Dipanggil manual via tombol jika tidak ada.

## 6. Perbaikan & Peningkatan Implementasi

### A. Python — Chart & Indikator

| Item | Deskripsi |
|------|-----------|
| A1 | Overlay SMA/EMA/BB di panel utama chart |
| A2 | Gambar support/resistance & Fibonacci sebagai horizontal line + label |
| A3 | `_swing_points()` return `(index, harga)` — marker swing tidak pernah salah tempat |

### B. Python — Logika Sinyal & Trading Plan

| Item | Deskripsi |
|------|-----------|
| B1 | Bobot sinyal berdasar relevansi horizon (tren > oscillator) |
| B2 | Filter tren jangka panjang: `trend_filter_passed` dari SMA200 |
| B3 | `confidence: high/medium/low` |
| B4 | ATR multiplier configurable, default 2.0–2.5× |
| B5 | `time_stop_days` default 20 |
| B6 | `suggested_lots` (nullable, butuh `--equity`) |
| B7 | **Position-aware trading plan**: avg_price, current_vs_avg_pct, BE target, position-based entry_note |

### C. Python — Robustness

| Item | Deskripsi |
|------|-----------|
| C1 | `_safe_series()` log exception, jangan silent-swallow |
| C2 | Validasi input: kolom wajib, urutan ascending, min 30 bar → error JSON jelas |
| C3 | Pin versi dependency di `requirements.txt` |

### D. Go — Integrasi & Robustness

| Item | Deskripsi |
|------|-----------|
| D1 | `pythonBin` configurable dari env, auto-detect python3/py/python |
| D2 | Propagasi `ctx` ke `GetSheet` / `SearchTickers` / `GetTicker` |
| D3 | `os.MkdirTemp` per-request ganti `os.TempDir()` manual |
| D4 | Timeout subprocess configurable dari constructor |
| D5 | Jangan swallow error `GetTicker` — minimal log warning |
| D6 | Field baru Go struct: `Confidence`, `TrendFilterPassed`, `TimeStopDays`, `SuggestedLots`, `Snapshot` |
| D7 | `gofmt` konsisten |

### E. AI Insight

| Item | Deskripsi |
|------|-----------|
| E1 | NVIDIA DeepSeek V4 Flash sebagai primary, Gemini fallback |
| E2 | Cache in-memory per `ticker:date` dengan daily expiry |
| E3 | Konteks posisi dikirim ke AI (`position`, `position_review`, `signal`) untuk insight relevan — `AiInsightRequest` memiliki field `Position`, `PositionReview`, `SignalResult` |
| E4 | Auto AI di backend **dihapus** dari `GetPositionAnalysis` — AI hanya dipanggil client-side via tab |

### F. Frontend — UI/UX

| Item | Deskripsi |
|------|-----------|
| F1 | **Tab system** di halaman analisis dan positions (review, indikator, snapshot, AI, chart) |
| F2 | Sticky bottom bar diganti jadi tab bar di **atas** area hasil |
| F3 | `isAdmin` gates dihapus — semua user lihat semua tab |
| F4 | **SessionStorage cache** untuk analisis posisi (5 menit expiry) |
| F5 | **Re-Analyze button** untuk refresh paksa |
| F6 | **D3 chart** redraw otomatis saat tab chart diaktifkan (fix zero-width SVG) |
| F7 | AI insight pre-loaded dari response, fallback ke generate button |

### G. Nightly Screening

| Item | Deskripsi |
|------|-----------|
| G1 | `screening_pass.py` — lightweight script (indicators + signals only, <1s/ticker) untuk screening massal |
| G2 | `deep_pass.py` — medium script (indicators + signals + trading_plan, no chart) untuk top 10 |
| G3 | `fetchOHLCV()` diekstrak dari `RunAnalisis` jadi method sendiri, reusable untuk screening |
| G4 | Screening pass **semua ticker paralel** via goroutines — tiap goroutine acquire worker pool slot |
| G5 | Filter liquid (min volume 1jt) + non-bearish, sort score → Top 10 |
| G6 | Deep pass sequential (AI insight mahal) — indicators + trading_plan + AI untuk top 10 |
| G7 | **Upsert** (`ON CONFLICT DO UPDATE`) — aman trigger ulang tanpa duplikasi |
| G8 | **Scheduler in-process** — hitung next 21:00 WIB, `time.After()`, trigger, loop. Tidak perlu cron external |
| G9 | Manual trigger via `POST /api/screening/trigger` |
| G10 | Tabel `daily_screening_results` dengan JSONB untuk indicators/signal/trading_plan — flexible, no schema changes saat indikator bertambah |

## 7. Keputusan Implementasi

1. **Timing refresh GOOGLEFINANCE** — Backend **poll** sheet `chart` setiap 500ms, max 20× (10s timeout). Jika timeout, return error.
2. **Handling `#N/A`** — Tampil "N/A" di UI. Nilai numerik `0` via `*float64` untuk bedakan N/A vs 0 riil.
3. **Concurrency** — `config.selected_ticker` adalah shared single cell. Tooling **single-user**, aman. Upgrade path: per-session sheet atau queue.
4. **Python output** — stdout JSON + PNG file via `--out`. Go baca PNG sebagai base64.
5. **Python discovery** — Auto-detect: `python3` → `py` → `python`, diverifikasi dengan `--version`.
6. **Error propagation** — Python error (field `error` di stdout JSON) diteruskan ke response API.
7. **AI insight dipisah** — Tidak blocking analisis utama. Dipanggil terpisah hanya saat diminta user.
8. **Cache analisis** — sessionStorage frontend (5 menit), bukan backend. Hindari re-fetch data yang sama dalam短期.
9. **Position-aware trading plan** — Target TP dihitung dari avg price, bukan current price. BE target selalu disertakan.
10. **Nightly screening parallel** — Semua ticker di-screen paralel via goroutines. Worker pool (10 slot) mencegah overload Python subprocess. Filter liquid + non-bearish sebelum sort. Deep pass + AI hanya untuk top 10 — hemat resource.

## 8. Metrik Keberhasilan

- User bisa selesaikan satu siklus pencarian → chart, indikator, sinyal, trading plan tampil tanpa error
- Waktu end-to-end rata-rata < 10 detik
- Autocomplete terasa instant (filter lokal, tanpa network)
- Tidak ada crash saat data mengandung `#N/A`, data historis pendek, atau ticker tidak lengkap
- Trading plan konsisten matematis (RR ratio, SL/TP wajar relatif terhadap ATR & support/resistance)
- Position-aware trading plan menampilkan avg price dan target yang relevan dengan posisi user
- Chart PNG menampilkan seluruh overlay dan anotasi dengan benar
- AI insight mengandung konteks posisi (bukan analisis umum) saat dipanggil dari halaman positions
- Nightly screening selesai < 30 menit untuk seluruh ticker (parallel screening pass)
- Top 10 menampilkan ticker yang liquid, trending, dan memiliki skor confluence tinggi
- Deep pass memberikan trading plan + AI insight yang relevan untuk tiap ticker di top 10
