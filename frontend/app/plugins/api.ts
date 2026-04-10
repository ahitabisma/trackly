export default defineNuxtPlugin(() => {
    const config = useRuntimeConfig();
    const apiUrl = config.public.apiUrl as string;

    // Make API URL available globally through window object for services
    if (process.client) {
        (globalThis as any).__API_URL__ = apiUrl;
    }

    return {
        provide: {
            apiUrl: apiUrl,
        },
    };
});
