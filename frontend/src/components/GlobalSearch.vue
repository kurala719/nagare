<template>
  <div class="global-search">
    <el-popover
      v-model:visible="searchVisible"
      placement="bottom-start"
      :width="500"
      trigger="click"
      :teleported="true"
      popper-class="global-search-popper"
    >
      <template #reference>
        <el-input
          v-model="searchQuery"
          :placeholder="$t('common.searchPlaceholder')"
          clearable
          class="global-search-input"
          @focus="openSearch"
          @click="openSearch"
          @input="handleSearch"
          @keydown.esc="closeSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </template>
      
      <div class="global-search-results">
        <div v-if="searching" class="search-loading">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span>{{ $t('common.search') }}...</span>
        </div>
        
        <div v-else-if="searchQuery && results.length === 0" class="no-results">
          No results found
        </div>
        
        <div v-else-if="results.length > 0" class="results-list">
          <div v-for="(group, category) in groupedResults" :key="category" class="result-group">
            <div class="result-category">{{ category }}</div>
            <div
              v-for="item in group"
              :key="item.id"
              class="result-item"
              @mousedown.prevent="navigateToItem(item)"
            >
              <el-icon class="result-icon" :color="item.color">
                <component :is="item.icon" />
              </el-icon>
              <div class="result-content">
                <div class="result-title">{{ item.title }}</div>
                <div class="result-subtitle">{{ item.subtitle }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </el-popover>
  </div>
</template>

<script>
import { ref, computed, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { Search, Loading, Monitor, Connection, DataAnalysis, Bell, User, Setting } from '@element-plus/icons-vue';
import { fetchMonitorData } from '@/api/monitors';
import { fetchHostData } from '@/api/hosts';
import { fetchItemData } from '@/api/items';
import { fetchAlertData } from '@/api/alerts';

export default {
  name: 'GlobalSearch',
  components: {
    Search,
    Loading,
    Monitor,
    Connection,
    DataAnalysis,
    Bell,
    User,
    Setting,
  },
  setup() {
    const router = useRouter();
    const { t } = useI18n();
    const searchQuery = ref('');
    const searchVisible = ref(false);
    const searching = ref(false);
    const results = ref([]);
    let searchTimeout = null;

    const groupedResults = computed(() => {
      const grouped = {};
      results.value.forEach(item => {
        if (!grouped[item.category]) {
          grouped[item.category] = [];
        }
        grouped[item.category].push(item);
      });
      return grouped;
    });

    const openSearch = () => {
      searchVisible.value = true;
    };

    const closeSearch = () => {
      searchVisible.value = false;
    };

    const handleSearch = () => {
      if (!searchQuery.value || searchQuery.value.length < 2) {
        results.value = [];
        return;
      }

      clearTimeout(searchTimeout);
      searchTimeout = setTimeout(async () => {
        await performSearch();
      }, 300);
    };

    const performSearch = async () => {
      searching.value = true;
      const query = searchQuery.value.toLowerCase();
      const foundResults = [];

      try {
        // Search monitors
        const monitorsData = await fetchMonitorData({ q: query, limit: 5, offset: 0 }).catch(() => ({ data: [] }));
        const monitors = Array.isArray(monitorsData) ? monitorsData : (monitorsData.data || monitorsData.monitors || []);
        monitors
          .filter(m => 
            m.name?.toLowerCase().includes(query) ||
            m.url?.toLowerCase().includes(query) ||
            m.type?.toLowerCase().includes(query)
          )
          .slice(0, 5)
          .forEach(m => {
            foundResults.push({
              id: `monitor-${m.id}`,
              category: 'Monitors',
              title: m.name || m.Name,
              subtitle: m.url || m.URL,
              icon: 'Monitor',
              color: '#409EFF',
              type: 'monitor',
              data: m,
            });
          });

        // Search hosts
        const hostsData = await fetchHostData({ q: query, limit: 5, offset: 0 }).catch(() => ({ data: [] }));
        const hosts = Array.isArray(hostsData) ? hostsData : (hostsData.data || hostsData.hosts || []);
        hosts
          .filter(h => 
            (h.name || h.Name)?.toLowerCase().includes(query) ||
            (h.ip_addr || h.IPAddr || h.ip)?.toLowerCase().includes(query) ||
            (h.description || h.Description || h.comment || h.Comment)?.toLowerCase().includes(query)
          )
          .slice(0, 5)
          .forEach(h => {
            foundResults.push({
              id: `host-${h.id}`,
              category: 'Hosts',
              title: h.name || h.Name,
              subtitle: h.ip_addr || h.IPAddr || h.ip || '-',
              icon: 'Connection',
              color: '#67C23A',
              type: 'host',
              data: h,
            });
          });

        // Search items
        const itemsData = await fetchItemData({ q: query, limit: 5, offset: 0 }).catch(() => ({ data: [] }));
        const items = Array.isArray(itemsData) ? itemsData : (itemsData.data || itemsData.items || []);
        items
          .filter(i => 
            (i.name || i.Name)?.toLowerCase().includes(query) ||
            (i.value || i.Value || i.last_value || i.LastValue)?.toLowerCase().includes(query) ||
            (i.comment || i.Comment || '')?.toLowerCase().includes(query)
          )
          .slice(0, 5)
          .forEach(i => {
            foundResults.push({
              id: `item-${i.id}`,
              category: 'Items',
              title: i.name || i.Name,
              subtitle: `Value: ${i.value || i.Value || i.last_value || i.LastValue || 'N/A'}`,
              icon: 'DataAnalysis',
              color: '#E6A23C',
              type: 'item',
              data: i,
            });
          });

        // Search alerts
        const alertsData = await fetchAlertData({ q: query, limit: 5, offset: 0 }).catch(() => ({ data: [] }));
        const alerts = Array.isArray(alertsData) ? alertsData : (alertsData.data || alertsData.alerts || []);
        alerts
          .filter(a => 
            (a.message || a.Message)?.toLowerCase().includes(query)
          )
          .slice(0, 5)
          .forEach(a => {
            foundResults.push({
              id: `alert-${a.id}`,
              category: 'Alerts',
              title: a.message || a.Message,
              subtitle: `Severity: ${a.severity ?? a.Severity ?? 'N/A'}`,
              icon: 'Bell',
              color: '#F56C6C',
              type: 'alert',
              data: a,
            });
          });

        results.value = foundResults;
      } catch (err) {
        console.error('Search error:', err);
        results.value = [];
      } finally {
        searching.value = false;
      }
    };

    const navigateToItem = (item) => {
      searchVisible.value = false;
      searchQuery.value = '';
      results.value = [];

      const queryText = String(item.title || item.subtitle || '').trim();
      const query = queryText ? { q: queryText } : {};
      
      switch (item.type) {
        case 'monitor':
          router.push({ path: '/monitor', query });
          break;
        case 'host':
          router.push({ path: '/host', query });
          break;
        case 'item':
          router.push({ path: '/item', query });
          break;
        case 'alert':
          router.push({ path: '/alert', query });
          break;
        default:
          break;
      }
    };

    watch(searchQuery, (newVal) => {
      if (!newVal) {
        results.value = [];
      }
    });

    return {
      searchQuery,
      searchVisible,
      searching,
      results,
      groupedResults,
      openSearch,
      closeSearch,
      handleSearch,
      navigateToItem,
    };
  },
};
</script>

<style scoped>
.global-search {
  min-width: 200px;
  max-width: 400px;
}

.global-search-input {
  width: 100%;
}

.global-search-results {
  max-height: 500px;
  overflow-y: auto;
}

:deep(.global-search-popper) {
  z-index: 3000;
}

.search-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 20px;
  color: #909399;
}

.no-results {
  padding: 20px;
  text-align: center;
  color: #909399;
}

.results-list {
  padding: 8px 0;
}

.result-group {
  margin-bottom: 12px;
}

.result-group:last-child {
  margin-bottom: 0;
}

.result-category {
  padding: 8px 16px 4px;
  font-size: 12px;
  font-weight: 600;
  color: #909399;
  text-transform: uppercase;
}

.result-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 16px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.result-item:hover {
  background-color: #f5f7fa;
}

.result-icon {
  font-size: 20px;
  flex-shrink: 0;
}

.result-content {
  flex: 1;
  min-width: 0;
}

.result-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.result-subtitle {
  font-size: 12px;
  color: #909399;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-top: 2px;
}

:deep(.el-input__wrapper) {
  transition: all 0.3s;
}

:deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px var(--el-input-hover-border-color) inset;
}
</style>
