import { describe, it, expect, beforeEach, vi } from 'vitest';
import { mount } from '@vue/test-utils';
import { createVuetify } from 'vuetify';
import * as components from 'vuetify/components';
import * as directives from 'vuetify/directives';
import type { AxiosResponse } from 'axios';
import { endpoints } from '../../config/api';
import DocumentsView from '../DocumentsView.vue';

// Mock the API endpoints
vi.mock('../../config/api', () => ({
  endpoints: {
    tags: {
      list: vi.fn(),
    },
    documents: {
      list: vi.fn(),
      create: vi.fn(),
      delete: vi.fn(),
      download: vi.fn(),
    },
  },
}));

// Create Vuetify instance
const vuetify = createVuetify({
  components,
  directives,
});

// Helper to create mock axios responses
const createAxiosResponse = <T>(data: T): AxiosResponse<T> => ({
  data,
  status: 200,
  statusText: 'OK',
  headers: {},
  config: {} as any,
});

describe('DocumentsView', () => {
  const mockTags = [
    { id: '1', name: 'design', color: '#ff0000' },
    { id: '2', name: 'materials', color: '#00ff00' },
  ] as const;

  const mockDocuments = [
    {
      id: '1',
      title: 'House Design',
      description: 'Initial house design documents',
      tags: ['design'],
      filePath: '/storage/documents/1.pdf',
      originalName: 'house-design.pdf',
      fileSize: 1024 * 1024,
      createdAt: '2025-09-19',
    },
    {
      id: '2',
      title: 'Material Choices',
      description: 'Material selection documentation',
      tags: ['materials'],
      filePath: '/storage/documents/2.pdf',
      originalName: 'materials.pdf',
      fileSize: 512 * 1024,
      createdAt: '2025-09-19',
    },
  ] as const;

  beforeEach(() => {
    vi.clearAllMocks();

    // Mock API responses
    vi.mocked(endpoints.tags.list).mockResolvedValue(createAxiosResponse(mockTags));
    vi.mocked(endpoints.documents.list).mockResolvedValue(createAxiosResponse(mockDocuments));

    // Mock the notifications inject
    vi.mock('vue', async () => {
      const actual = await vi.importActual('vue');
      return {
        ...actual,
        inject: () => (message: string) => console.log(message),
      };
    });

    // Create a more complete document mock
    const elements: { [key: string]: HTMLElement } = {};
    const mockDocument = {
      getElementById: (id: string) => elements[id] || null,
      createElement: (tag: string) => {
        const element = {
          style: {},
          className: '',
          id: '',
          setAttribute: vi.fn(),
          getAttribute: vi.fn(),
          appendChild: vi.fn(),
          removeChild: vi.fn(),
          remove: vi.fn(),
          click: vi.fn(),
          textContent: '',
          href: '',
          download: '',
        };
        return element;
      },
      head: {
        appendChild: vi.fn(),
      },
      body: {
        appendChild: vi.fn(),
        removeChild: vi.fn(),
      },
    };

    // Create a more complete window mock
    const mockWindow = {
      document: mockDocument,
      URL: {
        createObjectURL: vi.fn().mockReturnValue('blob:mock-url'),
        revokeObjectURL: vi.fn(),
      },
      visualViewport: {
        width: 1024,
        height: 768,
      },
      getComputedStyle: vi.fn().mockReturnValue({
        getPropertyValue: vi.fn().mockReturnValue(''),
      }),
    };

    // Set up global mocks
    Object.defineProperty(global, 'document', {
      value: mockDocument,
      writable: true,
    });

    Object.defineProperty(global, 'window', {
      value: mockWindow,
      writable: true,
    });

    // Additional required globals
    global.ResizeObserver = vi.fn().mockImplementation(() => ({
      observe: vi.fn(),
      unobserve: vi.fn(),
      disconnect: vi.fn(),
    }));
  });

  it('renders properly', async () => {
    const wrapper = mount(DocumentsView, {
      global: {
        plugins: [vuetify],
      },
    });

    await wrapper.vm.$nextTick();
    await new Promise(resolve => setTimeout(resolve, 100));

    expect(wrapper.find('.text-h4').text()).toBe('Documents');
    expect(endpoints.tags.list).toHaveBeenCalled();
    expect(endpoints.documents.list).toHaveBeenCalled();
  });

  it('loads and displays documents', async () => {
    const wrapper = mount(DocumentsView, {
      global: {
        plugins: [vuetify],
      },
    });

    await wrapper.vm.$nextTick();
    await new Promise(resolve => setTimeout(resolve, 100));

    const table = wrapper.findComponent({ name: 'v-data-table' });
    expect(table.exists()).toBe(true);
    expect(table.props('items')).toHaveLength(mockDocuments.length);
    
    const tableHtml = wrapper.html();
    expect(tableHtml).toContain('House Design');
    expect(tableHtml).toContain('Material Choices');
  });

  it('creates a new document', async () => {
    const newDocument = {
      title: 'New Document',
      description: 'Test description',
      tags: ['test'],
      file: new File(['test content'], 'test.pdf', { type: 'application/pdf' }),
    };

    const createResponse = {
      id: '3',
      ...newDocument,
      filePath: '/storage/documents/3.pdf',
      originalName: 'test.pdf',
      fileSize: 12,
      createdAt: '2025-09-19',
    };

    vi.mocked(endpoints.documents.create).mockResolvedValue(createAxiosResponse(createResponse));

    const wrapper = mount(DocumentsView, {
      global: {
        plugins: [vuetify],
      },
    });

    await wrapper.vm.$nextTick();
    await new Promise(resolve => setTimeout(resolve, 100));

    await wrapper.find('input[type="text"]').setValue(newDocument.title);
    await wrapper.find('textarea').setValue(newDocument.description);
    
    const tagsSelect = wrapper.findComponent({ name: 'v-select' });
    await tagsSelect.setValue(newDocument.tags);
    
    const fileInput = wrapper.findComponent({ name: 'v-file-input' });
    await fileInput.setValue(newDocument.file);

    await wrapper.find('form').trigger('submit');
    await wrapper.vm.$nextTick();

    expect(endpoints.documents.create).toHaveBeenCalled();
    const createCall = vi.mocked(endpoints.documents.create).mock.calls[0];
    const formData = createCall?.[0] as FormData;
    
    expect(formData).toBeDefined();
    if (formData) {
      expect(formData.get('title')).toBe(newDocument.title);
      expect(formData.get('description')).toBe(newDocument.description);
      expect(formData.get('tags')).toBe(JSON.stringify(newDocument.tags));
      expect(formData.get('file')).toEqual(newDocument.file);
    }

    expect(endpoints.documents.list).toHaveBeenCalled();
  });

  it('deletes a document', async () => {
    vi.mocked(endpoints.documents.delete).mockResolvedValue(
      createAxiosResponse({ message: 'Document deleted' })
    );

    const wrapper = mount(DocumentsView, {
      global: {
        plugins: [vuetify],
      },
    });

    await wrapper.vm.$nextTick();
    await new Promise(resolve => setTimeout(resolve, 100));

    await (wrapper.vm as any).confirmDelete(mockDocuments[0]);
    await wrapper.vm.$nextTick();
    await (wrapper.vm as any).handleDelete();
    await wrapper.vm.$nextTick();

    expect(endpoints.documents.delete).toHaveBeenCalledWith(mockDocuments[0].id);
    expect(endpoints.documents.list).toHaveBeenCalled();
  });

  it('downloads a document', async () => {
    const blob = new Blob(['test content'], { type: 'application/pdf' });
    vi.mocked(endpoints.documents.download).mockResolvedValue(createAxiosResponse(blob));

    const wrapper = mount(DocumentsView, {
      global: {
        plugins: [vuetify],
      },
    });

    await wrapper.vm.$nextTick();
    await new Promise(resolve => setTimeout(resolve, 100));

    await (wrapper.vm as any).downloadFile(mockDocuments[0]);
    await wrapper.vm.$nextTick();

    expect(endpoints.documents.download).toHaveBeenCalledWith(mockDocuments[0].id);
    expect(window.document.createElement).toHaveBeenCalledWith('a');
    expect(window.URL.createObjectURL).toHaveBeenCalledWith(blob);
  });
});