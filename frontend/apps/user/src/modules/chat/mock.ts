import type { Session } from '@private-kb/shared';
import { ASSISTANT_NAME } from '@private-kb/shared/constants/app';

import type { UiChatMessage } from './types';

export function buildPreviewConversation(session?: Session | null): UiChatMessage[] {
  const knowledgeBaseName = session?.knowledge_base_name || '当前未绑定知识库';
  const isKnowledgeBound = knowledgeBaseName !== '当前未绑定知识库';

  return [
    {
      id: 'welcome',
      role: 'assistant',
      content: isKnowledgeBound
        ? `欢迎回来，我是 ${ASSISTANT_NAME}。当前会话已绑定“${knowledgeBaseName}”，你可以直接提问，我会先检索知识库，再在回答下方展示真实引用来源。`
        : `欢迎回来，我是 ${ASSISTANT_NAME}。当前会话还没有绑定知识库，因此你现在就可以直接发起实时问答，消息也会保留在历史记录里。`,
      grounded: isKnowledgeBound,
      isPreview: true,
      tag: isKnowledgeBound ? '知识库检索已接通' : '通用问答已接通',
      tagTone: isKnowledgeBound ? 'success' : 'info'
    },
    {
      id: 'user-preview',
      role: 'user',
      content: isKnowledgeBound ? '这个知识库里有哪些和系统设计相关的资料？' : '这个版本目前已经可以做哪些前置操作？',
      isPreview: true
    },
    {
      id: 'assistant-preview',
      role: 'assistant',
      content: isKnowledgeBound
        ? '你现在可以直接发起知识库问答，命中资料时我会标记“命中知识库”并展示引用文档名；如果没有命中，也会自动回退到通用回答。'
        : '你已经可以创建会话、管理知识库、上传并轮询文档状态，还可以维护个人资料与密码，并在普通会话里直接体验实时问答。',
      grounded: isKnowledgeBound,
      isPreview: true,
      tag: isKnowledgeBound ? '命中知识库时将展示引用' : undefined,
      tagTone: isKnowledgeBound ? 'success' : undefined
    }
  ];
}
