<template>
  <div class="dashboard-layout">
    <el-container>
      <el-header>
        <div class="dashboard-header">
          <el-text size="large" tag="b">{{ $t('dashboard.title') }}</el-text>
          <span v-if="hasDashboardRefresh" class="dashboard-updated">
            {{ $t('dashboard.summaryLastUpdated') }}: {{ formattedDashboardRefresh }}
          </span>
        </div>
      </el-header>
      <el-main>
        <!-- Loading State -->
        <div v-if="loading" class="loading-container">
          <el-icon class="is-loading" size="50" color="#409EFF">
            <Loading />
          </el-icon>
          <p>{{ $t('dashboard.loading') }}</p>
        </div>

        <!-- Dashboard Content -->
        <div v-else class="dashboard-content">
          <!-- Summary Cards Row -->
          <el-row :gutter="20" class="summary-row">
            <el-col :span="6">
              <el-card class="summary-card" shadow="hover">
                <div class="summary-icon alerts-icon">
                  <el-icon size="40"><Bell /></el-icon>
                </div>
                <div class="summary-info">
                  <div class="summary-value">{{ summary.alerts.total }}</div>
                  <div class="summary-label">{{ $t('dashboard.alerts') }}</div>
                  <div class="summary-detail">
                    <el-tag type="danger" size="small">{{ summary.alerts.critical }} {{ $t('dashboard.critical') }}</el-tag>
                  </div>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card class="summary-card" shadow="hover">
                <div class="summary-icon hosts-icon">
                  <el-icon size="40"><Monitor /></el-icon>
                </div>
                <div class="summary-info">
                  <div class="summary-value">{{ summary.hosts.total }}</div>
                  <div class="summary-label">{{ $t('dashboard.hosts') }}</div>
                  <div class="summary-detail">
                    <el-tag type="success" size="small">{{ summary.hosts.online }} {{ $t('dashboard.online') }}</el-tag>
                  </div>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card class="summary-card" shadow="hover">
                <div class="summary-icon monitors-icon">
                  <el-icon size="40"><DataLine /></el-icon>
                </div>
                <div class="summary-info">
                  <div class="summary-value">{{ summary.monitors.total }}</div>
                  <div class="summary-label">{{ $t('dashboard.monitors') }}</div>
                  <div class="summary-detail">
                    <el-tag type="success" size="small">{{ summary.monitors.active }} {{ $t('dashboard.active') }}</el-tag>
                  </div>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card class="summary-card" shadow="hover">
                <div class="summary-icon providers-icon">
                  <el-icon size="40"><Connection /></el-icon>
                </div>
                <div class="summary-info">
                  <div class="summary-value">{{ summary.providers.total }}</div>
                  <div class="summary-label">{{ $t('dashboard.providers') }}</div>
                  <div class="summary-detail">
                    <el-tag type="success" size="small">{{ summary.providers.active }} {{ $t('dashboard.active') }}</el-tag>
                  </div>
                </div>
              </el-card>
            </el-col>
          </el-row>

          <!-- System Health Overview -->
          <div class="system-overview">
            <el-alert
              v-if="healthError"
              :title="healthError"
              type="error"
              show-icon
              class="overview-alert"
              :closable="false"
            />
            <el-row :gutter="20" class="summary-row">
              <el-col :span="6">
                <el-card class="summary-card" shadow="hover">
                  <div class="summary-icon score-icon">
                    <el-icon size="32"><DataLine /></el-icon>
                  </div>
                  <div class="summary-info">
                    <div class="summary-value">{{ healthLoading ? '--' : (healthScore.score ?? 0) }}</div>
                    <div class="summary-label">{{ $t('dashboard.healthScore') }}</div>
                    <el-tag :type="healthStatus.tagType" size="small">{{ healthStatus.label }}</el-tag>
                  </div>
                </el-card>
              </el-col>
              <el-col :span="6">
                <el-card class="summary-card" shadow="hover">
                  <div class="summary-icon monitors-icon">
                    <el-icon size="32"><Monitor /></el-icon>
                  </div>
                  <div class="summary-info">
                    <div class="summary-value">
                      {{ healthLoading ? '--' : `${healthScore.monitor_active ?? 0} / ${healthScore.monitor_total ?? 0}` }}
                    </div>
                    <div class="summary-label">{{ $t('dashboard.activeMonitors') }}</div>
                  </div>
                </el-card>
              </el-col>
              <el-col :span="6">
                <el-card class="summary-card" shadow="hover">
                  <div class="summary-icon hosts-icon">
                    <el-icon size="32"><Monitor /></el-icon>
                  </div>
                  <div class="summary-info">
                    <div class="summary-value">
                      {{ healthLoading ? '--' : `${healthScore.host_active ?? 0} / ${healthScore.host_total ?? 0}` }}
                    </div>
                    <div class="summary-label">{{ $t('dashboard.activeHosts') }}</div>
                  </div>
                </el-card>
              </el-col>
              <el-col :span="6">
                <el-card class="summary-card" shadow="hover">
                  <div class="summary-icon item-icon">
                    <el-icon size="32"><Collection /></el-icon>
                  </div>
                  <div class="summary-info">
                    <div class="summary-value">
                      {{ healthLoading ? '--' : `${healthScore.item_active ?? 0} / ${healthScore.item_total ?? 0}` }}
                    </div>
                    <div class="summary-label">{{ $t('dashboard.activeItems') }}</div>
                  </div>
                </el-card>
              </el-col>
            </el-row>

            <el-card class="detail-card" shadow="hover">
              <template #header>
                <div class="card-header">
                  <span>{{ $t('dashboard.healthTrendTitle') }}</span>
                  <div class="card-actions">
                    <el-switch
                      v-model="healthCompareMode"
                      size="small"
                      :active-text="$t('common.comparePrevious')"
                      style="margin-right: 8px;"
                    />
                    <el-date-picker
                      v-model="healthTrendRange"
                      type="datetimerange"
                      :shortcuts="healthTrendShortcuts"
                      :start-placeholder="$t('common.startTime')"
                      :end-placeholder="$t('common.endTime')"
                      size="small"
                      class="trend-range"
                    />
                    <el-button size="small" @click="loadHealthHistory" :loading="healthTrendLoading">
                      {{ $t('common.refresh') }}
                    </el-button>
                  </div>
                </div>
              </template>

              <el-skeleton v-if="healthTrendLoading" animated :rows="6" />
              <el-alert
                v-else-if="healthTrendError"
                :title="healthTrendError"
                type="error"
                show-icon
                :closable="false"
                class="trend-alert"
              />
              <el-empty v-else-if="healthTrendEmpty" :description="$t('common.noHistoryData')" />
              <div v-else ref="healthTrendChartRef" class="trend-chart"></div>
            </el-card>

            <el-card class="detail-card" shadow="hover">
              <template #header>
                <div class="card-header">
                  <span>{{ $t('dashboard.networkMetrics') }}</span>
                  <div class="card-actions">
                    <el-input
                      v-model="metricsQuery"
                      clearable
                      size="small"
                      class="metrics-search"
                      :placeholder="$t('dashboard.metricsSearch')"
                      @keyup.enter="loadMetrics"
                    />
                    <el-input
                      v-model.number="metricsLimit"
                      type="number"
                      size="small"
                      class="metrics-limit"
                      :min="10"
                      :max="500"
                    />
                    <el-button
                      size="small"
                      @click="refreshSystemOverview"
                      :loading="healthLoading || metricsLoading"
                    >
                      {{ $t('dashboard.refreshAll') }}
                    </el-button>
                    <span v-if="hasSystemRefresh" class="metrics-updated">
                      {{ $t('dashboard.lastUpdated') }}: {{ formattedLastRefresh }}
                    </span>
                    <el-button size="small" @click="loadMetrics" :loading="metricsLoading">
                      {{ $t('dashboard.refresh') }}
                    </el-button>
                  </div>
                </div>
              </template>

              <div v-if="metricsLoading" class="loading-container">
                <el-icon class="is-loading" size="40" color="#409EFF">
                  <Loading />
                </el-icon>
                <p>{{ $t('dashboard.loadingMetrics') }}</p>
              </div>

              <el-alert
                v-else-if="metricsError"
                :title="metricsError"
                type="error"
                show-icon
                :closable="false"
              />

              <template v-else>
                <el-table :data="metrics" style="width: 100%" max-height="320">
                  <el-table-column prop="host_name" :label="$t('dashboard.host')" min-width="140" />
                  <el-table-column prop="item_name" :label="$t('dashboard.metric')" min-width="160" />
                  <el-table-column :label="$t('dashboard.value')" min-width="140">
                    <template #default="{ row }">
                      <span>{{ row.value }} {{ row.units }}</span>
                    </template>
                  </el-table-column>
                  <el-table-column :label="$t('dashboard.status')" width="120">
                    <template #default="{ row }">
                      <el-tag :type="getStatusInfo(row.status).type" size="small">
                        {{ getStatusInfo(row.status).label }}
                      </el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column :label="$t('dashboard.updated')" min-width="160">
                    <template #default="{ row }">
                      {{ formatTimestamp(row.updated_at) }}
                    </template>
                  </el-table-column>
                </el-table>
                <el-empty v-if="metrics.length === 0" :description="$t('dashboard.noMetrics')" />
              </template>
            </el-card>

            <el-card class="detail-card topology-card" shadow="hover">
              <template #header>
                <div class="card-header">
                  <span>{{ $t('dashboard.topologyTitle') }}</span>
                  <div class="card-actions">
                    <el-button size="small" @click="loadTopologyData" :loading="topologyLoading">
                      {{ $t('dashboard.refresh') }}
                    </el-button>
                  </div>
                </div>
              </template>

              <div v-if="topologyLoading" class="loading-container">
                <el-icon class="is-loading" size="40" color="#409EFF">
                  <Loading />
                </el-icon>
                <p>{{ $t('dashboard.loadingTopology') }}</p>
              </div>

              <el-alert
                v-else-if="topologyError"
                :title="topologyError"
                type="error"
                show-icon
                :closable="false"
                class="topology-alert"
              />

              <el-empty v-else-if="topologyEmpty" :description="$t('dashboard.noTopology')" />
              <div v-else ref="topologyChartRef" class="topology-chart"></div>
            </el-card>

            <el-row :gutter="20" class="detail-row">
              <el-col :xs="24" :lg="12">
                <el-card class="detail-card voice-card" shadow="hover">
                  <template #header>
                    <div class="card-header">
                      <span>{{ $t('dashboard.voiceTitle') }}</span>
                      <div class="card-actions">
                        <el-button size="small" type="primary" :disabled="!voiceSupported" @click="toggleVoiceListening">
                          {{ voiceListening ? $t('dashboard.voiceStop') : $t('dashboard.voiceStart') }}
                        </el-button>
                      </div>
                    </div>
                  </template>

                  <div class="voice-body">
                    <el-alert
                      v-if="!voiceSupported"
                      type="warning"
                      :title="$t('dashboard.voiceNotSupported')"
                      :closable="false"
                      show-icon
                    />
                    <div v-else class="voice-status">
                      <div class="voice-status-pill" :class="voiceListening ? 'active' : 'idle'">
                        {{ voiceListening ? $t('dashboard.voiceListening') : $t('dashboard.voiceIdle') }}
                      </div>
                      <p class="voice-hint">{{ $t('dashboard.voiceHint') }}</p>
                      <div v-if="voiceTranscript" class="voice-transcript">
                        <span class="voice-label">{{ $t('dashboard.voiceTranscript') }}:</span>
                        <span>{{ voiceTranscript }}</span>
                      </div>
                      <div v-if="voiceLastAction" class="voice-action">
                        <span class="voice-label">{{ $t('dashboard.voiceAction') }}:</span>
                        <span>{{ voiceLastAction }}</span>
                      </div>
                      <div v-if="voiceError" class="voice-error">{{ voiceError }}</div>
                    </div>
                  </div>
                </el-card>
              </el-col>
              <el-col :xs="24" :lg="12">
                <el-card class="detail-card matrix-card" shadow="hover">
                  <template #header>
                    <div class="card-header">
                      <span>{{ $t('dashboard.matrixTitle') }}</span>
                      <div class="card-actions">
                        <el-button size="small" @click="toggleMatrixStream">
                          {{ matrixStreaming ? $t('dashboard.matrixPause') : $t('dashboard.matrixResume') }}
                        </el-button>
                      </div>
                    </div>
                  </template>
                  <div class="matrix-stream" ref="matrixStreamRef">
                    <div v-for="(line, index) in matrixLogs" :key="`${line.id}-${index}`" class="matrix-line">
                      <span class="matrix-time">{{ line.time }}</span>
                      <span class="matrix-text">{{ line.text }}</span>
                    </div>
                  </div>
                </el-card>
              </el-col>
            </el-row>
          </div>

          <!-- Recent Alerts Section -->
          <el-row :gutter="20" class="detail-row">
            <el-col :span="12">
              <el-card class="detail-card" shadow="hover">
                <template #header>
                  <div class="card-header">
                    <span>{{ $t('dashboard.recentAlerts') }}</span>
                    <el-button type="primary" text @click="$router.push('/alert')">{{ $t('dashboard.viewAll') }}</el-button>
                  </div>
                </template>
                <el-table :data="recentAlerts" style="width: 100%" max-height="250">
                  <el-table-column prop="id" :label="$t('dashboard.id')" width="60" />
                  <el-table-column prop="message" :label="$t('dashboard.message')" show-overflow-tooltip />
                  <el-table-column prop="severity" :label="$t('dashboard.severity')" width="100">
                    <template #default="{ row }">
                      <el-tag :type="getSeverityType(row.severity)" size="small">{{ row.severity }}</el-tag>
                    </template>
                  </el-table-column>
                </el-table>
                <el-empty v-if="recentAlerts.length === 0" :description="$t('dashboard.noAlerts')" />
              </el-card>
            </el-col>
            <el-col :span="12">
              <el-card class="detail-card" shadow="hover">
                <template #header>
                  <div class="card-header">
                    <span>{{ $t('dashboard.hostStatus') }}</span>
                    <el-button type="primary" text @click="$router.push('/host')">{{ $t('dashboard.viewAll') }}</el-button>
                  </div>
                </template>
                <el-table :data="recentHosts" style="width: 100%" max-height="250">
                  <el-table-column prop="id" :label="$t('dashboard.id')" width="60" />
                  <el-table-column prop="name" :label="$t('dashboard.name')" show-overflow-tooltip />
                  <el-table-column prop="ip" :label="$t('dashboard.ip')" width="120" />
                  <el-table-column prop="status" :label="$t('dashboard.status')" width="100">
                    <template #default="{ row }">
                      <el-tooltip :content="getStatusInfo(row.status).reason" placement="top">
                        <el-tag :type="getStatusInfo(row.status).type" size="small">
                          {{ getStatusInfo(row.status).label }}
                        </el-tag>
                      </el-tooltip>
                    </template>
                  </el-table-column>
                </el-table>
                <el-empty v-if="recentHosts.length === 0" :description="$t('dashboard.noHosts')" />
              </el-card>
            </el-col>
          </el-row>

          <!-- Providers Section -->
          <el-row :gutter="20" class="detail-row">
            <el-col :span="24">
              <el-card class="detail-card" shadow="hover">
                <template #header>
                  <div class="card-header">
                    <span>{{ $t('dashboard.aiProviders') }}</span>
                    <el-button type="primary" text @click="$router.push('/provider')">{{ $t('dashboard.viewAll') }}</el-button>
                  </div>
                </template>
                <el-table :data="recentProviders" style="width: 100%" max-height="200">
                  <el-table-column prop="id" :label="$t('dashboard.id')" width="60" />
                  <el-table-column prop="name" :label="$t('dashboard.name')" />
                  <el-table-column prop="model" :label="$t('dashboard.model')" />
                  <el-table-column prop="status" :label="$t('dashboard.status')" width="100">
                    <template #default="{ row }">
                      <el-tooltip :content="getStatusInfo(row.status).reason" placement="top">
                        <el-tag :type="getStatusInfo(row.status).type" size="small">
                          {{ getStatusInfo(row.status).label }}
                        </el-tag>
                      </el-tooltip>
                    </template>
                  </el-table-column>
                </el-table>
                <el-empty v-if="recentProviders.length === 0" :description="$t('dashboard.noProviders')" />
              </el-card>
            </el-col>
          </el-row>
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script>
import { fetchAlertData } from '@/api/alerts';
import { fetchHostData } from '@/api/hosts';
import { fetchSiteData } from '@/api/sites';
import { fetchProviderData } from '@/api/providers';
import { fetchMonitorData } from '@/api/monitors';
import { fetchNetworkStatusHistory } from '@/api/system';
import { authFetch } from '@/utils/authFetch';
import * as echarts from 'echarts';
import { ElMessage } from 'element-plus';
import { Loading, Bell, Monitor, DataLine, Connection, Collection } from '@element-plus/icons-vue';

export default {
  name: 'Databoard',
  components: {
    Loading,
    Bell,
    Monitor,
    DataLine,
    Connection,
    Collection,
  },
  data() {
    return {
      loading: false,
      summary: {
        alerts: { total: 0, critical: 0 },
        hosts: { total: 0, online: 0 },
        monitors: { total: 0, active: 0 },
        providers: { total: 0, active: 0 },
      },
      recentAlerts: [],
      recentHosts: [],
      recentProviders: [],
      healthLoading: false,
      healthError: null,
      healthScore: {
        score: 0,
        monitor_total: 0,
        monitor_active: 0,
        host_total: 0,
        host_active: 0,
        item_total: 0,
        item_active: 0,
      },
      healthTrendRange: [],
      healthTrendLoading: false,
      healthTrendError: null,
      healthTrendEmpty: false,
      healthTrendChart: null,
      healthCompareMode: false,
      healthTrendShortcuts: [
        {
          text: '1h',
          value: () => {
            const end = new Date();
            const start = new Date(end.getTime() - 60 * 60 * 1000);
            return [start, end];
          },
        },
        {
          text: '6h',
          value: () => {
            const end = new Date();
            const start = new Date(end.getTime() - 6 * 60 * 60 * 1000);
            return [start, end];
          },
        },
        {
          text: '24h',
          value: () => {
            const end = new Date();
            const start = new Date(end.getTime() - 24 * 60 * 60 * 1000);
            return [start, end];
          },
        },
        {
          text: '7d',
          value: () => {
            const end = new Date();
            const start = new Date(end.getTime() - 7 * 24 * 60 * 60 * 1000);
            return [start, end];
          },
        },
        {
          text: '30d',
          value: () => {
            const end = new Date();
            const start = new Date(end.getTime() - 30 * 24 * 60 * 60 * 1000);
            return [start, end];
          },
        },
      ],
      metricsLoading: false,
      metricsError: null,
      metrics: [],
      metricsQuery: '',
      metricsLimit: 200,
      lastHealthRefresh: null,
      lastMetricsRefresh: null,
      lastDashboardRefresh: null,
      topologyLoading: false,
      topologyError: null,
      topologyEmpty: false,
      topologyChart: null,
      matrixLogs: [],
      matrixLogSeed: 0,
      matrixStreamTimer: null,
      matrixStreaming: true,
      voiceSupported: false,
      voiceListening: false,
      voiceTranscript: '',
      voiceLastAction: '',
      voiceError: '',
      voiceRecognizer: null,
    };
  },
  created() {
    this.loadDashboardData();
    this.loadHealthScore();
    this.loadMetrics();
  },
  mounted() {
    this.setDefaultTrendRange();
    this.loadHealthHistory();
    this.loadTopologyData();
    this.initVoiceRecognition();
    this.startMatrixStream();
    window.addEventListener('resize', this.onResize);
  },
  beforeUnmount() {
    window.removeEventListener('resize', this.onResize);
    if (this.healthTrendChart) {
      this.healthTrendChart.dispose();
      this.healthTrendChart = null;
    }
    if (this.topologyChart) {
      this.topologyChart.dispose();
      this.topologyChart = null;
    }
    if (this.matrixStreamTimer) {
      clearInterval(this.matrixStreamTimer);
      this.matrixStreamTimer = null;
    }
    if (this.voiceRecognizer && this.voiceListening) {
      try {
        this.voiceRecognizer.stop();
      } catch (err) {
        console.warn('Failed to stop voice recognition:', err);
      }
    }
  },
  computed: {
    healthStatus() {
      const score = Number(this.healthScore?.score ?? 0);
      if (score >= 85) {
        return { label: this.$t('dashboard.healthGood'), tagType: 'success' };
      }
      if (score >= 60) {
        return { label: this.$t('dashboard.healthWarn'), tagType: 'warning' };
      }
      return { label: this.$t('dashboard.healthBad'), tagType: 'danger' };
    },
    formattedLastRefresh() {
      if (!this.lastHealthRefresh && !this.lastMetricsRefresh) return '--';
      const latest = [this.lastHealthRefresh, this.lastMetricsRefresh]
        .filter(Boolean)
        .map((value) => new Date(value))
        .sort((a, b) => b.getTime() - a.getTime())[0];
      if (!latest || Number.isNaN(latest.getTime())) return '--';
      return latest.toLocaleString();
    },
    formattedDashboardRefresh() {
      if (!this.lastDashboardRefresh) return '--';
      const date = new Date(this.lastDashboardRefresh);
      if (Number.isNaN(date.getTime())) return '--';
      return date.toLocaleString();
    },
    hasSystemRefresh() {
      return Boolean(this.lastHealthRefresh || this.lastMetricsRefresh);
    },
    hasDashboardRefresh() {
      return Boolean(this.lastDashboardRefresh);
    }
  },
  watch: {
    healthTrendRange() {
      this.loadHealthHistory();
    },
    healthCompareMode() {
      this.loadHealthHistory();
    },
  },
  methods: {
    resolveTrendWindow() {
      const range = this.healthTrendRange;
      if (Array.isArray(range) && range.length === 2 && range[0] && range[1]) {
        return [new Date(range[0]), new Date(range[1])];
      }
      const end = new Date();
      const start = new Date(end.getTime() - 24 * 60 * 60 * 1000);
      return [start, end];
    },
    buildHealthTrendChart(series, prevSeries = []) {
      const chartRef = this.$refs.healthTrendChartRef;
      if (!chartRef) return;
      if (!this.healthTrendChart) {
        this.healthTrendChart = echarts.init(chartRef);
      }
      const chartSeries = [
        {
          name: this.$t('common.currentPeriod'),
          type: 'line',
          smooth: true,
          showSymbol: false,
          data: series,
          itemStyle: { color: '#409EFF' },
          areaStyle: { opacity: 0.1 },
        }
      ];
      if (prevSeries.length > 0) {
        chartSeries.push({
          name: this.$t('common.previousPeriod'),
          type: 'line',
          smooth: true,
          showSymbol: false,
          data: prevSeries,
          itemStyle: { color: '#909399' },
          lineStyle: { type: 'dashed' },
          areaStyle: { opacity: 0.05 },
        });
      }
      this.healthTrendChart.setOption({
        tooltip: {
          trigger: 'axis',
          formatter: (params) => {
            let tip = '';
            params.forEach((point) => {
              const value = point.data?.[1];
              const time = new Date(point.data?.[0]);
              const timeLabel = Number.isNaN(time.getTime()) ? '-' : time.toLocaleString();
              tip += `<div style="margin-bottom:4px;"><strong>${point.seriesName}</strong><br/>${timeLabel}<br/>Score: ${value ?? '-'}</div>`;
            });
            return tip;
          },
        },
        legend: {
          show: prevSeries.length > 0,
          data: prevSeries.length > 0 ? [this.$t('common.currentPeriod'), this.$t('common.previousPeriod')] : [],
        },
        grid: { left: 40, right: 20, top: prevSeries.length > 0 ? 40 : 20, bottom: 40 },
        xAxis: { type: 'time' },
        yAxis: { type: 'value', min: 0, max: 100 },
        series: chartSeries,
      });
    },
    async loadHealthHistory() {
      this.healthTrendLoading = true;
      this.healthTrendError = null;
      this.healthTrendEmpty = false;
      try {
        const [start, end] = this.resolveTrendWindow();
        const from = Math.floor(start.getTime() / 1000);
        const to = Math.floor(end.getTime() / 1000);
        
        // Fetch current period
        const response = await fetchNetworkStatusHistory({
          from,
          to,
          limit: 500,
        });
        const payload = response?.data || response || [];
        const rows = Array.isArray(payload) ? payload : [];
        
        if (rows.length === 0) {
          this.healthTrendEmpty = true;
          this.buildHealthTrendChart([], []);
          return;
        }
        
        const series = [];
        rows.forEach((row) => {
          const sampledAt = row.sampled_at || row.SampledAt;
          const score = row.score ?? row.Score;
          if (!sampledAt || score === undefined) return;
          const time = new Date(sampledAt).getTime();
          if (Number.isNaN(time)) return;
          series.push([time, Number(score)]);
        });
        
        // Fetch previous period if compare mode is enabled
        let prevSeries = [];
        if (this.healthCompareMode) {
          const duration = end.getTime() - start.getTime();
          const prevStart = new Date(start.getTime() - duration);
          const prevEnd = new Date(start.getTime());
          const prevFrom = Math.floor(prevStart.getTime() / 1000);
          const prevTo = Math.floor(prevEnd.getTime() / 1000);
          
          try {
            const prevResponse = await fetchNetworkStatusHistory({
              from: prevFrom,
              to: prevTo,
              limit: 500,
            });
            const prevPayload = prevResponse?.data || prevResponse || [];
            const prevRows = Array.isArray(prevPayload) ? prevPayload : [];
            
            prevRows.forEach((row) => {
              const sampledAt = row.sampled_at || row.SampledAt;
              const score = row.score ?? row.Score;
              if (!sampledAt || score === undefined) return;
              const time = new Date(sampledAt).getTime();
              if (Number.isNaN(time)) return;
              // Shift timestamps to align with current period
              const shiftedTime = time + duration;
              prevSeries.push([shiftedTime, Number(score)]);
            });
          } catch (err) {
            console.warn('Failed to load previous period:', err);
          }
        }
        
        this.healthTrendEmpty = series.length === 0;
        this.buildHealthTrendChart(series, prevSeries);
      } catch (err) {
        this.healthTrendError = err?.message || this.$t('common.historyLoadFailed');
        this.buildHealthTrendChart([], []);
      } finally {
        this.healthTrendLoading = false;
      }
    },
    setDefaultTrendRange() {
      if (Array.isArray(this.healthTrendRange) && this.healthTrendRange.length === 2) return;
      const end = new Date();
      const start = new Date(end.getTime() - 24 * 60 * 60 * 1000);
      this.healthTrendRange = [start, end];
    },
    async loadDashboardData() {
      this.loading = true;
      try {
        await Promise.all([
          this.loadAlerts(),
          this.loadHosts(),
          this.loadProviders(),
        ]);
      } catch (err) {
        console.error('Error loading dashboard data:', err);
      } finally {
        this.loading = false;
        this.lastDashboardRefresh = new Date().toISOString();
      }
    },
    async loadAlerts() {
      try {
        const response = await fetchAlertData();
        const data = Array.isArray(response) ? response : (response.data || response.alerts || []);
        const alerts = data.map((a) => ({
          id: a.ID || a.id,
          message: a.Message || a.message || '',
          severity: a.Severity || a.severity || '',
          status: a.Status || a.status || '',
        }));
        this.summary.alerts.total = alerts.length;
        this.summary.alerts.critical = alerts.filter(a => 
          a.severity?.toLowerCase() === 'critical' || a.severity?.toLowerCase() === 'high'
        ).length;
        this.recentAlerts = alerts.slice(0, 5);
      } catch (err) {
        console.error('Error loading alerts:', err);
      }
    },
    async loadHosts() {
      try {
        const response = await fetchHostData();
        const data = Array.isArray(response) ? response : (response.data || response.hosts || []);
        const hosts = data.map((h) => ({
          id: h.ID || h.id,
          name: h.Name || h.name || '',
          ip: h.IP || h.ip || '',
          status: h.Status ?? h.status ?? 0,
        }));
        this.summary.hosts.total = hosts.length;
        this.summary.hosts.online = hosts.filter(h => h.status === 1).length;
        this.recentHosts = hosts.slice(0, 5);
      } catch (err) {
        console.error('Error loading hosts:', err);
      }
    },
    async loadProviders() {
      try {
        const response = await fetchProviderData();
        const data = Array.isArray(response) ? response : (response.data || response.providers || []);
        const providers = data.map((p) => ({
          id: p.ID || p.id,
          name: p.Name || p.name || '',
          model: p.Model || p.model || '',
          status: p.Status ?? p.status ?? 0,
        }));
        this.summary.providers.total = providers.length;
        this.summary.providers.active = providers.filter(p => p.status === 1).length;
        this.recentProviders = providers.slice(0, 5);
      } catch (err) {
        console.error('Error loading providers:', err);
      }
    },
    async loadHealthScore() {
      this.healthLoading = true;
      this.healthError = null;
      try {
        const response = await authFetch('/api/v1/system/health', {
          method: 'GET',
          headers: { 'Accept': 'application/json' },
        });
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json().catch(() => ({}));
        if (data?.success === false) {
          throw new Error(data.error || this.$t('dashboard.healthLoadFailed'));
        }
        const payload = data?.data || data || {};
        this.healthScore = {
          score: payload.score ?? 0,
          monitor_total: payload.monitor_total ?? 0,
          monitor_active: payload.monitor_active ?? 0,
          host_total: payload.host_total ?? 0,
          host_active: payload.host_active ?? 0,
          item_total: payload.item_total ?? 0,
          item_active: payload.item_active ?? 0,
        };
        this.lastHealthRefresh = new Date().toISOString();
      } catch (err) {
        this.healthError = err.message || this.$t('dashboard.healthLoadFailed');
        console.error('Error loading health score:', err);
      } finally {
        this.healthLoading = false;
      }
    },
    async loadMetrics() {
      this.metricsLoading = true;
      this.metricsError = null;
      try {
        const params = new URLSearchParams();
        if (this.metricsQuery.trim()) {
          params.set('q', this.metricsQuery.trim());
        }
        if (this.metricsLimit) {
          params.set('limit', String(this.metricsLimit));
        }
        const url = `/api/v1/system/metrics${params.toString() ? `?${params}` : ''}`;
        const response = await authFetch(url, {
          method: 'GET',
          headers: { 'Accept': 'application/json' },
        });
        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
        const data = await response.json().catch(() => ({}));
        if (data?.success === false) {
          throw new Error(data.error || this.$t('dashboard.metricsLoadFailed'));
        }
        const payload = data?.data || data || [];
        this.metrics = Array.isArray(payload) ? payload : [];
        this.lastMetricsRefresh = new Date().toISOString();
      } catch (err) {
        this.metricsError = err.message || this.$t('dashboard.metricsLoadFailed');
        console.error('Error loading metrics:', err);
      } finally {
        this.metricsLoading = false;
      }
    },
    async refreshSystemOverview() {
      await Promise.all([
        this.loadHealthScore(),
        this.loadMetrics(),
      ]);
    },
    onResize() {
      if (this.healthTrendChart) {
        this.healthTrendChart.resize();
      }
      if (this.topologyChart) {
        this.topologyChart.resize();
      }
    },
    async loadTopologyData() {
      this.topologyLoading = true;
      this.topologyError = null;
      this.topologyEmpty = false;
      try {
        const [siteResponse, hostResponse, monitorResponse] = await Promise.all([
          fetchSiteData({ limit: 200 }),
          fetchHostData({ limit: 500 }),
          fetchMonitorData({ limit: 200 }),
        ]);
        const sitePayload = siteResponse?.data || siteResponse?.sites || siteResponse || [];
        const hostPayload = hostResponse?.data || hostResponse?.hosts || hostResponse || [];
        const monitorPayload = monitorResponse?.data || monitorResponse?.monitors || monitorResponse || [];

        const sites = Array.isArray(sitePayload) ? sitePayload : [];
        const hosts = Array.isArray(hostPayload) ? hostPayload : [];
        const monitors = Array.isArray(monitorPayload) ? monitorPayload : [];
        const { nodes, links } = this.buildTopologyGraph(sites, hosts, monitors);
        this.topologyEmpty = nodes.length === 0;
        this.topologyLoading = false;
        await this.$nextTick();
        this.buildTopologyChart(nodes, links);
      } catch (err) {
        this.topologyError = err?.message || this.$t('dashboard.topologyLoadFailed');
        this.topologyLoading = false;
        await this.$nextTick();
        this.buildTopologyChart([], []);
      } finally {
        this.topologyLoading = false;
      }
    },
    buildTopologyGraph(sites, hosts, monitors) {
      const nodes = [];
      const links = [];
      const siteMap = new Map();
      const monitorMap = new Map();

      monitors.forEach((monitor) => {
        const id = Number(
          monitor.ID || monitor.Id || monitor.id ||
          monitor.MID || monitor.Mid || monitor.mid ||
          monitor.MonitorID || monitor.MonitorId || monitor.monitor_id || monitor.monitorId || monitor.monitorID || 0
        );
        if (!id) return;
        const name = monitor.Name || monitor.name || `${this.$t('dashboard.topologyMonitor')} ${id}`;
        const status = monitor.Status ?? monitor.status ?? 0;
        const nodeId = `monitor-${id}`;
        monitorMap.set(id, nodeId);
        nodes.push({
          id: nodeId,
          name,
          category: 0,
          symbolSize: 64,
          value: { status },
          itemStyle: { color: this.getStatusColor(status) },
          label: { show: true },
        });
      });

      sites.forEach((site) => {
        const id = site.ID || site.id;
        if (!id) return;
        const name = site.Name || site.name || `${this.$t('dashboard.topologySite')} ${id}`;
        const status = site.Status ?? site.status ?? 0;
        const nodeId = `site-${id}`;
        siteMap.set(id, nodeId);
        nodes.push({
          id: nodeId,
          name,
          category: 1,
          symbolSize: 48,
          value: { status },
          itemStyle: { color: this.getStatusColor(status) },
          label: { show: true },
        });
      });

      const siteHostCount = new Map();
      const monitorSiteLinks = new Set();
      hosts.forEach((host) => {
        const hostId = host.ID || host.id;
        if (!hostId) return;
        const siteId = host.SiteID || host.site_id || 0;
        const monitorId = Number(
          host.MID || host.Mid || host.mid ||
          host.MonitorID || host.MonitorId || host.monitor_id || host.monitorId || host.monitorID || host.m_id || 0
        );
        const name = host.Name || host.name || `${this.$t('dashboard.host')} ${hostId}`;
        const status = host.Status ?? host.status ?? 0;
        const nodeId = `host-${hostId}`;
        nodes.push({
          id: nodeId,
          name,
          category: 2,
          symbolSize: 30,
          value: { status, ip: host.IPAddr || host.ip_addr || host.IP || host.ip || '' },
          itemStyle: { color: this.getStatusColor(status) },
          label: { show: true },
        });

        if (siteId && siteMap.has(siteId)) {
          links.push({ source: siteMap.get(siteId), target: nodeId });
          siteHostCount.set(siteId, (siteHostCount.get(siteId) || 0) + 1);
          if (monitorId && monitorMap.has(monitorId)) {
            const linkKey = `${monitorId}-${siteId}`;
            if (!monitorSiteLinks.has(linkKey)) {
              links.push({ source: monitorMap.get(monitorId), target: siteMap.get(siteId) });
              monitorSiteLinks.add(linkKey);
            }
          }
        } else if (monitorId && monitorMap.has(monitorId)) {
          links.push({ source: monitorMap.get(monitorId), target: nodeId });
        }
      });

      nodes.forEach((node) => {
        if (!node.id.startsWith('site-')) return;
        const siteId = Number(node.id.replace('site-', ''));
        const count = siteHostCount.get(siteId) || 0;
        node.symbolSize = Math.min(72, 40 + count * 3);
      });

      return { nodes, links };
    },
    buildTopologyChart(nodes, links) {
      const chartRef = this.$refs.topologyChartRef;
      if (!chartRef) return;
      if (!this.topologyChart) {
        this.topologyChart = echarts.init(chartRef);
      }
      this.topologyChart.setOption({
        tooltip: {
          formatter: (params) => {
            if (params.dataType !== 'node') return '';
            const status = params.data?.value?.status ?? 0;
            const statusLabel = this.getStatusInfo(status).label;
            const ip = params.data?.value?.ip;
            const ipLine = ip ? `<br/>IP: ${ip}` : '';
            return `<strong>${params.data?.name || '-'}</strong><br/>${statusLabel}${ipLine}`;
          }
        },
        legend: [
          {
            data: [
              this.$t('dashboard.topologyMonitor'),
              this.$t('dashboard.topologySite'),
              this.$t('dashboard.topologyHost'),
            ],
            top: 8,
          }
        ],
        series: [
          {
            type: 'graph',
            layout: 'force',
            roam: true,
            data: nodes,
            links,
            categories: [
              { name: this.$t('dashboard.topologyMonitor') },
              { name: this.$t('dashboard.topologySite') },
              { name: this.$t('dashboard.topologyHost') },
            ],
            force: {
              repulsion: 160,
              edgeLength: 120,
              gravity: 0.2,
            },
            label: {
              position: 'right',
              formatter: '{b}',
              fontSize: 12,
            },
            lineStyle: {
              color: 'source',
              width: 1.5,
              opacity: 0.7,
              curveness: 0.2,
            },
          }
        ],
      });
    },
    initVoiceRecognition() {
      const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
      if (!SpeechRecognition) {
        this.voiceSupported = false;
        return;
      }
      this.voiceSupported = true;
      const recognizer = new SpeechRecognition();
      recognizer.lang = 'zh-CN';
      recognizer.interimResults = false;
      recognizer.maxAlternatives = 1;
      recognizer.continuous = false;
      recognizer.onresult = (event) => {
        const transcript = event?.results?.[0]?.[0]?.transcript || '';
        this.voiceTranscript = transcript;
        this.executeVoiceAction(transcript);
      };
      recognizer.onend = () => {
        this.voiceListening = false;
      };
      recognizer.onerror = (event) => {
        this.voiceError = event?.error || this.$t('dashboard.voiceFailed');
        this.voiceListening = false;
      };
      this.voiceRecognizer = recognizer;
    },
    toggleVoiceListening() {
      if (!this.voiceSupported || !this.voiceRecognizer) {
        this.voiceError = this.$t('dashboard.voiceNotSupported');
        return;
      }
      if (this.voiceListening) {
        this.voiceRecognizer.stop();
        this.voiceListening = false;
        return;
      }
      this.voiceError = '';
      this.voiceTranscript = '';
      this.voiceLastAction = '';
      try {
        this.voiceListening = true;
        this.voiceRecognizer.start();
      } catch (err) {
        this.voiceListening = false;
        this.voiceError = this.$t('dashboard.voiceFailed');
      }
    },
    async executeVoiceAction(text) {
      const normalized = String(text || '').toLowerCase();
      if (!normalized) return;
      if (normalized.includes('health') || normalized.includes('健康') || normalized.includes('状态')) {
        this.voiceLastAction = this.$t('dashboard.voiceActionHealth');
        await this.fetchQuickHealth();
        return;
      }
      if (normalized.includes('alert') || normalized.includes('告警')) {
        this.voiceLastAction = this.$t('dashboard.voiceActionAlerts');
        this.$router.push('/alert');
        return;
      }
      if (normalized.includes('switch') || normalized.includes('交换机')) {
        this.voiceLastAction = this.$t('dashboard.voiceActionSwitch');
        this.$router.push({ path: '/host', query: { q: 'switch' } });
        return;
      }
      if (normalized.includes('topology') || normalized.includes('拓扑')) {
        this.voiceLastAction = this.$t('dashboard.voiceActionTopology');
        const node = this.$refs.topologyChartRef;
        if (node && typeof node.scrollIntoView === 'function') {
          node.scrollIntoView({ behavior: 'smooth', block: 'center' });
        }
        return;
      }
      this.voiceLastAction = this.$t('dashboard.voiceNoMatch');
    },
    async fetchQuickHealth() {
      try {
        const response = await authFetch('/api/v1/system/health', {
          method: 'GET',
          headers: { 'Accept': 'application/json' },
        });
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}`);
        }
        const data = await response.json().catch(() => ({}));
        const payload = data?.data || data || {};
        const score = payload.score ?? '--';
        ElMessage.success(`${this.$t('dashboard.voiceHealthResult')}: ${score}`);
      } catch (err) {
        ElMessage.error(this.$t('dashboard.voiceHealthFailed'));
      }
    },
    startMatrixStream() {
      if (this.matrixStreamTimer) return;
      this.matrixStreamTimer = setInterval(() => {
        if (!this.matrixStreaming) return;
        this.appendMatrixLog();
      }, 200);
    },
    toggleMatrixStream() {
      this.matrixStreaming = !this.matrixStreaming;
      if (this.matrixStreaming) {
        this.appendMatrixLog();
      }
    },
    appendMatrixLog() {
      const line = this.buildMatrixLogLine();
      this.matrixLogs.push({
        id: this.matrixLogSeed++,
        time: new Date().toLocaleTimeString(),
        text: line,
      });
      if (this.matrixLogs.length > 200) {
        this.matrixLogs.splice(0, this.matrixLogs.length - 200);
      }
      this.$nextTick(() => {
        const container = this.$refs.matrixStreamRef;
        if (container) {
          container.scrollTop = container.scrollHeight;
        }
      });
    },
    buildMatrixLogLine() {
      const metric = this.metrics.length ? this.metrics[Math.floor(Math.random() * this.metrics.length)] : null;
      if (metric && Math.random() > 0.45) {
        const hostName = metric.host_name || metric.host || this.$t('dashboard.host');
        const label = this.getStatusInfo(metric.status ?? 0).label;
        return `Ping ${hostName}: ${metric.value ?? '--'} ${metric.units ?? ''} | ${label}`;
      }
      const phrases = [
        'Analyzing AI context window',
        'Syncing monitor node',
        'Optimizing token usage',
        'Indexing anomaly signatures',
        'Merging health telemetry',
        'Calibrating alert thresholds',
        'Reconciling host heartbeat',
        'Streaming topology edges',
        'Normalizing metric deltas',
        'Rebuilding signal graph',
      ];
      const targets = [
        this.recentHosts?.[0]?.name,
        this.recentHosts?.[1]?.name,
        this.recentProviders?.[0]?.name,
        this.recentAlerts?.[0]?.message,
      ].filter(Boolean);
      const phrase = phrases[Math.floor(Math.random() * phrases.length)];
      const target = targets.length ? ` :: ${targets[Math.floor(Math.random() * targets.length)]}` : '';
      return `${phrase}${target}`;
    },
    getStatusColor(status) {
      const palette = {
        0: '#909399',
        1: '#67c23a',
        2: '#f56c6c',
        3: '#e6a23c',
      };
      return palette[status] || palette[0];
    },
    formatTimestamp(value) {
      if (!value) return '--';
      const date = new Date(value);
      if (Number.isNaN(date.getTime())) return value;
      return date.toLocaleString();
    },
    getSeverityType(severity) {
      const s = (severity || '').toLowerCase();
      if (s === 'critical' || s === 'high') return 'danger';
      if (s === 'medium' || s === 'warning') return 'warning';
      if (s === 'low' || s === 'info') return 'info';
      return 'info';
    },
    getStatusInfo(status) {
      const map = {
        0: { label: this.$t('common.statusInactive'), reason: this.$t('common.reasonInactive'), type: 'info' },
        1: { label: this.$t('common.statusActive'), reason: this.$t('common.reasonActive'), type: 'success' },
        2: { label: this.$t('common.statusError'), reason: this.$t('common.reasonError'), type: 'danger' },
        3: { label: this.$t('common.statusSyncing'), reason: this.$t('common.reasonSyncing'), type: 'warning' },
      };
      return map[status] || map[0];
    },
  },
};
</script>

<style scoped>
.dashboard-layout {
  padding: 0;
}

.dashboard-layout .el-header {
  background-color: transparent;
  padding: 20px;
}

.dashboard-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.dashboard-updated {
  font-size: 12px;
  color: #909399;
}

.loading-container {
  text-align: center;
  padding: 60px;
  color: #909399;
}

.dashboard-content {
  padding: 0 20px 20px;
}

.summary-row {
  margin-bottom: 20px;
}

.summary-card {
  display: flex;
  align-items: center;
  padding: 10px;
}

.summary-card :deep(.el-card__body) {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 20px;
}

.summary-icon {
  width: 60px;
  height: 60px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16px;
  color: white;
}

.alerts-icon {
  background: linear-gradient(135deg, #f56c6c, #e74c3c);
}

.hosts-icon {
  background: linear-gradient(135deg, #67c23a, #2ecc71);
}

.monitors-icon {
  background: linear-gradient(135deg, #409eff, #3498db);
}

.providers-icon {
  background: linear-gradient(135deg, #e6a23c, #f39c12);
}

.summary-info {
  flex: 1;
}

.summary-value {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
}

.summary-label {
  font-size: 14px;
  color: #909399;
  margin-bottom: 8px;
}

.system-overview {
  margin-bottom: 20px;
}

.overview-alert {
  margin-bottom: 16px;
}

.score-icon {
  background: linear-gradient(135deg, #409eff, #36d1dc);
}

.item-icon {
  background: linear-gradient(135deg, #f56c6c, #e74c3c);
}

.card-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.trend-range {
  width: 260px;
}

.trend-chart {
  width: 100%;
  height: 260px;
}

.trend-alert {
  margin-bottom: 12px;
}

.metrics-search {
  width: 180px;
}

.metrics-limit {
  width: 90px;
}

.metrics-updated {
  font-size: 12px;
  color: #909399;
}

.topology-card {
  margin-top: 20px;
}

.topology-chart {
  width: 100%;
  height: 420px;
}

.topology-alert {
  margin-bottom: 12px;
}

.summary-detail {
  margin-top: 4px;
}

.detail-row {
  margin-bottom: 20px;
}

.detail-card {
  height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header span {
  font-weight: bold;
  font-size: 16px;
}

.voice-card :deep(.el-card__body),
.matrix-card :deep(.el-card__body) {
  padding: 18px;
}

.voice-body {
  min-height: 180px;
}

.voice-status {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.voice-status-pill {
  width: fit-content;
  padding: 6px 12px;
  border-radius: 999px;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.06em;
}

.voice-status-pill.idle {
  background: #f3f4f6;
  color: #4b5563;
}

.voice-status-pill.active {
  background: linear-gradient(135deg, #10b981, #22c55e);
  color: #ffffff;
  box-shadow: 0 6px 14px rgba(34, 197, 94, 0.35);
}

.voice-hint {
  color: #6b7280;
  font-size: 13px;
}

.voice-transcript,
.voice-action {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  font-size: 13px;
  color: #1f2937;
}

.voice-label {
  font-weight: 600;
  color: #111827;
}

.voice-error {
  color: #b91c1c;
  font-size: 12px;
}

.matrix-card {
  background: radial-gradient(circle at top left, rgba(16, 185, 129, 0.12), transparent 60%),
    radial-gradient(circle at bottom right, rgba(14, 116, 144, 0.18), transparent 55%),
    #050b08;
}

.matrix-card :deep(.el-card__body) {
  background: transparent;
}

.matrix-stream {
  height: 240px;
  overflow-y: auto;
  padding: 12px;
  border-radius: 12px;
  background: rgba(2, 9, 6, 0.9);
  border: 1px solid rgba(16, 185, 129, 0.2);
  font-family: "JetBrains Mono", "Fira Code", monospace;
  font-size: 12px;
  color: #7bf7b0;
  text-shadow: 0 0 6px rgba(16, 185, 129, 0.4);
}

.matrix-line {
  display: flex;
  gap: 10px;
  padding: 2px 0;
  opacity: 0.9;
}

.matrix-time {
  color: rgba(123, 247, 176, 0.7);
  min-width: 78px;
}

.matrix-text {
  flex: 1;
}
</style>