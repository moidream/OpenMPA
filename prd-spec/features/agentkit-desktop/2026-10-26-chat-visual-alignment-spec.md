# 对话页视觉对齐与重复输出修复

状态：已实现，待实机目视确认；日期：2026-10-26；类型：桌面 UI 修复 + 消息投影正确性，L2；基线为当前工作区（含 Region/SQLite hotfix 未提交改动）。

采用方向：可视化对比中的 **A · Quiet Canvas**。本 spec 只覆盖“最基础的对话页可用性与整体视觉对齐”，不包含助手设置（AGENTS.md / Skill）和页面编辑器。

## 1. 问题与证据

| ID  | 现象                                               | 代码证据                                                                                                                                                                              | 判断                                       |
| --- | -------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------ |
| F1  | 输入框与发送键错位、输入区大小不稳                 | `src/client/desktop/Chat.tsx:306-340` 把 textarea 与 send 放在同一 flex 行；`src/client/style.css:2274-2300` textarea 无边框、send `margin-bottom:5px`，textarea 增高时两者基线不一致 | 已由源码确认，属于布局缺陷                 |
| F2  | 输入区没有清晰的容器边界，和页面其他表单风格不一致 | `.chat-composer` 只有顶部分隔线；`.desktop-form input` 有 1px 边框和 10/12 padding，两套输入语言并存                                                                                  | 已由源码确认                               |
| F3  | 对话页缺少统一的内容宽度与排版基线                 | `.live-chat` max-width 850、padding 20/40；`.chat-bubble` font-size 12、line-height 1.95；assistant 颜色 `#717787` 偏灰                                                               | 已由源码确认，长文密度与页面其他区域不一致 |
| F4  | 同一条用户消息出现两次                             | `ChatState.tsx:151-171` 提交成功后先本地插入 user 消息；SSE 的 user author 事件又经 `projectEvent` 追加一次                                                                           | 机制假设，需用 fixture 复现后确认          |
| F5  | 助手回复被拆成多段重复输出                         | `projection.ts:63-99` partial 消息 id 依赖 `blockId`；ADK 每个 chunk 的 `wire.ID` 不同时会产生多个独立气泡                                                                            | 机制假设，需用 fixture 复现后确认          |
| F6  | 运行中/停止确认/未知结果提示叠加，页面跳动         | `Chat.tsx:176-296` 多个 `document-save-notice` 与 `chat-error` 同时渲染                                                                                                               | 已由源码确认                               |

F4/F5 在实现前必须先写失败用例复现；若复现结果与假设不同，以运行时证据为准调整修复，不按假设硬改。

## 2. 目标与非目标

目标：

- 对话页在 1280、900、720 三种窗口宽度下保持同一阅读列、同一输入容器、同一发送键位置。
- 输入框支持 1–6 行自动增高，发送键固定在底栏右侧，不随文本行数漂移。
- 消息以服务端历史为唯一权威，流式过程不重复渲染同一条用户或助手内容。
- 状态提示收敛为一条，不叠加、不造成输入区跳动。
- 视觉语言与现有 OpenMPA 一致：系统字体、暖白背景、细边框、低饱和青绿色强调色、120–160ms 颜色反馈。

非目标：

- 不修改 Runtime/Worker、不新增后端接口。
- 不重做侧边栏、顶部原生壳层、页面编辑器。
- 不实现助手设置、AGENTS.md、火山 Skill 目录（单独 spec）。
- 不引入新的动画库或全局主题改造。

## 3. 视觉规格

### 3.1 布局

| 元素       | 规格                                                                                                                                                                |
| ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 阅读列     | `max-width: 760px`，水平 padding 24px（<760px 时 16px），垂直 padding 20px 0 24px                                                                                   |
| Persona 头 | 高 52px；mascot 32px；标题 13px/600；副标题 11px muted；底部 1px `--line`；`position: sticky; top: 0`，背景 `rgba(255,255,255,.86)` + `backdrop-filter: blur(12px)` |
| 消息间距   | 18px；assistant 与 user 均为单列，不使用卡片包裹 assistant                                                                                                          |
| 用户消息   | 右对齐，`max-width:76%`，padding 9px 13px，圆角 `16px 16px 4px 16px`，背景 `#35766f`                                                                                |
| 助手消息   | 左对齐，13px / line-height 1.72，颜色 `#4b515c`；段间距 10px；代码块圆角 10px                                                                                       |
| 工具事件   | 单行 11px muted，左侧 2px 竖线，不占独立气泡                                                                                                                        |
| 空状态     | 居中，标题 22px/550，副标题 12px muted；可选 3 个示例提示，点击只填入输入框、不自动发送                                                                             |

### 3.2 输入容器

| 项       | 规格                                                                                                        |
| -------- | ----------------------------------------------------------------------------------------------------------- |
| 容器     | 1px `#e2e3e8`，圆角 16px，背景 `#fff`，`box-shadow: 0 8px 24px rgba(28,30,36,.06)`，padding `10px 12px 8px` |
| 聚焦     | `:focus-within` 边框 `#9fbdb6` + 3px `rgba(117,185,180,.18)` ring；不使用位移                               |
| textarea | 无边框、透明背景、13px、行高 22px、`min-height:44px`、`max-height:168px`、自动增高、`resize:none`           |
| 底栏     | 高 32px，`display:flex; align-items:center; justify-content:space-between`；左提示 10px muted；右发送键     |
| 发送键   | 32px 圆形；可用 `#5f8f87`，禁用 `#cfd6e6`；图标 16px；距右边缘 0（由容器 padding 控制）                     |
| 运行中   | 发送键替换为同尺寸停止键，位置不变，避免布局跳动                                                            |

### 3.3 状态与提示

- 只保留一个 `chat-status` 区域，位于输入容器上方；内容优先级：错误 > 未知结果 > 恢复中 > 运行中。
- `pending` 操作、profile 未验证、unknown 结果不再各占一块，统一为一条可展开的说明。
- 错误使用 `role="alert"`；普通状态使用 `role="status"`。
- 提示出现/消失不改变输入容器的宽度和位置。
- 自动滚动仅在距底部 80px 内生效；用户向上翻阅时保持位置，并显示“回到底部”按钮。

### 3.4 动效与可达性

- 只对 `background-color`、`border-color`、`box-shadow` 使用 140ms `cubic-bezier(0.23,1,0.32,1)`。
- 键盘触发的高频动作不加位移动画。
- `prefers-reduced-motion: reduce` 时取消非必要过渡。
- 焦点环使用现有 `#b8c4fa` 体系；输入容器获得 `:focus-within` 可见反馈。

## 4. 消息投影规则

1. 本地乐观 user 消息带稳定 `clientId` 与 `invocationId`；同一 invocation 只对应一条用户消息，收到该 invocation 的 user 事件时**替换**本地消息（服务端内容为权威），不按文本相似度判断。
2. partial 助手文本按 invocation 维护一个“流式槽位”，槽位 id 为 `${invocationId}:stream:${author}:${slot}`；同一槽位内的 chunk 追加到同一条消息，不再用 chunk 的 `blockId` 作为身份。
3. 以下事件关闭当前槽位：non-partial 文本事件、toolCall/toolResult 事件、terminal 事件。关闭后下一个 partial 文本开启新槽位并递增 `slot`，因此“文本→工具→文本”不会合并成一条。
4. 历史读取结果整体替换 projection，不与流式结果拼接。
5. 保留现有 `eventId + kind + blockId` 去重和 invocation 作用域检查。
6. 空文本事件不生成消息；工具事件单独渲染，不进入正文。

## 5. 验收

| ID   | Given                                                      | When       | Then                                                                            | 验证                                  |
| ---- | ---------------------------------------------------------- | ---------- | ------------------------------------------------------------------------------- | ------------------------------------- |
| AC-1 | 1280 / 900 / 720 宽窗口                                    | 打开对话页 | 阅读列 ≤760px 且居中；无横向滚动；输入容器与列同宽                              | 组件测试 + 构建后实机（用户目视确认） |
| AC-2 | 输入 1 行 / 6 行                                           | 输入变化   | 发送键始终在底栏右侧垂直居中；容器只向下增高；不遮挡消息                        | 组件测试 + 构建后实机（用户目视确认） |
| AC-3 | 运行中                                                     | 点击发送后 | 发送键原位变为停止键；无布局位移                                                | 组件测试                              |
| AC-4 | 乐观 user + SSE user 回显 + 20 个 partial chunk + 最终全文 | 流式完成   | 恰好 1 条 user、1 条 assistant                                                  | projection fixture 测试（先红后绿）   |
| AC-5 | 完成一次运行                                               | 回读历史   | 消息内容与流式结束时一致，无重复、无闪烁                                        | 集成测试                              |
| AC-6 | pending / unknown / profile 未验证                         | 同时存在   | 只显示一条状态，可展开查看原操作                                                | 组件测试                              |
| AC-7 | 键盘 Tab / Enter                                           | 聚焦与发送 | 焦点可见；Enter 发送、Shift+Enter 换行；IME 组合输入不误发                      | 组件测试                              |
| AC-8 | 系统减少动态效果                                           | 交互       | 无位移动画，状态仍可见                                                          | 组件测试 + 样式检查                   |
| AC-9 | 长对话且右侧用户气泡接近阅读列边缘                         | 滚动对话   | 滚动条位于整个聊天视口最右侧；阅读列保留 ≥16px gutter，气泡不被滚动槽覆盖或挤压 | CSS 结构回归 + 构建后实机             |

## 6. 任务拆分

| 序号 | 任务                                         | 文件                                                              | 完成标志                      |
| ---- | -------------------------------------------- | ----------------------------------------------------------------- | ----------------------------- |
| T1   | 对话页局部样式变量与阅读列/Persona 头        | `src/client/style.css`、`src/client/desktop/style.css`            | 三个断点布局一致              |
| T2   | 重构输入容器：固定底栏、自动增高、停止键原位 | `src/client/desktop/Chat.tsx`、`src/client/desktop/ChatState.tsx` | AC-2/AC-3                     |
| T3   | 收敛状态提示                                 | `src/client/desktop/Chat.tsx`                                     | AC-6                          |
| T4   | 消息投影去重与乐观消息对账                   | `src/client/desktop/projection.ts`、`ChatState.tsx`               | AC-4/AC-5（先写失败 fixture） |
| T5   | 回归与实机验收                               | `tests/desktop-*`、构建、安装                                     | 全量 Node/Go 门禁通过         |

## 7. 风险与回滚

- 风险：自动增高在 WebView 的 `scrollHeight` 读取时机不一致，可能在输入法组合期抖动。处理：仅在 `draft` 变化后测量，组合输入期间保持当前高度。
- 风险：历史回读整体替换可能导致滚动位置跳变。处理：替换后保持 `scrollTop` 相对底部距离，只有距底部 80px 内才自动贴底。
- 回滚：仅涉及前端 CSS 与对话组件/投影，无数据与协议变更；回滚即恢复旧样式与旧投影逻辑。

## 8. 实施与验证

实现文件：

- `src/client/desktop/projection.ts`：流式槽位、工具切槽、乐观用户消息替换、事件去重。
- `src/client/desktop/Chat.tsx`：单一状态区、紧凑工具行、输入容器与固定底栏、运行中停止键原位。
- `src/client/desktop/ChatState.tsx`：贴底滚动与“回到最新”。
- `src/client/style.css`、`src/client/desktop/style.css`：760px 阅读列、52px Persona 头、输入容器与状态样式。
- `desktop/locales/{zh-Hans,en}.json` 与生成物：`jumpToLatest`。

焦点黑框根因与修复：

- 截图取样确认输入框聚焦时出现 2px、RGB(36,36,36) 的圆角边框，颜色等于产品 `--accent/--ink: #242424`。
- 来源是既有规则 `.template-app :is(button,a,input,textarea,select):not(.call-view *):focus-visible { outline: 2px solid var(--accent) }`；其选择器特异性 (0,3,1) 高于 `.composer-field textarea:focus-visible { outline: none }` 的 (0,2,1)，所以多行输入框被画成黑色硬框。
- 修复：在 `src/client/desktop/style.css` 增加同特异性且更晚加载的 `.desktop-app .composer-field textarea:focus/:focus-visible { outline: none }`，并设置 `caret-color: #5f8f87`；焦点反馈只保留容器的青绿色 ring。构建产物已确认包含该规则与深色主题孪生规则。

滚动条挤压根因与修复：

- 根因：`live-chat` 自身同时承担 760px 阅读列和纵向滚动容器，滚动槽因此落在阅读列右边界，与右对齐的用户气泡共用空间。
- 修复：`persistent-chat.live-chat` 改为全宽外层，`chat-transcript` 负责整个视口滚动；消息通过动态 `padding-inline` 保持 760px 居中阅读列，Persona、状态区和 Composer 单独约束到同宽。滚动条固定在外层最右 gutter，窄窗口保留 16px 内容间距。
- 设计影响：遵循 `apple-design` 的空间一致性和 `emil-design-eng` 的不可见细节原则；不增加动画，不改变消息投影、滚动贴底或输入行为。

验证结果：

| 检查                                                        | 结果                                                           |
| ----------------------------------------------------------- | -------------------------------------------------------------- |
| `tests/desktop-projection-dedupe.test.ts`                   | 先红 3 项失败，修复后 5/5 通过                                 |
| `tests/desktop-chat-layout.test.tsx`                        | 3/3 通过                                                       |
| `npm test -- --run`                                         | 78 文件 / 392 项通过                                           |
| `npm run typecheck`、`npm run lint`、`npm run check-format` | 通过                                                           |
| `go test ./...`（desktop）                                  | 通过                                                           |
| `npm run build:desktop:frontend` + MyGo `build -skip-dmg`   | 通过，OpenMPA.app 21.4 MB，签名校验通过                        |
| 本机安装与启动                                              | 已安装到 `/Applications/OpenMPA.app`，进程启动，无应用级 crash |

未覆盖：实机 3 个断点的目视确认、真实 Runtime 流式对话的重复输出复验（需要用户观察或授权测试会话）。

## 9. 后续

- 助手设置：`GET/PUT /api/v1/agents` 读写 AGENTS.md、model、skills、MCP；火山 Skill 目录走 AgentKit OpenAPI `ListSkillSpaces` / `ListSkillsBySkillSpace`。单独 spec 与验收。
