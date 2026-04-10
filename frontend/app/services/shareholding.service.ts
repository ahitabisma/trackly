import { $fetch } from 'ofetch';
import type { FilterParams } from '~/types/filter.type';
import type { Response } from '~/types/response.type';

export class ShareholdingService {
    private getBaseURL(): string {
        const url = (typeof globalThis !== 'undefined' && (globalThis as any).__API_URL__)
            || (typeof process !== 'undefined' && process.env.NUXT_PUBLIC_API_URL)
            || 'https://api-trackly.aksanara.id';
        return url;
    }

    async getAllShareholdings(params: FilterParams = {}): Promise<Response> {
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

            const baseURL = this.getBaseURL();
            const url = `${baseURL}/shareholdings${queryParams.toString() ? '?' + queryParams.toString() : ''}`;

            const response = await $fetch<Response>(url, {
                method: 'GET',
            });

            return response;
        } catch (error) {
            console.error('Failed to fetch shareholdings:', error);
            throw error;
        }
    }
}

export const shareholdingService = new ShareholdingService();