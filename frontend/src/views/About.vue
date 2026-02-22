<template>
  <div class="nagare-container">
    <div class="page-header">
      <h1 class="page-title">{{ $t('about.title') }}</h1>
      <p class="page-subtitle">{{ $t('about.subtitle') }}</p>
    </div>

    <div class="about-content">
      <el-row :gutter="24">
        <el-col :lg="16">
          <el-card class="about-card main-card">
            <div class="brand-section">
              <div class="brand-logo">
                <el-icon :size="64" color="var(--brand-600)"><Monitor /></el-icon>
              </div>
              <h2 class="brand-name">Nagare</h2>
              <p class="brand-description">{{ $t('about.description') }}</p>
            </div>

            <el-divider />

            <div class="info-grid">
              <div class="info-item">
                <div class="info-label">{{ $t('about.version') }}</div>
                <div class="info-value">v1.1.0</div>
              </div>
              <div class="info-item">
                <div class="info-label">{{ $t('about.license') }}</div>
                <div class="info-value">MIT License</div>
              </div>
              <div class="info-item">
                <div class="info-label">{{ $t('about.github') }}</div>
                <div class="info-value">
                  <el-link 
                    type="primary" 
                    href="https://github.com/kurala719/nagare" 
                    target="_blank"
                    :icon="Share"
                  >
                    github.com/kurala719/nagare
                  </el-link>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>

        <el-col :lg="8">
          <el-card class="about-card contact-card">
            <template #header>
              <div class="card-header">
                <span>{{ $t('about.contact') }}</span>
              </div>
            </template>
            
            <div class="contact-list">
              <div class="contact-item">
                <div class="contact-icon author">
                  <el-icon><User /></el-icon>
                </div>
                <div class="contact-info">
                  <div class="contact-label">{{ $t('about.author') }}</div>
                  <div class="contact-value">kurala</div>
                </div>
              </div>

              <div class="contact-item">
                <div class="contact-icon email">
                  <el-icon><Message /></el-icon>
                </div>
                <div class="contact-info">
                  <div class="contact-label">{{ $t('about.email') }}</div>
                  <div class="contact-value">
                    <el-link type="primary" href="mailto:kurala719@gmail.com">kurala719@gmail.com</el-link>
                  </div>
                </div>
              </div>

              <div class="contact-item">
                <div class="contact-icon github">
                  <el-icon><Link /></el-icon>
                </div>
                <div class="contact-info">
                  <div class="contact-label">GitHub</div>
                  <div class="contact-value">
                    <el-link type="primary" href="https://github.com/kurala719" target="_blank">@kurala719</el-link>
                  </div>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup>
import { Monitor, User, Message, Link, Share } from '@element-plus/icons-vue'
</script>

<style scoped>
.about-content {
  margin-top: 24px;
}

.about-card {
  border: 1px solid var(--border-1);
  height: 100%;
}

.main-card {
  padding: 24px;
}

.brand-section {
  text-align: center;
  padding: 40px 0;
}

.brand-logo {
  margin-bottom: 24px;
  display: inline-flex;
  padding: 24px;
  background: var(--brand-50);
  border-radius: 24px;
  border: 1px solid var(--brand-100);
}

.brand-name {
  font-size: 32px;
  font-weight: 800;
  margin: 0 0 16px 0;
  color: var(--text-strong);
  font-family: var(--font-display);
  letter-spacing: -0.5px;
}

.brand-description {
  max-width: 600px;
  margin: 0 auto;
  line-height: 1.6;
  color: var(--text-muted);
  font-size: 16px;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 32px;
  padding: 24px 0;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.info-label {
  font-size: 13px;
  font-weight: 700;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.info-value {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-strong);
}

.contact-list {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.contact-item {
  display: flex;
  align-items: center;
  gap: 16px;
}

.contact-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
}

.contact-icon.author { background: var(--brand-50); color: var(--brand-600); }
.contact-icon.email { background: #f0fdf4; color: #16a34a; }
.contact-icon.github { background: #f8fafc; color: #334155; }

.contact-info {
  display: flex;
  flex-direction: column;
}

.contact-label {
  font-size: 12px;
  color: var(--text-tertiary);
}

.contact-value {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-strong);
}

.card-header {
  font-weight: 700;
  color: var(--text-strong);
}
</style>
