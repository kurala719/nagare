<template>
  <div class="global-search">
    <el-popover
      :visible="searchVisible"
      placement="bottom-start"
      :width="500"
      trigger="manual"
      :teleported="true"
      popper-class="global-search-popper"
    >
      <template #reference>
        <el-input
          v-model="searchQuery"
          :placeholder="$t('common.searchPlaceholder')"
          clearable
          class="global-search-input"
          @focus="handleInputFocus"
          @input="handleSearch"
          @blur="handleInputBlur"
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
        
        <div v-else-if="searchQuery && searchQuery.trim().length >= 2 && results.length === 0" class="no-results">
          {{ $t('common.noMore') || 'No results found' }}
        </div>
        
        <div v-else-if="results.length > 0" class="results-list">
          <div v-for="(group, categoryKey) in groupedResults" :key="categoryKey" class="result-group">
            <div class="result-category">{{ $t(`menu.${categoryKey}`) }}</div>
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
import { 
  Search, Loading, Monitor, Connection, DataAnalysis, 
  Bell, User, Setting, Box, Opportunity 
} from '@element-plus/icons-vue';
import { fetchMonitorData } from '@/api/monitors';
import { fetchHostData } from '@/api/hosts';
import { fetchItemData } from '@/api/items';
import { fetchAlertData } from '@/api/alerts';
import { fetchGroupData } from '@/api/groups';
import { fetchAlarmData } from '@/api/alarms';

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
    Box,
    Opportunity
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
        if (!grouped[item.categoryKey]) {
          grouped[item.categoryKey] = [];
        }
        grouped[item.categoryKey].push(item);
      });
      return grouped;
    });

    const handleInputFocus = () => {
      if (searchQuery.value && searchQuery.value.trim().length >= 2) {
        searchVisible.value = true;
      }
    };

    const handleInputBlur = () => {
      // Small delay to allow click events on results to trigger first
      setTimeout(() => {
        searchVisible.value = false;
      }, 200);
    };

    const closeSearch = () => {
      searchVisible.value = false;
    };

    const handleSearch = () => {
      if (!searchQuery.value || searchQuery.value.trim().length < 2) {
        results.value = [];
        searchVisible.value = false;
        return;
      }

      searchVisible.value = true;
      clearTimeout(searchTimeout);
      searchTimeout = setTimeout(async () => {
        await performSearch();
      }, 300);
    };

    const extractArray = (res) => {
      if (!res) return [];
      if (Array.isArray(res)) return res;
      if (res.success && res.data) {
        if (Array.isArray(res.data)) return res.data;
        if (Array.isArray(res.data.items)) return res.data.items;
      }
      return [];
    };

    const performSearch = async () => {
      searching.value = true;
      const query = searchQuery.value.trim().toLowerCase();
      const foundResults = [];

      try {
        const [monitorsRes, alarmsRes, groupsRes, hostsRes, itemsRes, alertsRes] = await Promise.all([
          fetchMonitorData({ q: query, limit: 5 }).catch(() => []),
          fetchAlarmData({ q: query, limit: 5 }).catch(() => []),
          fetchGroupData({ q: query, limit: 5 }).catch(() => []),
          fetchHostData({ q: query, limit: 5 }).catch(() => []),
          fetchItemData({ q: query, limit: 5 }).catch(() => []),
          fetchAlertData({ q: query, limit: 5 }).catch(() => [])
        ]);

        // 1. Monitors (Inventory Sources)
        extractArray(monitorsRes).forEach(m => {
          foundResults.push({
            id: `m-${m.id || m.ID}`,
            categoryKey: 'monitor',
            title: m.name || m.Name,
            subtitle: m.url || m.URL,
            icon: 'Monitor',
            color: '#409EFF',
            type: 'monitor'
          });
        });

        // 2. Alarms (Alert Sources)
        extractArray(alarmsRes).forEach(a => {
          foundResults.push({
            id: `a-${a.id || a.ID}`,
            categoryKey: 'alarm',
            title: a.name || a.Name,
            subtitle: a.url || a.URL,
            icon: 'Opportunity',
            color: '#E6A23C',
            type: 'alarm'
          });
        });

        // 3. Groups (Host Groups)
        extractArray(groupsRes).forEach(g => {
          foundResults.push({
            id: `g-${g.id || g.ID}`,
            categoryKey: 'group',
            title: g.name || g.Name,
            subtitle: g.description || g.Description || '-',
            icon: 'Box',
            color: '#909399',
            type: 'group'
          });
        });

        // 4. Hosts (End-point Assets)
        extractArray(hostsRes).forEach(h => {
          foundResults.push({
            id: `h-${h.id || h.ID}`,
            categoryKey: 'host',
            title: h.name || h.Name,
            subtitle: h.ip_addr || h.IPAddr || h.ip || '-',
            icon: 'Connection',
            color: '#67C23A',
            type: 'host'
          });
        });

        // 5. Items (Metrics)
        extractArray(itemsRes).forEach(i => {
          foundResults.push({
            id: `i-${i.id || i.ID}`,
            categoryKey: 'item',
            title: i.name || i.Name,
            subtitle: `Value: ${i.last_value || i.LastValue || 'N/A'} ${i.units || i.Units || ''}`,
            icon: 'DataAnalysis',
            color: '#409EFF',
            type: 'item'
          });
        });

        // 6. Alerts
        extractArray(alertsRes).forEach(a => {
          foundResults.push({
            id: `al-${a.id || a.ID}`,
            categoryKey: 'alert',
            title: a.message || a.Message,
            subtitle: `Severity: ${a.severity ?? a.Severity ?? 'N/A'}`,
            icon: 'Bell',
            color: '#F56C6C',
            type: 'alert'
          });
        });

        results.value = foundResults;
        
      } catch (err) {
        console.error('Global Search failed:', err);
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
      handleInputFocus,
      handleInputBlur,
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
