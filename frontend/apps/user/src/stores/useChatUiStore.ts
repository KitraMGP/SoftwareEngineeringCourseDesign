import { defineStore } from 'pinia';

export const useChatUiStore = defineStore('chat-ui', {
  state: () => ({
    knowledgeBaseDrawerOpen: false,
    preferredKnowledgeBaseId: '' as string | null,
    preferredKnowledgeBaseName: '' as string | null
  }),
  actions: {
    openKnowledgeBaseDrawer() {
      this.knowledgeBaseDrawerOpen = true;
    },
    closeKnowledgeBaseDrawer() {
      this.knowledgeBaseDrawerOpen = false;
    },
    selectKnowledgeBase(id: string, name: string) {
      this.preferredKnowledgeBaseId = id;
      this.preferredKnowledgeBaseName = name;
    },
    clearKnowledgeBase() {
      this.preferredKnowledgeBaseId = null;
      this.preferredKnowledgeBaseName = null;
    }
  }
});
