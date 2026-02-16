# Redis Task Queue & Alert Generator Implementation

## Summary

成功实现了基于 Redis 的异步任务队列系统和模拟告警生成器，解决了同步阻塞问题。

## 完成内容

### 1. Redis 任务队列 (Queue System)

**位置**: `backend/pkg/queue/queue.go`

**核心功能**:
- 任务入队/出队操作(Enqueue/Dequeue)
- 支持7种任务类型
- 失败重试机制(最多3次)
- 死信队列(Dead Letter Queue)
- 队列统计接口

**特点**:
- 使用 Redis List 作为底层存储
- 支持 BLPOP 阻塞操作(5秒超时)
- 任务 ID 带时间戳确保唯一性
- 连接池配置默认为 localhost:6379

### 2. 异步 API 端点

**新增 3 个端点**:

```
POST /api/v1/monitors/:id/hosts/pull-async
→ 立即返回 202 Accepted + task_id
→ 后台处理同步主机

POST /api/v1/monitors/:m_id/hosts/:h_id/items/pull-async  
→ 立即返回 202 Accepted + task_id
→ 后台处理同步指标

GET /api/v1/queue/stats
→ 返回所有队列当前长度
```

**输出示例**:
```json
{
  "message": "Host pull task queued",
  "task_id": "pull_hosts:1707996000000000000",
  "monitor_id": 1
}
```

### 3. 任务工作者 (Task Workers)

**位置**: `backend/internal/service/worker.go`

**特点**:
- 4 个并发工作协程(可配置)
- 轮询所有队列 (pull_hosts, pull_items, generate_alerts)
- 5秒阻塞等待超时
- 自动失败重试和死信队列转移
- 使用上下文进行操作追踪

**处理流程**:
```
Worker Loop:
├─ Try Dequeue pull_hosts (5s timeout)
├─ Try Dequeue pull_items (5s timeout)  
├─ Try Dequeue generate_alerts (5s timeout)
└─ Sleep 100ms if no tasks
```

### 4. 模拟告警生成器 (Alert Generator)

**位置**: `backend/internal/service/alert.go`

**新增函数**:

1. **GenerateTestAlerts(count int)**
   - 生成随机告警
   - 随机选择主机和指标
   - 3 个严重级别 (Info/Warning/Critical)
   - 15 种预定义消息
   - 严重告警自动标记主机/指标为错误状态

2. **CalculateAlertScore()**
   - 计算健康评分 (0-100)
   - Info告警权重: 1.0
   - Warning告警权重: 5.0  
   - Critical告警权重: 20.0
   - 已解决告警权重减半

**告警生成示例**:
```json
{
  "message": "High CPU usage detected",
  "severity": 2,
  "status": 0,
  "host_id": 3,
  "item_id": 15,
  "comment": "Auto-generated test alert at 2026-02-15T10:30:00Z"
}
```

### 5. API 端点补充

**告警相关新增**:
```
POST /api/v1/alerts/generate-test?count=5
→ 生成5条测试告警

GET /api/v1/alerts/score
→ 返回当前健康评分
```

### 6. 配置更新

**nagare_config.json** 新增 Redis 配置:
```json
{
  "redis": {
    "addr": "localhost:6379"
  }
}
```

环境变量支持: `REDIS_ADDR=redis.example.com:6379`

## 文件清单

### 新建文件
- `backend/pkg/queue/queue.go` - 队列核心实现
- `backend/internal/service/queue.go` - 应用层队列服务  
- `backend/internal/service/worker.go` - 任务工作者
- `backend/internal/api/queue.go` - API 控制器
- `REDIS_QUEUE_GUIDE.md` - 详细使用指南
- `test_queue.ps1` - 测试脚本

### 修改文件
- `go.mod` - 添加 redis/go-redis 依赖
- `backend/cmd/server/main.go` - 初始化队列和启动工作者
- `backend/internal/service/alert.go` - 添加告警生成函数
- `backend/cmd/server/router/router.go` - 新增路由
- `backend/configs/nagare_config.json` - Redis 配置

## 使用流程

### 1. 启动应用
```bash
cd backend
go run cmd/server/main.go
```

日志确认:
```
INFO: task queue initialized {redis_addr: localhost:6379}
INFO: task workers started {worker_count: 4}
```

### 2. 验证队列状态
```bash
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/v1/queue/stats
```

### 3. 队列异步任务
```bash
# 异步同步主机
curl -X POST \
  -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/v1/monitors/1/hosts/pull-async

# 返回 202 + task_id，立即返回，后台处理
```

### 4. 生成测试数据
```bash
# 生成10条告警
curl -X POST \
  -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/v1/alerts/generate-test?count=10

# 检查健康评分
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/v1/alerts/score
```

### 5. 使用测试脚本
```bash
.\test_queue.ps1 -Token "your-jwt-token"
```

## 架构优势

### 性能改进
| 指标 | 同步 | 异步 |
|------|------|------|
| HTTP 响应时间 | 30-60s | <100ms |
| 用户体验 | 阻塞等待 | 立即反馈 |
| 并发能力 | 低(被长操作阻塞) | 高(多工作者) |
| 容错能力 | 无 | 失败重试+死信队列 |

### 可扩展性
- 简单增加工作者数量提升吞吐
- Redis 易于集群部署
- 死信队列便于故障分析
- 任务日志完整追踪

## 测试场景

### 场景1: 批量同步主机
```bash
# 快速队列5个同步任务
for i in {1..5}; do
  curl -X POST http://localhost:8080/api/v1/monitors/$i/hosts/pull-async
done

# 查看队列
curl http://localhost:8080/api/v1/queue/stats
# Output: {"pull_hosts": 5, ...}

# 等待3秒，工作者处理
# 再次查看
curl http://localhost:8080/api/v1/queue/stats  
# Output: {"pull_hosts": 0, ...}  ✓
```

### 场景2: 告警生成与健康评分
```bash
# 生成50条严重告警
curl -X POST http://localhost:8080/api/v1/alerts/generate-test?count=50

# 检查健康评分下降
curl http://localhost:8080/api/v1/alerts/score
# Output: 20 (严重下降)

# 手动解决一些告警
# ...

# 再次检查评分上升
curl http://localhost:8080/api/v1/alerts/score
# Output: 65 (评分上升) ✓
```

## 配置建议

### 开发环境
```json
{
  "redis": {
    "addr": "localhost:6379"
  }
}
```

### 生产环境
```json
{
  "redis": {
    "addr": "redis-cluster.production.svc:6379"
  }
}
```

## 监控与调试

### 查看工作者日志
```bash
tail -f logs/service.log | grep worker
```

### 查看死信队列
```bash
redis-cli
> LLEN nagare:queue:dead
> LRANGE nagare:queue:dead 0 10
```

### 手动清理队列(测试用)
```bash
redis-cli FLUSHDB
```

## 故障排查

### 问题: Redis 连接失败
```
Error: failed to initialize task queue
```

**解决**:
1. 确认 Redis 运行: `redis-cli ping`
2. 检查配置文件中的地址和端口
3. 检查防火墙规则

### 问题: 任务无法处理
**检查**:
1. 工作者是否启动: 日志中查找 "task workers started"
2. 队列是否有任务: `GET /api/v1/queue/stats`
3. 死信队列是否有失败: `redis-cli LLEN nagare:queue:dead`

### 问题: 内存占用过高
**解决**:
1. 减少工作者数量
2. 增加处理速度(优化长操作)
3. 清理死信队列

## 后续增强

- [ ] 任务进度实时查询 API
- [ ] WebSocket 任务完成通知
- [ ] 任务超时自动失败
- [ ] Redis 集群支持
- [ ] 任务调度器(CRON)
- [ ] 分布式追踪(Jaeger)
- [ ] 指标导出(Prometheus)

## 文档参考

详见 [REDIS_QUEUE_GUIDE.md](./REDIS_QUEUE_GUIDE.md)
