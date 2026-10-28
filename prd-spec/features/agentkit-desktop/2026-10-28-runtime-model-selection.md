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

与 Studio 一致，宿主通过当前登录云账号的 Ark 目录，将 Runtime 模型列表与
账号可用模型的精确版本 ID 求交集，再补入 Runtime 默认模型。目录读取遵循
Studio 的默认 API Key 选择规则：排除 Restricted Key，优先匹配
`MODEL_AGENT_API_KEY_NAME`，否则按名称、ID 排序选取第一个 Key。
宿主将 `ListModelActivations` 的 Available 状态与 `/api/v3/models` 版本目录
关联，排除 Shutdown、非聊天模型和 API Key 权限不允许的模型。使用 Ark 的
北京区域目录，与 Studio 火山引擎账号目录保持一致；原始 API Key 不返回前端。
目录不可用时沿用 Studio 的 Runtime 能力列表回退；成功返回的空目录仍执行
交集，仅保留默认模型。查询中切换账号或 Runtime 时拒绝迟到响应。
没有迁移 Studio 的 A2A 逐条消息发送协议，也没有为 REST RunRequest
添加服务端不支持的 modelId 字段。保存仍作用于后续消息的会话配置。
实际执行模型的优先级由 Runtime 决定，包括 Main Profile 的模型设置。

## 验证

- 账号目录对齐：契约测试覆盖 Ark 固定签名区域、分页、默认 API Key、
  开通状态和精确版本关联、API Key 允许/拒绝规则、下线与非聊天模型过滤。
  会话集成测试覆盖成功交集、空目录保留默认模型、目录失败回退和账号/Runtime
  切换隔离；Go 全量 race 测试与相关 vet、14 项前端测试、类型检查和前端构建通过。
  使用隔离 HTTP fixture 验证，未访问真实云账号目录或修改模型开通状态。
- Go：Agent Card 提取、默认模型补入、去重、旧连接拒绝、无能力不回退、
  非法能力响应拒绝；全部 Go 测试通过。
- 前端：新 Runtime 模型无需快照、切换连接的迟到响应、失败重试、
  无隐式配置写入、保留 MCP 与 ETag、搜索 Enter 不发消息；534 个测试通过。
- TypeScript、ESLint、Prettier、design:check、桌面前端构建通过。
- 浏览器使用实际组件和样式、模拟 Runtime 目录检查单一焦点框、搜索过滤、
  选中回显与菜单收起。未向真实 Runtime 发送测试聊天或修改模型配置。
