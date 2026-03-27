# Nagare (流) - 基于 AI 的 IT 基础设施智能大脑

[**English (EN)**](./README.md)

**Nagare**（日语写作「ながれ」，意为“流”）是一个监视服务器和应用程序的智能平台。与仅仅在出问题时“报警”的传统系统不同，Nagare 使用人工智能来理解**为什么**出故障，并告诉您**如何**修复它。

## 🌟 Nagare 有何特别之处？
- **倾听**：集成 Zabbix，监听系统的每一次“心跳”。
- **记忆**：拥有“知识库”(RAG)。出现问题时，它会查阅过去的笔记和手册寻找解决方案。
- **思考**：使用先进的 AI (Google Gemini) 分析错误。它就像一位 7x24 小时待命的高级工程师。
- **行动**：直接从浏览器使用内置的“指挥中心”(WebSSH) 修复问题。

---

## 📖 用户手册
Nagare 新手？请查看我们的 **[Nagare 用户手册 (通用指南)](./NAGARE_USER_MANUAL.md)**。
它用通俗易懂的语言解释了系统的每一个部分。

---

## 📂 技术导航
如果您是开发人员或工程师，请查阅我们的深度技术手册：

| 手册 | 核心概念 | 技术重点 |
| :--- | :--- | :--- |
| [**架构设计**](./doc/ARCHITECTURE.md) | 神经系统 | Go 1.24, Gin, 高并发。 |
| [**数据库模式**](./doc/DATABASE_SCHEMA.md) | 存储引擎 | MySQL/GORM, ERD, 历史追踪。 |
| [**部署指南**](./doc/DEPLOYMENT_GUIDE.md) | 生产与测试 | Nginx, systemd, JWT Secrets, HTTPS。 |
| [**开发指南**](./doc/DEVELOPER_GUIDE.md) | 代码规范 | MVC 分层, Vue 3 Composition API。 |
| [**集成指南**](./doc/INTEGRATIONS.md) | 连接监控源 | Zabbix Webhooks, 自定义集成。 |
| [**AI 配置**](./doc/AI_CONFIGURATION.md) | 大脑设置 | Gemini, OpenAI, RAG 调优。 |
| [**安全与 RBAC**](./doc/RBAC_SECURITY_MODEL.md) | 访问控制 | 权限等级, JWT, WebSSH 安全性。 |
| [**故障排除**](./doc/TROUBLESHOOTING.md) | 解决问题 | 常见错误, 性能调优。 |
| [**WebSSH 与安全**](./doc/WEBSSH_SECURITY.md) | 指挥中心 | WebSocket 代理, xterm.js, XSS 防御。 |
| [**报表系统**](./doc/REPORTING_SYSTEM.md) | 定期检查 | PDF 渲染, Go-Charts, Cron 任务。 |
| [**前端指南**](./doc/FRONTEND_GUIDE.md) | 交互界面 | Vue 3, Vite, 感知速度优化。 |
| [**通信通知**](./doc/COMMUNICATION_NOTIFICATIONS.md) | 通知系统 | WebSockets, QQ 机器人, 白名单安全。 |
| [**API 参考**](./doc/API_REFERENCE.md) | 系统语言 | RESTful 接口, RBAC, MCP 协议。 |

---

## ⚡ 工程指标
- **极速性能**：优化的 JSON 处理 (`jsoniter`) 比标准工具快 30%。
- **兼容传统系统**：无缝对接现有运维系统（如 Zabbix），为传统基础设施提供 AI 增强能力。
- **面向未来**：作为 **MCP 客户端**，允许 Nagare AI 动态加载并无缝使用外部工具和脚本。

## 📄 许可证
Apache License 2.0
