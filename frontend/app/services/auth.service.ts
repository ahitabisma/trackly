import { $fetch } from 'ofetch'
import type { LoginRequest, RegisterRequest, LoginResponse, User, ApiResponse } from '~/types/auth.type'

export class AuthService {
    private getBaseURL(): string {
        return (typeof globalThis !== 'undefined' && (globalThis as any).__API_URL__)
            || (typeof process !== 'undefined' && process.env.NUXT_PUBLIC_API_URL)
            || 'https://api-trackly.aksanara.id'
    }

    async login(req: LoginRequest): Promise<LoginResponse> {
        const url = `${this.getBaseURL()}/auth/login`
        const response = await $fetch<ApiResponse<LoginResponse>>(url, {
            method: 'POST',
            body: req,
        })
        return response.data
    }

    async register(req: RegisterRequest): Promise<User> {
        const url = `${this.getBaseURL()}/auth/register`
        const response = await $fetch<ApiResponse<User>>(url, {
            method: 'POST',
            body: req,
        })
        return response.data
    }
}

export const authService = new AuthService()
