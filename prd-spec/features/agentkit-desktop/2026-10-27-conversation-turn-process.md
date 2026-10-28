# React Desktop 对话执行过程与 Turn Trace 渐进设计

本文档按职责拆分为以下分卷：

- [用户与需求](./2026-10-27-conversation-turn-process-requirements.md)
- [方案设计](./2026-10-27-conversation-turn-process-design.md)
- [失败恢复与运维](./2026-10-27-conversation-turn-process-operations.md)
- [实施、验收与追踪](./2026-10-27-conversation-turn-process-verification.md)

## 1. 文档状态

| 字段         | 内容                                                                                                   |
| ------------ | ------------------------------------------------------------------------------------------------------ |
| 标题         | React Desktop 对话执行过程与 Turn Trace 渐进设计                                                       |
| 状态         | 已接受                                                                                                 |
| 创建日期     | 2026-10-27                                                                                             |
| 变更类型     | 功能                                                                                                   |
| 风险等级     | L3。变更跨越 Runtime SSE、Go 到 WebView 事件协议和 React 展示，并涉及 reasoning 与工具数据的安全处理。 |
| 关联任务     | 无                                                                                                     |
| 关联历史设计 | [React Desktop Turn Trace 入口与可视化设计](2026-10-27-react-desktop-turn-trace.md)                    |
| 版本关系     | 补充 React Desktop Turn Trace 设计。保留完整 Trace，新增对话内按 Turn 折叠的执行过程。                 |

结论：日常对话采用“答案是主视觉，执行过程按 Turn 内联且可折叠，完整 Trace 负责深度诊断”的产品结构。实现先修复事件数据丢失，再实现 Turn process、Markdown 和 Trace 联动。本文不授权代码实现、提交、发布或 Runtime 后端变更。

## 2. 背景、问题与证据

### 2.1 当前问题

OpenMPA React Desktop 已经能流式显示文本、把工具事件显示为单行，并提供独立的全屏 Trace 面板。用户在普通对话中仍看不到完整执行过程。当前 Go 投影主动跳过 `Thought`，工具调用和结果也只保留名称与 ID。React 因此无法显示 reasoning 摘要、工具输入、工具输出或一个工具从运行到完成的状态变化。

完整 Trace 不能填补这个缺口。普通会话的 Runtime Trace 可能返回 404，而对话 SSE 仍然可用。若执行过程只存在于 Trace 面板，用户在最常见的对话路径中无法判断系统正在做什么，也无法在失败时从当前工具步骤跳到诊断详情。

不处理的后果如下：

- 运行中的对话只显示最终文本或工具名称，用户无法判断进度。
- 工具调用与结果显示为两条无关联文本，失败和停止状态不清楚。
- 完整 Trace 与对话割裂。用户必须离开阅读路径才能排障。
- `react-markdown` 在每个流式更新中接收完整累计文本，但当前没有长回复基线来判断是否需要增量解析。

### 2.2 已验证事实

证据采集基于 `superops/react-composer-codex-style` worktree 的 2026-10-27 状态，以及 deepseek-harness 提交 `22ddee3792402c3e70ec5f86a63c3e3e895a2f60`。

| ID      | 类型       | 结论/假设                                                                                                                           | 精确证据                                                                                                                                                                                                                                                    | 设计影响                                                                    | 验证/关闭方式                                 |
| ------- | ---------- | ----------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------- | --------------------------------------------- |
| FACT-1  | 已验证事实 | Go SSE 投影识别 `Thought`，但直接跳过该 part。                                                                                      | `desktop/internal/session/events.go:23-35`、`:71-81`。                                                                                                                                                                                                      | 只改 React 不能恢复 reasoning。必须先改 Go 事件投影。                       | 源码复核。                                    |
| FACT-2  | 已验证事实 | Go 仅从 function call 与 response 中读取 `id` 和 `name`。`ClientEvent` 也没有参数、结果、part 序号或事件时间字段。                  | `desktop/internal/session/events.go:27-34`、`:83-95`；`desktop/internal/dto/types.go:238-255`。                                                                                                                                                             | 工具卡缺少形成摘要、详情和状态关联所需的数据。                              | 源码复核。                                    |
| FACT-3  | 已验证事实 | React 投影使用扁平 `DisplayMessage`，通过字符串 `kind` 分支。它能去重事件、合并流式文本，并在工具事件前后切断文本 slot。            | `src/client/desktop/projection.ts:6-38`、`:73-104`、`:109-180`；`tests/desktop-projection-dedupe.test.ts:43-174`。                                                                                                                                          | 新模型必须保留现有去重、流式替换和工具前后顺序语义。                        | 现有投影测试。                                |
| FACT-4  | 已验证事实 | `DesktopChat` 把 tool call 和 tool result 渲染成无详情的单行，把其他消息交给 `react-markdown + remark-gfm`。                        | `src/client/desktop/Chat.tsx:19-30`、`:230-271`；`package.json:58-64`。                                                                                                                                                                                     | 对话执行过程需要 typed Turn 和专用行组件。Markdown 可以先沿用现有依赖。     | 源码与依赖复核。                              |
| FACT-5  | 已验证事实 | 现有组件测试只要求工具事件是紧凑行，并证明 GFM 表格可渲染。                                                                         | `tests/desktop-chat-layout.test.tsx:148-230`。                                                                                                                                                                                                              | 新测试要扩展行为，不应丢失现有紧凑展示和 GFM 能力。                         | 现有组件测试。                                |
| FACT-6  | 已验证事实 | `TracePanel` 独立读取 Runtime Trace，默认跟随当前 invocation，并明确处理 404 与 401/403。                                           | `src/client/desktop/TracePanel.tsx:51-69`、`:72-175`、`:230-343`。                                                                                                                                                                                          | 对话内过程不能依赖 Trace 成功。联动必须使用精确 focus，并保留当前失败语义。 | 源码与 `tests/desktop-trace-panel.test.tsx`。 |
| FACT-7  | 已验证事实 | Trace 详情当前在 React 侧按敏感键、深度、字符串和数组长度过滤。                                                                     | `src/client/desktop/trace-model.ts:334-351`；`tests/trace-model.test.ts` 的 `formats durations and redacts sensitive detail keys`。                                                                                                                         | 对话事件详情不得比 Trace 更弱。安全处理应前移到 Go，React 保留防御。        | 源码与单元测试。                              |
| FACT-8  | 已验证事实 | Desktop 前端在非 MyGo 环境只显示占位页，不能直接用普通 Vite 页面完成真实视觉验收。                                                  | `src/client/desktop/main.tsx:12-29`。                                                                                                                                                                                                                       | Phase 0 需要独立 fixture。最终仍必须在真实 MyGo 应用验收。                  | 源码复核。                                    |
| FACT-9  | 已验证事实 | 当前聚焦基线包含 Trace、模型、流式和聊天布局 23 个测试，全部通过。                                                                  | `npx vitest run tests/desktop-trace-panel.test.tsx tests/trace-model.test.ts tests/desktop-stream.test.tsx tests/desktop-chat-layout.test.tsx`，2026-10-27 结果为 4 files、23 tests passed。                                                                | 后续阶段以该结果为回归基线，不把现存失败归因于新设计。                      | 重跑相同命令。                                |
| FACT-10 | 已验证事实 | deepseek-harness 使用 typed assistant blocks、独立 reasoning disclosure、关联工具树、共享 Conversation 投影和独立 Trajectory 视图。 | `packages/client/ui-chat/src/client/chat/AssistantMarkdown.tsx`、`ReasoningRow.tsx`、`packages/client/ui-tool/src/client/tool/ToolCallTree.tsx`、`packages/client/ui-trajectory/src/client/index.ts`，来源提交 `22ddee3792402c3e70ec5f86a63c3e3e895a2f60`。 | OpenMPA 吸收信息层级和数据建模，不移植其插件框架。                          | 固定提交源码复核。                            |
| FACT-11 | 已验证事实 | 用户在视觉对比中选择“对话内折叠过程 + 完整 Trace”，并批准数据模型、交互规则和五阶段验收方案。                                       | 本次 2026-10-27 设计评审记录。                                                                                                                                                                                                                              | 方案 B 是本设计的产品基线。                                                 | 用户批准。                                    |

### 2.3 假设与待确认项

| ID     | 类型     | 结论/假设                                                              | 精确证据                                                                   | 设计影响                                                         | 验证/关闭方式                                             |
| ------ | -------- | ---------------------------------------------------------------------- | -------------------------------------------------------------------------- | ---------------------------------------------------------------- | --------------------------------------------------------- |
| ASM-1  | 假设     | ADK 实际 SSE 的 function call 与 response 含有可解析的参数和结果字段。 | 当前 `adkEvent` 只声明 `id` 和 `name`，未保留原始 part。                   | 决定 schema v2 能提供哪些安全预览。                              | Phase 0 捕获脱敏 fixture，并与 Runtime schema 对照。      |
| OPEN-1 | 待确认项 | 哪一层为 reasoning 内容提供“可展示”授权。                              | 当前只有 `Thought bool`，没有可见性字段。                                  | 未关闭前，只能显示通用“思考中”状态，不能默认暴露原始 reasoning。 | Phase 1 前由 Runtime 合同或产品策略给出显式字段和默认值。 |
| OPEN-2 | 待确认项 | Session `toolId` 是否能稳定映射到 Runtime Trace step。                 | 当前 `TraceStepVM` 有 correlation 数据，但本文尚未验证所有工具类型的映射。 | 决定工具行能否精确聚焦 step，或只能打开 invocation。             | 用真实 Trace fixture 覆盖至少一个成功和一个失败工具调用。 |
| OPEN-3 | 待确认项 | 长回复与长执行记录的当前渲染基线和可接受预算。                         | 当前没有同一场景的分块渲染、滚动和输入响应测量。                           | 决定是否需要增量 Markdown 和虚拟列表。                           | Phase 0 记录基线，Phase 3 与 Phase 4 用相同场景复测。     |

## 3. 目标、非目标与范围

### 3.1 目标

1. 用户在对话中可以看到当前 Turn 的 reasoning 摘要、工具步骤、运行状态和最终结果，无需打开 Trace。
2. Turn 完成后，执行过程折叠为摘要，最终答案保持主视觉。
3. 工具调用与工具结果通过稳定 ID 原位更新，重连或历史回放不产生重复行。
4. 用户可以从工具行打开完整 Trace，并在有稳定关联时定位到对应 step。
5. Markdown 在流式、长内容、宽表格和窄窗口场景中保持可读，并用测量结果决定是否引入增量解析。

### 3.2 非目标

- 不移植 deepseek-harness 的 Cordis、slot、插件包或完整 Markdown 解析器。
- 不在本文新增普通 Session 的 Runtime Trace 路由或采集策略。
- 不把普通 Session 改成 `admin_debug`。
- 不在未确认可见性规则前展示原始 reasoning。
- 不在没有 Phase 0 基线时承诺具体帧率、耗时或 DOM 数量阈值。
- 不发布版本，不创建提交，不清理任何 worktree。

### 3.3 范围

范围内：

- Runtime SSE 到 Go `ClientEvent` 的展示事件合同。
- React Turn 投影、执行过程、工具行、reasoning 行和 Markdown 展示。
- 对话工具行到 `TracePanel` 的 focus 联动。
- 本地 fixture、单元测试、组件测试、构建检查和真实 MyGo 验收。

范围外：

- Runtime Trace 后端的数据采集与 owner 路由。
- AgentKit 模型或工具执行语义。
- 原生 MyGo chat 重写。
- 会话持久化格式和服务端历史格式迁移。
- 发布流水线、灰度环境和远端告警配置。
