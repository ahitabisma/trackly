# PRD: Fitur Menu "Analisis" (Search Ticker + Visualisasi Chart + Rekomendasi Trading)

## 1. Latar Belakang

Aplikasi saat ini sudah punya frontend + backend (Go) yang terhubung ke Google Sheets (via Apps Script API) sebagai sumber data Google Finance, dengan 4 sheet utama:

- `master` — daftar seluruh emiten (Kode, Nama Perusahaan, Tanggal Pencatatan, Saham, Papan Pencatatan)
- `config` — key-value control cell: `selected_ticker`, `date_start`, `date_end`
- `google_finance` — snapshot harga & fundamental per ticker (price, high52, low52, volume, marketcap, pe, eps, currency, company_name)
- `chart` — data historis OHLCV mengikuti ticker & rentang tanggal yang aktif di `config`

Dibutuhkan menu baru **Analisis** di mana user bisa mencari ticker, memilih rentang tanggal, lalu melihat grafik historis lengkap dengan indikator teknikal, serta output **trading plan** (entry, stop loss, take profit) sebagai bantuan keputusan.

> **Catatan penting**: Output analisis teknikal & trading plan bersifat sinyal otomatis berbasis indikator historis, bukan rekomendasi investasi/finansial yang terjamin. Aplikasi wajib menampilkan disclaimer ini di UI.

## 2. Tujuan

- User bisa mencari ticker dengan cepat (autocomplete dari `master`)
- User bisa melihat snapshot fundamental terkini (price, 52w high/low, market cap, PE, EPS)
- User bisa melihat 1 chart komprehensif berisi price action + overlay indikator + panel indikator terpisah
- Sistem menghasilkan **skor/sinyal teknikal** (bullish/bearish/netral) dari confluence beberapa indikator
- Sistem menghasilkan **trading plan** otomatis: entry zone, stop loss, take profit (multi-target), risk/reward ratio, position sizing
- Proses end-to-end (search → update config → refresh sheet → generate chart) berjalan dalam waktu yang wajar (target < 5 detik)

## 3. Non-Goals

- Tidak membuat sistem multi-user/session terisolasi di versi awal (lihat Open Decision #3)
- Tidak membangun data warehouse/cache historis sendiri di luar Google Sheets pada versi awal
- Tidak mencakup fitur watchlist/portfolio/eksekusi order otomatis di PRD ini
- Bukan pengganti nasihat keuangan profesional — trading plan yang dihasilkan adalah output algoritmik dari data historis

## 4. User Flow

1. User membuka menu **Analisis**
2. User mengetik nama/kode ticker di search box → autocomplete muncul (`Kode — Nama Perusahaan`)
3. User memilih ticker → memilih rentang tanggal (default: 1 tahun terakhir)
4. User klik "Analisis" → tampil loading state
5. Backend:
   - Update `config.selected_ticker`, `date_start`, `date_end` via Apps Script
   - Tunggu/poll sampai sheet `chart` selesai recalculate
   - Ambil data `chart` terbaru
   - Jalankan Python untuk generate chart image, hitung seluruh indikator, hitung sinyal, dan generate trading plan
6. Frontend menampilkan: snapshot fundamental → chart lengkap dengan indikator → ringkasan sinyal (bullish/bearish/netral + skor) → trading plan (entry/SL/TP/RR) → disclaimer

## 5. Functional Requirements

### 5.1 Backend (Go)

| Endpoint            | Method | Deskripsi                                                                                                                                                         |
| ------------------- | ------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `/api/tickers`   | GET    | Autocomplete ticker dari sheet `master`, in-memory filter + cache TTL                                                                                             |
| `/api/analisis`     | POST   | Body `{ticker, date_start, date_end}` → orkestrasi update config, refresh chart, panggil Python, return hasil lengkap (chart + indikator + sinyal + trading plan) |

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
  "chart_image": "base64 atau URL",
  "indicators": {
    "trend": {
      "sma20": 0,
      "sma50": 0,
      "sma200": 0,
      "ema20": 0,
      "ema50": 0,
      "adx": 0,
      "di_plus": 0,
      "di_minus": 0
    },
    "momentum": {
      "rsi14": 0,
      "macd": 0,
      "macd_signal": 0,
      "macd_hist": 0,
      "stoch_k": 0,
      "stoch_d": 0
    },
    "volatility": {
      "bollinger_upper": 0,
      "bollinger_mid": 0,
      "bollinger_lower": 0,
      "atr14": 0
    },
    "volume": {
      "obv": 0,
      "volume_sma20": 0
    },
    "price_levels": {
      "support": [0, 0],
      "resistance": [0, 0],
      "fibonacci": { "0.236": 0, "0.382": 0, "0.5": 0, "0.618": 0, "0.786": 0 }
    }
  },
  "signal": {
    "overall": "bullish | bearish | netral",
    "score": 0,
    "breakdown": [
      { "indicator": "MA Crossover", "signal": "bullish", "note": "..." },
      { "indicator": "RSI", "signal": "netral", "note": "..." },
      { "indicator": "MACD", "signal": "bullish", "note": "..." }
    ]
  },
  "trading_plan": {
    "bias": "buy | sell | hold",
    "entry_zone": { "low": 0, "high": 0 },
    "stop_loss": 0,
    "take_profit": [
      { "level": 1, "price": 0, "rr_ratio": 1.5 },
      { "level": 2, "price": 0, "rr_ratio": 2.5 },
      { "level": 3, "price": 0, "rr_ratio": 3.5 }
    ],
    "risk_reward_ratio": 0,
    "suggested_position_size_pct": 0,
    "invalidation_note": "Rencana batal jika harga close di bawah stop loss / support kunci"
  },
  "disclaimer": "Analisis ini dihasilkan otomatis dari data historis, bukan rekomendasi finansial."
}
```

### 5.2 Python (chart, indikator & trading plan)

**Input**: data OHLCV dari sheet `chart` (format: `Date, Open, High, Low, Close, Volume`)

**Indikator yang dihitung (selengkap mungkin, dikelompokkan):**

- **Trend**: SMA 20/50/200, EMA 20/50, ADX + DI+/DI- (kekuatan tren)
- **Momentum**: RSI(14), MACD (12,26,9) + histogram, Stochastic Oscillator (%K/%D)
- **Volatilitas**: Bollinger Bands (20, 2 std dev), ATR(14)
- **Volume**: OBV (On-Balance Volume), Volume MA(20), deteksi volume spike
- **Price levels**: Support/Resistance (dari swing high/low lokal), Fibonacci Retracement dari swing terakhir

**Chart output** (satu figure komprehensif, multi-panel):

1. Panel utama: candlestick + overlay SMA20/50/200, EMA20/50, Bollinger Bands, garis support/resistance, level Fibonacci
2. Panel volume: bar volume + Volume MA20, highlight volume spike
3. Panel RSI (dengan garis 30/70)
4. Panel MACD (line + signal + histogram)
5. Panel Stochastic (dengan garis 20/80)
6. Anotasi entry/SL/TP dari trading plan langsung di panel utama (garis horizontal + label)

**Logika sinyal (confluence-based, sederhana & transparan):**

- Tiap indikator memberi vote: bullish (+1) / bearish (-1) / netral (0)
- Contoh aturan:
  - MA: harga > SMA50 > SMA200 → bullish; sebaliknya bearish
  - RSI: <30 oversold (bullish bias), >70 overbought (bearish bias), 30-70 netral
  - MACD: MACD line > signal line & histogram naik → bullish
  - Stochastic: %K silang %D di bawah 20 → bullish; di atas 80 → bearish
  - ADX > 25 -> tren kuat (menguatkan arah sinyal), ADX < 20 -> tren lemah (turunkan confidence)
- Skor total dinormalisasi -> label `bullish`/`bearish`/`netral`

**Logika trading plan:**

- **Entry zone**: berbasis area support terdekat (untuk buy) atau resistance terdekat (untuk sell), atau pullback ke SMA20/EMA20 saat tren kuat
- **Stop loss**: di bawah support kunci / swing low terakhir, dengan buffer dari ATR(14) (misal `support - 1xATR`)
- **Take profit**: multi-level berdasarkan resistance berikutnya dan/atau kelipatan risk (1.5R, 2.5R, 3.5R)
- **Risk/reward ratio**: `(take_profit - entry) / (entry - stop_loss)`
- **Position sizing**: dihitung sebagai persentase modal berdasarkan risk per trade standar (misal 1-2% dari modal per trade), backend/Python hanya memberi rekomendasi persentase, bukan nominal (karena modal user tidak diketahui)

**Output**: PNG chart (multi-panel) + JSON berisi `indicators`, `signal`, `trading_plan` sesuai schema di atas

```
python3 generate_chart.py --input chart_data.json --ticker ANTM --out chart_ANTM.png
```

### 5.3 Frontend

- Search box dengan autocomplete (debounced)
- Date range picker (default 1 tahun terakhir)
- Tombol "Analisis" + loading state
- Area hasil:
  - Snapshot card (price, 52w high/low, market cap, PE, EPS)
  - Chart multi-panel (image dari backend)
  - Ringkasan sinyal: badge bullish/bearish/netral + skor + breakdown per indikator
  - Kartu **Trading Plan**: bias, entry zone, stop loss, target profit (multi-level + RR ratio), saran ukuran posisi (%), catatan invalidasi
  - Disclaimer permanen di bawah trading plan: "Bukan nasihat finansial, gunakan sebagai referensi tambahan"
- Handling untuk nilai `#N/A` (tampil sebagai "N/A", bukan crash/kosong)

## 6. Non-Functional Requirements

- Response time end-to-end target < 5 detik (bergantung latency Google Sheets recalculation + kompleksitas perhitungan indikator)
- Autocomplete search < 300ms
- Error handling jelas jika Apps Script/Google Sheets timeout, ticker tidak ditemukan, atau data historis terlalu pendek untuk menghitung indikator (misal SMA200 butuh minimal 200 data point - beri fallback/pesan jelas)

## 7. Open Decisions (perlu jawaban sebelum implementasi)

1. **Timing refresh `GOOGLEFINANCE`** - apakah Apps Script sudah bisa force-recalculate (`SpreadsheetApp.flush()`), atau backend perlu retry/poll loop menunggu sheet `chart` ter-update?
2. **Handling `#N/A`** - ditampilkan sebagai "N/A" di UI, atau field tersebut di-exclude dari response?
3. **Concurrency** - `config.selected_ticker` adalah shared single cell. Apakah ini tooling internal single-user (aman diabaikan dulu), atau perlu isolasi per-request (queue/sheet per session)?
4. **Bobot/aturan sinyal** - apakah aturan confluence di atas (vote per indikator) sudah sesuai preferensi, atau ada aturan/threshold spesifik yang mau dipakai (misal RSI 20/80 alih-alih 30/70)?
5. **Risk per trade** - persentase default untuk position sizing (1%? 2%?) - perlu dikonfirmasi atau dibuat configurable di settings.

## 8. Metrik Keberhasilan

- User bisa menyelesaikan satu siklus pencarian -> tampil chart + trading plan tanpa error
- Waktu rata-rata end-to-end di bawah target yang disepakati
- Tidak ada crash saat data mengandung `#N/A`, data historis pendek, atau ticker tidak lengkap
- Trading plan yang dihasilkan konsisten secara matematis (RR ratio, SL/TP masuk akal relatif terhadap ATR & support/resistance)
