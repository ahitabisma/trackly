// Use the global API_URL set by plugin, or environment variable
export const API_URL = (typeof globalThis !== 'undefined' && (globalThis as any).__API_URL__)
    || process.env.NUXT_PUBLIC_API_URL
    || 'https://api-trackly.aksanara.id';

export const APP_URL = process.env.NUXT_PUBLIC_APP_URL
    || 'https://trackly.aksanara.id';