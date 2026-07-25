export interface Shareholding {
    id: number;
    company_id: number;
    company_kode: string;
    company_name: string;
    investor_id: number;
    investor_name: string;
    investor_type: string;
    local_foreign: string;
    nationality: string;
    domicile: string;
    holdings_scripless: number;
    holdings_scrip: number;
    total_holding_shares: number;
    percentage: number;
    date: string;
    source: string | null;
}