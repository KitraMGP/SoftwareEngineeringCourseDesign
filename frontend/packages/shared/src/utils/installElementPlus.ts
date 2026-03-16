import type { App } from 'vue';

import {
  ElButton,
  ElCheckbox,
  ElDialog,
  ElDrawer,
  ElForm,
  ElFormItem,
  ElIcon,
  ElInput,
  ElInputNumber,
  ElOption,
  ElPagination,
  ElSelect,
  ElTable,
  ElTableColumn,
  ElUpload,
  provideGlobalConfig
} from 'element-plus';
import zhCn from 'element-plus/es/locale/lang/zh-cn';

const elementComponents = [
  ElButton,
  ElCheckbox,
  ElDialog,
  ElDrawer,
  ElForm,
  ElFormItem,
  ElIcon,
  ElInput,
  ElInputNumber,
  ElOption,
  ElPagination,
  ElSelect,
  ElTable,
  ElTableColumn,
  ElUpload
] as const;

export function installElementPlus(app: App) {
  provideGlobalConfig({ locale: zhCn }, app, true);

  for (const component of elementComponents) {
    if (component.name) {
      app.component(component.name, component);
    }
  }
}
