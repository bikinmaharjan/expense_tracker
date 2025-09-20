<template>
  <div>
    <!-- Header -->
    <v-row>
      <v-col cols="12">
        <h1 class="text-h4 mb-4">Payments</h1>
      </v-col>
    </v-row>

    <!-- Payment Creation Form -->
    <v-row>
      <v-col cols="12">
        <v-card :loading="loading">
          <v-card-title>New Payment</v-card-title>
          <v-card-text>
            <v-form @submit.prevent="handleSubmit" ref="form">
              <v-row>
                <v-col cols="12" md="6">
                  <v-text-field
                    v-model="newPayment.info"
                    label="Payment Information"
                    required
                    :rules="[(v) => !!v || 'Payment information is required']"
                  />
                </v-col>
                <v-col cols="12" md="6">
                  <v-text-field
                    v-model="newPayment.amount"
                    label="Amount"
                    type="number"
                    prefix="$"
                    required
                    :rules="[
                      (v) => !!v || 'Amount is required',
                      (v) => v > 0 || 'Amount must be greater than 0',
                    ]"
                  />
                </v-col>
                <v-col cols="12" md="6">
                  <v-select
                    v-model="newPayment.tags"
                    :items="tags"
                    label="Tags"
                    item-title="name"
                    item-value="id"
                    multiple
                    chips
                  >
                    <template #chip="{ props, item }">
                      <v-chip v-bind="props" :color="getTagColor(item.raw.id)">
                        {{ item.raw.name }}
                      </v-chip>
                    </template>
                  </v-select>
                </v-col>
                <v-col cols="12" md="6">
                  <v-switch
                    v-model="newPayment.fullyPaid"
                    label="Fully Paid"
                    color="success"
                  />
                </v-col>
                <v-col cols="12" md="6">
                  <v-file-input
                    v-model="newPayment.invoice"
                    label="Invoice"
                    accept="application/pdf,image/*"
                    :show-size="true"
                  />
                </v-col>
                <v-col cols="12" md="6">
                  <v-text-field
                    v-model="newPayment.datePaid"
                    label="Date Paid"
                    type="date"
                    required
                  />
                </v-col>
              </v-row>
              <v-btn
                type="submit"
                color="primary"
                :loading="loading"
                :disabled="loading"
              >
                Create Payment
              </v-btn>
            </v-form>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <!-- Payments List -->
    <v-row class="mt-4">
      <v-col cols="12">
        <v-card :loading="loading">
          <v-card-title>Payment History</v-card-title>
          <v-card-text>
            <v-data-table
              :headers="headers"
              :items="payments"
              :loading="loading"
            >
              <template #[`item.amount`]="{ item }">
                {{ formatCurrency(item.amount) }}
              </template>
              <template #[`item.datePaid`]="{ item }">
                {{ formatDate(item.datePaid) }}
              </template>
              <template #[`item.fullyPaid`]="{ item }">
                <v-chip
                  :color="item.fullyPaid ? 'success' : 'warning'"
                  size="small"
                >
                  {{ item.fullyPaid ? 'Paid' : 'Pending' }}
                </v-chip>
              </template>
              <template #[`item.tags`]="{ item }">
                <v-chip
                  v-for="tag in item.tags"
                  :key="tag"
                  small
                  class="mr-1"
                  :style="{
                    backgroundColor: getTagColor(tag),
                    color: isLightColor(getTagColor(tag)) ? '#111' : '#fff',
                  }"
                >
                  {{ getTagName(tag) }}
                </v-chip>
              </template>
              <template #[`item.documents`]="{ item }">
                <v-chip
                  v-if="item.invoicePath"
                  small
                  class="mr-1"
                  color="grey lighten-4"
                  @click="viewInvoice(item)"
                  style="cursor: pointer"
                  aria-label="Open document"
                >
                  <v-icon small>mdi-file-document</v-icon>
                </v-chip>
              </template>
              <template #[`item.actions`]="{ item }">
                <v-btn
                  icon="mdi-pencil"
                  size="small"
                  color="primary"
                  class="mr-2"
                  @click="editPayment(item)"
                />
                <v-btn
                  icon="mdi-delete"
                  size="small"
                  color="error"
                  @click="confirmDelete(item)"
                />
              </template>
            </v-data-table>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <!-- Edit Dialog -->
    <v-dialog v-model="editDialog" max-width="600px">
      <v-card>
        <v-card-title>Edit Payment</v-card-title>
        <v-card-text>
          <v-form @submit.prevent="handleEdit" ref="editForm">
            <v-row>
              <v-col cols="12" md="6">
                <v-text-field
                  v-model="editingPayment.info"
                  label="Payment Information"
                  required
                  :rules="[(v) => !!v || 'Payment information is required']"
                />
              </v-col>
              <v-col cols="12" md="6">
                <v-text-field
                  v-model="editingPayment.amount"
                  label="Amount"
                  type="number"
                  prefix="$"
                  required
                  :rules="[
                    (v) => !!v || 'Amount is required',
                    (v) => v > 0 || 'Amount must be greater than 0',
                  ]"
                />
              </v-col>
              <v-col cols="12" md="6">
                <v-select
                  v-model="editingPayment.tags"
                  :items="tags"
                  label="Tags"
                  item-title="name"
                  item-value="id"
                  multiple
                  chips
                >
                  <template #chip="{ props, item }">
                    <v-chip v-bind="props" :color="getTagColor(item.raw.id)">
                      {{ item.raw.name }}
                    </v-chip>
                  </template>
                </v-select>
              </v-col>
              <v-col cols="12" md="6">
                <v-switch
                  v-model="editingPayment.fullyPaid"
                  label="Fully Paid"
                  color="success"
                />
              </v-col>
              <v-col cols="12" md="6">
                <v-file-input
                  v-model="newInvoice"
                  label="New Invoice"
                  accept="application/pdf,image/*"
                  :show-size="true"
                />
              </v-col>
              <v-col cols="12" md="6">
                <v-text-field
                  v-model="editingPayment.datePaid"
                  label="Date Paid"
                  type="date"
                  required
                />
              </v-col>
            </v-row>
          </v-form>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn color="error" variant="text" @click="editDialog = false">
            Cancel
          </v-btn>
          <v-btn
            color="primary"
            @click="handleEdit"
            :loading="loading"
            :disabled="loading"
          >
            Save
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Delete Confirmation Dialog -->
    <v-dialog v-model="deleteDialog" max-width="400px">
      <v-card>
        <v-card-title>Delete Payment</v-card-title>
        <v-card-text>
          Are you sure you want to delete this payment?
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn
            color="primary"
            variant="text"
            @click="deleteDialog = false"
            :disabled="loading"
          >
            Cancel
          </v-btn>
          <v-btn
            color="error"
            @click="handleDelete"
            :loading="loading"
            :disabled="loading"
          >
            Delete
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script lang="ts" setup>
import { ref, inject, onMounted } from 'vue';
import dayjs from 'dayjs';
import { endpoints } from '../config/api';

// Types
interface Tag {
  id: string;
  name: string;
  color: string;
}

interface Payment {
  id: string;
  info: string;
  amount: number;
  tags: string[];
  datePaid: string;
  fullyPaid: boolean;
  invoicePath?: string;
}

interface EditingPayment extends Payment {
  invoice?: File;
}

// Notification injected from App.vue
const showNotification = inject('showNotification') as (
  text: string,
  color?: string
) => void;

// Form refs
const form = ref();
const editForm = ref();

// Table headers
const headers = [
  { title: 'Date', key: 'datePaid' },
  { title: 'Information', key: 'info' },
  { title: 'Amount', key: 'amount' },
  { title: 'Tags', key: 'tags' },
  { title: 'Documents', key: 'documents' },
  { title: 'Status', key: 'fullyPaid' },
  { title: 'Actions', key: 'actions' },
];

// State
const loading = ref(false);
const payments = ref<Payment[]>([]);
const tags = ref<Tag[]>([]);
const newPayment = ref({
  info: '',
  amount: 0,
  tags: [] as string[],
  datePaid: dayjs().format('YYYY-MM-DD'),
  fullyPaid: false,
  invoice: null as File | null,
});

// Edit dialog state
const editDialog = ref(false);
const editingPayment = ref<EditingPayment>({
  id: '',
  info: '',
  amount: 0,
  tags: [],
  datePaid: dayjs().format('YYYY-MM-DD'),
  fullyPaid: false,
});
const newInvoice = ref<File | null>(null);

// Delete dialog state
const deleteDialog = ref(false);
const paymentToDelete = ref<Payment | null>(null);

// Format helpers
const formatCurrency = (amount: number) => {
  return new Intl.NumberFormat('en-AU', {
    style: 'currency',
    currency: 'AUD',
  }).format(amount);
};

const formatDate = (date: string) => {
  return dayjs(date).format('DD/MM/YYYY');
};

// Tag helpers
const getTagColor = (tagId: string) => {
  const tag = tags.value.find((t) => t.id === tagId);
  return tag?.color || 'default';
};

const getTagName = (tagId: string) => {
  const tag = tags.value.find((t) => t.id === tagId);
  return tag?.name || tagId;
};

// Helper: determine if a hex color is light
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

// Methods
const fetchTags = async () => {
  try {
    const response = await endpoints.tags.list();
    tags.value = response.data;
  } catch (error) {
    console.error('Failed to fetch tags:', error);
    showNotification('Failed to load tags', 'error');
  }
};

const fetchPayments = async () => {
  loading.value = true;
  try {
    const response = await endpoints.payments.list();
    // Ensure payments.value is always an array
    const paymentData = response.data?.results || response.data || [];
    payments.value = Array.isArray(paymentData) ? paymentData : [];
  } catch (error) {
    console.error('Failed to fetch payments:', error);
    showNotification('Failed to load payments', 'error');
  } finally {
    loading.value = false;
  }
};

const handleSubmit = async () => {
  if (!form.value?.validate()) return;

  loading.value = true;
  try {
    // Create payment without invoice first
    const paymentData = {
      info: newPayment.value.info,
      amount: Number(newPayment.value.amount),
      datePaid: newPayment.value.datePaid,
      fullyPaid: newPayment.value.fullyPaid,
      tags: newPayment.value.tags,
    };

    const response = await endpoints.payments.create(paymentData);

    // If invoice exists, upload it separately
    if (newPayment.value.invoice) {
      const formData = new FormData();
      formData.append('invoice', newPayment.value.invoice);
      await endpoints.payments.uploadInvoice(
        response.data.id,
        newPayment.value.invoice
      );
    }
    showNotification('Payment created successfully');
    newPayment.value = {
      info: '',
      amount: 0,
      tags: [],
      datePaid: dayjs().format('YYYY-MM-DD'),
      fullyPaid: false,
      invoice: null,
    };
    await fetchPayments();
  } catch (error) {
    console.error('Failed to create payment:', error);
    showNotification('Failed to create payment', 'error');
  } finally {
    loading.value = false;
  }
};

const editPayment = (payment: Payment) => {
  editingPayment.value = {
    ...payment,
    // Ensure all required fields are present
    id: payment.id || '',
    info: payment.info || '',
    amount: payment.amount || 0,
    tags: payment.tags || [],
    // Normalize incoming date to YYYY-MM-DD for the date input
    datePaid: payment.datePaid
      ? dayjs(payment.datePaid).format('YYYY-MM-DD')
      : dayjs().format('YYYY-MM-DD'),
    fullyPaid: payment.fullyPaid || false,
  };
  newInvoice.value = null;
  editDialog.value = true;
};

const handleEdit = async () => {
  if (!editForm.value?.validate() || !editingPayment.value) return;

  loading.value = true;
  try {
    // Backend expects a time.Time for datePaid when updating.
    // We accept editingPayment.value.datePaid as either an ISO string or YYYY-MM-DD.
    let isoDate: string | undefined;
    if (editingPayment.value.datePaid) {
      // If it already contains a 'T', assume it's an ISO-like string
      if (String(editingPayment.value.datePaid).includes('T')) {
        const d = new Date(editingPayment.value.datePaid as any);
        if (!isNaN(d.getTime())) isoDate = d.toISOString();
      } else {
        // Treat as YYYY-MM-DD
        const d = new Date(
          String(editingPayment.value.datePaid) + 'T00:00:00Z'
        );
        if (!isNaN(d.getTime())) isoDate = d.toISOString();
      }
    }

    const payload: any = {
      info: editingPayment.value.info,
      amount: Number(editingPayment.value.amount),
      tags: editingPayment.value.tags,
      fullyPaid: editingPayment.value.fullyPaid,
    };
    if (isoDate) payload.datePaid = isoDate;

    await endpoints.payments.update(editingPayment.value.id, payload);

    if (newInvoice.value) {
      await endpoints.payments.uploadInvoice(
        editingPayment.value.id,
        newInvoice.value
      );
    }

    showNotification('Payment updated successfully');
    editDialog.value = false;
    await fetchPayments();
  } catch (error) {
    console.error('Failed to update payment:', error);
    showNotification('Failed to update payment', 'error');
  } finally {
    loading.value = false;
  }
};

const confirmDelete = (payment: Payment) => {
  paymentToDelete.value = payment;
  deleteDialog.value = true;
};

const handleDelete = async () => {
  if (!paymentToDelete.value) return;

  loading.value = true;
  try {
    await endpoints.payments.delete(paymentToDelete.value.id);
    showNotification('Payment deleted successfully');
    deleteDialog.value = false;
    await fetchPayments();
  } catch (error) {
    console.error('Failed to delete payment:', error);
    showNotification('Failed to delete payment', 'error');
  } finally {
    loading.value = false;
  }
};

const viewInvoice = async (payment: Payment) => {
  loading.value = true;
  try {
    // Fetch the latest payment record to inspect invoicePath
    const pResp = await endpoints.payments.get(payment.id);
    const invoicePath = pResp.data?.invoicePath;
    if (!invoicePath) {
      showNotification('No document attached to this payment', 'warning');
      return;
    }

    // Try to download the invoice blob
    const response = await endpoints.payments.downloadInvoice(payment.id);
    const blob =
      response.data instanceof Blob ? response.data : new Blob([response.data]);
    const url = window.URL.createObjectURL(blob);
    window.open(url, '_blank');
  } catch (error: any) {
    console.error('Failed to open invoice:', error);
    // Provide a clearer message when server returns 404
    if (error?.response?.status === 404) {
      showNotification(
        'Document not found on server (404). It may have been deleted.',
        'error'
      );
    } else {
      showNotification('Failed to open document', 'error');
    }
  } finally {
    loading.value = false;
  }
};

// Load data when component mounts
onMounted(async () => {
  await Promise.all([fetchTags(), fetchPayments()]);
});
</script>