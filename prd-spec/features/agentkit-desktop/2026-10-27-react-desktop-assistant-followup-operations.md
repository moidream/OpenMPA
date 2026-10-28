# React Desktop 助手、定时任务与主题：收口实施与验收

## 7. 边界、失败与恢复

| 场景                                  | 系统行为                        | 状态是否改变 | 可观察信号   | 恢复方式               |
| ------------------------------------- | ------------------------------- | ------------ | ------------ | ---------------------- |
| 未选择助手就打开 Scheduled & activity | `loadTasks` 返回 `invalidInput` | 否           | 页面错误提示 | 先在侧边栏选择助手     |
| 定时任务第 2 页失败                   | 返回该页错误，不返回部分结果    | 否           | 错误提示     | 重试或缩小查询         |
| 匹配任务超过 1000                     | 返回 `boundedResult`            | 否           | 错误提示     | 缩小查询或等待后端过滤 |
| 助手配置保存 401/403                  | 保留本地编辑                    | 否           | 错误提示     | 重新认证后重试         |
| 助手配置保存 412                      | 保留本地编辑                    | 否           | 错误提示     | 重新读取后再保存       |
| skill 重复添加                        | 返回 `conflict`，不发网络请求   | 否           | 错误提示     | 修改 ID                |
| 主题拉取失败                          | 静默保留当前主题                | 否           | 无阻塞       | 下次启动重试           |
| 主题事件丢失                          | 挂载拉取仍能恢复                | 否           | 主题最终一致 | 重新挂载               |
| 已删除的 `pages` 导航仍被旧调用方发送 | `validShellState` 拒绝 `pages`  | 否           | 无效导航错误 | 升级调用方             |

## 8. 兼容、迁移与回滚

兼容：

- `display.theme` 是新增键，旧数据库默认 `system`，不需要迁移脚本。
- `ShellState` 新增字段均为可选；旧 WebView 不发送时保留原生状态。
- `Host.Theme` 是新增绑定，旧 WebView 不调用不影响运行。
- 删除 `Pages.tsx` 只影响 React Desktop 内部引用；Web 版 `SpaceWorkspace` 不受影响。

迁移：

1. 部署包含 `Host.Theme` 与助手设置的构建。
2. 首次启动读取主题并写入 React localStorage。
3. 旧 Spaces 导航不再出现；旧 `pages` 视图被拒绝。

回滚：

- 代码回滚即可恢复旧侧边栏与设置页。
- `display.theme` 键可保留，不影响旧版本。
- 已删除的 `Pages.tsx` 通过 Git 回滚恢复。
- 定时任务过滤回滚会恢复当前页过滤，属于已知退化，需在回滚说明中标注。

## 9. 安全、隐私与权限

- 信任边界：Runtime 连接层继续负责认证与 scope 校验；客户端不自行推断权限。
- 助手配置写入只发送 `agentsMd` 与 skills，不发送 AK/SK、Runtime key、JWT。
- skill `id` 与 `name` 在客户端做长度和换行校验；服务端仍是最终校验者。
- 定时任务 `agentId` 来自当前会话摘要；客户端只按该值过滤。服务端仍返回当前用户可见的任务集合，不新增跨用户读取。
- 主题不含敏感信息，只存 `system`、`light`、`dark`。
- 日志与错误提示不包含完整 URL、query、token 或 skill 内容。

## 10. 可观测性、发布与运维

本地桌面应用，不新增远端指标。

成功信号：

- 助手保存成功后显示 “Assistant configuration saved.”。
- 主题保存成功后显示 “Display theme updated.”。
- 定时任务列表显示 “Scheduled tasks for this assistant.”。

失败信号：

- 助手读取/保存失败显示错误横幅或页面错误。
- 定时任务读取失败显示页面错误。
- 主题拉取失败静默，不阻塞首屏。

发布检查：`go test ./...`、`go vet ./...`、`npx tsc --noEmit`、`npx vitest run`、`npm run build:desktop:frontend`、`git diff --check`。

观察窗口：首次启动后检查主题、助手设置、定时任务三项主路径。无远端告警入口，属于本地应用。

## 11. 实施计划与影响范围

| 阶段 | 变更边界               | 前置条件                    | 完成证据                                                   | 安全停止点                           |
| ---- | ---------------------- | --------------------------- | ---------------------------------------------------------- | ------------------------------------ |
| 1    | 编写本 PRD             | 现有事实已复核              | 本文档                                                     | 文档评审                             |
| 2    | 助手设置与 agents 服务 | 现有 `/api/v1/agents` 契约  | `go test ./internal/agents ./internal/nativeui`            | 若接口形状不符，停止并记录           |
| 3    | 定时任务跨页过滤       | 现有分页契约                | `go test ./internal/schedule -run TestListForAgent`        | 超过 1000 条返回错误，不返回部分结果 |
| 4    | 主题拉取绑定           | `Host.Theme` 生成           | `npx vitest run tests/desktop-navigation.test.tsx`         | 拉取失败不阻塞首屏                   |
| 5    | Spaces 死代码删除      | `rg DesktopPages` 为空      | `npx vitest run`、构建通过                                 | 构建失败则恢复文件                   |
| 6    | 原生 PNG 烟雾测试      | `ui.NewTester.Image()` 可用 | `go test ./internal/host -run TestSidebarPreviewImages -v` | PNG 为空或尺寸错误则失败             |
| 7    | 全量门禁               | 前六阶段通过                | `go test ./...`、`go vet ./...`、`tsc`、`vitest`、`build`  | 任一失败停止                         |

影响范围：

- `desktop/internal/agents`、`desktop/internal/nativeui/settings_assistant.go`、`desktop/internal/nativeui/settings.go`
- `desktop/internal/schedule`、`desktop/internal/nativeui/schedules.go`
- `desktop/internal/host/rail.go`、`desktop/internal/configstore`
- `src/client/desktop/DesktopAppState.tsx`、`src/client/desktop/models.ts`、`src/client/desktop/services.ts`
- `src/client/desktop/Pages.tsx` 与 `tests/desktop-page-review-latch.test.tsx` 删除
- `desktop/internal/host/rail_test.go`、`desktop/internal/schedule/agent_list_test.go`

## 12. 验收标准与验证计划

| ID   | Given                                    | When                                | Then                                                                | 验证方法/命令                                                                                                                                                                                                                                                                                         | 环境 |
| ---- | ---------------------------------------- | ----------------------------------- | ------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---- |
| AC-1 | 已连接 Runtime，agent 已读取             | 筛选 skills、添加 skill、删除 skill | 显示集合正确；保存请求只含 `agentsMd` 与 skills                     | `go test ./internal/nativeui -run TestAssistantSettings`                                                                                                                                                                                                                                              | 本地 |
| AC-2 | Assistant 设置已打开                     | 查看页面                            | 出现 “Configured skills only” 说明；没有 skill space 导入按钮或调用 | `rg "Configured skills only" desktop/internal/nativeui/settings_assistant.go`；`rg "Import skill" src/client/desktop desktop/internal/nativeui` 无匹配；`rg "skillSpace" src/client/desktop desktop/internal/nativeui` 无匹配；`rg "skill_space" src/client/desktop desktop/internal/nativeui` 无匹配 | 本地 |
| AC-3 | 服务端两页任务，agent A 分布在第 1、2 页 | 加载 agent A 的定时任务             | 两页中 agent A 任务全部出现；`CountAvailable=false`                 | `go test ./internal/schedule -run TestListForAgentReadsEveryPageBeforeFiltering`                                                                                                                                                                                                                      | 本地 |
| AC-4 | `Host.Theme` 返回 Dark                   | React 首次挂载                      | `themePreference()` 为 Dark                                         | `npx vitest run tests/desktop-navigation.test.tsx`                                                                                                                                                                                                                                                    | 本地 |
| AC-5 | `Pages.tsx` 已删除                       | 搜索生产引用                        | `rg DesktopPages src tests` 为空；构建通过                          | `rg DesktopPages src tests`；`npm run build:desktop:frontend`                                                                                                                                                                                                                                         | 本地 |
| AC-6 | 侧边栏 chat 与 settings 状态             | 运行预览测试                        | 两张 PNG 尺寸 240x700，且大于 1000 字节                             | `go test ./internal/host -run TestSidebarPreviewImages -v`                                                                                                                                                                                                                                            | 本地 |
| AC-7 | 全量代码                                 | 运行门禁                            | Go、TS、Vitest、构建全部通过                                        | `go test ./...`；`go vet ./...`；`npx tsc --noEmit`；`npx vitest run`；`npm run build:desktop:frontend`                                                                                                                                                                                               | 本地 |
| AC-8 | 已安装 OpenMPA 正在运行                  | 尝试启动开发实例                    | 明确记录被单实例锁拦截，不用假截图冒充验收                          | `pgrep -af OpenMPA`                                                                                                                                                                                                                                                                                   | 本机 |

## 13. 风险与待确认项

| ID     | 触发条件                           | 概率/影响 | 缓解措施                                     | 监测信号              | 兜底/回滚              |
| ------ | ---------------------------------- | --------- | -------------------------------------------- | --------------------- | ---------------------- |
| RISK-1 | 用户期望从 skill space 筛选导入    | 高/中     | 页面明确说明只支持已配置 skills；保留 OPEN-1 | 用户反馈；OPEN-1 状态 | 后端契约可用后新增 FR  |
| RISK-2 | 任务总数很大，跨页读取超过 1000    | 中/中     | 有界读取，超限返回 `boundedResult`           | 错误提示              | 缩小查询或等待后端过滤 |
| RISK-3 | 主题拉取与事件同时到达             | 低/低     | 拉取先执行，事件后覆盖；两者都写同一偏好     | 主题测试              | 重新挂载               |
| RISK-4 | 删除 `Pages.tsx` 影响未知调用方    | 低/中     | `rg DesktopPages` 与构建检查                 | 构建失败              | Git 回滚文件           |
| RISK-5 | PNG 烟雾测试不能证明真实字体与合成 | 中/中     | 保留人工 GUI 验收 AC-8                       | 人工验收记录          | 关闭应用后运行开发构建 |

| ID     | 问题                               | 为什么重要                       | 影响的决策      | 确认方式                        | 最晚确认阶段            |
| ------ | ---------------------------------- | -------------------------------- | --------------- | ------------------------------- | ----------------------- |
| OPEN-1 | AgentKit skill space 目录 API 契约 | 决定能否实现真正的筛选添加与导入 | FR-2、DEC-3     | 后端提供 OpenAPI 或确认外部服务 | 下一轮 skill 导入实现前 |
| OPEN-2 | 是否删除 Go workspace page API     | 决定 Spaces 后端清理范围         | FR-5 的后续阶段 | 确认无其他调用方                | 下一次 workspace 重构前 |
| OPEN-3 | 真实 GUI 截图验收执行人            | 决定视觉验收能否关闭             | AC-8            | 关闭已安装应用或提供独立构建    | 发布前                  |

## 14. 追踪矩阵

| 需求  | 方案位置/DEC             | 验收标准   | 验证方法                                                 | 覆盖状态 |
| ----- | ------------------------ | ---------- | -------------------------------------------------------- | -------- |
| FR-1  | 6.2 助手配置流程 / DEC-1 | AC-1       | `go test ./internal/nativeui -run TestAssistantSettings` | 已覆盖   |
| FR-2  | 6.2 助手配置流程 / DEC-3 | AC-2       | 文本与代码搜索                                           | 已覆盖   |
| FR-3  | 6.2 定时任务流程 / DEC-2 | AC-3       | `go test ./internal/schedule -run TestListForAgent`      | 已覆盖   |
| FR-4  | 6.2 主题流程 / DEC-4     | AC-4       | `npx vitest run tests/desktop-navigation.test.tsx`       | 已覆盖   |
| FR-5  | 6.2 Spaces 流程 / DEC-5  | AC-5       | `rg DesktopPages` 与构建                                 | 已覆盖   |
| FR-6  | 6.2 视觉烟雾 / DEC-5     | AC-6       | `go test ./internal/host -run TestSidebarPreviewImages`  | 已覆盖   |
| FR-7  | 6.2 助手选择             | AC-6、AC-7 | 侧边栏测试与构建                                         | 已覆盖   |
| FR-8  | 6.2 知识作用域           | AC-7       | `go test ./internal/nativeui`                            | 已覆盖   |
| NFR-1 | 6.2 定时任务流程 / DEC-2 | AC-3       | `go test ./internal/schedule`                            | 已覆盖   |
| NFR-2 | 6.2 主题流程 / DEC-4     | AC-4       | `npx vitest run tests/desktop-navigation.test.tsx`       | 已覆盖   |
| NFR-3 | 9 安全                   | AC-7       | 现有 secret gate 与代码审查                              | 已覆盖   |
| NFR-4 | 11 实施计划              | AC-7       | 全量门禁                                                 | 已覆盖   |
