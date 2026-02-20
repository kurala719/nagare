<template>
  <div class="site-wrapper">
    <!-- Navbar -->
    <nav class="navbar">
      <div class="container nav-content">
        <div class="logo">
          <span class="logo-text">Nagare</span>
        </div>
        <div class="nav-links">
          <a href="#about">About</a>
          <a href="#features">Features</a>
          <a href="#documentation">Docs</a>
          <a href="#installation">Install</a>
          <a href="#community">Community</a>
          <a href="https://github.com/kurala719" target="_blank" class="github-link">
            <svg height="24" width="24" viewBox="0 0 16 16" fill="currentColor"><path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.18-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"></path></svg>
          </a>
        </div>
      </div>
    </nav>

    <!-- Hero Section -->
    <header class="hero">
      <div class="container hero-content">
        <h1>Intelligent Infrastructure <span class="gradient-text">Nagare</span></h1>
        <p class="hero-subtitle">A next-generation AI-powered operations platform for unified monitoring, diagnosis, and reporting.</p>
        <div class="hero-actions">
          <a href="#installation" class="btn btn-primary">Get Started</a>
          <a href="https://github.com/kurala719" target="_blank" class="btn btn-outline">View on GitHub</a>
        </div>
      </div>
    </header>

    <!-- About -->
    <section id="about" class="section">
      <div class="container">
        <h2 class="section-title">About Nagare</h2>
        <div class="about-grid">
          <div class="about-text">
            <p>Nagare is a comprehensive IT infrastructure management system designed to bridge the gap between complex monitoring data and actionable insights. Built with a Go backend and a Vue 3 frontend, it provides a seamless experience for DevOps engineers and SREs.</p>
            <p>The platform doesn't just show you that a system is down; it uses AI to tell you <strong>why</strong> and how to fix it, drawing from both global LLM knowledge and your own private localized Knowledge Base.</p>
          </div>
          <div class="about-stats">
            <div class="stat-card">
              <span class="stat-num">AI</span>
              <span class="stat-label">Driven Diagnosis</span>
            </div>
            <div class="stat-card">
              <span class="stat-num">RAG</span>
              <span class="stat-label">Knowledge Base</span>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Features -->
    <section id="features" class="section surface">
      <div class="container">
        <h2 class="section-title">Core Features</h2>
        <div class="features-grid">
          <div class="feature-card">
            <div class="feature-icon">🤖</div>
            <h3>AI Alert Diagnosis</h3>
            <p>Integrated with Gemini and OpenAI to provide real-time analysis of monitoring alerts and logs.</p>
          </div>
          <div class="feature-card">
            <div class="feature-icon">📚</div>
            <h3>RAG Knowledge Base</h3>
            <p>Upload your own documentation and technical manuals. Our RAG system uses them to provide localized context for AI analysis.</p>
          </div>
          <div class="feature-card">
            <div class="feature-icon">📊</div>
            <h3>Unified Monitoring</h3>
            <p>Seamlessly synchronize hosts and metrics from Zabbix, Prometheus, and other providers into a single dashboard.</p>
          </div>
          <div class="feature-card">
            <div class="feature-icon">🐚</div>
            <h3>Integrated WebSSH</h3>
            <p>Access your servers directly from the browser with a high-performance terminal supporting legacy algorithms.</p>
          </div>
          <div class="feature-card">
            <div class="feature-icon">📝</div>
            <h3>Automated Reporting</h3>
            <p>Generate weekly and monthly operations reports in PDF format with AI-generated executive summaries.</p>
          </div>
          <div class="feature-card">
            <div class="feature-icon">🔔</div>
            <h3>Multi-Channel Alerts</h3>
            <p>Configurable triggers for Email, Webhooks, and QQ integration with whitelist management.</p>
          </div>
        </div>
      </div>
    </section>

    <!-- Documentation -->
    <section id="documentation" class="section">
      <div class="container">
        <h2 class="section-title">Architecture</h2>
        <div class="docs-content">
          <div class="arch-diagram">
            <div class="node">Frontend (Vue 3 + Vite)</div>
            <div class="arrow">↓</div>
            <div class="node">Backend (Go + Gin)</div>
            <div class="arrow">↓</div>
            <div class="db-grid">
              <div class="node small">MySQL (Core)</div>
            </div>
          </div>
          <div class="tech-stack">
            <h3>Tech Stack</h3>
            <ul>
              <li><strong>Backend:</strong> Golang, Gin, GORM, Viper</li>
              <li><strong>Frontend:</strong> Vue 3, Element Plus, xterm.js, Vue-i18n</li>
              <li><strong>AI:</strong> LangChain-style RAG, WebSocket-based Chat</li>
              <li><strong>Database:</strong> MySQL 8.0</li>
            </ul>
          </div>
        </div>
      </div>
    </section>

    <!-- Installation -->
    <section id="installation" class="section surface">
      <div class="container">
        <h2 class="section-title">Installation</h2>
        <div class="install-steps">
          <h3>1. Backend Setup</h3>
          <pre><code># Clone the repository
git clone https://github.com/kurala719/nagare.git
cd nagare/backend

# Configure your settings
cp configs/nagare_config.json.example configs/nagare_config.json

# Run the server
go run cmd/server/main.go</code></pre>

          <h3>2. Frontend Setup</h3>
          <pre><code>cd ../frontend

# Install dependencies
npm install

# Run development server
npm run dev</code></pre>
        </div>
      </div>
    </section>

    <!-- Community -->
    <section id="community" class="section">
      <div class="container">
        <h2 class="section-title">Community & Contribution</h2>
        <p class="community-text">Nagare is an open-source project. We welcome contributions, bug reports, and feature requests!</p>
        <div class="community-actions">
          <a href="https://github.com/kurala719" target="_blank" class="btn btn-outline">Open an Issue</a>
          <a href="https://github.com/kurala719" target="_blank" class="btn btn-outline">Submit a Pull Request</a>
        </div>
      </div>
    </section>

    <!-- Footer -->
    <footer class="footer">
      <div class="container footer-content">
        <div class="footer-brand">
          <span class="logo-text">Nagare</span>
          <p>© 2026 kurala719. All rights reserved.</p>
        </div>
        <div class="footer-contact">
          <h4>Contact</h4>
          <p>Author: kurala719</p>
          <p>Email: <a href="mailto:kurala719@gmail.com">kurala719@gmail.com</a></p>
          <p>GitHub: <a href="https://github.com/kurala719" target="_blank">@kurala719</a></p>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup>
// Modern Vue 3 script setup
</script>

<style scoped>
.navbar {
  position: fixed;
  top: 0;
  width: 100%;
  background-color: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(8px);
  border-bottom: 1px solid var(--border-color);
  z-index: 1000;
  height: 64px;
  display: flex;
  align-items: center;
}

.nav-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.logo-text {
  font-size: 24px;
  font-weight: 800;
  color: var(--brand-600);
  letter-spacing: -1px;
}

.nav-links {
  display: flex;
  gap: 32px;
  align-items: center;
}

.nav-links a {
  font-weight: 500;
  color: var(--text-secondary);
}

.nav-links a:hover {
  color: var(--brand-600);
}

.hero {
  padding: 160px 0 100px;
  text-align: center;
  background: radial-gradient(circle at top right, var(--brand-50), transparent);
}

.hero h1 {
  font-size: 64px;
  margin-bottom: 24px;
  letter-spacing: -2px;
}

.gradient-text {
  background: linear-gradient(135deg, var(--brand-500), var(--brand-700));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.hero-subtitle {
  font-size: 20px;
  color: var(--text-secondary);
  max-width: 700px;
  margin: 0 auto 40px;
}

.hero-actions {
  display: flex;
  gap: 16px;
  justify-content: center;
}

.section-title {
  font-size: 36px;
  margin-bottom: 48px;
  text-align: center;
}

.surface {
  background-color: var(--surface-color);
}

.about-grid {
  display: grid;
  grid-template-columns: 1.5fr 1fr;
  gap: 60px;
  align-items: center;
}

.about-text p {
  font-size: 18px;
  color: var(--text-secondary);
  margin-bottom: 24px;
}

.about-stats {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.stat-card {
  background-color: white;
  padding: 32px;
  border-radius: 16px;
  box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  text-align: center;
}

.stat-num {
  display: block;
  font-size: 32px;
  font-weight: 800;
  color: var(--brand-600);
  margin-bottom: 8px;
}

.stat-label {
  color: var(--text-muted);
  font-weight: 600;
  font-size: 14px;
}

.features-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 32px;
}

.feature-card {
  background-color: white;
  padding: 40px;
  border-radius: 20px;
  border: 1px solid var(--border-color);
  transition: all 0.3s;
}

.feature-card:hover {
  transform: translateY(-8px);
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
}

.feature-icon {
  font-size: 40px;
  margin-bottom: 24px;
}

.feature-card h3 {
  margin-bottom: 16px;
}

.feature-card p {
  color: var(--text-secondary);
}

.arch-diagram {
  text-align: center;
  margin-bottom: 40px;
}

.node {
  display: inline-block;
  padding: 16px 32px;
  background-color: var(--brand-600);
  color: white;
  border-radius: 12px;
  font-weight: 600;
}

.node.small {
  padding: 12px 24px;
  font-size: 14px;
  background-color: var(--text-primary);
}

.db-grid {
  display: flex;
  gap: 20px;
  justify-content: center;
}

.arrow {
  font-size: 24px;
  margin: 10px 0;
  color: var(--text-muted);
}

.tech-stack ul {
  list-style: none;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.tech-stack li {
  padding: 12px;
  background-color: var(--surface-color);
  border-radius: 8px;
}

.community-text {
  text-align: center;
  font-size: 18px;
  margin-bottom: 32px;
}

.community-actions {
  display: flex;
  justify-content: center;
  gap: 16px;
}

.footer {
  padding: 80px 0;
  border-top: 1px solid var(--border-color);
  background-color: var(--surface-color);
}

.footer-content {
  display: flex;
  justify-content: space-between;
}

.footer-brand p {
  color: var(--text-muted);
  margin-top: 12px;
}

.footer-contact h4 {
  margin-bottom: 16px;
}

.footer-contact p {
  color: var(--text-secondary);
  margin-bottom: 8px;
}

@media (max-width: 768px) {
  .hero h1 { font-size: 40px; }
  .about-grid { grid-template-columns: 1fr; }
  .nav-links { display: none; }
  .footer-content { flex-direction: column; gap: 40px; }
}
</style>
