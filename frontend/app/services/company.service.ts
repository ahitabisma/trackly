import { $fetch } from 'ofetch';
import { API_URL } from '~/lib/constants';

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
}

export class CompanyService {
    private baseURL: string;

    constructor() {
        this.baseURL = API_URL;
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

            const url = `${this.baseURL}/companies${queryParams.toString() ? '?' + queryParams.toString() : ''}`;

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
            const url = `${this.baseURL}/companies?search=${encodeURIComponent(query)}&limit=${limit}`;

            const response = await $fetch<CompanyResponse>(url, {
                method: 'GET',
            });

            return response;
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
            const url = `${this.baseURL}/companies?filters=${JSON.stringify({ kode })}`;

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