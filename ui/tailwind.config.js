/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['"Space Grotesk"', 'sans-serif'],
      },
      colors: {
        main: '#B829EA', // Phantom Purple
        'main-foreground': '#FFFFFF',
        background: '#050505', // Deep space dark
        'secondary-background': '#111111',
        foreground: '#FFFFFF',
        border: '#B829EA', // Hard purple borders
        neonCyan: '#00F3FF', // Connected status
        electricPink: '#FF0055', // Offline status
        busyOrange: '#FF3300', // Busy status
      },
      boxShadow: {
        shadow: '4px 4px 0 0 #FFFFFF', // White hard shadow
        nav: '2px 2px 0 0 #FFFFFF',
        cardLayer: '8px 8px 0 0 #B829EA', // Colored plate behind cards
      },
      translate: {
        boxShadowX: '4px',
        boxShadowY: '4px',
        cardLayerX: '8px',
        cardLayerY: '8px',
      },
      borderRadius: {
        base: '0px', // Neo-Brutalism is usually sharp
      },
      transitionTimingFunction: {
        spring: 'cubic-bezier(0.175, 0.885, 0.32, 1.275)', // Brutalist snap back
      },
    }
  },
  plugins: []
};
