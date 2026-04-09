/** @type {import('tailwindcss').Config} */
export default {
    content: [
        './app/components/**/*.{js,vue,ts}',
        './app/layouts/**/*.vue',
        './app/pages/**/*.vue',
        './app/plugins/**/*.{js,ts}',
        './app.vue',
    ],
    theme: {
        extend: {
            colors: {
                cream: '#f2ede4',
                cream2: '#ede8de',
                card: '#f7f3ec',
                ink: '#1a1612',
                ink2: '#3d3730',
                muted: '#8a8178',
                bdr: '#e0d9cf',
                bdr2: '#ccc5b9',
                blue: '#6ab0e8',
                bluebg: '#daedf9',
                bluedk: '#2d7ab5',
            },
            fontFamily: {
                serif: ['Instrument Serif', 'Georgia', 'serif'],
                sans: ['DM Sans', 'system-ui', 'sans-serif'],
                mono: ['Share Tech Mono', 'monospace'],
            },
        },
    },
    plugins: [],
}
