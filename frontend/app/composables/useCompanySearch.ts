import { ref, computed } from 'vue';
import { companyService } from '~/services/company.service';
import { shareholdingService } from '~/services/shareholding.service';
import type { Company } from '~/types/company.type';
import type { Shareholding } from '~/types/shareholding.type';

export const useCompanySearch = () => {
    const searchQuery = ref('');
    const suggestions = ref<Company[]>([]);
    const loading = ref(false);
    const error = ref<string | null>(null);

    /**
     * Search companies as user types
     */
    const searchCompanies = async (query: string = searchQuery.value) => {
        if (!query.trim()) {
            suggestions.value = [];
            return;
        }

        loading.value = true;
        error.value = null;

        try {
            const response = await companyService.searchCompanies(query, 10);
            suggestions.value = response.data;
        } catch (err) {
            error.value = err instanceof Error ? err.message : 'Failed to search companies';
            console.error('Error searching companies:', err);
            suggestions.value = [];
        } finally {
            loading.value = false;
        }
    };

    /**
     * Clear suggestions
     */
    const clearSuggestions = () => {
        suggestions.value = [];
        searchQuery.value = '';
    };

    const hasSuggestions = computed(() => suggestions.value.length > 0);

    return {
        searchQuery,
        suggestions,
        loading,
        error,
        searchCompanies,
        clearSuggestions,
        hasSuggestions,
    };
};

/**
 * Composable for company shareholdings view
 */
export const useCompanyShareholdings = () => {
    const selectedCompany = ref<Company | null>(null);
    const shareholdings = ref<Shareholding[]>([]);
    const loading = ref(false);
    const error = ref<string | null>(null);
    const pagination = ref({
        total: 0,
        page: 1,
        limit: 10,
        total_pages: 0,
        has_next_page: false,
        has_prev_page: false,
    });

    /**
     * Load shareholdings for selected company with percentage ordering
     */
    const loadCompanyShareholdings = async (company: Company) => {
        selectedCompany.value = company;
        loading.value = true;
        error.value = null;

        try {
            const response = await shareholdingService.getAllShareholdings({
                page: 1,
                limit: 1000,
                filters: { company_kode: company.kode },
                orderKey: 'percentage',
                orderRule: 'desc',
            });

            shareholdings.value = response.data;
            pagination.value = response.meta?.pagination || {
                total: 0,
                page: 1,
                limit: 10,
                total_pages: 0,
                has_next_page: false,
                has_prev_page: false,
            };
        } catch (err) {
            error.value = err instanceof Error ? err.message : 'Failed to fetch shareholdings';
            console.error('Error fetching shareholdings:', err);
        } finally {
            loading.value = false;
        }
    };

    /**
     * Go to specific page
     */
    const goToPage = async (page: number) => {
        if (!selectedCompany.value) return;

        loading.value = true;
        error.value = null;

        try {
            const response = await shareholdingService.getAllShareholdings({
                page,
                limit: 1000,
                filters: { company_kode: selectedCompany.value.kode },
                orderKey: 'percentage',
                orderRule: 'desc',
            });

            shareholdings.value = response.data;
            pagination.value = response.meta?.pagination || {
                total: 0,
                page: 1,
                limit: 10,
                total_pages: 0,
                has_next_page: false,
                has_prev_page: false,
            };
        } catch (err) {
            error.value = err instanceof Error ? err.message : 'Failed to fetch shareholdings';
            console.error('Error fetching shareholdings:', err);
        } finally {
            loading.value = false;
        }
    };

    /**
     * Clear selected company and shareholdings
     */
    const clear = () => {
        selectedCompany.value = null;
        shareholdings.value = [];
        pagination.value = {
            total: 0,
            page: 1,
            limit: 10,
            total_pages: 0,
            has_next_page: false,
            has_prev_page: false,
        };
    };

    const hasData = computed(() => shareholdings.value.length > 0);

    return {
        selectedCompany,
        shareholdings,
        loading,
        error,
        pagination,
        loadCompanyShareholdings,
        goToPage,
        clear,
        hasData,
    };
};
