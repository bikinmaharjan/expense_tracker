import { createApp } from 'vue';
import { createPinia } from 'pinia';
import './style.css';

// Import the main app component
import App from './App.vue';
// Import router configuration
import router from './router';
// Import Vuetify configuration
import vuetify from './plugins/vuetify';

// Create the app instance
const app = createApp(App);

// Install plugins
app.use(createPinia());
app.use(router);
app.use(vuetify);

// Mount the app
app.mount('#app');
