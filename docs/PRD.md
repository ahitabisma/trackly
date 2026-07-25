# PRD: Fitur Menu "Analisis" (Search Ticker + Visualisasi Chart)

## 1. Latar Belakang

Aplikasi saat ini sudah punya frontend + backend (Go) yang terhubung ke Google Sheets (via Apps Script API) sebagai sumber data Google Finance, dengan 4 sheet utama:

- `master` — daftar seluruh emiten (Kode, Nama Perusahaan, Tanggal Pencatatan, Saham, Papan Pencatatan)
- `config` — key-value control cell: `selected_ticker`, `date_start`, `date_end`
- `google_finance` — snapshot harga & fundamental per ticker (price, high52, low52, volume, marketcap, pe, eps, currency, company_name)
- `chart` — data historis OHLCV mengikuti ticker & rentang tanggal yang aktif di `config`

Dibutuhkan menu baru **Analisis** di mana user bisa mencari ticker, memilih rentang tanggal, lalu melihat grafik historis beserta indikator teknikal yang di-generate via Python.

## 2. Tujuan

- User bisa mencari ticker dengan cepat (autocomplete dari `master`)
- User bisa melihat snapshot fundamental terkini (price, 52w high/low, market cap, PE, EPS)
- User bisa melihat grafik historis (candlestick/line) beserta indikator dasar (MA20/MA50, RSI, MACD, support-resistance)
- Proses end-to-end (search → update config → refresh sheet → generate chart) berjalan dalam waktu yang wajar (target < 5 detik)

## 3. Non-Goals

- Tidak membuat sistem multi-user/session terisolasi di versi awal (lihat Open Decision #3)
- Tidak membangun data warehouse/cache historis sendiri di luar Google Sheets pada versi awal
- Tidak mencakup fitur watchlist/portfolio di PRD ini

## 4. User Flow

1. User membuka menu **Analisis**
2. `onMounted` → frontend fetch semua ticker dari `GET /api/tickers` (tanpa query param), simpan di memori
3. User mengetik nama/kode ticker → filter lokal via `computed` (instant, tanpa debounce/API call)
4. User memilih ticker → memilih rentang tanggal (default: 1 tahun terakhir)
5. User klik "Analisis" → tampil loading state
6. Backend:
   - Update `config.selected_ticker`, `date_start`, `date_end` via Apps Script (`POST ?action=setValueByKey`)
   - Poll `?action=chart` setiap 500ms (max 20× = 10s) sampai data tersedia
   - Ambil data `chart` terbaru
   - Jalankan Python (stdout JSON) untuk hitung indikator MA20, MA50, RSI, MACD, support/resistance
7. Frontend menampilkan: snapshot fundamental + D3 candlestick chart (SVG) + ringkasan indikator + tabel OHLCV

## 5. Functional Requirements

### 5.1 Backend (Go)

| Endpoint              | Method | Deskripsi                                                                                            |
| --------------------- | ------ | ---------------------------------------------------------------------------------------------------- |
| `/api/tickers`        | GET    | Seluruh daftar ticker dari sheet `master` (cache TTL 5 menit)                                        |
| `/api/ticker/{kode}`  | GET    | Snapshot dari sheet `google_finance`                                                                 |
| `/api/analisis`       | POST   | Body `{ticker, date_start, date_end}` → orkestrasi update config, poll chart, panggil Python, hasil  |

Apps Script URL pattern: `?action=master`, `?action=google_finance`, `?action=chart`, `?action=setValueByKey` (POST JSON body `{key, value}`).

Response `/api/analisis`:

```json
{
  "snapshot": {
    "kode": "ANTM",
    "company_name": "...",
    "price": 3000,
    "high52": 3250,
    "low52": 2770,
    "volume": 61092600,
    "marketcap": 0,
    "pe": null,
    "eps": 0,
    "currency": "IDR"
  },
  "ohlcv": [
    {"date": "2025-07-25", "open": 3040, "high": 3040, "low": 2950, "close": 2970, "volume": 61092600}
  ],
  "indicators": {
    "ma20": 0,
    "ma50": 0,
    "rsi": 0,
    "macd": 0,
    "support": 0,
    "resistance": 0
  }
}
```

### 5.2 Python (indikator + sinyal + trading plan + chart)

- Input: JSON file array of objects `[{date, open, high, low, close, volume}]`
- Dependencies: `pandas`, `numpy`, `ta` (technical analysis library), `matplotlib`, `mplfinance`
- CLI: `python generate_chart.py --input <temp.json> --ticker <KODE> --out <chart.png>`
- stdout: JSON `{indicators: {...}, signal: {...}, trading_plan: {...}}`
- Modul:
  - `indicators.py` — SMA20/50/200, EMA20/50, ADX+DI, RSI, MACD+signal+histogram, Stochastic %K/%D, Bollinger Bands (upper/mid/lower/width), ATR, OBV, Volume MA20, volume spike flag, swing high/low S/R, Fibonacci retracement
  - `signals.py` — rule-based confluence: tiap indikator vote bullish(+1)/bearish(-1)/neutral(0), skor dinormalisasi, output `overall: bullish/bearish/neutral` + breakdown per indikator
  - `trading_plan.py` — bias (buy/sell/hold) dari signal overall, entry zone, stop loss (ATR×1.5), multi-level TP (1.5R/2.5R/3.5R), R:R ratio tiap target, suggested position size (1% risk), invalidation note, disclaimer permanen
  - `chart_renderer.py` — multi-panel matplotlib PNG: (1) candlestick + SMA/EMA/BB + swing markers + entry/SL/TP lines, (2) volume + MA, (3) RSI + 30/70, (4) MACD + signal + histogram, (5) Stochastic + 20/80

### 5.3 Frontend

- Search box dengan autocomplete — **fetch semua ticker sekali** (`onMounted`), filter lokal via `computed` (instant)
- Date range picker (`<input type="date">`, default 1 tahun terakhir)
- Tombol "Analisis" + loading state
- Area hasil (berurutan):
  1. **Snapshot card** — price, 52w high/low, market cap, PE, EPS, currency
  2. **D3 candlestick chart** (SVG interaktif, tooltip hover)
  3. **Multi-panel chart** dari Python (base64 PNG)
  4. **Signal badge** — overall (bullish/bearish/neutral) + skor + expandable breakdown per indikator
  5. **Indikator grid** — SMA20/50/200, EMA20, RSI, MACD, ADX, ATR, BB Width, volume spike, S/R, Fibonacci
  6. **Trading plan card** — bias, entry price, stop loss, TP multi-level table (dengan R:R ratio), ukuran posisi (%), invalidation note
  7. **Disclaimer** — permanen, non-dismissible, selalu terlihat di bawah trading plan
  8. **OHLCV table** — collapsible (`<details>`)
- Handling untuk nilai `#N/A` / null (tampil sebagai "N/A")

## 6. Non-Functional Requirements

- Response time end-to-end target < 10 detik (poll loop 500ms × 20 retries)
- Autocomplete search *instant* karena filter lokal (tanpa network)
- Error handling jelas jika Apps Script/Google Sheets timeout atau ticker tidak ditemukan

## 7. Keputusan Implementasi

1. **Timing refresh `GOOGLEFINANCE`** — Apps Script perlu waktu recalculate. Backend **poll** sheet `chart` setiap 500ms, max 20× (total 10s timeout). Jika timeout, return error.
2. **Handling `#N/A`** — ditampilkan sebagai **"N/A"** di UI. Nilai numerik `0` yang berasal dari `#N/A` tetap ditampilkan sebagai "N/A" (dibedakan dari nilai 0 asli via `*float64` untuk field PE).
3. **Concurrency** — `config.selected_ticker` adalah shared single cell. Tooling internal **single-user**, aman diabaikan dulu. Upgrade path: per-session sheet atau queue per request.
4. **Python output** — stdout JSON (`indicators` + `signal` + `trading_plan`) + **PNG file** (multi-panel chart via `--out`). Go baca PNG sebagai base64 dan kirim ke frontend.
5. **Frontend chart** — 2 chart: D3 candlestick (SVG interaktif) + Python multi-panel PNG (lengkap dengan overlay indikator dan trading plan).

## 8. Metrik Keberhasilan

- User bisa menyelesaikan satu siklus pencarian → chart & indikator tampil tanpa error
- Ticker autocomplete terasa instant (filter lokal tanpa network)
- Tidak ada crash saat data mengandung `#N/A` atau ticker tidak lengkap
