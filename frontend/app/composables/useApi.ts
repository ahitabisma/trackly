export const useApi = () => {
    const config = useRuntimeConfig();
    const apiUrl = config.public.apiUrl as string;

    return {
        apiUrl,
    };
};
