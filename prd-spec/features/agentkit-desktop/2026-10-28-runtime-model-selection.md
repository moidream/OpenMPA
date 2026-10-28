# Runtime 模型选择与搜索框修复

## 行为

- 从当前连接 Runtime 的 `/.well-known/agent-card.json` 读取
  `urn:veadk:mpa:model-selection:v1` 扩展的 `models` 与 `defaultModel`，
  与 Studio 普通 MPA 聊天的 Runtime 模型来源一致。
- 删除客户端内置模型快照及其未经 Runtime 确认的厂商、上下文窗口说明。
  新模型无需发布客户端即可进入列表；默认模型缺失于列表时补入列表。
- 切换 Runtime 后清空旧列表，拒绝迟到的旧连接响应。读取失败可重试；
  未声明该能力的 Runtime 显示说明，不编造可选模型，也不阻止正常聊天。
- 打开、搜索或刷新菜单不会修改会话。明确选择模型后，继续使用已有会话
  execution-config 接口及 ETag 保存，仅替换 model，保留 MCP 配置。
- 搜索框仅由外层绘制焦点框，内层 input 不再继承通用表单的边框、圆角和
  最小高度。菜单和列表高度受视口约束；搜索框内按 Enter 不提交聊天表单。

## 与 Studio 的边界

Studio 可进一步将 Runtime 模型列表与账号已开通模型求交集；本客户端当前
没有该账号模型目录服务，采用 Studio 在该服务不可用时使用的 Runtime
能力列表。没有迁移 Studio 的 A2A 逐条消息发送协议，也没有为 REST RunRequest
添加服务端不支持的 modelId 字段。保存仍作用于后续消息的会话配置。
实际执行模型的优先级由 Runtime 决定，包括 Main Profile 的模型设置。

## 验证

- Go：Agent Card 提取、默认模型补入、去重、旧连接拒绝、无能力不回退、
  非法能力响应拒绝；全部 Go 测试通过。
- 前端：新 Runtime 模型无需快照、切换连接的迟到响应、失败重试、
  无隐式配置写入、保留 MCP 与 ETag、搜索 Enter 不发消息；534 个测试通过。
- TypeScript、ESLint、Prettier、design:check、桌面前端构建通过。
- 浏览器使用实际组件和样式、模拟 Runtime 目录检查单一焦点框、搜索过滤、
  选中回显与菜单收起。未向真实 Runtime 发送测试聊天或修改模型配置。
