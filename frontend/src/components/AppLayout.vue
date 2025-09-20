<template>
  <v-app>
    <!-- Navigation drawer -->
    <v-navigation-drawer v-model="drawer" app>
      <v-list>
        <v-list-item
          v-for="route in routes"
          :key="route.path"
          :to="route.path"
          :prepend-icon="route.icon"
          :title="route.title"
        />
      </v-list>
    </v-navigation-drawer>

    <!-- App bar -->
    <v-app-bar app>
      <v-app-bar-nav-icon @click="drawer = !drawer" />
      <v-app-bar-title>Expense Tracker</v-app-bar-title>
    </v-app-bar>

    <!-- Main content -->
    <v-main>
      <v-container fluid>
        <router-view v-slot="{ Component }">
          <v-fade-transition mode="out-in">
            <component :is="Component" />
          </v-fade-transition>
        </router-view>
      </v-container>
    </v-main>

    <!-- Snackbar for notifications -->
    <v-snackbar
      v-model="snackbar.show"
      :color="snackbar.color"
      :timeout="snackbar.timeout"
    >
      {{ snackbar.text }}
      <template v-slot:actions>
        <v-btn variant="text" @click="snackbar.show = false"> Close </v-btn>
      </template>
    </v-snackbar>
  </v-app>
</template>

<script lang="ts">
import { defineComponent, ref, computed } from 'vue';
import { useRouter } from 'vue-router';

export default defineComponent({
  name: 'AppLayout',
  setup() {
    // Router setup
    const router = useRouter();
    const menuRoutes = computed(() =>
      (router.options.routes || [])
        .filter((route) => route.meta && route.meta.title)
        .map((route: any) => ({
          path: route.path,
          icon: route.meta?.icon || '',
          title: route.meta?.title || '',
        }))
    );

    // Drawer state
    const drawer = ref(true);

    // Snackbar state
    const snackbar = ref({
      show: false,
      text: '',
      color: 'success',
      timeout: 3000,
    });

    // Show notification method
    const showNotification = (text: string, color = 'success') => {
      snackbar.value = {
        show: true,
        text,
        color,
        timeout: 3000,
      };
    };

    return {
      routes: menuRoutes,
      drawer,
      snackbar,
      showNotification,
    };
  },
});
</script>

<style scoped>
.v-main {
  background-color: var(--bg);
}

/* Ensure the navigation drawer and app bar use surface color */
.v-application .v-navigation-drawer,
.v-application .v-app-bar {
  background: var(--surface) !important;
  color: var(--text) !important;
  box-shadow: 0 6px 18px rgba(15, 23, 36, 0.06);
}
</style>