import './assets/main.css'

import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import i18n from './i18n'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import 'element-plus/dist/index.css'

//import api from './api/index.js';
const app = createApp(App);
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}
//app.use(ElementPlus);
app.use(router);
app.use(i18n);
app.mount("#app");
//createApp(App).use(router).mount('#app')
