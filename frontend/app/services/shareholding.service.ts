import { $fetch } from 'ofetch';
import { API_URL } from '~/lib/constants';
import type { FilterParams } from '~/types/filter.type';
import type { Response } from '~/types/response.type';

export class ShareholdingService {
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

            const url = `${API_URL}/shareholdings${queryParams.toString() ? '?' + queryParams.toString() : ''}`;

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