import { defineComponent, h } from 'vue'

export const useAuthIcons = () => {
    const EyeIcon = defineComponent({
        render: () => h('svg', {
            width: 18, height: 18, viewBox: '0 0 24 24', fill: 'none',
            stroke: 'currentColor', 'stroke-width': 2, 'stroke-linecap': 'round',
        }, [
            h('path', { d: 'M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7-10-7-10-7z' }),
            h('circle', { cx: 12, cy: 12, r: 3 }),
        ]),
    })

    const EyeOffIcon = defineComponent({
        render: () => h('svg', {
            width: 18, height: 18, viewBox: '0 0 24 24', fill: 'none',
            stroke: 'currentColor', 'stroke-width': 2, 'stroke-linecap': 'round',
        }, [
            h('path', { d: 'M17.94 17.94A10.07 10.07 0 0112 20c-7 0-11-8-11-8a18.45 18.45 0 015.06-5.94' }),
            h('path', { d: 'M9.9 4.24A9.12 9.12 0 0112 4c7 0 11 8 11 8a18.5 18.5 0 01-2.16 3.19' }),
            h('line', { x1: 1, y1: 1, x2: 23, y2: 23 }),
        ]),
    })

    const GoogleIcon = defineComponent({
        render: () => h('svg', { width: 20, height: 20, viewBox: '0 0 24 24', xmlns: 'http://www.w3.org/2000/svg' }, [
            h('path', { d: 'M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z', fill: '#4285F4' }),
            h('path', { d: 'M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z', fill: '#34A853' }),
            h('path', { d: 'M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l3.66-2.84z', fill: '#FBBC05' }),
            h('path', { d: 'M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z', fill: '#EA4335' }),
        ]),
    })

    const SpinnerIcon = defineComponent({
        render: () => h('svg', {
            width: 15, height: 15, viewBox: '0 0 24 24', fill: 'none',
            stroke: 'currentColor', 'stroke-width': 2,
            style: 'animation: spin .6s linear infinite',
        }, [
            h('path', { d: 'M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83' }),
        ]),
    })

    return { EyeIcon, EyeOffIcon, GoogleIcon, SpinnerIcon }
}
