# React Desktop 对话执行过程：失败恢复与运维

## 7. 边界、失败与恢复

| 场景                      | 系统行为                                                                      | 状态是否改变       | 可观察信号                   | 恢复方式                    |
| ------------------------- | ----------------------------------------------------------------------------- | ------------------ | ---------------------------- | --------------------------- |
| 事件没有 invocation       | 使用订阅请求 invocation；两者都缺失则拒绝形成 Turn                            | 不创建未知 Turn    | 安全错误或 fallback 事件     | 刷新权威历史                |
| 重复事件                  | reducer 忽略已经应用的事件身份                                                | 否                 | segment 数量不变             | 无需用户操作                |
| 同一工具重复 call         | 使用 `toolId` 更新既有 segment，不新增第二行                                  | 工具状态可更新     | 原行状态变化                 | 无需用户操作                |
| tool result 早于 call     | 创建待关联 segment，call 到达后补齐；terminal 前仍缺 call 时使用通用 fallback | 是                 | 显示结果状态与缺失摘要       | 重放历史可补齐              |
| terminal 到达时工具仍运行 | 将该工具收口为 stopped 或 failed，具体取决于 terminal reason                  | 是                 | Turn 摘要不再 running        | 用户可打开详情或 Trace      |
| reasoning 没有可见性      | 不显示文本，只显示允许的通用状态                                              | 否                 | DOM 无 reasoning 文本        | 关闭 OPEN-1 后启用内容      |
| 工具详情无法安全解析      | 只显示名称、状态和安全 summary                                                | 否                 | 详情按钮缺失或显示不可用说明 | 完整 Trace 仍按现有权限工作 |
| SSE 中断                  | 保留已投影 segment，按现有 cursor 恢复；不把中断当完成                        | 运行状态保持可验证 | 现有错误与恢复 UI            | 重连并读取权威历史          |
| Trace 404                 | 对话内过程不变，Trace 显示 admin_debug 限制                                   | 否                 | 现有 Trace error banner      | 等待后端路由，不伪造数据    |
| Trace step 无稳定关联     | 打开 invocation，不自动选择 step                                              | 否                 | Trace 默认当前 Turn          | 用户在 Trace 内搜索         |
| Markdown 单个 block 失败  | 该 block 降级为纯文本，其他 segment 继续渲染                                  | 否                 | 局部 fallback                | 用户仍可复制原文            |
| 用户向上阅读              | 停止自动跟随，新内容继续追加                                                  | 否                 | “回到最新”控制可见           | 用户回到底部后恢复跟随      |

## 8. 兼容、迁移与回滚

### 8.1 兼容窗口

- schema v1 与 v2 在客户端共存。v1 使用现有文本和工具摘要，v2 使用 typed segment。
- 新 Go 与旧 React 混跑时，旧 React 忽略新增可选字段。
- 新 React 与旧 Go 混跑时，新 React 按 v1 降级，不显示 reasoning 或结构化工具详情。
- 本设计不改变服务端历史持久化格式，也不回填历史数据。

### 8.2 迁移顺序

1. 增加 fixture 和 v1 回归基线。
2. 增加可选的 schema v2 字段，但保持 React 使用 v1。
3. 上线 typed reducer，并用 v1 fixture 证明降级路径。
4. 启用 Turn process 组件。
5. 启用 Markdown 和 Trace focus 增强。
6. 只有基线需要时才增加增量解析或虚拟列表。

每一步都必须保持 worktree 可构建。任一步失败时停止在上一阶段，不同时清理 v1 路径。

### 8.3 回滚

- Phase 1 可以停止发送 v2 可选字段，恢复 v1 展示。没有持久化数据需要恢复。
- Phase 2 到 Phase 4 可以关闭新 Turn process 渲染并恢复现有 `DisplayMessage` 视图。事件 DTO 的可选字段可以保留。
- 只有 v1 回归、真实 MyGo 会话和安全负控均通过后，才能考虑删除旧投影。删除旧投影不属于本文阶段。
- 回滚不得删除或重置用户现有 worktree 改动。

## 9. 安全、隐私与权限

信任边界位于 Runtime SSE 进入 Go 的位置。Runtime 数据是外部输入。Go 负责验证结构、限制大小，并形成 display-safe 事件。WebView 不接收 Runtime key、JWT、Authorization、cookie 或未经处理的工具 payload。

安全规则：

- reasoning 默认隐藏。只有显式可展示策略允许内容进入 WebView。
- 工具 presenter 使用 allowlisted 字段形成摘要。通用 presenter 不显示原始输入和输出。
- 复用并加强当前敏感键过滤。嵌套对象、数组、深度和字符串长度都受限。
- 字符串内容可能包含没有键名的秘密。没有上游 display-safe 标记时，Go 不把原始长文本当作安全详情。
- 外部链接只允许已验证的 HTTP(S) 目标。本地文件链接通过 Go 文件服务验证路径与权限。
- raw HTML 不进入 React DOM。
- 安全过滤失败时关闭详情，不开放原始数据。
- 日志只记录事件 kind、schema version、计数和脱敏身份。日志不记录 reasoning、参数、结果或原始 URL。

滥用负控至少覆盖：嵌套 token、Authorization、cookie、password、API key、超长字符串、深层对象、恶意链接和 raw HTML。

## 10. 可观测性、发布与运维

本文不授权发布。实施阶段需要产出以下本地与真实产品信号：

- 投影信号：schema version、输入事件数、去重数、fallback 数和未关联工具数。
- 用户结果：Turn process 是否完成收口，是否有永久 running segment，Trace focus 是否精确。
- 降级信号：v1 fallback、reasoning hidden、tool detail unavailable 和 Trace 404。
- 性能信号：固定 fixture 的分块渲染耗时、滚动位置变化、输入响应和挂载节点数量。

发布前置条件由未来发布任务决定。没有现有 dashboard、告警或灰度流程证据，因此本文不虚构远端阈值和 owner。
