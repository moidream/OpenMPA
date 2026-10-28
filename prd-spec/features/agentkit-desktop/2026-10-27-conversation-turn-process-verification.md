# React Desktop 对话执行过程：实施、验收与追踪

## 11. 实施计划与影响范围

实施目标为当前隔离的 `<repo-worktree>`。主 checkout 与其他 worktree 不在写入范围。

| 阶段                       | 变更边界                                          | 前置条件                          | 完成证据                                                                  | 安全停止点                             |
| -------------------------- | ------------------------------------------------- | --------------------------------- | ------------------------------------------------------------------------- | -------------------------------------- |
| Phase 0：基线              | 事件 fixture、DesktopChat fixture、当前性能与截图 | 已批准本文                        | 当前缺口可重复，23 个聚焦测试保持通过，fixture 能展示运行、完成和失败状态 | 没有稳定 fixture 时不改协议            |
| Phase 1：事件合同 v2       | Go SSE 投影、DTO、生成绑定、安全预览              | 关闭 OPEN-1，捕获真实工具 fixture | Go 测试覆盖 v1、v2、重复、乱序、敏感数据和 terminal 收口                  | 任一安全负控失败时不把详情送入 WebView |
| Phase 2：Turn process      | typed reducer、Turn process、reasoning 与工具行   | Phase 1 通过                      | reducer 与组件用例覆盖运行、成功、失败、停止和回放                        | v1 降级失败时不切换主渲染              |
| Phase 3：Markdown 与工具卡 | `AssistantMarkdown`、专用 presenter、滚动行为     | Phase 2 通过，OPEN-3 有基线       | Markdown 边界、窄屏和滚动用例通过；相同基线复测                           | 没有测量证据时不引入增量解析           |
| Phase 4：Trace 联动        | focus 合同、404 降级、大记录策略                  | 关闭 OPEN-2，Phase 3 通过         | 工具行能精确或安全降级地打开 Trace；大记录策略有测量依据                  | 关联不稳定时只打开 invocation          |
| Phase 5：真实验收          | 全量门禁、MyGo 真实会话、截图和交接               | Phase 0 到 Phase 4 稳定           | 真实运行、完成、失败、长 Markdown 和 Trace 联动证据齐全                   | 真实产品未验证时不得宣称完成           |

确认影响范围：

- `desktop/internal/session` 的 SSE 投影与测试。
- `desktop/internal/dto` 的展示事件 DTO。
- MyGo 生成绑定和 React service adapter。
- React 对话投影、状态、组件、样式和本地化。
- Trace focus 输入与相关测试。
- DesktopChat fixture、组件测试、性能场景和设计文档。

不确定的精确文件布局由实现计划在读取当前 diff 后决定。本文不要求复制 deepseek-harness 的文件结构。

## 12. 验收标准与验证计划

| ID    | Given                                                                   | When                                     | Then                                                                      | 验证方法/命令                                                                                                                        | 环境        |
| ----- | ----------------------------------------------------------------------- | ---------------------------------------- | ------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ | ----------- |
| AC-1  | fixture 包含 text、可展示 reasoning、tool call、tool result 和 terminal | 事件依次进入 Go 与 React 投影            | 一个 Turn 按原顺序包含对应 segment，工具调用与结果占同一行                | Go `Project` 测试、Turn reducer 单元测试                                                                                             | 本地        |
| AC-2  | 相同 fixture 被完整重放                                                 | reducer 再次消费事件                     | segment 数量、身份、顺序和内容不变                                        | 扩展 `tests/desktop-projection-dedupe.test.ts` 或新 Turn reducer 测试                                                                | 本地        |
| AC-3  | Turn 从 running 进入 completed、failed 或 stopped                       | terminal 到达                            | 过程组标题和默认 disclosure 符合 6.5 状态表，没有永久 running 工具        | React 组件测试                                                                                                                       | 本地        |
| AC-4  | reasoning 分别标记 hidden、summary 和 detail                            | 渲染 Turn                                | hidden 文本不在 DOM；summary 只有摘要；detail 可展开完整可展示内容        | reasoning 组件与安全负控                                                                                                             | 本地        |
| AC-5  | 工具详情含敏感键、深层对象、超长字符串和恶意链接                        | Go 形成 display-safe 事件并由 React 渲染 | 敏感值和未授权链接不在 DOM，工具名称与状态仍可见                          | Go sanitizer 测试、React 负控                                                                                                        | 本地        |
| AC-6  | schema v1 历史包含文本和工具摘要                                        | 新客户端读取历史                         | 最终答案和工具摘要仍显示，不要求 v2 字段                                  | v1 fixture 回归测试                                                                                                                  | 本地        |
| AC-7  | 助手消息含 GFM 表格、代码围栏、任务列表、CJK 粗体、长路径和 raw HTML    | Markdown 渲染                            | 语义元素可读，宽内容不撑破列，raw HTML 不注入 DOM                         | Vitest 组件测试与 DesktopChat fixture 截图                                                                                           | 本地浏览器  |
| AC-8  | 用户向上滚动离开底部，助手继续流式输出                                  | 新 chunk 到达                            | 当前滚动位置保持，回到最新控制可见                                        | 浏览器交互测试                                                                                                                       | 本地浏览器  |
| AC-9  | 工具 segment 有已验证 Trace step 关联                                   | 用户选择“在完整 Trace 中查看”            | `TracePanel` 打开同一 invocation 并选中该 step                            | TracePanel 组件测试与真实 Trace fixture                                                                                              | 本地和 MyGo |
| AC-10 | Trace 对普通 Session 返回 404                                           | 用户打开 Trace                           | 对话内 segment 保持不变，Trace 显示 admin_debug 限制                      | `tests/desktop-trace-panel.test.tsx` 加联合断言                                                                                      | 本地        |
| AC-11 | 已完成 Phase 0 到 Phase 4                                               | 运行仓库门禁                             | 聚焦测试、TypeScript、Desktop 前端构建、Go Desktop 测试和 diff check 通过 | `npx vitest run <affected tests>`；`npm run typecheck`；`npm run build:desktop:frontend`；`npm run desktop:test`；`git diff --check` | 本地        |
| AC-12 | MyGo 连接真实 Runtime，并准备运行、成功、失败和长 Markdown 会话         | 操作真实 Desktop 应用                    | 执行顺序、折叠、工具状态、滚动和 Trace 联动与本文一致                     | `npm run desktop:dev` 或等价开发构建，保存步骤与截图                                                                                 | macOS MyGo  |
| AC-13 | Phase 0 已记录长回复与长执行基线                                        | Phase 3 与 Phase 4 使用相同输入复测      | 报告给出可复查的前后数据，并据此决定增量解析与虚拟列表                    | 固定浏览器场景与结果文件                                                                                                             | 本地浏览器  |

## 13. 风险与待确认项

### 13.1 风险

| ID     | 触发条件                                   | 概率/影响 | 缓解措施                                                         | 监测信号                                      | 兜底/回滚                           |
| ------ | ------------------------------------------ | --------- | ---------------------------------------------------------------- | --------------------------------------------- | ----------------------------------- |
| RISK-1 | reasoning 内容没有明确可见性却进入 WebView | 中/高     | 默认 hidden，Phase 1 前关闭 OPEN-1，安全负控检查 DOM             | hidden reasoning DOM 命中                     | 禁用 reasoning 文本，仅保留通用状态 |
| RISK-2 | 工具输入或输出包含没有敏感键名的秘密       | 中/高     | Go 只发送 allowlisted display detail，通用工具不发送原始 payload | sanitizer fallback 和 detail unavailable 计数 | 关闭详情，保留名称与状态            |
| RISK-3 | v2 reducer 改变现有流式合并或顺序          | 中/高     | 固定 v1 与 v2 fixture，先保留双路径，重放测试                    | 重复 segment、顺序断言失败                    | 切回 v1 视图，不删除 v2 可选字段    |
| RISK-4 | 完整累计 Markdown 重解析造成长回复卡顿     | 中/中     | Phase 0 基线，隔离 Markdown 渲染，测量后才引入尾部增量解析       | 分块渲染、输入响应和滚动基线回退              | 限制重渲染范围，另立增量设计        |
| RISK-5 | Turn process 在长工具链中占据过多空间      | 中/中     | 完成后折叠，运行中只展开当前状态，详情按需显示                   | fixture 截图与实际反馈                        | 收紧摘要，保留完整 Trace            |
| RISK-6 | `toolId` 与 Trace step 不能稳定关联        | 中/中     | 关闭 OPEN-2，不猜测 step                                         | focus 只有 invocation                         | 降级为 invocation 级 Trace          |
| RISK-7 | 现有 dirty worktree 中的用户改动被覆盖     | 低/高     | 只在目标 worktree 工作，编辑前复核当前 diff，不重置、不清理      | `git status --short` 与文件 diff              | 停止修改并报告重叠                  |

### 13.2 待确认项

| ID     | 问题                                                           | 为什么重要                     | 影响的决策             | 确认方式                               | 最晚确认阶段   |
| ------ | -------------------------------------------------------------- | ------------------------------ | ---------------------- | -------------------------------------- | -------------- |
| OPEN-1 | reasoning 可见性由 Runtime、AgentKit 还是 OpenMPA 产品策略决定 | 决定内容是否可以进入 WebView   | FR-3、6.3、6.6、RISK-1 | 获取权威合同或产品决策，并加入 fixture | Phase 1 开始前 |
| OPEN-2 | 哪些 Trace step 稳定携带 Session `toolId` 关联                 | 决定是否可以精确聚焦           | FR-6、6.4、RISK-6      | 真实成功与失败 Trace fixture           | Phase 4 开始前 |
| OPEN-3 | 长回复和长执行的基线与预算是什么                               | 决定是否需要增量解析或虚拟列表 | NFR-4、6.7、RISK-4     | Phase 0 固定场景测量并记录结果         | Phase 3 开始前 |

## 14. 追踪矩阵

| 需求  | 方案位置/DEC           | 验收标准          | 验证方法                                 | 覆盖状态 |
| ----- | ---------------------- | ----------------- | ---------------------------------------- | -------- |
| FR-1  | 6.1、6.2、DEC-1        | AC-1、AC-3、AC-12 | reducer、组件和 MyGo 验收                | 已覆盖   |
| FR-2  | 6.5                    | AC-3、AC-12       | 组件状态表与真实会话                     | 已覆盖   |
| FR-3  | 6.3、6.6               | AC-4、AC-12       | reasoning 组件与可见性负控               | 已覆盖   |
| FR-4  | 6.2、6.6、DEC-2、DEC-4 | AC-1、AC-2、AC-5  | Go 投影、reducer 和安全测试              | 已覆盖   |
| FR-5  | 6.7、DEC-5             | AC-7、AC-8、AC-13 | Markdown、浏览器交互和基线复测           | 已覆盖   |
| FR-6  | 6.4、DEC-1             | AC-9              | TracePanel 组件与真实 fixture            | 已覆盖   |
| FR-7  | 6.1、7                 | AC-10、AC-12      | 404 联合测试与 MyGo 验收                 | 已覆盖   |
| FR-8  | 6.3、8.1               | AC-6、AC-11       | v1 fixture 与仓库门禁                    | 已覆盖   |
| NFR-1 | 6.3、6.6、9、DEC-4     | AC-4、AC-5        | Go 与 React 安全负控                     | 已覆盖   |
| NFR-2 | 6.2、7                 | AC-1、AC-2、AC-3  | reducer 幂等、顺序和 terminal 测试       | 已覆盖   |
| NFR-3 | 6.5、6.6               | AC-3、AC-4、AC-7  | 组件可访问性、键盘和 reduced-motion 断言 | 已覆盖   |
| NFR-4 | 6.7、10、DEC-5         | AC-8、AC-13       | 固定浏览器基线与复测                     | 已覆盖   |
| NFR-5 | 6.7、DEC-3、DEC-5      | AC-11             | 依赖 diff、构建和审查                    | 已覆盖   |
| NFR-6 | 11、Phase 5            | AC-12             | 真实 MyGo 会话与截图                     | 已覆盖   |

## 15. 文档完成边界

本文已记录用户批准的产品方向、事件与状态模型、交互规则、兼容、安全、阶段、风险和验收。OPEN-1 到 OPEN-3 有明确关闭阶段，不阻止 Phase 0。实现必须从 Phase 0 开始，并在每个安全停止点保留可构建状态。

本文没有开始产品实现，没有创建提交，没有发布，也没有改变 worktree 注册状态或其他 checkout。
