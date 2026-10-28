# React Desktop 助手、定时任务与主题：方案设计

## 6. 方案设计

### 6.1 方案总览

本期采用“现有契约内收口”的方案。助手配置复用 `GET/PUT /api/v1/agents`，不新增 skill space 目录；定时任务在客户端读取有界页集后按 `agentId` 过滤；主题增加挂载时拉取 `Host.Theme`，保留事件用于实时变化；Spaces 只删除 React 死代码，保留关闭刷新；视觉验证增加 headless PNG 烟雾测试。

### 6.2 关键流程、状态与不变量

助手配置流程：

1. 进入 Settings > Assistant。
2. 若已连接 Runtime 且 `agent == nil`，调用 `Agents.Read`。
3. 读取成功后填充 `agentMD` 与 `agentSkills`。
4. 用户筛选只影响显示，不改变 `agentSkills` 原始集合。
5. 添加或删除只修改内存集合，保存时才调用 `Agents.Update`。
6. 保存成功后用服务端返回值覆盖本地状态。

不变量：`agentSkills` 的每个条目必须有非空 `id` 与 `name`；保存请求不包含未识别的字段；失败时保留用户编辑。

定时任务流程：

1. `loadTasks` 从当前选中会话读取 `agentId`。
2. `ListForAgent` 以 200 条为一页读取，最多 20 页。
3. 每页返回后只保留 `task.AgentID == agentId` 的任务。
4. 任一页失败立即返回错误，不返回部分结果。
5. 超过 1000 个匹配任务返回 `boundedResult`。
6. 返回合成 `TaskPage`，`CountAvailable=false`，不设置 `HasMore` 或 `NextOffset`。

主题流程：

1. React 挂载时调用 `Host.Theme`。
2. 成功则 `setThemePreference(theme)`。
3. 失败则静默保留当前主题。
4. 原生设置变化继续通过 `themeChanged` 事件推送。

Spaces 流程：

1. React 侧边栏不再发送 `pages` 或 `createSpace`。
2. `DesktopAppState` 不再加载 spaces。
3. `DesktopPages` 与对应测试删除。
4. `workspace.flush()` 保留在窗口关闭流程。
