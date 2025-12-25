/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        brand: {
          red: '#E50914', // Netflix Red
          dark: '#141414',
          gray: '#2F2F2F',
        }
      },
      fontFamily: {
        sans: ['Inter', 'sans-serif'], // Make sure to import Inter in index.html or css
      },
    },
  },
  plugins: [],
}
