import { describe, it, expect, beforeEach, vi } from 'vitest';
import { mount } from '@vue/test-utils';
import { createVuetify } from 'vuetify';
import * as components from 'vuetify/components';
import * as directives from 'vuetify/directives';
import type { AxiosResponse } from 'axios';
import { endpoints } from '../../config/api';
import TagsView from '../TagsView.vue';

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

describe('TagsView', () => {
  const mockTags = [
    { id: '1', name: 'Test Tag', color: '#FF0000', createdAt: '2025-09-19', updatedAt: '2025-09-19' },
    { id: '2', name: 'Another Tag', color: '#00FF00', createdAt: '2025-09-19', updatedAt: '2025-09-19' },
  ];

  beforeEach(() => {
    vi.clearAllMocks();
    // Mock the notifications inject
    vi.mock('vue', async () => {
      const actual = await vi.importActual('vue');
      return {
        ...actual,
        inject: () => (message: string) => console.log(message),
      };
    });
  });

  it('renders properly', () => {
    const wrapper = mount(TagsView, {
      global: {
        plugins: [vuetify],
      },
    });

    expect(wrapper.find('h1').text()).toBe('Tags');
  });

  it('loads and displays tags', async () => {
    // Mock the API response
    vi.mocked(endpoints.tags.list).mockResolvedValue(createAxiosResponse(mockTags));

    const wrapper = mount(TagsView, {
      global: {
        plugins: [vuetify],
      },
    });

    // Wait for tags to load
    await wrapper.vm.$nextTick();

    // Check if tags are displayed
    const rows = wrapper.findAll('tbody tr');
    expect(rows).toHaveLength(mockTags.length);
    expect(rows[0]?.text()).toContain('Test Tag');
    expect(rows[1]?.text()).toContain('Another Tag');
  });

  it('creates a new tag', async () => {
    const newTag = { name: 'New Tag', color: '#0000FF' };
    vi.mocked(endpoints.tags.create).mockResolvedValue(
      createAxiosResponse({ ...newTag, id: '3' })
    );
    vi.mocked(endpoints.tags.list).mockResolvedValue(createAxiosResponse(mockTags));

    const wrapper = mount(TagsView, {
      global: {
        plugins: [vuetify],
      },
    });

    // Fill in the form
    await wrapper.find('[data-test="tag-name-input"]').setValue(newTag.name);
    await wrapper.find('[data-test="tag-color-picker"]').setValue(newTag.color);

    // Submit the form
    await wrapper.find('form').trigger('submit.prevent');

    // Verify API was called with correct data
    expect(endpoints.tags.create).toHaveBeenCalledWith(newTag);

    // Verify list was refreshed
    expect(endpoints.tags.list).toHaveBeenCalled();
  });

  it('deletes a tag', async () => {
    vi.mocked(endpoints.tags.list).mockResolvedValue(createAxiosResponse(mockTags));
    vi.mocked(endpoints.tags.delete).mockResolvedValue(
      createAxiosResponse({ message: 'Tag deleted' })
    );

    const wrapper = mount(TagsView, {
      global: {
        plugins: [vuetify],
      },
    });

    await wrapper.vm.$nextTick();

    // Click delete button for first tag
    await wrapper.find('[data-test="delete-tag-btn"]').trigger('click');

    // Confirm deletion in dialog
    await wrapper.find('[data-test="confirm-delete-btn"]').trigger('click');

    // Verify API was called
    expect(endpoints.tags.delete).toHaveBeenCalledWith(mockTags[0]?.id);

    // Verify list was refreshed
    expect(endpoints.tags.list).toHaveBeenCalled();
  });

  it('updates a tag', async () => {
    const updatedTag = { ...mockTags[0]!, name: 'Updated Tag' };
    vi.mocked(endpoints.tags.list).mockResolvedValue(createAxiosResponse(mockTags));
    vi.mocked(endpoints.tags.update).mockResolvedValue(createAxiosResponse(updatedTag));

    const wrapper = mount(TagsView, {
      global: {
        plugins: [vuetify],
      },
    });

    await wrapper.vm.$nextTick();

    // Click edit button
    await wrapper.find('[data-test="edit-tag-btn"]').trigger('click');

    // Update name in dialog
    await wrapper.find('[data-test="edit-tag-name-input"]').setValue(updatedTag.name);

    // Save changes
    await wrapper.find('[data-test="save-tag-btn"]').trigger('click');

    // Verify API was called with updated data
    expect(endpoints.tags.update).toHaveBeenCalledWith(updatedTag.id, {
      name: updatedTag.name,
      color: updatedTag.color,
    });

    // Verify list was refreshed
    expect(endpoints.tags.list).toHaveBeenCalled();
  });
});