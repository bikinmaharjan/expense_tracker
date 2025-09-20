import { vi } from 'vitest';
import { createVuetify } from 'vuetify';
import * as components from 'vuetify/components';
import * as directives from 'vuetify/directives';
import dayjs from 'dayjs';

// Create Vuetify instance
export const vuetify = createVuetify({
  components,
  directives,
});

// Mock dayjs
vi.mock('dayjs', () => {
  return {
    default: vi.fn(() => ({
      format: vi.fn(() => '2025-09-19'),
    })),
  };
});

// Mock window.URL.createObjectURL
if (typeof window !== 'undefined') {
  window.URL.createObjectURL = vi.fn();
}

// Mock API endpoints
vi.mock('../config/api', () => ({
  endpoints: {
    tags: {
      list: vi.fn(),
      create: vi.fn(),
      update: vi.fn(),
      delete: vi.fn(),
      stats: vi.fn(),
    },
    payments: {
      list: vi.fn(),
      create: vi.fn(),
      update: vi.fn(),
      delete: vi.fn(),
      uploadInvoice: vi.fn(),
      downloadInvoice: vi.fn(),
    },
    documents: {
      list: vi.fn(),
      create: vi.fn(),
      update: vi.fn(),
      delete: vi.fn(),
      download: vi.fn(),
    },
  },
}));