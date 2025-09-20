<template>
  <div>
    <v-row>
      <v-col cols="12">
        <h1 class="text-h4 mb-4">Documents</h1>
      </v-col>
    </v-row>

    <!-- Document Upload Form -->
    <v-row>
      <v-col cols="12">
        <v-card>
          <v-card-title>Upload Document</v-card-title>
          <v-card-text>
            <v-form @submit.prevent="handleSubmit" ref="form">
              <v-row>
                <v-col cols="12" md="6">
                  <v-text-field
                    v-model="newDocument.title"
                    label="Title"
                    required
                    :rules="[(v) => !!v || 'Title is required']"
                  />
                </v-col>
                <v-col cols="12" md="6">
                  <v-textarea
                    v-model="newDocument.description"
                    label="Description"
                    rows="3"
                  />
                </v-col>
                <v-col cols="12" md="6">
                  <v-select
                    v-model="newDocument.tags"
                    :items="tags"
                    label="Tags"
                    item-title="name"
                    item-value="id"
                    multiple
                    chips
                  >
                    <template #chip="{ props, item }">
                      <v-chip
                        v-bind="props"
                        :style="{
                          backgroundColor: getTagColor(item.raw.id),
                          color: isLightColor(getTagColor(item.raw.id))
                            ? '#111'
                            : '#fff',
                        }"
                      >
                        {{ item.raw.name }}
                      </v-chip>
                    </template>
                  </v-select>
                </v-col>
                <v-col cols="12" md="6">
                  <v-file-input
                    v-model="newDocument.file"
                    label="Document File"
                    accept="application/pdf,image/*,.doc,.docx,.txt"
                    :show-size="true"
                    required
                    :rules="[
                      (v) => !!v || 'Document file is required',
                      (v) =>
                        !v ||
                        v.size < 10 * 1024 * 1024 ||
                        'File size should be less than 10MB',
                      (v) => !v || validateFileType(v) || 'Invalid file type',
                    ]"
                  />
                </v-col>
              </v-row>
              <v-btn type="submit" color="primary" :loading="loading">
                Upload Document
              </v-btn>
            </v-form>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <!-- Documents List -->
    <v-row class="mt-4">
      <v-col cols="12">
        <v-card>
          <v-card-title class="d-flex align-center">
            Documents List
            <v-spacer />
            <v-text-field
              v-model="search"
              append-icon="mdi-magnify"
              label="Search"
              single-line
              hide-details
              density="compact"
              class="ml-4"
              style="max-width: 300px"
            />
          </v-card-title>
          <v-card-text>
            <v-data-table
              :headers="headers"
              :items="documents"
              :search="search"
              :loading="loading"
            >
              <template #[`item.type`]="{ item }">
                <v-icon
                  :color="
                    isPDF(item.type)
                      ? 'error'
                      : isImage(item.type)
                      ? 'primary'
                      : 'grey'
                  "
                >
                  {{ getFileTypeIcon(item.type) }}
                </v-icon>
              </template>

              <template #[`item.createdAt`]="{ item }">
                {{ formatDate(item.createdAt) }}
              </template>
              <template #[`item.originalName`]="{ item }">
                <v-tooltip>
                  <template #activator="{ props }">
                    <v-btn
                      v-bind="props"
                      variant="text"
                      @click="viewDocument(item)"
                      aria-label="Open file"
                    >
                      <v-icon small>mdi-file</v-icon>
                    </v-btn>
                  </template>
                  <span>{{ item.originalName }}</span>
                </v-tooltip>
              </template>
              <template #[`item.fileSize`]="{ item }">
                {{ formatFileSize(item.fileSize) }}
              </template>
              <template #[`item.tags`]="{ item }">
                <v-chip
                  v-for="tagId in item.tags"
                  :key="tagId"
                  small
                  class="mr-1"
                  :style="{
                    backgroundColor: getTagColor(tagId),
                    color: isLightColor(getTagColor(tagId)) ? '#111' : '#fff',
                  }"
                >
                  {{ getTagName(tagId) }}
                </v-chip>
              </template>
              <template #[`item.actions`]="{ item }">
                <v-btn
                  icon="mdi-pencil"
                  size="small"
                  color="primary"
                  class="mr-2"
                  @click="editDocument(item)"
                  :title="'Edit ' + item.title"
                />
                <v-btn
                  icon="mdi-delete"
                  size="small"
                  color="error"
                  @click="confirmDelete(item)"
                  :title="'Delete ' + item.title"
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
        <v-card-title>Edit Document</v-card-title>
        <v-card-text>
          <v-form @submit.prevent="handleEdit" ref="editForm">
            <v-row>
              <v-col cols="12" md="6">
                <v-text-field
                  v-model="editingDocument.title"
                  label="Title"
                  required
                  :rules="[(v) => !!v || 'Title is required']"
                />
              </v-col>
              <v-col cols="12" md="6">
                <v-textarea
                  v-model="editingDocument.description"
                  label="Description"
                  rows="3"
                />
              </v-col>
              <v-col cols="12" md="6">
                <v-select
                  v-model="editingDocument.tags"
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
                <v-file-input
                  v-model="newFile"
                  label="New Document File"
                  accept="application/pdf,image/*,.doc,.docx,.txt"
                  :show-size="true"
                  :rules="[
                    (v) =>
                      !v ||
                      v.size < 10 * 1024 * 1024 ||
                      'File size should be less than 10MB',
                    (v) => !v || validateFileType(v) || 'Invalid file type',
                  ]"
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
          <v-btn color="primary" @click="handleEdit" :loading="loading">
            Save
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <!-- Delete Confirmation Dialog -->
    <v-dialog v-model="deleteDialog" max-width="400px">
      <v-card>
        <v-card-title>Delete Document</v-card-title>
        <v-card-text>
          Are you sure you want to delete this document?
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn color="primary" variant="text" @click="deleteDialog = false">
            Cancel
          </v-btn>
          <v-btn color="error" @click="handleDelete" :loading="loading">
            Delete
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <!-- Document Viewer Dialog -->
    <DocumentViewer
      v-model="viewerDialog"
      :document-id="selectedDocument?.id || ''"
      :title="selectedDocument?.title || ''"
      :mime-type="getMimeType(selectedDocument?.originalName || '')"
    />
  </div>
</template>

<script lang="ts" setup>
import { ref, inject, onMounted } from 'vue';
import dayjs from 'dayjs';
import { endpoints } from '../config/api';
import DocumentViewer from '../components/DocumentViewer.vue';

// Types
interface Tag {
  id: string;
  name: string;
  color: string;
}

interface Document {
  id: string;
  title: string;
  description: string;
  tags: string[];
  filePath: string;
  originalName: string;
  fileSize: number;
  createdAt: string;
  type: string;
}

interface EditingDocument extends Document {
  file?: File;
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
// Helper function to get file type icon
const getFileTypeIcon = (mimeType: string | undefined) => {
  if (!mimeType) return 'mdi-file';
  if (mimeType.includes('pdf')) return 'mdi-file-pdf';
  if (mimeType.startsWith('image/')) return 'mdi-file-image';
  if (mimeType.includes('msword') || mimeType.includes('wordprocessingml'))
    return 'mdi-file-word';
  if (mimeType.startsWith('text/')) return 'mdi-file-document';
  return 'mdi-file';
};

const headers = [
  {
    title: 'Type',
    key: 'type',
    width: '48px',
    sortable: false,
  },
  { title: 'Title', key: 'title' },
  { title: 'File', key: 'originalName' },
  { title: 'Description', key: 'description' },
  { title: 'Tags', key: 'tags' },
  { title: 'Date Added', key: 'createdAt' },
  { title: 'File Size', key: 'fileSize' },
  { title: 'Actions', key: 'actions' },
];

// State
const loading = ref(false);
const search = ref('');
const documents = ref<Document[]>([]);
const tags = ref<Tag[]>([]);
// Initialize with proper types
interface NewDocument {
  title: string;
  description: string;
  tags: string[];
  file: File | null;
}

const newDocument = ref<NewDocument>({
  title: '',
  description: '',
  tags: [],
  file: null,
});

// Edit dialog state
const editDialog = ref(false);
const editingDocument = ref<EditingDocument>({
  id: '',
  title: '',
  description: '',
  tags: [],
  filePath: '',
  originalName: '',
  fileSize: 0,
  createdAt: '',
  type: '',
});
const newFile = ref<File | null>(null);

// Delete dialog state
const deleteDialog = ref(false);
const documentToDelete = ref<Document | null>(null);
const viewerDialog = ref(false);
const selectedDocument = ref<Document | null>(null);

// Helper function to get MIME type from filename
const isPDF = (mimeType: string | undefined) => mimeType?.includes('pdf');
const isImage = (mimeType: string | undefined) =>
  mimeType?.startsWith('image/');

const getMimeType = (filename: string) => {
  const ext = filename.toLowerCase().split('.').pop() || '';
  const mimeTypes: { [key: string]: string } = {
    pdf: 'application/pdf',
    jpg: 'image/jpeg',
    jpeg: 'image/jpeg',
    png: 'image/png',
    gif: 'image/gif',
    txt: 'text/plain',
    doc: 'application/msword',
    docx: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  };
  return mimeTypes[ext] || 'application/octet-stream';
};

// Show a shortened display name for filenames: prefer the leading part before
// a UUID/suffix. Examples:
//   invoice-83074afc-1782-44e1-acee-182a306aee9b.pdf -> invoice
//   receipt_company-2025-09-01.pdf -> receipt_company
const formatDisplayName = (originalName: string) => {
  if (!originalName) return '';
  // Strip extension
  const base = originalName.replace(/\.[^/.]+$/, '');
  // Split by common separators and by UUID-like patterns
  // Try to find a meaningful leading token (before long hex tokens)
  const parts = base.split(/[-_ ]+/);
  if (parts.length === 0) return base;

  // If first part is short and meaningful, use it. Otherwise use the first
  // token that is not a UUID-like long hex string.
  for (const p of parts) {
    // treat as UUID-like if it contains multiple dashes or long hex
    if (/^[0-9a-fA-F]{8,}$/.test(p)) continue;
    if (p.length >= 2 && p.length <= 30) return p;
  }

  // fallback: use the first part trimmed to 24 chars (defensive)
  return (parts[0] || base).slice(0, 24);
};

// File type validation helper
const validateFileType = (file: File) => {
  const allowedTypes = [
    'application/pdf',
    'image/jpeg',
    'image/png',
    'image/gif',
    'text/plain',
    'application/msword',
    'application/vnd.openxmlformats-officedocument.wordprocessingml.document',
  ];
  return allowedTypes.includes(file.type);
};

// View document
const viewDocument = (doc: Document) => {
  selectedDocument.value = doc;
  viewerDialog.value = true;
};

// Format helpers
const formatDate = (date: string) => {
  return dayjs(date).format('DD/MM/YYYY HH:mm');
};

const formatFileSize = (bytes: number) => {
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  if (bytes < 1024 * 1024 * 1024)
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB';
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
    // Ensure tags is always an array and handle both response formats
    const tagsData = response.data?.results || response.data || [];
    const tagsArray = Array.isArray(tagsData) ? tagsData : [];
    tags.value = tagsArray
      .filter((tag: any) => tag && tag.id)
      .map((tag: any) => ({
        id: tag.id,
        name: tag.name || '',
        color: tag.color || 'default',
      }));
  } catch (error) {
    console.error('Failed to fetch tags:', error);
    showNotification('Failed to load tags', 'error');
  }
};

const fetchDocuments = async () => {
  loading.value = true;
  try {
    const response = await endpoints.documents.list();
    // Ensure documents.value is always an array
    const docsData = response.data?.results || response.data || [];
    const docs = Array.isArray(docsData) ? docsData : [];
    documents.value = docs.map((doc: any): Document & { type?: string } => {
      // Convert tag names to IDs if needed
      let tagIds = doc.tags || [];
      if (
        tagIds.length > 0 &&
        typeof tagIds[0] === 'string' &&
        !tags.value.find((t) => t.id === tagIds[0])
      ) {
        tagIds = tagIds.map((tagName: string) => {
          const tag = tags.value.find((t) => t.name === tagName);
          return tag ? tag.id : tagName;
        });
      }

      return {
        id: doc.id || doc.ID || '',
        title: doc.title || doc.title || '',
        description: doc.description || doc.description || '',
        tags: tagIds,
        // Accept either camelCase or snake_case keys from the API
        filePath: doc.filePath || doc.file_path || '',
        originalName: doc.originalName || doc.original_name || '',
        fileSize: Number(doc.fileSize ?? doc.file_size ?? 0) || 0,
        createdAt: doc.createdAt || doc.created_at || '',
        type: getMimeType(doc.originalName || doc.original_name || ''),
      };
    });
  } catch (error) {
    console.error('Failed to fetch documents:', error);
    showNotification('Failed to load documents', 'error');
  } finally {
    loading.value = false;
  }
};

const handleSubmit = async () => {
  if (!form.value?.validate()) {
    showNotification('Please fix the validation errors', 'error');
    return;
  }

  loading.value = true;
  try {
    // Validate required fields
    if (!newDocument.value.title.trim()) {
      showNotification('Title is required', 'error');
      return;
    }

    if (!newDocument.value.file) {
      showNotification('Document file is required', 'error');
      return;
    }

    const formData = new FormData();
    formData.append('title', newDocument.value.title.trim());
    formData.append('description', newDocument.value.description.trim());

    // Ensure tags is a proper array of valid tag IDs and append each as 'tags'
    const tagIds = Array.isArray(newDocument.value.tags)
      ? newDocument.value.tags.filter((id) =>
          tags.value.some((t) => t.id === id)
        )
      : [];
    tagIds.forEach((id) => formData.append('tags', id.toString()));
    formData.append('file', newDocument.value.file);

    await endpoints.documents.create(formData);
    showNotification('Document uploaded successfully');
    // Reset form with proper typing
    newDocument.value = {
      title: '',
      description: '',
      tags: [],
      file: null,
    } as NewDocument;
    await fetchDocuments();
  } catch (error) {
    console.error('Failed to upload document:', error);
    // Show backend error details when available
    const details =
      (error as any)?.response?.data?.details || (error as any)?.message;
    showNotification(`Failed to upload document: ${details}`, 'error');
  } finally {
    loading.value = false;
  }
};

const editDocument = (doc: Document) => {
  const tagIds = doc.tags || [];
  editingDocument.value = {
    ...doc,
    id: doc.id || '',
    title: doc.title || '',
    description: doc.description || '',
    tags: tagIds,
    filePath: doc.filePath || '',
    originalName: doc.originalName || '',
    fileSize: Number(doc.fileSize) || 0,
    createdAt: doc.createdAt || '',
    type: doc.type || getMimeType(doc.originalName || ''),
  };
  newFile.value = null;
  editDialog.value = true;
};

const handleEdit = async () => {
  if (!editForm.value?.validate() || !editingDocument.value) {
    showNotification('Please fix the validation errors', 'error');
    return;
  }

  loading.value = true;
  try {
    if (!editingDocument.value.title.trim()) {
      showNotification('Title is required', 'error');
      return;
    }

    const formData = new FormData();
    formData.append('title', editingDocument.value.title.trim());
    formData.append('description', editingDocument.value.description.trim());

    // Ensure tags is a proper array of valid tag IDs and append each as 'tags'
    const tagIds = Array.isArray(editingDocument.value.tags)
      ? editingDocument.value.tags.filter((id) =>
          tags.value.some((t) => t.id === id)
        )
      : [];
    tagIds.forEach((id) => formData.append('tags', id.toString()));

    if (newFile.value) {
      formData.append('file', newFile.value);
    }

    await endpoints.documents.update(editingDocument.value.id, formData);

    showNotification('Document updated successfully');
    editDialog.value = false;
    await fetchDocuments();
  } catch (error) {
    console.error('Failed to update document:', error);
    const details =
      (error as any)?.response?.data?.details || (error as any)?.message;
    showNotification(`Failed to update document: ${details}`, 'error');
  } finally {
    loading.value = false;
  }
};

const confirmDelete = (doc: Document) => {
  documentToDelete.value = doc;
  deleteDialog.value = true;
};

const handleDelete = async () => {
  if (!documentToDelete.value) return;

  loading.value = true;
  try {
    await endpoints.documents.delete(documentToDelete.value.id);
    showNotification('Document deleted successfully');
    deleteDialog.value = false;
    await fetchDocuments();
  } catch (error) {
    console.error('Failed to delete document:', error);
    showNotification('Failed to delete document', 'error');
  } finally {
    loading.value = false;
  }
};

// Load data when component mounts
onMounted(async () => {
  await Promise.all([fetchTags(), fetchDocuments()]);
});
</script>