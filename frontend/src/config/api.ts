import axios from 'axios';

export const API_URL = 'http://localhost:8080/api';

export const api = axios.create({
  baseURL: API_URL,
});

// Response interceptor
api.interceptors.response.use(
  (response) => response,
  (error) => {
    console.error('API Error:', error.response?.data || error.message);
    return Promise.reject(error);
  }
);

// Request interceptor for handling file uploads
api.interceptors.request.use((config) => {
  if (config.data instanceof FormData) {
    // Let the browser set the Content-Type including the multipart boundary.
    // Ensure we remove any Content-Type that may have been set by defaults.
    if (config.headers) {
      delete config.headers['Content-Type'];
      delete config.headers['content-type'];
    }
  } else {
    // For non-FormData requests ensure Content-Type is application/json
    if (config.headers) config.headers['Content-Type'] = 'application/json';
  }
  return config;
});

export interface ApiResponse<T = any> {
  data: T;
  message?: string;
  error?: string;
}

// API endpoints
export const endpoints = {
  // Tags
  tags: {
    list: () => api.get('/tags'),
    create: (data: any) => api.post('/tags', data),
    update: (id: string, data: any) => api.put(`/tags/${id}`, data),
    delete: (id: string) => api.delete(`/tags/${id}`),
    stats: () => api.get('/tags/stats'),
  },
  
  // Payments
  payments: {
    list: (params?: any) => api.get('/payments', { params }),
    analytics: () => api.get('/payments/analytics'),
    create: (data: {
      info: string;
      amount: number;
      datePaid: string;
      fullyPaid: boolean;
      tags: string[];
    }) => api.post('/payments', data),
    get: (id: string) => api.get(`/payments/${id}`),
    update: (id: string, data: any) => api.put(`/payments/${id}`, data),
    delete: (id: string) => api.delete(`/payments/${id}`),
    uploadInvoice: (id: string, file: File) => {
      const formData = new FormData();
      formData.append('invoice', file);
      return api.post(`/payments/${id}/invoice`, formData);
    },
    downloadInvoice: (id: string) => api.get(`/payments/${id}/invoice`, { responseType: 'blob' }),
  },
  
  // Documents
  documents: {
    list: (params?: any) => api.get('/documents', { params }),
    create: (data: FormData) => api.post('/documents', data),
    get: (id: string) => api.get(`/documents/${id}`),
    update: (id: string, data: any) => api.put(`/documents/${id}`, data),
    delete: (id: string) => api.delete(`/documents/${id}`),
    download: (id: string) => api.get(`/documents/${id}/download`, { responseType: 'blob' }),
  },
};