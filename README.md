# 高性能加密货币市场实时监控系统

## 项目概述

本项目是一个基于 Go 语言开发的高性能、低延迟加密货币市场实时监控系统，专门用于 Binance U本位永续合约市场的技术分析信号捕捉与即时通知。系统通过实时监测 BTC/USDT 和 ETH/USDT 交易对的多个时间粒度数据，运用 EMA 指标交叉策略生成交易信号，并通过 Webhook 推送富媒体消息卡片辅助交易员决策。

## 核心功能

### 1. 市场监控
- **交易对**：BTC/USDT、ETH/USDT（Binance U本位永续合约）
- **数据源**：Binance WebSocket API 实时 K 线数据
- **订阅频道**：`kline_5m`、`kline_15m`、`kline_1h`、`kline_4h`

### 2. 信号捕捉策略
- **技术指标**：EMA12（12周期指数移动平均线）、EMA144（144周期指数移动平均线）
- **信号触发条件**：
  - **金叉（Bullish Signal）**：
    1. EMA12 从下方上穿 EMA144
    2. 当前 K 线收盘价位于 EMA144 之上
  - **死叉（Bearish Signal）**：
    1. EMA12 从上方下穿 EMA144
    2. 当前 K 线收盘价位于 EMA144 之下
- **多时间周期支持**：系统同时监控 5分钟、15分钟、1小时、4小时四个时间粒度，独立计算信号

### 3. 信号处理流程
1. **数据接收**：实时接收 WebSocket K 线数据
2. **指标计算**：维护每个交易对、每个时间周期的 K 线队列，实时计算 EMA12 和 EMA144
3. **信号检测**：检测 EMA 交叉事件，验证收盘价位置条件
4. **信号过滤**：
   - 去重处理：防止同一信号在短时间内重复触发
   - 有效性验证：确保收盘价条件满足
5. **消息生成**：将信号转换为结构化消息数据

### 4. 即时通知
- **推送方式**：HTTP Webhook POST 请求
- **消息格式**：富媒体消息卡片（Message Card），支持以下平台：
  - Microsoft Teams
  - Slack
  - 钉钉
  - **飞书（Lark）** - 支持飞书群机器人 Webhook
  - 自定义 Webhook 接口
- **消息内容**：
  - 交易对与时间周期
  - 信号类型（金叉/死叉）
  - 当前价格与 EMA 值
  - 时间戳
  - 建议操作提示
- **飞书集成**：系统提供专门适配飞书消息卡片的格式，支持按钮、交互式消息和@提醒功能。

## 系统架构

### 组件模块
```
├── config/              # 配置文件管理
├── data/               # 数据采集层
│   ├── websocket/      # Binance WebSocket 客户端
│   └── kline/          # K 线数据处理
├── indicator/          # 技术指标计算
│   ├── ema.go          # EMA 计算引擎
│   └── crossover.go    # 交叉检测逻辑
├── signal/             # 信号处理层
│   ├── detector.go     # 信号检测器
│   └── filter.go       # 信号过滤器
├── notification/       # 通知服务
│   ├── webhook.go      # Webhook 发送器
│   └── messagecard.go  # 消息卡片生成
├── monitor/            # 系统监控
│   ├── metrics.go      # Prometheus 指标
│   └── healthcheck.go  # 健康检查
└── cmd/                # 应用程序入口
```

### 数据流设计
```
Binance WebSocket → K线数据解析 → 指标计算引擎 → 信号检测器 → 信号过滤器 → 消息生成器 → Webhook 推送
```

## 技术选型

### 编程语言
- **Go 1.21+**：高性能、高并发、低内存占用，适合实时系统

### 核心依赖
- **WebSocket 客户端**：`gorilla/websocket` 或自定义实现
- **配置管理**：`spf13/viper`（支持 YAML、JSON、环境变量）
- **HTTP 客户端**：标准库 `net/http` 或 `valyala/fasthttp`（高性能）
- **日志系统**：`uber-go/zap`（结构化日志）
- **指标监控**：`prometheus/client_golang`（性能指标暴露）

### 部署环境
- **操作系统**：Linux/Windows/macOS
- **容器化**：Docker 支持
- **进程管理**：Systemd 或 Kubernetes

## 配置说明

### 配置文件示例 (`config/config.yaml`)
```yaml
# Binance 配置
binance:
  websocket_url: "wss://fstream.binance.com/ws"
  reconnect_interval: 5s
  ping_interval: 30s

# 交易对配置
symbols:
  - "btcusdt"
  - "ethusdt"

# 时间周期配置
intervals:
  - "5m"
  - "15m"
  - "1h"
  - "4h"

# EMA 参数
indicators:
  ema_short_period: 12
  ema_long_period: 144

# 信号过滤
signal:
  deduplication_window: "10m"  # 信号去重时间窗口
  min_volume: 1000.0           # 最小交易量过滤（可选）

# Webhook 配置
webhook:
  enabled: true
  url: "https://your-webhook-endpoint.com"
  timeout: "10s"
  retry_count: 3
  retry_backoff: "1s"
  # 飞书 Webhook 配置示例
  lark:
    enabled: false
    webhook_url: "https://open.feishu.cn/open-apis/bot/v2/hook/{your_token}"
    secret: ""  # 可选，签名密钥
    msg_type: "interactive"  # 交互式消息卡片

# 消息卡片模板
message_card:
  title: "🎯 交易信号警报"
  theme_color: "0078D7"
  include_price: true
  include_ema_values: true
  include_timestamp: true
  # 飞书特定配置
  lark_specific:
    at_all: false  # 是否 @ 所有人
    at_users: []   # 要 @ 的用户ID列表
    buttons:
      - text: "查看详情"
        url: "https://binance.com/zh-CN/futures/{symbol}"
      - text: "忽略信号"
        action: "ignore"

# 监控配置
monitoring:
  prometheus_enabled: true
  prometheus_port: 9090
  healthcheck_port: 8080
  log_level: "info"
```

### 环境变量覆盖
所有配置项均支持环境变量覆盖，格式：`FIBO_<SECTION>_<KEY>`，例如：
- `FIBO_BINANCE_WEBSOCKET_URL`
- `FIBO_WEBHOOK_URL`
- `FIBO_MONITORING_LOG_LEVEL`

## 性能指标

系统暴露以下 Prometheus 指标：
- `fibo_websocket_connections`：WebSocket 连接状态
- `fibo_kline_received_total`：接收的 K 线数量
- `fibo_ema_calculated_total`：EMA 计算次数
- `fibo_signals_detected_total`：检测到的信号数量
- `fibo_webhook_sent_total`：Webhook 发送次数
- `fibo_webhook_errors_total`：Webhook 错误次数
- `fibo_processing_latency_seconds`：处理延迟直方图

## 部署与运行

### 本地运行
```bash
# 克隆项目
git clone <repository-url>
cd Fibo_ws

# 安装依赖
go mod download

# 编辑配置文件
cp config/config.example.yaml config/config.yaml
# 修改 config.yaml 中的 Webhook URL 等配置

# 构建
go build -o fibo-monitor ./cmd

# 运行
./fibo-monitor
```

### Docker 运行
```bash
# 构建镜像
docker build -t fibo-monitor .

# 运行容器（基础配置）
docker run -d \
  -v $(pwd)/config:/app/config \
  -p 8080:8080 \
  -p 9090:9090 \
  --name fibo-monitor \
  fibo-monitor

# 运行容器（带飞书环境变量配置）
docker run -d \
  -v $(pwd)/config:/app/config \
  -e FIBO_WEBHOOK_LARK_ENABLED=true \
  -e FIBO_WEBHOOK_LARK_WEBHOOK_URL="https://open.feishu.cn/open-apis/bot/v2/hook/{your_token}" \
  -e FIBO_WEBHOOK_LARK_SECRET="your_secret" \
  -p 8080:8080 \
  -p 9090:9090 \
  --name fibo-monitor-lark \
  fibo-monitor

### 使用 GitHub Container Registry (GHCR) 直接运行
本项目已配置 GitHub Actions 自动构建 Docker 镜像并发布到 GHCR。您可以直接拉取并运行最新镜像，无需本地构建。

```bash
# 1. 拉取最新镜像
# 注意：替换 <username> 为 GitHub 用户名，<repo> 为仓库名（需小写）
docker pull ghcr.io/<username>/<repo>:latest

# 2. 运行容器
docker run -d \
  -v $(pwd)/config:/app/config \
  -p 8080:8080 \
  -p 9090:9090 \
  --name fibo-monitor \
  ghcr.io/<username>/<repo>:latest
```

# 使用 Docker Compose（推荐）
docker-compose up -d
```

### Docker Compose 配置示例 (`docker-compose.yml`)
```yaml
version: '3.8'

services:
  fibo-monitor:
    image: ghcr.io/uykb/fibo_ws:latest # 推荐使用 GHCR 镜像
    container_name: fibo-monitor
    restart: unless-stopped
    ports:
      - "8080:8080"   # 健康检查端口
      - "9090:9090"   # Prometheus 指标端口
    volumes:
      - ./config:/app/config:ro
      - ./logs:/app/logs
    environment:
      # Binance 配置
      - FIBO_BINANCE_WEBSOCKET_URL=wss://fstream.binance.com/ws
      - FIBO_BINANCE_RECONNECT_INTERVAL=5s
      # 飞书 Webhook 配置 (推荐使用环境变量覆盖配置文件)
      - FIBO_WEBHOOK_LARK_ENABLED=true
      - FIBO_WEBHOOK_LARK_WEBHOOK_URL=https://open.feishu.cn/open-apis/bot/v2/hook/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
      - FIBO_WEBHOOK_LARK_SECRET=  # 如果开启了签名校验，请填写密钥
      # 监控配置
      - FIBO_MONITORING_LOG_LEVEL=info
      - FIBO_MONITORING_PROMETHEUS_ENABLED=true
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

## 飞书 (Lark) 集成指南

本系统原生支持飞书群机器人的 Webhook 推送，并支持富媒体消息卡片。

### 1. 获取 Webhook 地址
1.  在飞书群组中，点击右上角设置 -> 群机器人 -> 添加机器人 -> 自定义机器人。
2.  添加后，您将获得一个 Webhook 地址，格式如下：
    `https://open.feishu.cn/open-apis/bot/v2/hook/xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`
3.  (可选) 在安全设置中，您可以勾选 "签名校验"，并获取 **签名密钥 (Secret)**。

### 2. 配置环境变量
为了安全起见，建议通过环境变量配置 Webhook 地址和密钥，而不是直接写入配置文件。

| 环境变量名称 | 描述 | 示例值 |
| :--- | :--- | :--- |
| `FIBO_WEBHOOK_LARK_ENABLED` | 是否启用飞书推送 | `true` |
| `FIBO_WEBHOOK_LARK_WEBHOOK_URL` | 飞书机器人的 Webhook 地址 | `https://open.feishu.cn/...` |
| `FIBO_WEBHOOK_LARK_SECRET` | (可选) 签名密钥 | `your_secret_string` |

### 3. 测试运行
配置完成后，您可以直接启动 Docker 容器：

```bash
docker run -d \
  -e FIBO_WEBHOOK_LARK_ENABLED=true \
  -e FIBO_WEBHOOK_LARK_WEBHOOK_URL="https://open.feishu.cn/open-apis/bot/v2/hook/您的Token" \
  --name fibo-monitor \
  ghcr.io/uykb/fibo_ws:latest
```

## 故障排除

### 常见问题
1. **WebSocket 连接断开**
   - 检查网络连接和防火墙设置
   - 确认 Binance API 状态
   - 查看日志中的重连记录

2. **Webhook 发送失败**
   - 验证 Webhook URL 是否正确
   - 检查目标服务是否可访问
   - 查看重试日志

3. **信号未触发**
   - 确认 EMA 参数设置
   - 检查收盘价条件是否满足
   - 验证去重窗口设置

### 日志查看
```bash
# 查看实时日志
tail -f logs/fibo-monitor.log

# 按级别过滤
grep "ERROR" logs/fibo-monitor.log
grep "SIGNAL" logs/fibo-monitor.log
```

## 开发指南

### 代码结构规范
- **包组织**：按功能模块划分，避免循环依赖
- **错误处理**：使用 Go 1.13+ 的错误包装
- **并发安全**：合理使用 sync 包或 channel
- **测试覆盖**：单元测试覆盖率 >80%

### 添加新的技术指标
1. 在 `indicator/` 目录下创建新指标实现
2. 实现 `Calculator` 接口
3. 在信号检测器中注册新指标
4. 添加相应的配置参数

### 扩展新的交易所
1. 在 `data/exchange/` 下创建新的适配器
2. 实现 `ExchangeClient` 接口
3. 更新配置支持

## 路线图

### 短期计划（V1.0）
- [x] 项目需求分析与设计
- [ ] 基础框架搭建
- [ ] Binance WebSocket 集成
- [ ] EMA 指标计算引擎
- [ ] 交叉信号检测逻辑
- [ ] Webhook 通知系统
- [ ] 基础监控与日志

### 中期计划（V1.1）
- [ ] 多交易所支持（Bybit、OKX）
- [ ] 更多技术指标（MACD、RSI、布林带）
- [ ] 信号回测框架
- [ ] 图形化仪表盘
- [ ] 移动端通知（Telegram、微信）

### 长期计划（V2.0）
- [ ] 机器学习信号预测
- [ ] 自适应参数优化
- [ ] 分布式部署支持
- [ ] 实时风险控制模块
- [ ] API 服务暴露

## 免责声明

本项目仅为技术分析和决策辅助工具，不构成任何投资建议。加密货币交易具有高风险，用户应自行承担交易决策带来的风险。开发者不对因使用本系统而产生的任何直接或间接损失负责。

## 许可证

MIT License

## 联系方式

如有问题或建议，请通过以下方式联系：
- GitHub Issues: [项目 Issues 页面]
- 电子邮件: [联系邮箱]

---
*最后更新: 2025-12-17*
