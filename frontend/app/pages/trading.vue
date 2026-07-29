<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useTrading } from '~/composables/useTrading'
import type { TransactionRequest, UpdateTransactionRequest } from '~/services/trading.service'

definePageMeta({ layout: 'main', middleware: 'auth'})

const { transactions, loading, error, add, load, updateTransaction, removeTransaction } = useTrading()

const form = ref<TransactionRequest>({
  ticker: '', transaction_type: 'buy', lot: 1, price: 0, transaction_date: new Date().toISOString().split('T')[0], notes: '',
})
const formError = ref('')
const formSuccess = ref('')
const showForm = ref(false)
const filterTicker = ref('')
const filterDateStart = ref('')
const filterDateEnd = ref('')
const filterType = ref('')
const editingId = ref<number | null>(null)
const editForm = ref<UpdateTransactionRequest>({})
const deleteConfirmId = ref<number | null>(null)

const today = () => new Date().toISOString().split('T')[0]

const filteredTransactions = computed(() => {
  return transactions.value.filter(t => {
    if (filterTicker.value && !t.ticker.toUpperCase().includes(filterTicker.value.toUpperCase())) return false
    if (filterDateStart.value && t.transaction_date < filterDateStart.value) return false
    if (filterDateEnd.value && t.transaction_date > filterDateEnd.value) return false
    if (filterType.value && t.transaction_type !== filterType.value) return false
    return true
  })
})

const formatNum = (v: number) => new Intl.NumberFormat('id-ID', { minimumFractionDigits: 0, maximumFractionDigits: 2 }).format(v)
const fmtDate = (d: string) => {
  const dt = new Date(d)
  return dt.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' })
}

const submit = async () => {
  formError.value = ''; formSuccess.value = ''
  if (!form.value.ticker) { formError.value = 'Ticker wajib diisi'; return }
  if (!form.value.price || form.value.price <= 0) { formError.value = 'Harga wajib > 0'; return }
  if (!form.value.lot || form.value.lot <= 0) { formError.value = 'Lot wajib > 0'; return }
  if (!form.value.transaction_date) { formError.value = 'Tanggal wajib diisi'; return }

  try {
    await add({ ...form.value })
    formSuccess.value = `Transaksi ${form.value.transaction_type.toUpperCase()} ${form.value.ticker} berhasil`
    form.value = { ticker: '', transaction_type: 'buy', lot: 1, price: 0, transaction_date: today(), notes: '' }
    showForm.value = false
  } catch { formError.value = error.value || 'Gagal simpan' }
}

const startEdit = (t: any) => {
  editingId.value = t.id
  editForm.value = {
    ticker: t.ticker, transaction_type: t.transaction_type, lot: t.lot, price: t.price,
    transaction_date: t.transaction_date, notes: t.notes || '',
  }
}

const cancelEdit = () => {
  editingId.value = null; editForm.value = {}
}

const saveEdit = async (id: number) => {
  const body: UpdateTransactionRequest = {}
  if (editForm.value.ticker !== undefined) body.ticker = editForm.value.ticker
  if (editForm.value.transaction_type !== undefined) body.transaction_type = editForm.value.transaction_type as any
  if (editForm.value.lot !== undefined) body.lot = editForm.value.lot
  if (editForm.value.price !== undefined) body.price = editForm.value.price
  if (editForm.value.transaction_date !== undefined) body.transaction_date = editForm.value.transaction_date
  if (editForm.value.notes !== undefined) body.notes = editForm.value.notes
  try {
    await updateTransaction(id, body)
    editingId.value = null; editForm.value = {}
  } catch {}
}

const confirmDelete = (id: number) => { deleteConfirmId.value = id }
const cancelDelete = () => { deleteConfirmId.value = null }

const doDelete = async (id: number) => {
  try {
    await removeTransaction(id)
    deleteConfirmId.value = null
  } catch {}
}

const applyFilter = () => {
  load({ ticker: filterTicker.value || undefined, dateStart: filterDateStart.value || undefined, dateEnd: filterDateEnd.value || undefined, type: filterType.value || undefined })
}

const clearFilter = () => {
  filterTicker.value = ''; filterDateStart.value = ''; filterDateEnd.value = ''; filterType.value = ''
  load()
}

onMounted(() => { load() })
</script>

<template>
  <main class="min-h-screen bg-cream text-ink font-sans antialiased overflow-x-hidden pt-20 pb-24">
    <div class="max-w-5xl mx-auto px-5 sm:px-8">
      <div class="fu mb-8">
        <h1 class="font-serif text-4xl text-ink" style="letter-spacing:-1px">Riwayat Transaksi</h1>
        <p class="font-mono text-xs text-muted mt-1 uppercase tracking-wider">Catat &amp; kelola transaksi saham</p>
      </div>

      <div class="space-y-6">
        <!-- Add Transaction Button -->
        <div class="fu2">
          <button class="neo-btn-primary max-w-xs" @click="showForm = !showForm">
            {{ showForm ? 'Batal' : '+ Transaksi Baru' }}
          </button>
        </div>

        <!-- Transaction Form -->
        <div v-if="showForm" class="fu3 bg-card border-2 border-ink rounded-2xl p-5 sm:p-6"
          style="box-shadow:4px 4px 0 #1a1612">
          <h2 class="font-mono text-xs text-muted uppercase tracking-wider mb-4">Transaksi Baru</h2>
          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mb-4">
            <div>
              <label class="font-mono text-xs text-muted uppercase block mb-1">Ticker</label>
              <input v-model="form.ticker" class="neo-input" placeholder="ANTM" @input="formError='';formSuccess=''" />
            </div>
            <div>
              <label class="font-mono text-xs text-muted uppercase block mb-1">Tipe</label>
              <select v-model="form.transaction_type" class="neo-input">
                <option value="buy">Buy</option>
                <option value="sell">Sell</option>
              </select>
            </div>
            <div>
              <label class="font-mono text-xs text-muted uppercase block mb-1">Lot</label>
              <input v-model.number="form.lot" class="neo-input" type="number" min="1" step="1" />
            </div>
            <div>
              <label class="font-mono text-xs text-muted uppercase block mb-1">Harga per Lembar</label>
              <input v-model.number="form.price" class="neo-input" type="number" min="1" step="1" />
            </div>
            <div>
              <label class="font-mono text-xs text-muted uppercase block mb-1">Tanggal</label>
              <input v-model="form.transaction_date" class="neo-input" type="date" />
            </div>
            <div>
              <label class="font-mono text-xs text-muted uppercase block mb-1">Catatan</label>
              <input v-model="form.notes" class="neo-input" placeholder="Opsional" />
            </div>
          </div>

          <div v-if="formError" class="mb-3 font-mono text-xs text-red-600">{{ formError }}</div>
          <div v-if="formSuccess" class="mb-3 font-mono text-xs text-green-700">{{ formSuccess }}</div>

          <button class="neo-btn-primary max-w-[200px]" :disabled="loading" @click="submit">
            <span v-if="loading" class="inline-block w-4 h-4 border-2 border-cream border-t-transparent rounded-full animate-spin"></span>
            {{ loading ? 'Menyimpan...' : 'Simpan' }}
          </button>
        </div>

        <!-- Transaction List -->
        <div class="fu3 bg-card border-2 border-ink rounded-2xl p-5 sm:p-6"
          style="box-shadow:4px 4px 0 #1a1612">
          <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 mb-4">
            <h2 class="font-mono text-xs text-muted uppercase tracking-wider">Riwayat Transaksi</h2>
          </div>

          <!-- Filters -->
          <div class="flex flex-wrap gap-3 mb-4">
            <input v-model="filterTicker" class="neo-input max-w-[120px]" placeholder="Ticker" @input="applyFilter" />
            <input v-model="filterDateStart" class="neo-input max-w-[150px]" type="date" @change="applyFilter" />
            <input v-model="filterDateEnd" class="neo-input max-w-[150px]" type="date" @change="applyFilter" />
            <select v-model="filterType" class="neo-input max-w-[120px]" @change="applyFilter">
              <option value="">Semua</option>
              <option value="buy">Buy</option>
              <option value="sell">Sell</option>
            </select>
            <button class="neo-btn-sm text-muted" @click="clearFilter">Reset</button>
          </div>

          <div v-if="loading && !transactions.length" class="text-center py-8">
            <div class="inline-block w-6 h-6 border-2 border-ink border-t-transparent rounded-full animate-spin mb-2"></div>
            <p class="font-mono text-xs text-muted">Memuat...</p>
          </div>

          <div v-else-if="!filteredTransactions.length" class="text-center py-8">
            <p class="font-mono text-sm text-muted">Belum ada transaksi</p>
          </div>

          <div v-else class="overflow-x-auto">
            <table class="w-full text-left font-mono text-xs data-table">
              <thead>
                <tr class="text-muted uppercase tracking-wider border-b border-bdr">
                  <th class="py-2 pr-3">Tanggal</th>
                  <th class="py-2 pr-3">Ticker</th>
                  <th class="py-2 pr-3">Tipe</th>
                  <th class="py-2 pr-3">Lot</th>
                  <th class="py-2 pr-3">Harga</th>
                  <th class="py-2 pr-3">Total</th>
                  <th class="py-2 pr-3">Catatan</th>
                  <th class="py-2 pr-3">Aksi</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="t in filteredTransactions" :key="t.id" class="border-b border-bdr/50">
                  <template v-if="editingId === t.id">
                    <td class="py-2 pr-3"><input v-model="editForm.transaction_date" class="neo-input max-w-[130px]" type="date" /></td>
                    <td class="py-2 pr-3"><input v-model="editForm.ticker" class="neo-input max-w-[90px]" /></td>
                    <td class="py-2 pr-3">
                      <select v-model="editForm.transaction_type" class="neo-input max-w-[90px]">
                        <option value="buy">Buy</option>
                        <option value="sell">Sell</option>
                      </select>
                    </td>
                    <td class="py-2 pr-3"><input v-model.number="editForm.lot" class="neo-input max-w-[80px]" type="number" min="1" /></td>
                    <td class="py-2 pr-3"><input v-model.number="editForm.price" class="neo-input max-w-[100px]" type="number" min="1" /></td>
                    <td class="py-2 pr-3">{{ editForm.lot && editForm.price ? formatNum(editForm.lot * editForm.price * 100) : '-' }}</td>
                    <td class="py-2 pr-3"><input v-model="editForm.notes" class="neo-input max-w-[100px]" /></td>
                    <td class="py-2 pr-3 whitespace-nowrap">
                      <button class="neo-btn-sm text-green-700 mr-1" @click="saveEdit(t.id)">Simpan</button>
                      <button class="neo-btn-sm text-muted" @click="cancelEdit">Batal</button>
                    </td>
                  </template>
                  <template v-else>
                    <td class="py-2 pr-3 whitespace-nowrap">{{ fmtDate(t.transaction_date) }}</td>
                    <td class="py-2 pr-3 font-bold">{{ t.ticker }}</td>
                    <td class="py-2 pr-3">
                      <span class="font-bold uppercase" :class="t.transaction_type === 'buy' ? 'text-green-700' : 'text-red-700'">{{ t.transaction_type }}</span>
                    </td>
                    <td class="py-2 pr-3">{{ formatNum(t.lot) }}</td>
                    <td class="py-2 pr-3">{{ formatNum(t.price) }}</td>
                    <td class="py-2 pr-3">{{ formatNum(t.lot * t.price * 100) }}</td>
                    <td class="py-2 pr-3 text-muted">{{ t.notes || '-' }}</td>
                    <td class="py-2 pr-3 whitespace-nowrap">
                      <button class="neo-btn-sm text-blue-600 mr-1" @click="startEdit(t)">Edit</button>
                      <button v-if="deleteConfirmId !== t.id" class="neo-btn-sm text-red-600" @click="confirmDelete(t.id)">Hapus</button>
                      <span v-else class="inline-flex gap-1">
                        <button class="neo-btn-sm text-red-700 font-bold" @click="doDelete(t.id)">Yakin?</button>
                        <button class="neo-btn-sm text-muted" @click="cancelDelete">Batal</button>
                      </span>
                    </td>
                  </template>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Error -->
      <div v-if="error && !showForm" class="fu3 bg-red-50 border-2 border-red-300 rounded-xl px-5 py-4 mt-6"
        style="box-shadow:3px 3px 0 #e25757">
        <p class="font-mono text-sm text-red-700">{{ error }}</p>
      </div>
    </div>
  </main>
</template>