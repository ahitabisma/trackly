import { $fetch } from 'ofetch';

interface Company {
    id: number;
    kode: string;
    nama_perusahaan: string;
    tanggal_pencatatan: string;
    jumlah_saham: number;
    papan_pencatatan: string;
    created_at: string;
    updated_at: string;
}

interface CompanyResponse {
    success: boolean;
    message: string;
    data: Company[];
    meta: {
        request_id: string;
        timestamp: string;
        pagination: {
            total: number;
            page: number;
            limit: number;
            total_pages: number;
            has_next_page: boolean;
            has_prev_page: boolean;
        };
    };
}

interface FilterParams {
    page?: number;
    limit?: number;
    filters?: Record<string, any>;
    orderKey?: string;
    orderRule?: 'asc' | 'desc';
    search?: string;
    searchFilters?: Array<{ field: string; value: string }> | Record<string, any>;
}

export class CompanyService {
    private getBaseURL(): string {
        const url = (typeof globalThis !== 'undefined' && (globalThis as any).__API_URL__)
            || (typeof process !== 'undefined' && process.env.NUXT_PUBLIC_API_URL)
            || 'https://api-trackly.aksanara.id';
        return url;
    }

    /**
     * Get all companies with filters and pagination
     */
    async getAllCompanies(params: FilterParams = {}): Promise<CompanyResponse> {
        try {
            const queryParams = new URLSearchParams();

            // Pagination
            if (params.page) queryParams.append('page', params.page.toString());
            if (params.limit) queryParams.append('limit', params.limit.toString());

            // Filters
            if (params.filters && Object.keys(params.filters).length > 0) {
                queryParams.append('filters', JSON.stringify(params.filters));
            }

            // Order
            if (params.orderKey) queryParams.append('orderKey', params.orderKey);
            if (params.orderRule) queryParams.append('orderRule', params.orderRule);

            // Search
            if (params.search) queryParams.append('search', params.search);

            // Search filters (ILIKE on backend)
            if (params.searchFilters) {
                queryParams.append('searchFilters', JSON.stringify(params.searchFilters));
            }

            const baseURL = this.getBaseURL();
            const url = `${baseURL}/companies${queryParams.toString() ? '?' + queryParams.toString() : ''}`;

            const response = await $fetch<CompanyResponse>(url, {
                method: 'GET',
            });

            return response;
        } catch (error) {
            console.error('Failed to fetch companies:', error);
            throw error;
        }
    }

    /**
     * Search companies by kode or nama
     */
    async searchCompanies(query: string, limit: number = 10): Promise<CompanyResponse> {
        try {
            return this.getAllCompanies({
                page: 1,
                limit,
                searchFilters: [
                    { field: 'kode', value: query },
                    { field: 'nama_perusahaan', value: query },
                ],
            });
        } catch (error) {
            console.error('Failed to search companies:', error);
            throw error;
        }
    }

    /**
     * Get single company by kode
     */
    async getCompanyByKode(kode: string): Promise<CompanyResponse> {
        try {
            const baseURL = this.getBaseURL();
            const url = `${baseURL}/companies?filters=${JSON.stringify({ kode })}`;

            const response = await $fetch<CompanyResponse>(url, {
                method: 'GET',
            });

            return response;
        } catch (error) {
            console.error('Failed to fetch company:', error);
            throw error;
        }
    }
}

export const companyService = new CompanyService();