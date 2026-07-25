export type FiltersType = Record<string, any>;

export interface FilterParams {
    page?: number;
    limit?: number;
    filters?: FiltersType;
    orderKey?: string;
    orderRule?: 'asc' | 'desc';
}