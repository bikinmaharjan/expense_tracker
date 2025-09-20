<template>
  <div class="dashboard-root">
    <v-row>
      <v-col cols="12">
        <h1 class="text-h4 mb-4">Dashboard</h1>
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="12" md="4">
        <v-card>
          <v-card-title class="text-h6">Total Expenses</v-card-title>
          <v-card-text class="text-h4">
            <v-progress-circular v-if="loading" indeterminate />
            <template v-else>
              {{ formatCurrency(totalExpenses) }}
            </template>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" md="4">
        <v-card>
          <v-card-title class="text-h6">Pending Payments</v-card-title>
          <v-card-text class="text-h4">
            <v-progress-circular v-if="loading" indeterminate />
            <template v-else>
              {{ pendingPayments }}
            </template>
          </v-card-text>
        </v-card>
      </v-col>

      <v-col cols="12" md="4">
        <v-card>
          <v-card-title class="text-h6">Monthly Expenses</v-card-title>
          <v-card-subtitle>
            <v-select
              :items="monthOptions"
              v-model="selectedMonth"
              item-title="label"
              item-value="value"
              label="Month"
              dense
            />
          </v-card-subtitle>
          <v-card-text class="text-h4">
            <v-progress-circular v-if="loading" indeterminate />
            <template v-else>
              {{ formatCurrency(monthlyExpenses) }}
            </template>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-row class="mt-4">
      <v-col cols="12">
        <v-card>
          <v-card-title class="text-h6">Payments</v-card-title>
          <v-card-subtitle>
            <v-tabs v-model="activeTab">
              <v-tab :value="0">Recent Payments</v-tab>
              <v-tab :value="1">Payments for {{ selectedMonthLabel }}</v-tab>
            </v-tabs>
          </v-card-subtitle>
          <v-card-text>
            <div v-if="activeTab === 0">
              <v-table :loading="loading">
                <thead>
                  <tr>
                    <th>Date</th>
                    <th>Description</th>
                    <th>Tags</th>
                    <th>Amount</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="payment in recentPayments" :key="payment.id">
                    <td>{{ formatDate(payment.datePaid) }}</td>
                    <td>{{ payment.info }}</td>
                    <td>
                      <v-chip
                        v-for="tag in payment.tags || []"
                        :key="tag"
                        small
                        :style="{
                          backgroundColor: getTagColor(tag),
                          color: isLightColor(getTagColor(tag))
                            ? '#111'
                            : '#fff',
                        }"
                        class="mr-1"
                      >
                        {{ getTagName(tag) }}
                      </v-chip>
                    </td>
                    <td>{{ formatCurrency(payment.amount) }}</td>
                    <td>
                      <v-chip
                        :color="payment.fullyPaid ? 'success' : 'warning'"
                        size="small"
                      >
                        {{ payment.fullyPaid ? 'Paid' : 'Pending' }}
                      </v-chip>
                    </td>
                  </tr>
                </tbody>
              </v-table>
            </div>
            <div v-else>
              <v-table :loading="loading">
                <thead>
                  <tr>
                    <th>Date</th>
                    <th>Description</th>
                    <th>Tags</th>
                    <th>Amount</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="payment in monthPayments" :key="payment.id">
                    <td>{{ formatDate(payment.datePaid) }}</td>
                    <td>{{ payment.info }}</td>
                    <td>
                      <v-chip
                        v-for="tag in payment.tags || []"
                        :key="tag"
                        small
                        :style="{
                          backgroundColor: getTagColor(tag),
                          color: isLightColor(getTagColor(tag))
                            ? '#111'
                            : '#fff',
                        }"
                        class="mr-1"
                      >
                        {{ getTagName(tag) }}
                      </v-chip>
                    </td>
                    <td>{{ formatCurrency(payment.amount) }}</td>
                    <td>
                      <v-chip
                        :color="payment.fullyPaid ? 'success' : 'warning'"
                        size="small"
                      >
                        {{ payment.fullyPaid ? 'Paid' : 'Pending' }}
                      </v-chip>
                    </td>
                  </tr>
                </tbody>
              </v-table>
            </div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, inject, watch } from 'vue';
import dayjs from 'dayjs';
import { endpoints } from '../config/api';

const showNotification = inject('showNotification') as (
  text: string,
  color?: string
) => void;

const loading = ref(false);
interface Payment {
  id: string;
  info: string;
  amount: number;
  datePaid: string;
  fullyPaid: boolean;
  tags?: string[];
}

const totalExpenses = ref(0);
const pendingPayments = ref(0);
const monthlyExpenses = ref(0);
const recentPayments = ref<Payment[]>([]);

const tagsMap = ref<Record<string, { name: string; color: string }>>({});

const selectedMonth = ref(dayjs().format('YYYY-MM'));
const monthOptions = ref<{ label: string; value: string }[]>([]);
const monthlyStats = ref<any[]>([]);
const activeTab = ref(0);
const monthPayments = ref<Payment[]>([]);
const selectedMonthLabel = ref(dayjs().format('MMMM YYYY'));

const formatCurrency = (amount: number) =>
  new Intl.NumberFormat('en-AU', { style: 'currency', currency: 'AUD' }).format(
    amount
  );
const formatDate = (date: string) => dayjs(date).format('DD/MM/YYYY');

const fetchListAndRecent = async () => {
  try {
    const response = await endpoints.payments.list({
      stats: true,
      limit: 10,
      sort: '-datePaid',
    });
    const stats = response.data?.stats || {};
    pendingPayments.value = Number(stats.pending) || 0;

    const paymentsData = response.data?.results || response.data || [];
    recentPayments.value = (Array.isArray(paymentsData) ? paymentsData : [])
      .map((p: any) => ({
        id: p.id || '',
        info: p.info || '',
        amount: Number(p.amount) || 0,
        datePaid: p.datePaid || dayjs().format('YYYY-MM-DD'),
        fullyPaid: Boolean(p.fullyPaid),
        tags: Array.isArray(p.tags) ? p.tags : [],
      }))
      .filter((p) => p.id && p.info);
  } catch (err) {
    console.error('Failed to fetch payments list', err);
    showNotification('Failed to load recent payments', 'error');
  }
};

const fetchTags = async () => {
  try {
    const resp = await endpoints.tags.list();
    const list = resp.data || [];
    const map: Record<string, { name: string; color: string }> = {};
    for (const t of list) {
      if (t.id) map[t.id] = { name: t.name || '', color: t.color || '#1976D2' };
    }
    tagsMap.value = map;
  } catch (err) {
    console.error('Failed to fetch tags', err);
  }
};

const fetchAnalytics = async () => {
  try {
    const resp = await endpoints.payments.analytics();
    const data = resp.data || {};
    totalExpenses.value = Number(data.total_stats?.total_amount) || 0;
    monthlyStats.value = Array.isArray(data.monthly_stats)
      ? data.monthly_stats
      : [];

    monthOptions.value = monthlyStats.value.map((m: any) => ({
      label: dayjs(`${m.year}-${m.month}-01`).format('MMMM YYYY'),
      value: `${m.year}-${m.month}`,
    }));
    const current = dayjs().format('YYYY-MM');
    if (!monthOptions.value.find((o) => o.value === current))
      monthOptions.value.unshift({
        label: dayjs(current + '-01').format('MMMM YYYY'),
        value: current,
      });
    if (!selectedMonth.value) selectedMonth.value = current;

    const found = monthlyStats.value.find(
      (m: any) => `${m.year}-${m.month}` === selectedMonth.value
    );
    monthlyExpenses.value = found ? Number(found.amount) : 0;
  } catch (err) {
    console.error('Failed to fetch analytics', err);
    showNotification('Failed to load analytics', 'error');
  }
};

watch(selectedMonth, (val) => {
  const found = monthlyStats.value.find(
    (m: any) => `${m.year}-${m.month}` === val
  );
  monthlyExpenses.value = found ? Number(found.amount) : 0;
  selectedMonthLabel.value = dayjs(val + '-01').format('MMMM YYYY');
  // if the month tab is active, refresh month payments
  if (activeTab.value === 1) fetchMonthPayments();
});

const fetchMonthPayments = async () => {
  if (!selectedMonth.value) return;
  const [year, month] = selectedMonth.value.split('-');
  const startDate = `${year}-${month}-01`;
  const endDate = dayjs(startDate).endOf('month').format('YYYY-MM-DD');
  try {
    const resp = await endpoints.payments.list({
      start_date: startDate,
      end_date: endDate,
      limit: 100,
      sort: '-datePaid',
    });
    const paymentsData = resp.data?.results || resp.data || [];
    monthPayments.value = (Array.isArray(paymentsData) ? paymentsData : [])
      .map((p: any) => ({
        id: p.id || '',
        info: p.info || '',
        amount: Number(p.amount) || 0,
        datePaid: p.datePaid || dayjs().format('YYYY-MM-DD'),
        fullyPaid: Boolean(p.fullyPaid),
        tags: Array.isArray(p.tags) ? p.tags : [],
      }))
      .filter((p) => p.id && p.info);
  } catch (err) {
    console.error('Failed to fetch month payments', err);
  }
};

// fetch month payments when switching to the month tab
watch(activeTab, (val) => {
  if (val === 1) fetchMonthPayments();
});

// Helper to determine light/dark text for tag chips
const isLightColor = (hex: string) => {
  if (!hex) return false;
  const h = hex.replace('#', '');
  const bigint = parseInt(
    h.length === 3
      ? h
          .split('')
          .map((c) => c + c)
          .join('')
      : h,
    16
  );
  const r = (bigint >> 16) & 255;
  const g = (bigint >> 8) & 255;
  const b = bigint & 255;
  const luminance = 0.2126 * r + 0.7152 * g + 0.0722 * b;
  return luminance > 200;
};

onMounted(async () => {
  loading.value = true;
  await Promise.all([fetchListAndRecent(), fetchAnalytics(), fetchTags()]);
  loading.value = false;
});

const getTagName = (id: string) => tagsMap.value[id]?.name || id;
const getTagColor = (id: string) => tagsMap.value[id]?.color || '#1976D2';
</script>

<style scoped>
.v-card {
  height: 100%;
}
</style>


<style scoped>
.v-card {
  height: 100%;
}
</style>