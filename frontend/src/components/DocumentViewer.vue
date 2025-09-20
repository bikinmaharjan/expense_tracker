<template>
  <v-dialog
    :model-value="modelValue"
    @update:model-value="emit('update:modelValue', $event)"
    fullscreen
    hide-overlay
  >
    <v-card>
      <v-toolbar dark color="primary">
        <v-btn icon dark @click="$emit('update:modelValue', false)">
          <v-icon>mdi-close</v-icon>
        </v-btn>
        <v-toolbar-title>{{ title }}</v-toolbar-title>
        <v-spacer></v-spacer>
        <v-btn
          icon
          dark
          @click="() => downloadFile(true)"
          :loading="loading"
          :disabled="loading"
        >
          <v-icon v-if="!loading">mdi-download</v-icon>
          <template v-else>{{ downloadProgress }}%</template>
        </v-btn>
      </v-toolbar>

      <v-card-text class="pa-0">
        <!-- PDF Viewer -->
        <iframe
          v-if="isPDF"
          :src="fileUrl"
          class="document-viewer"
          type="application/pdf"
        ></iframe>

        <!-- Image Viewer -->
        <v-img
          v-else-if="isImage"
          :src="fileUrl"
          max-height="100%"
          contain
          class="document-viewer"
        ></v-img>

        <!-- Text Viewer -->
        <v-sheet v-else-if="isText" class="document-viewer pa-4 text-pre-wrap">
          {{ textContent || '' }}
        </v-sheet>

        <!-- Loading State -->
        <v-sheet
          v-if="loading"
          class="document-viewer d-flex align-center justify-center"
        >
          <v-progress-circular
            indeterminate
            color="primary"
            size="64"
          ></v-progress-circular>
        </v-sheet>

        <!-- Error State -->
        <v-sheet
          v-else-if="error"
          class="document-viewer d-flex align-center justify-center"
        >
          <div class="text-center">
            <v-icon size="64" color="error">mdi-alert-circle</v-icon>
            <div class="text-h6 mt-4">Failed to load document</div>
            <v-btn
              color="primary"
              class="mt-4"
              @click="loadFile"
              :loading="loading"
            >
              Retry
            </v-btn>
          </div>
        </v-sheet>

        <!-- Unsupported Format -->
        <v-sheet
          v-else-if="!isPDF && !isImage && !isText"
          class="document-viewer d-flex align-center justify-center"
        >
          <div class="text-center">
            <v-icon size="64" color="grey">mdi-file-alert</v-icon>
            <div class="text-h6 mt-4">
              Preview not available for this file type
            </div>
            <v-btn
              color="primary"
              class="mt-4"
              prepend-icon="mdi-download"
              @click="() => downloadFile(true)"
              :loading="loading"
              :disabled="loading"
            >
              {{
                loading ? `Downloading ${downloadProgress}%` : 'Download File'
              }}
            </v-btn>
          </div>
        </v-sheet>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script lang="ts">
export default {
  name: 'DocumentViewer',
};
</script>

<script setup lang="ts">
import { computed, inject, onUnmounted, ref, watch } from 'vue';
import { endpoints } from '../config/api';
import { api } from '../config/api';

const props = defineProps({
  modelValue: {
    type: Boolean,
    required: true,
  },
  documentId: {
    type: String,
    required: true,
  },
  title: {
    type: String,
    required: true,
  },
  mimeType: {
    type: String,
    required: false,
    default: undefined,
  },
});

const emit = defineEmits(['update:modelValue']);

// Inject notification function
const showNotification = inject('showNotification') as (
  text: string,
  color?: string
) => void;

// State
const loading = ref(false);
const error = ref(false);
const fileUrl = ref('');
const textContent = ref('');
const downloadProgress = ref(0);

// Cache for document data
const documentCache = new Map<
  string,
  {
    url: string;
    text?: string;
    mimeType: string;
    timestamp: number;
  }
>();

// Cache duration - 5 minutes
const CACHE_DURATION = 5 * 60 * 1000;

// Computed
const isPDF = computed(() => props.mimeType?.includes('pdf'));
const isImage = computed(() => props.mimeType?.startsWith('image/'));
const isText = computed(
  () =>
    props.mimeType?.startsWith('text/') ||
    props.mimeType?.includes('javascript') ||
    props.mimeType?.includes('json')
);

const loadFile = async () => {
  if (!props.documentId) return;

  loading.value = true;
  error.value = false;

  try {
    // Check cache first
    const cached = documentCache.get(props.documentId);
    if (cached && Date.now() - cached.timestamp < CACHE_DURATION) {
      if (isText.value && cached.text) {
        textContent.value = cached.text;
      } else if (cached.url) {
        fileUrl.value = cached.url;
      }
      loading.value = false;
      return;
    }
    const response = await endpoints.documents.download(props.documentId);

    if (isText.value) {
      // For text files, read the content
      const text = await response.data.text();
      textContent.value = text;

      // Cache the text content
      documentCache.set(props.documentId, {
        url: '',
        text,
        mimeType: props.mimeType || '',
        timestamp: Date.now(),
      });
    } else {
      // For other files, create an object URL
      const blob = new Blob([response.data], { type: props.mimeType });
      const url = URL.createObjectURL(blob);
      fileUrl.value = url;

      // Cache the blob URL
      documentCache.set(props.documentId, {
        url,
        mimeType: props.mimeType || '',
        timestamp: Date.now(),
      });
    }
  } catch (err) {
    console.error('Failed to load document:', err);
    error.value = true;
    showNotification('Failed to load document preview', 'error');
  } finally {
    loading.value = false;
  }
};

const downloadFile = async (retry = true) => {
  loading.value = true;
  try {
    downloadProgress.value = 0;
    const response = await api.get(`/documents/${props.documentId}/download`, {
      responseType: 'blob',
      onDownloadProgress: (progressEvent) => {
        if (progressEvent.total) {
          downloadProgress.value = Math.round(
            (progressEvent.loaded * 100) / progressEvent.total
          );
        }
      },
    });
    const url = URL.createObjectURL(new Blob([response.data]));
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', props.title);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url); // Clean up the URL immediately
  } catch (err) {
    console.error('Failed to download document:', err);
    showNotification('Failed to download document', 'error');
    error.value = true;
    // Add retry button to UI
    if (retry) {
      const shouldRetry = window.confirm(
        'Download failed. Would you like to try again?'
      );
      if (shouldRetry) {
        await downloadFile(false); // Prevent infinite retry loops
      }
    }
  } finally {
    loading.value = false;
  }
};

watch(
  () => props.modelValue,
  (newValue) => {
    if (newValue) {
      loadFile();
    } else {
      // Cleanup object URL when dialog is closed
      if (fileUrl.value) {
        URL.revokeObjectURL(fileUrl.value);
        fileUrl.value = '';
      }
      textContent.value = '';
    }
  }
);

// Cleanup on component unmount
onUnmounted(() => {
  // Cleanup all cached URLs
  documentCache.forEach((cache) => {
    if (cache.url) {
      URL.revokeObjectURL(cache.url);
    }
  });
  documentCache.clear();

  if (fileUrl.value) {
    URL.revokeObjectURL(fileUrl.value);
  }
});

// Clean up expired cache entries periodically
const cleanupCache = () => {
  const now = Date.now();
  documentCache.forEach((cache, id) => {
    if (now - cache.timestamp > CACHE_DURATION) {
      if (cache.url) {
        URL.revokeObjectURL(cache.url);
      }
      documentCache.delete(id);
    }
  });
};

// Run cache cleanup every minute
const cleanupInterval = setInterval(cleanupCache, 60000);
onUnmounted(() => {
  clearInterval(cleanupInterval);
});
</script>

<style scoped>
.document-viewer {
  width: 100%;
  height: calc(100vh - 64px); /* Subtract toolbar height */
}

.text-pre-wrap {
  white-space: pre-wrap;
  font-family: monospace;
  overflow-y: auto;
}
</style>