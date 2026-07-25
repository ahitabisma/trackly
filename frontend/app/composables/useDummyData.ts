export const DB: Record<string, any> = {
    BRMS: {
        ticker: 'BRMS',
        name: 'BUMI RESOURCES MINERALS Tbk',
        edges: [
            { target: 'EMIRATES TARIAN GLOBAL VENTURES SPC', pct: 25.1, type: 'OTHER', origin: 'FOREIGN', shares: 35592738434 },
            { target: 'GLAS TRUST (SINGAPORE) LTD.', pct: 7.61, type: 'INVESTMENT_BANK', origin: 'FOREIGN', shares: 10784607811 },
            { target: 'SUGIMAN HALIM', pct: 7.45, type: 'INDIVIDUAL', origin: 'LOCAL', shares: 10568888888 },
            { target: 'CGS INTL SEKURITAS INDONESIA', pct: 5.25, type: 'SECURITIES', origin: 'LOCAL', shares: 7439770171 },
            { target: 'PETROMINE ENERGY TRADING', pct: 4.28, type: 'CORPORATE', origin: 'LOCAL', shares: 6067891376 },
            { target: 'PT TRIMEGAH SEKURITAS INDONESIA TBK', pct: 3.91, type: 'SECURITIES', origin: 'LOCAL', shares: 5538637200 },
            { target: 'CGS INTL SECURITIES SINGAPORE PTE', pct: 3.77, type: 'INVESTMENT_BANK', origin: 'FOREIGN', shares: 5351412625 },
            { target: 'PT MAYBANK SEKURITAS INDONESIA', pct: 3.24, type: 'SECURITIES', origin: 'LOCAL', shares: 4589589548 },
            { target: 'BUMI RESOURCES TBK', pct: 3.08, type: 'CORPORATE', origin: 'LOCAL', shares: 4365383689 },
            { target: 'VANECK GOLD MINERS ETF', pct: 2.87, type: 'MUTUAL_FUND', origin: 'FOREIGN', shares: 4075687300 },
            { target: 'WEXLER CAPITAL PTE. LTD', pct: 2.19, type: 'CORPORATE', origin: 'FOREIGN', shares: 3100000032 },
        ],
        cross: [
            { investor: 'GLAS TRUST (SINGAPORE) LTD.', type: 'INVESTMENT_BANK', target: 'TRIO', pct: 25.53 },
            { investor: 'GLAS TRUST (SINGAPORE) LTD.', type: 'INVESTMENT_BANK', target: 'BUMI', pct: 2.08 },
            { investor: 'PETROMINE ENERGY TRADING', type: 'CORPORATE', target: 'DEWA', pct: 4.46 },
            { investor: 'PETROMINE ENERGY TRADING', type: 'CORPORATE', target: 'VKTR', pct: 2.83 },
            { investor: 'PETROMINE ENERGY TRADING', type: 'CORPORATE', target: 'ELTY', pct: 2.61 },
        ],
    },
}

export const EMITEN = [
    { ticker: 'BRMS', name: 'BUMI RESOURCES MINERALS Tbk' },
    { ticker: 'BBCA', name: 'BANK CENTRAL ASIA Tbk' },
    { ticker: 'TLKM', name: 'TELEKOMUNIKASI INDONESIA Tbk' },
    { ticker: 'BMRI', name: 'BANK MANDIRI Tbk' },
    { ticker: 'ASII', name: 'ASTRA INTERNATIONAL Tbk' },
    { ticker: 'GOTO', name: 'GOTO GOJEK TOKOPEDIA Tbk' },
    { ticker: 'BUMI', name: 'BUMI RESOURCES Tbk' },
    { ticker: 'DEWA', name: 'DARMA HENWA Tbk' },
    { ticker: 'TRIO', name: 'TRIKOMSEL OKE Tbk' },
]

export const TYPES = [
    { key: 'stock', label: 'Stock (Emiten)', fill: '#daedf9', stroke: '#2d7ab5' },
    { key: 'CORPORATE', label: 'Corporate', fill: '#e8e2f8', stroke: '#6040c0' },
    { key: 'INVESTMENT_BANK', label: 'Investment Bank', fill: '#ede0f8', stroke: '#8b50d8' },
    { key: 'INDIVIDUAL', label: 'Individual', fill: '#fce8cc', stroke: '#c07830' },
    { key: 'MUTUAL_FUND', label: 'Mutual Fund', fill: '#d4f0dc', stroke: '#3a8a58' },
    { key: 'SECURITIES', label: 'Securities', fill: '#fce4e4', stroke: '#b04040' },
    { key: 'OTHER', label: 'Other', fill: '#f0ece8', stroke: '#807870' },
]

export const TM: Record<string, typeof TYPES[0]> = {}
TYPES.forEach(t => { TM[t.key] = t })