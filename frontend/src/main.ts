import { createApp } from "vue";
import { createRouter, createWebHistory } from "vue-router";
import "./style.generated.css";
import App from "./App.vue";
import DocumentList from "./pages/DocumentList.vue";
import DocumentDetail from "./pages/DocumentDetail.vue";
import UploadPage from "./pages/UploadPage.vue";
import ActivityPage from "./pages/ActivityPage.vue";
import AnalyticsPage from "./pages/AnalyticsPage.vue";
import SearchPage from "./pages/SearchPage.vue";
import SettingsPage from "./pages/SettingsPage.vue";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: "/", component: DocumentList },
    { path: "/search", component: SearchPage },
    { path: "/upload", component: UploadPage },
    { path: "/activity", component: ActivityPage },
    { path: "/analytics", component: AnalyticsPage },
    { path: "/settings", component: SettingsPage },
    { path: "/documents/:id", component: DocumentDetail },
  ],
});

createApp(App).use(router).mount("#app");
