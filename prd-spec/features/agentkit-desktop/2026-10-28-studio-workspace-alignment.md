# Studio 工作区能力对齐

## 范围

补齐现有 A2A 协作、Worker 技能/MCP/共享 AGENTS.md、消息渠道绑定。
基于任务开始时的脏工作区继续适配，保留此前汉化、模型选择和其他功能修改。
不适配 Main 配置，不自动提交、推送，不对真实云端执行测试写入。

## 实现与来源

Studio 参考为本机 veadk-python 中的 A2ACenter、a2a_center.py、
MpaAgentInfoRail、skills/repository.py、RuntimeChannels 与 wecomAuthorization。
Runtime 参考为本机 agentkit-mpa-agent 中的 studio_metadata、collaborating_agents、
channels、mcp_registry、MCP 配置解析与 Agent Card 资源拓扑。
所有生产改动位于 OpenMPA。

### Worker 资源

- 技能空间从 Agent Card 中所有 agent 节点的 mounts 关系发现，仅接受 configured
  的 skill-space，去重并使用已绑定 Runtime 的地域。无需先选择会话。
- 使用 Studio 的 ListSkillsBySkillSpace 分页关系接口。仅在关系不存在时，读取
  GetSkillSpace 名称，再通过 ListSkillsBySpaceId/GetSkillInfo 恢复旧目录；
  不把权限错误当成空列表，不使用错误的 InnerTags 沙箱目录契约。
- MCP 页面读取 Runtime 与当前关联 Worker Tool 部署配置中的外部 MCP，支持
  MCP_SERVERS_JSON、MCP_URLS、AGENTKIT_MCP_SERVERS 和托管 MCP Toolset；
  Runtime 的 mcp-resources/current 只补充已发现资源及工具描述。
  不再读取会话 MCP，不修改会话配置或 Main Profile。
- MCP 地址去除查询与 fragment；认证 Header、密钥和工具 schema 不进入 UI。
- AGENTS.md 自动展示现有 Studio 共享文档，继续使用已有 revision 校验保存、
  冲突读回和保留草稿逻辑。Main Profile 的服务端逻辑和保存接口未修改。

### A2A

- 自动加载绑定、空间与成员；保存绑定和别名后可继续操作。
- 新增创建空间、意图路由设置、Runtime 注册（查询实际网络端点）、HTTPS Card
  URL/JSON 导入、成员详情/公开技能、注销成员和删除未绑定的空空间。
- 云端操作复用固定 AgentKit 签名入口及版本。地域、目标 Runtime、账号来自 Go
  的已验证绑定，前端不能指定其他凭据或目标连接。
- 注销前校验成员属于所选空间；删除空间前读回当前绑定、空间身份及成员总数。
  注销不删除原 Runtime。删除操作由页面明确确认。
- 写入不自动重试；异常或不完整回执要求刷新确认。注册结果未知时保留输入，
  阻止直接重复提交。切换空间关闭注册表单，防止向错误目标提交。
- 保留已保存别名的 @ 协作草稿流程，不自动发送消息。

### 消息渠道

- 使用当前 Runtime 的凭据和云端已验证账号身份连接渠道管理，与 Studio 的
  Runtime 管理方式一致；手动渠道凭据仍可覆盖。业务会话身份保持独立。
- 飞书、Lark、钉钉、企业微信入口；根据 Runtime 能力展示手动绑定，检查
  serverSideBinding、bindingReady 和所有者身份。
- 扫码绑定保持同一操作 ID 自动轮询，完成后停止并刷新账号。未知结果阻止新建；
  明确拒绝的请求可修正后重新提交。保留原有状态读回、重试与恢复入口。
- 企业微信使用 Studio 同版本 0.1.0 SDK，提前懒加载以保留点击时的弹窗激活，
  授权凭据通过现有服务端手动绑定接口提交，支持取消；也可手动输入凭据。
- 保留账号权限、私聊用户、模板和机器人配置功能。切换 Runtime 时隔离页面状态。

## 验收

- 前端 99 个文件、541 个测试通过；涵盖无会话 Worker MCP、A2A 创建/未知结果、
  别名立即可用、二维码轮询停止、未就绪拒绝、未知结果防重复、企业微信提交契约。
- Go 全量 race 测试与 vet 通过。新增契约测试覆盖签名请求、成员归属、空空间删除
  前置条件、Card 投影、模型外的凭据隔离、Worker 挂载和托管技能关系。
- TypeScript、ESLint、Prettier、design:check、桌面前端构建通过。
- 浏览器用实际组件、样式及模拟 IPC 验证：自动展示技能，无会话 MCP，自动展示
  AGENTS.md，A2A 创建空间后自动选中，渠道轮询完成并回显账号。
- 页面证据在本机忽略目录 `.openmpa/worker-resources-fixed.png` 和
  `.openmpa/channel-binding-fixed.png`。不包含真实账号和凭据。

## 验证边界

未执行真实云端注册/删除、共享文档写入或渠道绑定。真实账号权限、旧 Runtime
能力、企业微信在原生 WebView 中的授权弹窗及第三方扫码结果需要在对应环境确认。
旧 Runtime 不提供资源工具元数据时保留已部署的外部 MCP 展示，不编造工具列表。
新增 Go 接口和绑定身份需重启开发版并重新选择 Runtime 后加载。
