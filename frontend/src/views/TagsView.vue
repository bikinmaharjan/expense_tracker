<template>
  <div>
    <v-row>
      <v-col cols="12">
        <h1 class="text-h4 mb-4">Tags</h1>
      </v-col>
    </v-row>

    <!-- Tag Creation Form -->
    <v-row>
      <v-col cols="12">
        <v-card>
          <v-card-title>Create New Tag</v-card-title>
          <v-card-text>
            <v-form @submit.prevent="handleSubmit" ref="form">
              <v-row>
                <v-col cols="12" md="6">
                  <v-text-field
                    v-model="newTag.name"
                    label="Tag Name"
                    required
                    :rules="[(v) => !!v || 'Tag name is required']"
                    data-test="tag-name-input"
                  />
                </v-col>
                <v-col cols="12" md="6">
                  <v-color-picker
                    v-model="newTag.color"
                    mode="hex"
                    show-swatches
                    swatches-max-height="300"
                    data-test="tag-color-picker"
                  />
                </v-col>
              </v-row>
              <v-btn type="submit" color="primary" :loading="loading">
                Create Tag
              </v-btn>
            </v-form>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <!-- Tags List -->
    <v-row class="mt-4">
      <v-col cols="12">
        <v-card>
          <v-card-title>Existing Tags</v-card-title>
          <v-card-text>
            <v-table>
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Color</th>
                  <th>Usage Count</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="tag in tags" :key="tag.id">
                  <td>{{ tag.name }}</td>
                  <td>
                    <v-chip
                      v-bind:key="tag.id"
                      small
                      :style="{
                        backgroundColor: tag.color,
                        color: isLightColor(tag.color) ? '#111' : '#fff',
                      }"
                    >
                      {{ tag.name }}
                    </v-chip>
                  </td>
                  <td>{{ tag.usageCount || 0 }}</td>
                  <td>
                    <v-btn
                      icon="mdi-pencil"
                      size="small"
                      color="primary"
                      class="mr-2"
                      @click="editTag(tag)"
                      data-test="edit-tag-btn"
                    />
                    <v-btn
                      icon="mdi-delete"
                      size="small"
                      color="error"
                      @click="confirmDelete(tag)"
                      data-test="delete-tag-btn"
                    />
                  </td>
                </tr>
              </tbody>
            </v-table>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <!-- Edit Dialog -->
    <v-dialog v-model="editDialog" max-width="500px">
      <v-card>
        <v-card-title>Edit Tag</v-card-title>
        <v-card-text>
          <v-form @submit.prevent="handleEdit" ref="editForm">
            <v-text-field
              v-model="editingTag.name"
              label="Tag Name"
              required
              :rules="[(v) => !!v || 'Tag name is required']"
              data-test="edit-tag-name-input"
            />
            <v-color-picker
              v-model="editingTag.color"
              mode="hex"
              show-swatches
              swatches-max-height="300"
            />
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
            data-test="save-tag-btn"
          >
            Save
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Delete Confirmation Dialog -->
    <v-dialog v-model="deleteDialog" max-width="400px">
      <v-card>
        <v-card-title>Delete Tag</v-card-title>
        <v-card-text> Are you sure you want to delete this tag? </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn color="primary" variant="text" @click="deleteDialog = false">
            Cancel
          </v-btn>
          <v-btn
            color="error"
            @click="handleDelete"
            :loading="loading"
            data-test="confirm-delete-btn"
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
import { endpoints } from '../config/api';

// Types
interface Tag {
  id: string;
  name: string;
  color: string;
  usageCount?: number;
}

// Notification injected from App.vue
const showNotification = inject('showNotification') as (
  text: string,
  color?: string
) => void;

// Form refs
const form = ref();
const editForm = ref();

// State
const loading = ref(false);
const tags = ref<Tag[]>([]);
const newTag = ref({
  name: '',
  color: '#1976D2',
});

// Edit dialog state
const editDialog = ref(false);
// keep editingTag as a non-null object so template v-model bindings won't error
const editingTag = ref<Tag>({ id: '', name: '', color: '#1976D2' });

// Delete dialog state
const deleteDialog = ref(false);
const tagToDelete = ref<Tag | null>(null);

// Methods
const fetchTags = async () => {
  try {
    // Fetch base tags and stats (separate endpoints)
    const [listResp, statsResp] = await Promise.all([
      endpoints.tags.list(),
      endpoints.tags.stats(),
    ]);

    const baseTags = listResp.data || [];
    const stats = statsResp.data || [];

    // Build maps from tag id and name to stats (backend returns id and name)
    const statsById: Record<string, any> = {};
    const statsByName: Record<string, any> = {};
    for (const s of stats) {
      if (s.id) statsById[s.id] = s;
      if (s.name) statsByName[s.name] = s;
    }

    // Merge usage counts into tags list, prefer id-based mapping (safer)
    tags.value = baseTags.map((t: any) => {
      const s = statsById[t.id] || statsByName[t.name];
      const paymentCount = s?.payment_count || 0;
      const docCount = s?.document_count || 0;
      return {
        ...t,
        usageCount: paymentCount + docCount,
      };
    });
  } catch (error) {
    console.error('Failed to fetch tags:', error);
    showNotification('Failed to load tags', 'error');
  }
};

const handleSubmit = async () => {
  if (!form.value?.validate()) return;

  loading.value = true;
  try {
    await endpoints.tags.create(newTag.value);
    showNotification('Tag created successfully');
    newTag.value = { name: '', color: '#1976D2' };
    await fetchTags();
  } catch (error) {
    console.error('Failed to create tag:', error);
    showNotification('Failed to create tag', 'error');
  } finally {
    loading.value = false;
  }
};

const editTag = (tag: Tag) => {
  editingTag.value = { ...tag };
  editDialog.value = true;
};

const handleEdit = async () => {
  if (!editForm.value?.validate() || !editingTag.value) return;

  loading.value = true;
  try {
    await endpoints.tags.update(editingTag.value.id, {
      name: editingTag.value.name,
      color: editingTag.value.color,
    });
    showNotification('Tag updated successfully');
    editDialog.value = false;
    await fetchTags();
  } catch (error) {
    console.error('Failed to update tag:', error);
    showNotification('Failed to update tag', 'error');
  } finally {
    loading.value = false;
  }
};

const confirmDelete = (tag: Tag) => {
  tagToDelete.value = tag;
  deleteDialog.value = true;
};

const handleDelete = async () => {
  if (!tagToDelete.value) return;

  loading.value = true;
  try {
    await endpoints.tags.delete(tagToDelete.value.id);
    showNotification('Tag deleted successfully');
    deleteDialog.value = false;
    await fetchTags();
  } catch (error) {
    console.error('Failed to delete tag:', error);
    showNotification('Failed to delete tag', 'error');
  } finally {
    loading.value = false;
  }
};

// Load tags when component mounts
onMounted(fetchTags);

// Helper: determine if a hex color is light (so we can use dark text)
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
  // Perceived luminance
  const luminance = 0.2126 * r + 0.7152 * g + 0.0722 * b;
  return luminance > 200; // threshold: >200 considered light
};
</script>