# React Desktop 对话执行过程：方案设计

## 6. 方案设计

### 6.1 方案总览

Runtime SSE 先由 Go 规范化为版本化的展示事件。React reducer 再把事件投影为按 invocation 分组的 `ConversationTurn`。每个 Turn 持有有序的 text、reasoning、tool 和 status segment。对话渲染这份模型，完整 Trace 保持独立数据源，并只通过精确身份联动。

该方案不把 Trace 当作对话过程的数据源。Trace 返回 404 时，SSE 驱动的执行过程仍可用。

```text
Runtime SSE
    ↓
Go event projection and display-safe detail policy
    ↓ schema v1 or v2 ClientEvent
Turn reducer keyed by invocationId
    ├── inline TurnProcess
    │   ├── ReasoningRow
    │   ├── ToolRow
    │   └── RunStatusRow
    ├── AssistantMarkdown
    └── exact focus request to TracePanel
```

### 6.2 数据模型与所有权

逻辑模型如下。名称说明职责，不要求实现保留完全相同的文件布局。

```ts
interface ConversationTurn {
  invocationId: string;
  status: 'running' | 'completed' | 'failed' | 'stopped';
  startedAt?: string;
  endedAt?: string;
  segments: ConversationSegment[];
}

type ConversationSegment =
  TextSegment | ReasoningSegment | ToolSegment | StatusSegment;
```

不变量：

1. `invocationId` 是 Turn 身份。不得用文本或时间窗口合并 Turn。
2. 工具 segment 使用 `invocationId + toolId` 作为业务身份。tool result 更新同一 segment。
3. 文本与 reasoning 使用独立流式 slot。工具事件会关闭当前 slot，工具前后的文本不能合并。
4. 事件身份至少包含 invocation、event、part 和 kind。重复身份不得二次应用。
5. terminal 关闭所有开放 slot。未结束工具变为 `stopped` 或 `failed`，不能永久保持 `running`。
6. reducer 只处理展示事件。它不改变服务端 Session、工具执行或 Trace 数据。

### 6.3 Go 到 WebView 事件合同

schema v2 在现有 `ClientEvent` 上增加结构化展示字段。以下 JSON 只说明最小语义。字段名可以在实现评审中调整，但语义和默认行为不能弱化。

```json
{
  "schemaVersion": 2,
  "connectionEpoch": 3,
  "subscriptionEpoch": 9,
  "sessionId": "session-1",
  "invocationId": "invocation-1",
  "eventId": "event-12",
  "partIndex": 0,
  "kind": "toolCall",
  "completeness": "partial",
  "tool": {
    "id": "tool-7",
    "name": "read",
    "status": "running",
    "summary": "读取 Chat.tsx",
    "inputPreview": { "path": "src/client/desktop/Chat.tsx" }
  },
  "partial": true,
  "terminal": false
}
```

合同规则：

- v1 字段与语义保持可读。客户端按 `schemaVersion` 选择能力，不按字段猜测版本。
- 空 `invocationId` 继续使用订阅请求中的 invocation。
- `partIndex` 来自原始 part 顺序。缺失时按 v1 处理。
- reasoning 需要 `visibility`。缺失或未知值按 `hidden` 处理。
- `tool` 只携带展示安全的 summary 和 preview。Go 不把未经处理的原始 payload 发送给 WebView。
- 无法解析或无法安全预览的工具仍发送 name、id 和状态，不发送详情。
- 未识别 kind 进入 fallback segment。它不能导致整个 Turn 丢失。
- schema v2 的具体输入与输出大小限制在 Phase 0 后记录为独立决策。未确定前不得无限转发 payload。

### 6.4 React 组件边界

```text
DesktopChat
├── ConversationTurn
│   ├── UserMessage
│   ├── TurnProcess
│   │   ├── ReasoningRow
│   │   ├── ToolRow
│   │   └── RunStatusRow
│   └── AssistantMarkdown
└── Composer
```

职责：

- Turn reducer 负责事件归一化、去重、关联、顺序和终止收口。
- `ConversationTurn` 负责一个 invocation 的布局，不解析事件。
- `TurnProcess` 负责 disclosure 与步骤顺序，不读取 Runtime 服务。
- `ReasoningRow` 负责可见性、摘要和紧凑 Markdown。
- `ToolRow` 负责通用工具状态。专用 renderer 只把安全数据变成摘要或卡片。
- `AssistantMarkdown` 负责 Markdown 的安全渲染和流式策略。
- `TracePanel` 接受 `{ invocationId, toolId?, stepId? }` focus。未验证 `stepId` 时只聚焦 invocation。
- `DesktopChat` 组合组件和服务，不继续承担事件解析细节。

### 6.5 disclosure 与状态规则

| Turn 状态 | 过程组默认状态 | 标题                            | segment 详情                                                 |
| --------- | -------------- | ------------------------------- | ------------------------------------------------------------ |
| running   | 展开           | 正在执行 · N 个步骤             | 已完成步骤显示摘要，当前步骤显示运行状态，原始详情不自动展开 |
| completed | 折叠           | 执行了 N 步 · duration          | 用户展开后显示有序 segment                                   |
| failed    | 折叠           | 执行失败 · failed segment title | 错误行可见，安全详情按需展开                                 |
| stopped   | 折叠           | 已停止 · 完成 M/N 步            | 未完成工具显示 stopped                                       |

用户主动切换 disclosure 后，该 Turn 在当前挂载周期内保持用户选择。切换 Session 或重新加载后回到状态默认值。键盘触发不增加入场动画。频繁出现的步骤行只使用状态颜色、图标和短 opacity 变化。`prefers-reduced-motion` 下禁用非必要变化。

### 6.6 reasoning 与工具展示

reasoning 规则：

- `visibility=hidden` 时不保留或渲染文本。
- `visibility=summary` 时只渲染服务端提供或 Go 派生的安全摘要。
- `visibility=detail` 时，流式阶段显示最近一个完整段落的首行，完成后显示第一行摘要。用户展开后渲染完整的可展示内容。
- 半句话不作为流式摘要，避免每个 chunk 改变 collapsed row 宽度。

工具规则：

- `read`、`shell`、`search` 和文件写入类工具可以注册专用 presenter。
- 专用 presenter 只接收 `ToolDisplayDetail`。它不能访问原始 Runtime payload。
- 通用 presenter 显示工具名、状态和已有安全 summary。没有 summary 时不显示 payload。
- error 和 stopped 状态始终使用文字与图标，不只依赖颜色。
- 展开详情与“在完整 Trace 中查看”是独立操作。点击安全链接不能切换 disclosure。

### 6.7 Markdown 策略

Phase 3 继续使用 `react-markdown + remark-gfm`，补齐以下展示：

- 代码块的语言标签、复制、换行和水平滚动。
- 宽表格的容器滚动与窄窗口降级。
- 标题、列表、引用、任务列表、行内代码和 CJK 粗体。
- 长路径与无空格字符串的换行。
- HTTP(S) 链接的安全打开。
- 经过 Go 验证的本地文件与附件链接。
- raw HTML 不作为 DOM 注入。

先测量现有完整重解析。如果长回复场景不满足 Phase 0 形成的预算，再设计“冻结稳定 block，只重解析尾部”的增量方案。增量解析是后续独立设计，不是本设计的默认实现。

### 6.8 设计决策

| ID    | 决策                                                                            | 依据                              | 影响范围                        | 重新评估条件                                     |
| ----- | ------------------------------------------------------------------------------- | --------------------------------- | ------------------------------- | ------------------------------------------------ |
| DEC-1 | 采用对话内折叠过程与独立完整 Trace。                                            | FACT-6、FACT-11。                 | 对话信息层级、Trace 联动。      | 用户研究证明常驻 Inspector 更适合主要用户。      |
| DEC-2 | 使用 typed Turn 与 segment reducer，不继续扩展扁平 `DisplayMessage.kind` 分支。 | FACT-1、FACT-2、FACT-3。          | Go 事件合同、React 状态和测试。 | 上游提供一份更稳定且覆盖相同不变量的 Turn 模型。 |
| DEC-3 | 不移植 deepseek-harness 插件框架。                                              | FACT-10、OpenMPA 当前规模和依赖。 | 组件边界与依赖。                | OpenMPA 出现多个可独立安装的对话展示插件。       |
| DEC-4 | 工具详情在 Go 侧形成 display-safe 数据，React 只消费安全预览。                  | FACT-7、L3 安全边界。             | DTO、事件投影、工具 presenter。 | 上游协议提供等价的强类型安全展示合同。           |
| DEC-5 | 先保留 `react-markdown`，用基线决定增量解析。                                   | FACT-4、OPEN-3。                  | Markdown 与性能工作。           | 基线证明完整重解析阻塞输入或造成可观察抖动。     |

### 6.9 备选方案与取舍

| 方案                             | 用户阅读                     | 诊断能力                         | 实施与兼容                           | 结论   |
| -------------------------------- | ---------------------------- | -------------------------------- | ------------------------------------ | ------ |
| A. 对话与 Trace 完全分离         | 对话最安静                   | 日常过程不可见，失败必须离开对话 | 最接近现状，价值不足                 | 不采用 |
| B. 对话内折叠过程 + 完整 Trace   | 答案保持主视觉，过程按需展开 | 可从当前步骤进入深度诊断         | 需要事件合同与 typed Turn            | 采用   |
| C. 常驻双栏 Inspector            | 诊断信息始终可见             | 最强                             | 压缩 Markdown 与表格，普通用户负担高 | 不采用 |
| D. 移植 deepseek-harness UI 插件 | 能获得相近能力               | 强                               | 引入不匹配的插件、状态和依赖结构     | 不采用 |
