/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        // Emerald primary accent
        primary: {
          DEFAULT: '#10b981',
          50: '#ecfdf5',
          100: '#d1fae5',
          200: '#a7f3d0',
          300: '#6ee7b7',
          400: '#34d399',
          500: '#10b981',
          600: '#059669',
          700: '#047857',
          800: '#065f46',
          900: '#064e3b',
          950: '#022c22',
        },
        // Dark neutral surfaces
        ink: {
          black: '#050505',
          base: '#0f1115',
          surface: '#16181d',
          elevated: '#1c1f26',
          border: '#262932',
        },
      },
      fontFamily: {
        heading: ['Manrope', 'system-ui', 'sans-serif'],
        sans: ['Inter', 'system-ui', 'sans-serif'],
      },
      borderRadius: {
        xl: '0.875rem',
        '2xl': '1.125rem',
      },
      boxShadow: {
        soft: '0 1px 2px 0 rgba(0,0,0,0.04), 0 1px 3px 0 rgba(0,0,0,0.06)',
        card: '0 1px 3px rgba(0,0,0,0.06), 0 8px 24px -12px rgba(0,0,0,0.12)',
        'card-dark': '0 1px 3px rgba(0,0,0,0.4), 0 12px 32px -16px rgba(0,0,0,0.6)',
        glow: '0 0 0 1px rgba(16,185,129,0.2), 0 8px 24px -8px rgba(16,185,129,0.35)',
      },
      keyframes: {
        'fade-in': {
          '0%': { opacity: '0', transform: 'translateY(4px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        'slide-in-right': {
          '0%': { transform: 'translateX(100%)' },
          '100%': { transform: 'translateX(0)' },
        },
        'slide-in-left': {
          '0%': { transform: 'translateX(-100%)' },
          '100%': { transform: 'translateX(0)' },
        },
        shimmer: {
          '100%': { transform: 'translateX(100%)' },
        },
      },
      animation: {
        'fade-in': 'fade-in 0.2s ease-out',
        'slide-in-right': 'slide-in-right 0.25s cubic-bezier(0.16,1,0.3,1)',
        'slide-in-left': 'slide-in-left 0.25s cubic-bezier(0.16,1,0.3,1)',
      },
    },
  },
  plugins: [],
};
