import 'vuetify/styles';
import '@mdi/font/css/materialdesignicons.css';
import { createVuetify } from 'vuetify';
import * as components from 'vuetify/components';
import * as directives from 'vuetify/directives';

export default createVuetify({
  components,
  directives,
  theme: {
    defaultTheme: 'light',
    themes: {
      light: {
        colors: {
          // Nightshades / Material-inspired palette
          primary: '#5b21b6', // deep purple (nightshade)
          'primary-darken-1': '#4c1d9b',
          secondary: '#111827', // dark neutral
          accent: '#14b8a6', // teal accent
          error: '#ef4444',
          info: '#60a5fa',
          success: '#10b981',
          warning: '#f59e0b',
          background: '#f6f7fb',
          surface: '#ffffff',
        },
      },
    },
  },
});