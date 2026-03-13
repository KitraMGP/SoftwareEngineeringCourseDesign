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
        ? `欢迎回来，我是 ${ASSISTANT_NAME}。当前会话会以“${knowledgeBaseName}”作为资料上下文展示区。等知识库检索链路接通后，这里会展示真实引用和命中内容。`
        : `欢迎回来，我是 ${ASSISTANT_NAME}。当前会话还没有绑定知识库，因此你现在就可以直接发起实时问答，消息也会保留在历史记录里。`,
      citations: isKnowledgeBound
        ? [
            {
              id: 'citation-1',
              title: `${knowledgeBaseName} / onboarding.md`,
              snippet: '这里会在知识库问答链路完成后展示真实引用来源摘要。'
            }
          ]
        : [],
      grounded: isKnowledgeBound
    },
    {
      id: 'user-preview',
      role: 'user',
      content: '这个版本目前已经可以做哪些前置操作？'
    },
    {
      id: 'assistant-preview',
      role: 'assistant',
      content: isKnowledgeBound
        ? '你已经可以创建会话、管理知识库、上传并轮询文档状态，还可以维护个人资料与密码。当前绑定知识库的检索式问答仍在继续联调。'
        : '你已经可以创建会话、管理知识库、上传并轮询文档状态，还可以维护个人资料与密码，并在空会话里直接体验实时问答。',
      grounded: false,
      tag: isKnowledgeBound ? '知识库问答待开放' : undefined
    }
  ];
}
