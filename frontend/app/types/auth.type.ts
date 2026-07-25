export interface User {
    id: number
    name: string
    email: string
    avatar: string | null
    role: string
    created_at: string
    updated_at: string
}

export interface LoginRequest {
    email: string
    password: string
}

export interface RegisterRequest {
    name: string
    email: string
    password: string
}

export interface LoginResponse {
    access_token: string
    user: User
}

export interface Session {
    token: string
    user: User
}

export interface ApiResponse<T> {
    success: boolean
    message: string
    data: T
    errors?: any
    meta?: any
}
