import { defineStore } from 'pinia';

export const useUserAppStore = defineStore('user-app-ui', {
  state: () => ({
    sidebarCollapsed: false,
    mobileSidebarOpen: false
  }),
  actions: {
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed;
    },
    setMobileSidebarOpen(value: boolean) {
      this.mobileSidebarOpen = value;
    }
  }
});
