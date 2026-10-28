# React Desktop 助手、定时任务与主题收口设计

本文档按职责拆分为以下分卷：

- [需求与约束](./2026-10-27-react-desktop-assistant-followup-requirements.md)
- [方案设计](./2026-10-27-react-desktop-assistant-followup-design.md)
- [接口与配置契约](./2026-10-27-react-desktop-assistant-followup-contracts.md)
- [收口实施与验收](./2026-10-27-react-desktop-assistant-followup-operations.md)

## 1. 文档状态

| 字段         | 内容                                                                                         |
| ------------ | -------------------------------------------------------------------------------------------- |
| 标题         | React Desktop 助手、定时任务与主题收口设计                                                   |
| 状态         | 评审中                                                                                       |
| 创建日期     | 2026-10-27                                                                                   |
| 变更类型     | 功能与缺陷收口                                                                               |
| 关联任务     | 无                                                                                           |
| 关联历史设计 | [React Desktop 设计 token](../../../design/README.md)、[Native v1 合同](native-v1/README.md) |
| 版本关系     | 首次设计，补充当前 React Desktop 迁移后的未完成项                                            |

结论：本设计只关闭 React Desktop 已经暴露但尚未闭环的五类问题。AgentKit skill space 的目录与导入需要后端新契约，本期不伪造该能力，只把边界写清并保留为待确认项。

实现状态：FR-1、FR-3、FR-4、FR-5、FR-6、FR-7、FR-8 已实现并通过本地门禁。FR-2 以“只显示已配置 skills”的边界说明实现。真正的 skill space 筛选导入等待 OPEN-1。

## 2. 背景、问题与证据

### 2.1 当前问题或机会

React Desktop 已经完成侧边栏、设置分区、助手配置、定时任务和主题的第一轮改造。剩余问题集中在五处。

1. 助手配置只能读取和编辑当前 agent 的 `skills`，无法浏览 AgentKit skill space，也没有导入入口。
2. 定时任务只在当前页按 `agentId` 过滤，翻页后会漏掉其他页上的同助手任务，总数仍来自全量统计。
3. React 主题依赖一次性的 `themeChanged` 事件。事件早于监听注册时会丢失，React 与原生主题可能短暂不一致。
4. Spaces 的 React 入口已删除，但 `Pages.tsx` 和对应测试仍在仓库中，形成死代码。
5. 原生侧边栏和设置页没有可重复的视觉产物。当前已安装应用正在运行，临时开发实例会被单实例锁拦截。

不解决的后果是：助手配置看起来可用但能力不完整；定时任务出现跨页漏项；主题在重启或首次加载时可能闪回；Spaces 死代码继续被维护；视觉验收只能依赖人工记忆。

### 2.2 已验证事实

| ID      | 类型       | 结论/假设                                                                                                         | 精确证据                                                                                                                                                                                           | 设计影响                                                     | 验证/关闭方式                                                |
| ------- | ---------- | ----------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------ | ------------------------------------------------------------ |
| FACT-1  | 已验证事实 | `agentkit-mpa-agent` 只提供当前 agent 的读取、创建、更新和 profile 应用，没有 skill 目录或 skill space 列表路由。 | `agentkit-mpa-agent/app/api/v1/agents.py:128` GET，`:230` PUT；该文件路由集合见 `:110`、`:142`、`:185`、`:204`、`:215`、`:249`。                                                                   | skill space 浏览与导入不能由现有客户端 API 完成。            | 路由清单与 OpenAPI 复核。                                    |
| FACT-2  | 已验证事实 | agent 的 skills 是 `SkillConfig` 引用列表，字段为 `id`、`name`、`type`、`enabled`、`config`。                     | `agentkit-mpa-agent/app/schemas/agent_config.py:19`。                                                                                                                                              | 客户端可以增删引用，但不能上传 skill 文件内容。              | schema 复核。                                                |
| FACT-3  | 已验证事实 | MPA 的 skill space 是服务端缓存与 provider，使用 `skill_space_id`，没有公开的客户端目录路由。                     | `mpa-agent/app/services/skill_names_cache.py:23`、`:86`、`:91`；`mpa-agent/app/api/deps.py:553`。                                                                                                  | skill space 目录需要后端新增稳定契约。                       | 源码复核；后端确认。                                         |
| FACT-4  | 已验证事实 | 定时任务集合路由支持 `query`、分页和排序，没有 `agentId` 过滤参数。                                               | `agentkit-mpa-agent/app/api/v1/scheduled_tasks.py:88`；OpenMPA 客户端 `desktop/internal/schedule/service.go:122` 只发送 `limit`、`offset`、`includeDisabled`、`query`、`sortBy`、`sortDirection`。 | 按助手过滤必须读取所有页后执行，或等待后端新增过滤参数。     | 路由与客户端请求参数复核。                                   |
| FACT-5  | 已验证事实 | 定时任务定义归属 `agentId`，任务运行记录才有 `sessionId`。                                                        | `agentkit-mpa-agent/app/schemas/scheduled_task.py:299`、`:412`。                                                                                                                                   | 助手作用域应使用 `agentId`，不能把任务定义错误地按会话过滤。 | schema 复核。                                                |
| FACT-6  | 已验证事实 | 主题持久化在 `openmpa.db` 的 `display.theme`，原生 rail 读取该值。                                                | `desktop/internal/configstore/store.go:20`、`:242`；`desktop/internal/host/rail.go:454`。                                                                                                          | React 可以在挂载时拉取同一主题值。                           | 单元测试 `TestThemePreferencePersists`。                     |
| FACT-7  | 已验证事实 | React 已通过 `Host.Theme` 在挂载时拉取主题，同时保留 `themeChanged` 事件用于实时变化。                            | `desktop/internal/host/rail.go:454`；`src/client/desktop/DesktopAppState.tsx:113`；`src/client/desktop/DesktopAppState.tsx:230`。                                                                  | 首次加载不再依赖事件时序。                                   | `tests/desktop-navigation.test.tsx` 主题拉取用例。           |
| FACT-8  | 已验证事实 | Spaces 的 React 入口已移除，`Pages.tsx` 与 `desktop-page-review-latch.test.tsx` 已删除；工作区关闭刷新仍保留。    | `src/client/desktop/DesktopAppState.tsx:72` 调用 `workspace.flush()`；`desktop/internal/workspace/service.go:514`。                                                                                | 可以删除页面 UI，但不能删除关闭安全所需的 workspace flush。  | `rg DesktopPages` 为空；构建与测试。                         |
| FACT-9  | 已验证事实 | 原生侧边栏可以通过 MyGo headless tester 生成 PNG 预览。                                                           | MyGo `ui/headless.go:441` 的 `Tester.Image()`；OpenMPA `desktop/internal/host/rail_test.go:164`。                                                                                                  | 可以加入可重复的视觉烟雾测试。                               | `go test ./internal/host -run TestSidebarPreviewImages -v`。 |
| FACT-10 | 已验证事实 | 当前已安装的 OpenMPA 正在运行，临时开发实例会被单实例锁拦截。                                                     | 2026-10-27 `pgrep -af OpenMPA` 返回 PID 41293 与 93478；`/Applications/OpenMPA.app` 存在。                                                                                                         | 真实 GUI 验收需要关闭已安装应用或使用独立构建。              | 运行前检查与人工确认。                                       |
| FACT-11 | 已验证事实 | Runtime 目录只暴露 MPA 标签的 Runtime，agent 配置是当前单例。                                                     | `desktop/internal/cloud/client.go:334`；`agentkit-mpa-agent/app/services/agent_config_service.py:149`。                                                                                            | 本设计把 Runtime 建模为助手，Session 建模为会话。            | 目录过滤与 service 复核。                                    |
| FACT-12 | 已验证事实 | 知识库列表接口要求 `session_id`。                                                                                 | `agentkit-mpa-agent/app/api/v1/knowledge.py:228`。                                                                                                                                                 | 知识库必须绑定当前会话，不能做成全局目录。                   | 路由复核。                                                   |

### 2.3 假设与待确认项

| ID     | 类型     | 结论/假设                                                                             | 精确证据                                           | 设计影响                                        | 验证/关闭方式                            |
| ------ | -------- | ------------------------------------------------------------------------------------- | -------------------------------------------------- | ----------------------------------------------- | ---------------------------------------- |
| ASM-1  | 假设     | 添加 `SkillConfig` 引用后，Runtime 会在 apply 阶段验证 skill 是否存在于 skill space。 | 当前只有 `SkillConfig` 结构，没有 apply 结果证据。 | 客户端允许手动添加 skill ID，但不承诺立即可用。 | 在授权测试 Runtime 上读取 apply 结果。   |
| OPEN-1 | 待确认项 | AgentKit skill space 的目录 API 契约是什么。                                          | 当前无公开路由。                                   | 决定是否实现真正的筛选添加与导入。              | 后端提供 OpenAPI 或确认外部服务契约。    |
| OPEN-2 | 待确认项 | 是否彻底删除 Go workspace 的 page API。                                               | 当前 `workspace.Service` 仍服务关闭刷新。          | 决定 Spaces 后端清理范围。                      | 确认没有其他调用方后单独重构。           |
| OPEN-3 | 待确认项 | 真实 GUI 截图验收由谁在何时执行。                                                     | 已安装应用占用单实例锁。                           | 决定视觉验收的最终关闭方式。                    | 关闭应用后运行开发构建，或提供独立构建。 |

## 3. 目标、非目标与范围

### 3.1 目标

1. 助手配置在现有 API 范围内完整可用：AGENTS.md、已配置 skills 的筛选、添加、删除和保存。
2. 定时任务按当前助手跨页过滤，不再只过滤第一页，也不展示误导性的全量总数。
3. React 在挂载时从原生状态读取主题，主题变化事件仍可实时生效。
4. Spaces 的 React 死代码被删除，关闭刷新安全保持不变。
5. 原生侧边栏生成可重复的 PNG 预览，作为视觉烟雾证据。

### 3.2 非目标

- 不在本期新增 `agentkit-mpa-agent` 后端路由。
- 不伪造 skill space 导入，不把本地文件上传伪装成 skill space 添加。
- 不在未授权的真实 Runtime 上执行助手配置写入或定时任务写入。
- 不发布 DMG，不改变版本号。
- 不删除 Go workspace 的 page API，除非 OPEN-2 关闭。

### 3.3 范围

范围内：React Desktop 的助手设置、定时任务列表、主题同步、Spaces 死代码、原生侧边栏预览测试。

范围外：Web 版 Spaces、MPA Agent 后端、AgentKit 控制面、发布流水线、真实云资源变更。

## 4. 用户、调用方与场景

| 场景             | 角色     | Given                                     | When                                                 | Then                                                                     |
| ---------------- | -------- | ----------------------------------------- | ---------------------------------------------------- | ------------------------------------------------------------------------ |
| 配置助手         | 桌面用户 | 已连接 Runtime，进入 Settings > Assistant | 编辑 AGENTS.md，筛选并添加一个已配置 skill，点击保存 | 保存请求只提交当前 agent 的 `agentsMd` 与 skills；成功或失败都有明确提示 |
| 查看定时任务     | 桌面用户 | 已选择助手 A 的会话                       | 打开 Scheduled & activity                            | 列表只显示 agent A 的任务，跨页任务不丢失；总数不可用时明确提示          |
| 重启后主题       | 桌面用户 | 上次选择 Dark                             | 重新打开应用                                         | React 在首屏读取 Dark，不依赖一次性事件                                  |
| 无助手时新建会话 | 桌面用户 | 未选择助手或未连接 Runtime                | 点击 New conversation                                | 进入 Settings 的 Runtime 分区，不创建会话                                |
| skill space 浏览 | 桌面用户 | 进入 Assistant 设置                       | 期望浏览 skill space                                 | 界面明确说明当前客户端只显示已配置 skills，不展示假入口                  |
