import type { Session } from '@private-kb/shared';
import { ASSISTANT_NAME } from '@private-kb/shared/constants/app';

import type { UiChatMessage } from './types';

export function buildPreviewConversation(session?: Session | null): UiChatMessage[] {
  const knowledgeBaseName = session?.knowledge_base_name || '未绑定知识库';
  const isKnowledgeBound = knowledgeBaseName !== '未绑定知识库';

  return [
    {
      id: 'welcome',
      role: 'assistant',
      content: isKnowledgeBound
        ? `欢迎回来，我是 ${ASSISTANT_NAME}。当前会话已绑定“${knowledgeBaseName}”。`
        : `欢迎回来，我是 ${ASSISTANT_NAME}。你可以直接开始新的对话。`,
      grounded: isKnowledgeBound,
      isPreview: true,
      tag: isKnowledgeBound ? '知识库会话' : '普通会话',
      tagTone: isKnowledgeBound ? 'success' : 'info'
    },
    {
      id: 'user-preview',
      role: 'user',
      content: isKnowledgeBound ? '这个知识库里有哪些和系统设计相关的资料？' : '帮我整理一下今天的工作重点。',
      isPreview: true
    },
    {
      id: 'assistant-preview',
      role: 'assistant',
      content: isKnowledgeBound
        ? '你可以围绕当前知识库继续提问，我会结合资料来组织回答。'
        : '我可以帮助你整理问题、生成回复，并继续跟进后续对话。',
      grounded: isKnowledgeBound,
      isPreview: true,
      tag: isKnowledgeBound ? '知识库回答' : undefined,
      tagTone: isKnowledgeBound ? 'success' : undefined
    }
  ];
}
