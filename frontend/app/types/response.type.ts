export interface Response {
    success: boolean;
    message: string;
    data: any;
    meta: {
        request_id: string;
        timestamp: string;
        pagination?: Pagination;
    };
    errors?: any;
}

interface Pagination {
    total: number;
    page: number;
    limit: number;
    total_pages: number;
    has_next_page: boolean;
    has_prev_page: boolean;
}