<script setup lang="ts">
import type { FormInstance, FormRules } from 'element-plus';

import { useMutation, useQuery, useQueryClient } from '@tanstack/vue-query';
import { ElMessage, ElMessageBox } from 'element-plus';
import { computed, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';

import {
  DEFAULT_EMBEDDING_MODEL,
  EmptyStatePanel,
  SectionHeading,
  SurfaceCard,
  knowledgeBaseApi,
  queryKeys
} from '@private-kb/shared';
import { formatDateTime } from '@private-kb/shared/utils/date';
import { shortenText } from '@private-kb/shared/utils/format';
import { getErrorMessage, toFieldErrorMap } from '@private-kb/shared/utils/errors';

const router = useRouter();
const queryClient = useQueryClient();

const keyword = ref('');
const searchKeyword = ref('');

const dialogVisible = ref(false);
const dialogMode = ref<'create' | 'edit'>('create');
const editingKnowledgeBaseId = ref('');
const formRef = ref<FormInstance>();
const fieldErrors = ref<Record<string, string>>({});

const form = reactive({
  name: '',
  description: '',
  embedding_model: DEFAULT_EMBEDDING_MODEL,
  prompt_template: '',
  retrieval_top_k: 5,
  similarity_threshold: 0.3
});

const rules: FormRules<typeof form> = {
  name: [{ required: true, message: '请输入知识库名称', trigger: 'blur' }],
  embedding_model: [{ required: true, message: '请输入 embedding model 标识', trigger: 'blur' }]
};

const knowledgeBasesQuery = useQuery({
  queryKey: computed(() => queryKeys.knowledgeBases({ keyword: searchKeyword.value })),
  queryFn: () =>
    knowledgeBaseApi.listKnowledgeBases({
      page: 1,
      size: 50,
      keyword: searchKeyword.value || undefined
    })
});

const createMutation = useMutation({
  mutationFn: () =>
    knowledgeBaseApi.createKnowledgeBase({
      ...form,
      description: form.description || null,
      prompt_template: form.prompt_template || null
    }),
  onSuccess: async ({ knowledge_base_id }) => {
    dialogVisible.value = false;
    await queryClient.invalidateQueries({ queryKey: queryKeys.knowledgeBasesRoot });
    ElMessage.success('知识库已创建。');
    router.push({ name: 'knowledge-base-detail', params: { kbId: knowledge_base_id } });
  },
  onError: (error) => {
    fieldErrors.value = toFieldErrorMap(error);
    ElMessage.error(getErrorMessage(error));
  }
});

const updateMutation = useMutation({
  mutationFn: () =>
    knowledgeBaseApi.updateKnowledgeBase(editingKnowledgeBaseId.value, {
      name: form.name,
      description: form.description || null,
      prompt_template: form.prompt_template || null,
      retrieval_top_k: form.retrieval_top_k,
      similarity_threshold: form.similarity_threshold
    }),
  onSuccess: async () => {
    dialogVisible.value = false;
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: queryKeys.knowledgeBasesRoot }),
      queryClient.invalidateQueries({ queryKey: queryKeys.knowledgeBase(editingKnowledgeBaseId.value) })
    ]);
    ElMessage.success('知识库信息已更新。');
  },
  onError: (error) => {
    fieldErrors.value = toFieldErrorMap(error);
    ElMessage.error(getErrorMessage(error));
  }
});

const deleteMutation = useMutation({
  mutationFn: (knowledgeBaseId: string) => knowledgeBaseApi.deleteKnowledgeBase(knowledgeBaseId),
  onSuccess: async () => {
    await queryClient.invalidateQueries({ queryKey: queryKeys.knowledgeBasesRoot });
    ElMessage.success('知识库已删除。');
  },
  onError: (error) => {
    ElMessage.error(getErrorMessage(error));
  }
});

function openCreateDialog() {
  dialogMode.value = 'create';
  editingKnowledgeBaseId.value = '';
  fieldErrors.value = {};
  Object.assign(form, {
    name: '',
    description: '',
    embedding_model: DEFAULT_EMBEDDING_MODEL,
    prompt_template: '',
    retrieval_top_k: 5,
    similarity_threshold: 0.3
  });
  dialogVisible.value = true;
}

function openEditDialog(item: {
  id: string;
  name: string;
  description?: string | null;
  embedding_model: string;
  prompt_template?: string | null;
  retrieval_top_k: number;
  similarity_threshold?: number | null;
}) {
  dialogMode.value = 'edit';
  editingKnowledgeBaseId.value = item.id;
  fieldErrors.value = {};
  Object.assign(form, {
    name: item.name,
    description: item.description || '',
    embedding_model: item.embedding_model,
    prompt_template: item.prompt_template || '',
    retrieval_top_k: item.retrieval_top_k,
    similarity_threshold: item.similarity_threshold ?? 0.3
  });
  dialogVisible.value = true;
}

function submitSearch() {
  searchKeyword.value = keyword.value.trim();
}

async function submitDialog() {
  if (!formRef.value) {
    return;
  }

  fieldErrors.value = {};
  await formRef.value.validate();

  if (dialogMode.value === 'create') {
    createMutation.mutate();
    return;
  }

  updateMutation.mutate();
}

async function handleDelete(knowledgeBaseId: string) {
  try {
    await ElMessageBox.confirm(
      '删除知识库后，其下文档和后续清理任务会一并进入回收流程，是否继续？',
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

  deleteMutation.mutate(knowledgeBaseId);
}
</script>

<template>
  <div class="soft-scrollbar flex h-full flex-1 flex-col overflow-y-auto px-5 py-6 lg:px-8">
    <div class="space-y-6">
      <SectionHeading
        eyebrow="Knowledge bases"
        title="知识库"
        description="这里直接接通真实后端知识库接口。进入详情页后可以上传文档、查看状态轮询并触发重建索引。"
      >
        <template #actions>
          <button
            type="button"
            class="rounded-full bg-slate-900 px-5 py-3 text-sm font-semibold text-white transition hover:bg-slate-800"
            @click="openCreateDialog"
          >
            新建知识库
          </button>
        </template>
      </SectionHeading>

      <SurfaceCard>
        <div class="flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between">
          <div>
            <p class="text-sm font-medium text-slate-900">按名称搜索当前资料库</p>
            <p class="mt-2 text-sm leading-6 text-slate-500">
              列表页聚焦入口和基础配置，文档数量与处理状态请进入详情查看。
            </p>
          </div>
          <div class="flex w-full max-w-xl gap-3">
            <el-input
              v-model="keyword"
              placeholder="输入知识库名称关键字"
              clearable
              @keyup.enter="submitSearch"
            />
            <button
              type="button"
              class="min-w-[96px] whitespace-nowrap rounded-full border border-slate-200 px-5 py-3 text-sm font-medium text-slate-700 transition hover:border-slate-300 hover:bg-slate-50"
              @click="submitSearch"
            >
              搜索
            </button>
          </div>
        </div>
      </SurfaceCard>

      <div v-if="knowledgeBasesQuery.isLoading.value" class="grid gap-5 lg:grid-cols-2">
        <div
          v-for="item in 4"
          :key="item"
          class="h-[220px] animate-pulse rounded-[30px] bg-white/70"
        />
      </div>

      <div v-else-if="knowledgeBasesQuery.data.value?.items.length" class="grid gap-5 lg:grid-cols-2">
        <SurfaceCard
          v-for="item in knowledgeBasesQuery.data.value.items"
          :key="item.id"
          class="flex flex-col justify-between"
        >
          <div>
            <div class="flex items-start justify-between gap-4">
              <div class="space-y-2">
                <h3 class="font-serif text-2xl text-slate-900">{{ item.name }}</h3>
                <p class="text-sm leading-6 text-slate-500">
                  {{ shortenText(item.description || '当前知识库还没有摘要说明。', 96) }}
                </p>
              </div>
              <span class="rounded-full bg-slate-100 px-3 py-1 text-xs font-medium text-slate-500">
                Top K {{ item.retrieval_top_k }}
              </span>
            </div>

            <div class="mt-5 grid gap-3 text-sm text-slate-500 md:grid-cols-2">
              <div class="rounded-[18px] bg-slate-50 px-4 py-3">
                <p class="text-xs uppercase tracking-[0.26em] text-slate-400">Embedding</p>
                <p class="mt-2 font-medium text-slate-700">{{ item.embedding_model }}</p>
              </div>
              <div class="rounded-[18px] bg-slate-50 px-4 py-3">
                <p class="text-xs uppercase tracking-[0.26em] text-slate-400">更新时间</p>
                <p class="mt-2 font-medium text-slate-700">{{ formatDateTime(item.updated_at) }}</p>
              </div>
            </div>
          </div>

          <div class="mt-6 flex flex-wrap items-center gap-3">
            <router-link
              class="rounded-full bg-slate-900 px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-slate-800"
              :to="{ name: 'knowledge-base-detail', params: { kbId: item.id } }"
            >
              进入详情
            </router-link>
            <button
              type="button"
              class="rounded-full border border-slate-200 px-4 py-2.5 text-sm font-medium text-slate-700 transition hover:border-slate-300 hover:bg-slate-50"
              @click="openEditDialog(item)"
            >
              编辑
            </button>
            <button
              type="button"
              class="rounded-full border border-transparent px-4 py-2.5 text-sm font-medium text-rose-500 transition hover:bg-rose-50"
              @click="handleDelete(item.id)"
            >
              删除
            </button>
          </div>
        </SurfaceCard>
      </div>

      <EmptyStatePanel
        v-else
        title="先创建一个知识库"
        description="V1 前端已经可以直接管理知识库、上传文档并轮询状态。创建后即可进入详情页继续整理资料。"
        action-label="立即创建"
        @action="openCreateDialog"
      />
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '新建知识库' : '编辑知识库'"
      width="640px"
      destroy-on-close
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <el-form-item label="名称" prop="name" :error="fieldErrors.name">
          <el-input v-model="form.name" placeholder="例如：售前资料库" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="用一句话描述这个知识库适合承载哪些内容"
          />
        </el-form-item>
        <el-form-item label="Embedding Model" prop="embedding_model" :error="fieldErrors.embedding_model">
          <el-input
            v-model="form.embedding_model"
            :disabled="dialogMode === 'edit'"
            placeholder="填写与后端配置一致的标识"
          />
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
          <el-input
            v-model="form.prompt_template"
            type="textarea"
            :rows="4"
            placeholder="如需为该知识库设置专属回答模板，可在此填入。"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <div class="flex justify-end gap-3">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button
            type="primary"
            :loading="createMutation.isPending.value || updateMutation.isPending.value"
            @click="submitDialog"
          >
            {{ dialogMode === 'create' ? '创建知识库' : '保存更新' }}
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>
