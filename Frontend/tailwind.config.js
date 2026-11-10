/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: {
          DEFAULT: '#168F7A',
          hover: '#127564',
          light: '#E5F3E9',
          50: '#F0F9F6',
          100: '#E5F3E9',
          200: '#C2E5D6',
          500: '#20B8A5',
          600: '#168F7A',
          700: '#127564',
          800: '#0F5D50',
          900: '#183B35',
        },
        secondary: {
          DEFAULT: '#20B8A5',
          hover: '#199989',
        },
        nature: {
          green: '#6F9F78',
          sage: '#A8C5A5',
          mint: '#E5F3E9',
          dark: '#183B35',
        },
        surface: {
          DEFAULT: '#FFFFFF',
          muted: '#F5F8F5',
          subtle: '#FAFCFA',
        },
        text: {
          primary: '#183B35',
          secondary: '#71827B',
          muted: '#94A39D',
        },
        border: {
          DEFAULT: '#E2EAE4',
          subtle: '#EEF4F0',
        },
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', '-apple-system', 'sans-serif'],
      },
      boxShadow: {
        'soft-sm': '0 1px 3px 0 rgba(22, 143, 122, 0.04)',
        'soft-md': '0 4px 12px -2px rgba(22, 143, 122, 0.08)',
        'soft-lg': '0 12px 24px -4px rgba(24, 59, 53, 0.1)',
      },
    },
  },
  plugins: [],
}
