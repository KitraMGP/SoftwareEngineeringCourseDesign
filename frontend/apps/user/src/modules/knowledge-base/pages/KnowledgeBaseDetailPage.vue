<script setup lang="ts">
import type { FormInstance, FormRules, UploadRequestOptions } from 'element-plus';

import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query';
import { ElMessage, ElMessageBox } from 'element-plus';
import { computed, reactive, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import {
  AppStatusBadge,
  SectionHeading,
  StatPanel,
  SurfaceCard,
  knowledgeBaseApi,
  queryKeys
} from '@private-kb/shared';
import { formatDateTime } from '@private-kb/shared/utils/date';
import { formatBytes } from '@private-kb/shared/utils/format';
import { getErrorMessage, toFieldErrorMap } from '@private-kb/shared/utils/errors';

const route = useRoute();
const router = useRouter();
const queryClient = useQueryClient();

const kbId = computed(() => String(route.params.kbId || ''));

const editVisible = ref(false);
const detailDrawerVisible = ref(false);
const selectedDocumentId = ref('');
const formRef = ref<FormInstance>();
const fieldErrors = ref<Record<string, string>>({});

const form = reactive({
  name: '',
  description: '',
  prompt_template: '',
  retrieval_top_k: 5,
  similarity_threshold: 0.3
});

const rules: FormRules<typeof form> = {
  name: [{ required: true, message: '请输入知识库名称', trigger: 'blur' }]
};

const knowledgeBaseQuery = useQuery({
  queryKey: computed(() => queryKeys.knowledgeBase(kbId.value)),
  queryFn: () => knowledgeBaseApi.getKnowledgeBase(kbId.value),
  enabled: computed(() => !!kbId.value)
});

const documentsQuery = useQuery({
  queryKey: computed(() => queryKeys.documents(kbId.value)),
  queryFn: () => knowledgeBaseApi.listDocuments(kbId.value, { page: 1, size: 100 }),
  enabled: computed(() => !!kbId.value),
  refetchInterval: (query) => {
    const items = query.state.data?.items ?? [];
    return items.some((item) => ['pending', 'processing', 'deleting'].includes(item.status))
      ? 5000
      : false;
  }
});

const documentDetailQuery = useQuery({
  queryKey: computed(() => queryKeys.document(kbId.value, selectedDocumentId.value)),
  queryFn: () => knowledgeBaseApi.getDocument(kbId.value, selectedDocumentId.value),
  enabled: computed(() => !!kbId.value && !!selectedDocumentId.value)
});

const updateMutation = useMutation({
  mutationFn: () =>
    knowledgeBaseApi.updateKnowledgeBase(kbId.value, {
      name: form.name,
      description: form.description || null,
      prompt_template: form.prompt_template || null,
      retrieval_top_k: form.retrieval_top_k,
      similarity_threshold: form.similarity_threshold
    }),
  onSuccess: async () => {
    editVisible.value = false;
    await queryClient.invalidateQueries({ queryKey: queryKeys.knowledgeBase(kbId.value) });
    await queryClient.invalidateQueries({ queryKey: queryKeys.knowledgeBasesRoot });
    ElMessage.success('知识库配置已更新。');
  },
  onError: (error) => {
    fieldErrors.value = toFieldErrorMap(error);
    ElMessage.error(getErrorMessage(error));
  }
});

const deleteKnowledgeBaseMutation = useMutation({
  mutationFn: () => knowledgeBaseApi.deleteKnowledgeBase(kbId.value),
  onSuccess: async () => {
    await queryClient.invalidateQueries({ queryKey: queryKeys.knowledgeBasesRoot });
    ElMessage.success('知识库已删除。');
    router.push({ name: 'knowledge-bases' });
  },
  onError: (error) => {
    ElMessage.error(getErrorMessage(error));
  }
});

const reindexMutation = useMutation({
  mutationFn: () => knowledgeBaseApi.reindexKnowledgeBase(kbId.value),
  onSuccess: ({ task_id }) => {
    ElMessage.success(`已提交重建索引任务：${task_id}`);
  },
  onError: (error) => {
    ElMessage.error(getErrorMessage(error));
  }
});

const uploadMutation = useMutation({
  mutationFn: (file: File) => knowledgeBaseApi.uploadDocument(kbId.value, file),
  onSuccess: async (result) => {
    ElMessage.success(`文档已进入处理队列：${result.document_id}`);
    await queryClient.invalidateQueries({ queryKey: queryKeys.documentsRoot(kbId.value) });
  },
  onError: (error) => {
    ElMessage.error(getErrorMessage(error));
  }
});

const deleteDocumentMutation = useMutation({
  mutationFn: (documentId: string) => knowledgeBaseApi.deleteDocument(kbId.value, documentId),
  onSuccess: async () => {
    await queryClient.invalidateQueries({ queryKey: queryKeys.documentsRoot(kbId.value) });
    ElMessage.success('文档已删除。');
  },
  onError: (error) => {
    ElMessage.error(getErrorMessage(error));
  }
});

function openEditDialog() {
  const knowledgeBase = knowledgeBaseQuery.data.value;
  if (!knowledgeBase) {
    return;
  }

  fieldErrors.value = {};
  Object.assign(form, {
    name: knowledgeBase.name,
    description: knowledgeBase.description || '',
    prompt_template: knowledgeBase.prompt_template || '',
    retrieval_top_k: knowledgeBase.retrieval_top_k,
    similarity_threshold: knowledgeBase.similarity_threshold ?? 0.3
  });
  editVisible.value = true;
}

async function submitEdit() {
  if (!formRef.value) {
    return;
  }

  fieldErrors.value = {};
  await formRef.value.validate();
  updateMutation.mutate();
}

async function confirmDeleteKnowledgeBase() {
  try {
    await ElMessageBox.confirm(
      '删除知识库后，当前文档列表和回收任务会一并进入清理流程，是否继续？',
      '删除知识库',
      {
        type: 'warning',
        confirmButtonText: '删除',
        cancelButtonText: '取消'
      }
    );
  } catch {
    return;
  }

  deleteKnowledgeBaseMutation.mutate();
}

async function confirmDeleteDocument(documentId: string) {
  try {
    await ElMessageBox.confirm('文档删除后不可恢复，是否继续？', '删除文档', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    });
  } catch {
    return;
  }

  deleteDocumentMutation.mutate(documentId);
}

async function handleUploadRequest(options: UploadRequestOptions) {
  try {
    const response = await uploadMutation.mutateAsync(options.file as File);
    options.onSuccess(response);
  } catch (error) {
    const uploadError = Object.assign(new Error(getErrorMessage(error)), {
      name: 'UploadAjaxError',
      status: 0,
      method: options.method,
      url: options.action
    }) as Parameters<UploadRequestOptions['onError']>[0];

    options.onError(uploadError);
  }
}

function openDocumentDetail(documentId: string) {
  selectedDocumentId.value = documentId;
  detailDrawerVisible.value = true;
}

function statusTone(status: string): 'info' | 'success' | 'warning' | 'danger' {
  if (status === 'available') {
    return 'success';
  }
  if (status === 'failed') {
    return 'danger';
  }
  if (status === 'pending' || status === 'processing' || status === 'deleting') {
    return 'warning';
  }
  return 'info';
}
</script>

<template>
  <div class="soft-scrollbar flex h-full flex-1 flex-col overflow-y-auto px-5 py-6 lg:px-8">
    <div class="space-y-6">
      <SectionHeading
        eyebrow="Knowledge base detail"
        :title="knowledgeBaseQuery.data.value?.name || '知识库详情'"
        :description="knowledgeBaseQuery.data.value?.description || '这里可以继续维护知识库配置、上传文档并跟踪状态轮询。'"
      >
        <template #actions>
          <button
            type="button"
            class="rounded-full border border-slate-200 px-4 py-2.5 text-sm font-medium text-slate-700 transition hover:border-slate-300 hover:bg-slate-50"
            @click="openEditDialog"
          >
            编辑
          </button>
          <button
            type="button"
            class="rounded-full border border-slate-200 px-4 py-2.5 text-sm font-medium text-slate-700 transition hover:border-slate-300 hover:bg-slate-50"
            @click="reindexMutation.mutate()"
          >
            重建索引
          </button>
          <button
            type="button"
            class="rounded-full border border-transparent px-4 py-2.5 text-sm font-medium text-rose-500 transition hover:bg-rose-50"
            @click="confirmDeleteKnowledgeBase"
          >
            删除
          </button>
        </template>
      </SectionHeading>

      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <StatPanel
          label="Embedding"
          :value="knowledgeBaseQuery.data.value?.embedding_model || '--'"
          helper="当前与后端实际存储字段保持一致"
        />
        <StatPanel
          label="Documents"
          :value="String(documentsQuery.data.value?.items.length || 0)"
          helper="列表中会持续轮询处理中状态"
        />
        <StatPanel
          label="Last Indexed"
          :value="knowledgeBaseQuery.data.value?.last_indexed_at ? formatDateTime(knowledgeBaseQuery.data.value.last_indexed_at) : '未建立'"
          helper="重建索引会生成新的后台任务"
        />
        <StatPanel
          label="Updated"
          :value="knowledgeBaseQuery.data.value?.updated_at ? formatDateTime(knowledgeBaseQuery.data.value.updated_at) : '--'"
          helper="保留最近一次知识库配置更新时间"
        />
      </div>

      <SurfaceCard>
        <div class="grid gap-6 lg:grid-cols-[1.1fr_0.9fr]">
          <div class="space-y-4">
            <p class="text-xs uppercase tracking-[0.3em] text-slate-400">Upload documents</p>
            <h3 class="font-serif text-2xl text-slate-900">上传资料并等待处理完成</h3>
            <p class="text-sm leading-6 text-slate-500">
              当前后端会接受 txt、markdown、docx 和 pdf。需要注意的是，pdf
              目前会先完成上传，再在处理阶段失败，失败原因会回写到列表状态中。
            </p>
          </div>

          <el-upload
            drag
            action="#"
            :show-file-list="false"
            :http-request="handleUploadRequest"
            accept=".txt,.md,.markdown,.docx,.pdf"
            class="rounded-[28px] border border-dashed border-slate-200 bg-white/70"
          >
            <div class="py-4 text-center">
              <p class="text-base font-medium text-slate-900">拖拽文件到这里，或点击上传</p>
              <p class="mt-2 text-sm leading-6 text-slate-500">
                每次单文件上传，上传完成后会自动刷新文档列表。
              </p>
            </div>
          </el-upload>
        </div>
      </SurfaceCard>

      <SurfaceCard>
        <div class="flex items-center justify-between gap-4">
          <div>
            <p class="text-xs uppercase tracking-[0.3em] text-slate-400">Document list</p>
            <h3 class="mt-2 font-serif text-2xl text-slate-900">文档处理进度</h3>
          </div>
          <AppStatusBadge
            :tone="
              (documentsQuery.data.value?.items || []).some((item) =>
                ['pending', 'processing', 'deleting'].includes(item.status)
              )
                ? 'warning'
                : 'success'
            "
          >
            {{
              (documentsQuery.data.value?.items || []).some((item) =>
                ['pending', 'processing', 'deleting'].includes(item.status)
              )
                ? '轮询中'
                : '已稳定'
            }}
          </AppStatusBadge>
        </div>

        <div v-if="documentsQuery.data.value?.items.length" class="mt-5 overflow-x-auto">
          <table class="min-w-full border-separate border-spacing-y-3">
            <thead>
              <tr class="text-left text-xs uppercase tracking-[0.28em] text-slate-400">
                <th class="px-4">文档</th>
                <th class="px-4">状态</th>
                <th class="px-4">大小</th>
                <th class="px-4">分块数</th>
                <th class="px-4">更新时间</th>
                <th class="px-4">操作</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="document in documentsQuery.data.value.items"
                :key="document.id"
                class="rounded-[20px] bg-white"
              >
                <td class="rounded-l-[20px] px-4 py-4">
                  <div class="space-y-1">
                    <p class="font-medium text-slate-900">
                      {{ document.title || document.original_filename || '未命名文档' }}
                    </p>
                    <p class="text-sm text-slate-500">{{ document.original_filename || '系统生成文件名' }}</p>
                    <p v-if="document.error_message" class="text-sm text-rose-500">
                      {{ document.error_message }}
                    </p>
                  </div>
                </td>
                <td class="px-4 py-4">
                  <AppStatusBadge :tone="statusTone(document.status)">
                    {{ document.status }}
                  </AppStatusBadge>
                </td>
                <td class="px-4 py-4 text-sm text-slate-500">
                  {{ formatBytes(document.content_length) }}
                </td>
                <td class="px-4 py-4 text-sm text-slate-500">
                  {{ document.chunk_count }}
                </td>
                <td class="px-4 py-4 text-sm text-slate-500">
                  {{ formatDateTime(document.updated_at) }}
                </td>
                <td class="rounded-r-[20px] px-4 py-4">
                  <div class="flex gap-3 text-sm">
                    <button
                      type="button"
                      class="font-medium text-slate-700 transition hover:text-slate-900"
                      @click="openDocumentDetail(document.id)"
                    >
                      详情
                    </button>
                    <button
                      type="button"
                      class="font-medium text-rose-500 transition hover:text-rose-600"
                      @click="confirmDeleteDocument(document.id)"
                    >
                      删除
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div
          v-else
          class="mt-5 rounded-[24px] border border-dashed border-slate-200 bg-white/70 px-5 py-8 text-sm leading-6 text-slate-500"
        >
          当前还没有文档。上传首个文件后，系统会在这里展示 pending、processing、available 或 failed
          等状态变化。
        </div>
      </SurfaceCard>
    </div>

    <el-dialog v-model="editVisible" title="编辑知识库" width="640px" destroy-on-close>
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <el-form-item label="名称" prop="name" :error="fieldErrors.name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" />
        </el-form-item>
        <div class="grid gap-4 md:grid-cols-2">
          <el-form-item label="Retrieval Top K" prop="retrieval_top_k" :error="fieldErrors.retrieval_top_k">
            <el-input-number v-model="form.retrieval_top_k" :min="1" :max="20" class="!w-full" />
          </el-form-item>
          <el-form-item label="Similarity Threshold">
            <el-input-number
              v-model="form.similarity_threshold"
              :min="0"
              :max="1"
              :step="0.05"
              :precision="2"
              class="!w-full"
            />
          </el-form-item>
        </div>
        <el-form-item label="Prompt Template">
          <el-input v-model="form.prompt_template" type="textarea" :rows="4" />
        </el-form-item>
      </el-form>

      <template #footer>
        <div class="flex justify-end gap-3">
          <el-button @click="editVisible = false">取消</el-button>
          <el-button type="primary" :loading="updateMutation.isPending.value" @click="submitEdit">
            保存更新
          </el-button>
        </div>
      </template>
    </el-dialog>

    <el-drawer
      v-model="detailDrawerVisible"
      title="文档详情"
      size="420px"
      direction="rtl"
      destroy-on-close
    >
      <div v-if="documentDetailQuery.data.value" class="space-y-5">
        <div class="space-y-2">
          <p class="text-xs uppercase tracking-[0.3em] text-slate-400">Document</p>
          <h3 class="font-serif text-2xl text-slate-900">
            {{ documentDetailQuery.data.value.title || documentDetailQuery.data.value.original_filename }}
          </h3>
        </div>

        <div class="grid gap-3">
          <div class="rounded-[20px] bg-slate-50 px-4 py-4">
            <p class="text-xs uppercase tracking-[0.26em] text-slate-400">状态</p>
            <div class="mt-2">
              <AppStatusBadge :tone="statusTone(documentDetailQuery.data.value.status)">
                {{ documentDetailQuery.data.value.status }}
              </AppStatusBadge>
            </div>
          </div>
          <div class="rounded-[20px] bg-slate-50 px-4 py-4">
            <p class="text-xs uppercase tracking-[0.26em] text-slate-400">内容大小</p>
            <p class="mt-2 text-sm text-slate-700">
              {{ formatBytes(documentDetailQuery.data.value.content_length) }}
            </p>
          </div>
          <div class="rounded-[20px] bg-slate-50 px-4 py-4">
            <p class="text-xs uppercase tracking-[0.26em] text-slate-400">更新时间</p>
            <p class="mt-2 text-sm text-slate-700">
              {{ formatDateTime(documentDetailQuery.data.value.updated_at) }}
            </p>
          </div>
          <div v-if="documentDetailQuery.data.value.error_message" class="rounded-[20px] bg-rose-50 px-4 py-4">
            <p class="text-xs uppercase tracking-[0.26em] text-rose-400">错误信息</p>
            <p class="mt-2 text-sm leading-6 text-rose-500">
              {{ documentDetailQuery.data.value.error_message }}
            </p>
          </div>
        </div>
      </div>
    </el-drawer>
  </div>
</template>
