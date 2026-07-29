# PRD — Riwayat Trading (Neon Postgres), Analisis Posisi, API Hermes, AI Insight

## 1. Latar Belakang

Fitur Analisis (search ticker → indikator → signal → trading plan) sudah
jalan, datanya lewat Google Sheets + Apps Script. Untuk fitur riwayat
trading, data **TIDAK** disimpan di Sheets — pakai **Neon (Postgres
serverless)** sebagai database terpisah, karena data transaksi butuh
query relasional (agregasi per ticker, hitung average cost, dst) yang
lebih pas di database beneran daripada spreadsheet.

## 2. Tujuan

1. User bisa mencatat transaksi trading (beli & jual) lewat percakapan
   natural di Telegram (via Hermes) atau frontend, disimpan di Neon.
2. Kalau user beli ticker yang SAMA lebih dari sekali sebelum posisinya
   ditutup total, sistem otomatis **rata-ratakan (averaging)** harga
   beli untuk keperluan analisis posisi — TAPI tiap transaksi beli/jual
   tetap **dicatat terpisah sebagai baris sendiri** (ledger, bukan
   di-merge jadi 1 baris). Rata-rata dihitung on-the-fly dari ledger,
   bukan disimpan sebagai angka independen yang bisa nggak sinkron.
3. User bisa minta sistem menganalisis posisi yang sedang dipegang
   (hold atau sell, di harga berapa) — pakai mesin analisis yang sama
   (indicators/signals) tapi logic keputusan beda dari analisis entry
   baru, DI MENU TERPISAH dari menu Analisis.
4. Hermes akses semua ini lewat API dengan token bearer tetap yang
   **khusus role admin** (baca & tulis penuh) — user generate sendiri.
5. Menu Analisis dan menu Riwayat sama-sama dapat **AI insight** dari
   NVIDIA NIM (model `deepseek-ai/deepseek-v4-flash`, free tier).

## 3. Non-Goals

- Eksekusi order otomatis ke broker.
- Multi-user dengan role selain admin di iterasi ini (skema token role
  disiapkan extensible, tapi cuma 1 role yang dipakai sekarang).
- Short selling / margin trading (tetap long-only, konsisten dengan
  `trading_plan.py`).
- Perhitungan pajak/fee broker — P&L yang dihitung gross.
- FIFO/LIFO costing — averaging (weighted average cost) yang dipakai,
  bukan FIFO.

## 4. User Stories

- **US1**: "Saya beli RAJA 3 lot di 880 tanggal 20 Juli" → tercatat 1
  transaksi `buy` di Neon.
- **US2**: "Saya beli RAJA lagi 2 lot di 850 tanggal 22 Juli" → tercatat
  transaksi `buy` KEDUA (baris terpisah), tapi kalau saya tanya "posisi
  RAJA saya gimana", sistem kasih tau average buy price gabungan dari
  kedua transaksi itu (bukan cuma transaksi terakhir).
- **US3**: "Saya jual RAJA 3 lot di 900 tanggal 25 Juli" → tercatat
  transaksi `sell`, sisa posisi RAJA berkurang 3 lot dari total yang
  dipegang, average cost basis untuk lot yang TERSISA tetap sama
  (metode average cost, bukan FIFO).
- **US4**: "Analisis posisi RAJA saya" → dapat rekomendasi hold/sell
  berdasarkan average buy price gabungan, di menu terpisah dari Analisis
  biasa.
- **US5**: Saya lihat AI insight (ringkasan naratif) di hasil Analisis
  maupun Riwayat.
- **US6**: Saya generate 1 token admin sekali, taruh di config Hermes,
  semua request Hermes pakai token itu.

## 5. Data Model (Neon Postgres)

### Tabel `trade_transactions` — ledger mentah, immutable

| Kolom            | Tipe        | Keterangan                       |
| ---------------- | ----------- | -------------------------------- |
| id               | UUID (PK)   | `gen_random_uuid()`              |
| ticker           | TEXT        | mis. "RAJA"                      |
| transaction_type | TEXT        | `'buy'` atau `'sell'`            |
| lot              | NUMERIC     | 1 lot = 100 lembar (standar BEI) |
| price            | NUMERIC     | harga per lembar saat transaksi  |
| transaction_date | DATE        | tanggal transaksi                |
| notes            | TEXT        | opsional                         |
| created_at       | TIMESTAMPTZ | default `now()`                  |

**Prinsip penting**: tabel ini APPEND-ONLY. Transaksi yang salah input
di-KOREKSI dengan menambah transaksi pembalik + catatan di `notes`, BUKAN
di-UPDATE/DELETE baris lama (jaga integritas ledger, sama seperti prinsip
akuntansi — jangan hapus jejak audit).

### "Posisi" — TIDAK disimpan sebagai tabel terpisah, DIHITUNG dari ledger

Posisi per ticker (open lot, average buy price) adalah **derived data**,
dihitung dari `trade_transactions` pakai metode **moving-average-cost**:

- Tiap transaksi `buy`: average price baru = (lot lama × avg lama + lot
  baru × harga baru) / (lot lama + lot baru).
- Tiap transaksi `sell`: lot berkurang, average price TETAP SAMA untuk
  sisa lot (ciri khas average-cost method, beda dari FIFO).
- Posisi `open` kalau net lot > 0, `closed` kalau net lot = 0.

Ini SENGAJA tidak disimpan sebagai kolom independen supaya tidak ada
risiko data "posisi" nggak sinkron dengan ledger transaksi aslinya —
source of truth cuma 1 (ledger), posisi selalu dihitung ulang saat
dibutuhkan.

### Tabel `api_tokens` — untuk auth Hermes

| Kolom      | Tipe                   | Keterangan                               |
| ---------- | ---------------------- | ---------------------------------------- |
| id         | UUID (PK)              |                                          |
| token_hash | TEXT                   | SHA-256 hash dari token, BUKAN plaintext |
| role       | TEXT                   | saat ini cuma `'admin'` dipakai          |
| label      | TEXT                   | mis. "Hermes Telegram Bot"               |
| created_at | TIMESTAMPTZ            |                                          |
| revoked_at | TIMESTAMPTZ (nullable) | isi kalau token di-revoke                |

## 6. Fitur Detail

### 6.1 Catat Transaksi

Input: ticker, transaction_type (buy/sell), lot, price, transaction_date,
notes. Insert 1 row baru ke `trade_transactions`. Tidak ada validasi
"tidak boleh jual lebih dari yang dipegang" di v1 (bisa jadi validasi
lanjutan kalau perlu — untuk sekarang percaya input user).

### 6.2 Lihat Riwayat (ledger mentah)

List semua transaksi, filter by ticker/tanggal/type — ini yang tampil di
menu "Riwayat" sebagai histori apa adanya (bukan yang diagregasi).

### 6.3 Lihat Posisi Terbuka (agregat)

Hitung dari ledger (lihat algoritma section 5) untuk semua ticker yang
net lot > 0. Ini yang tampil sebagai daftar "posisi yang sedang dipegang".

### 6.4 Analisis Posisi Terbuka (menu terpisah dari Analisis)

Input: ticker (dari salah satu posisi open). Ambil average_buy_price dan
total_lot hasil agregasi (bukan dari 1 transaksi tunggal), lalu jalankan
fungsi analisis khusus posisi (reuse indicators/signals yang sudah ada,
fungsi keputusan baru — lihat prompt implementasi) → rekomendasi
hold/sell + AI insight.

### 6.5 AI Insight

Provider: **NVIDIA NIM**, model **`deepseek-ai/deepseek-v4-flash`**
(free tier, OpenAI-compatible API di `https://integrate.api.nvidia.com/v1`).
Insight 2-4 kalimat Bahasa Indonesia, MENJELASKAN angka yang sudah
dihitung sistem (indicators/signal/trading_plan atau position review),
BUKAN membuat rekomendasi independen yang bisa kontradiksi dengan hasil
perhitungan. Muncul di response menu Analisis dan menu Riwayat/Posisi.

## 7. API untuk Hermes (Telegram)

Semua endpoint di bawah butuh header `Authorization: Bearer <token>`,
token harus punya `role = 'admin'` di tabel `api_tokens` dan belum
di-revoke.

| Endpoint                          | Method | Fungsi                                          |
| --------------------------------- | ------ | ----------------------------------------------- |
| `/api/transactions`               | POST   | Tambah 1 transaksi (buy/sell)                   |
| `/api/transactions`               | GET    | List ledger mentah, filter `?ticker=`, `?type=` |
| `/api/positions`                  | GET    | List posisi open (agregat, average cost)        |
| `/api/positions/:ticker/analysis` | GET    | Analisis posisi (hold/sell) + AI insight        |
| `/api/analisis`                   | POST   | (sudah ada) — tambah field `ai_insight`         |

## 8. Non-Functional Requirements

- **Token security**: token disimpan sebagai HASH (SHA-256) di
  `api_tokens`, bukan plaintext — bahkan kalau database bocor, token asli
  tidak langsung ke-expose. Token asli cuma ditunjukkan SEKALI saat
  di-generate (tidak bisa dilihat ulang dari database setelahnya).
- **Koneksi Neon**: pakai **pooled connection string** dari dashboard
  Neon (biasanya hostname ber-suffix `-pooler`) untuk koneksi dari
  aplikasi web, bukan direct connection — supaya tidak menghabiskan
  connection limit Postgres kalau instance Go di-scale/restart berkali².
- **Ledger append-only**: tidak ada endpoint UPDATE/DELETE untuk
  `trade_transactions` di v1 — koreksi kesalahan input dilakukan dengan
  transaksi pembalik, bukan mengubah data historis.
- **AI insight rate limit**: NVIDIA NIM free tier ada rate limit
  (cek dashboard NVIDIA untuk angka pasti, biasanya puluhan
  request/menit) — cache insight per ticker+tanggal supaya tidak boros
  kuota untuk request berulang di hari yang sama.

## 9. Open Questions

- Kalau posisi sudah closed (net lot 0) lalu user beli ticker yang sama
  lagi nanti, apakah dianggap "posisi baru" (average cost mulai dari 0
  lagi) atau tetap nyambung ke histori lama? (Rekomendasi: posisi baru,
  average cost reset — closed = closed, ini konsisten dengan prinsip
  akuntansi standar.)
- Validasi "jangan jual lebih dari yang dipegang" — perlu di v1 atau
  nanti? (Saat ini: tidak divalidasi, dianggap tanggung jawab user
  input yang benar — bisa ditambah kalau ternyata sering salah input.)
