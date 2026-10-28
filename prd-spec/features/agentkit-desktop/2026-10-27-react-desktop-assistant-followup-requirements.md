# React Desktop 助手、定时任务与主题：需求与约束

## 5. 需求与约束

### 5.1 功能需求

| ID   | 来源                                    | 优先级 | 触发/适用范围        | 必须行为或约束                                                                       | 可观察结果/度量                                                                                                   |
| ---- | --------------------------------------- | ------ | -------------------- | ------------------------------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------------- |
| FR-1 | 用户原始请求 / FACT-1 / FACT-2          | 必须   | Assistant 设置       | 显示并编辑 AGENTS.md；显示已配置 skills；支持筛选、添加、删除和保存。                | 设置页可见 AGENTS.md、Filter skills、Add skill、Save assistant configuration；保存请求包含 `agentsMd` 与 skills。 |
| FR-2 | 用户原始请求 / FACT-1 / FACT-3 / OPEN-1 | 必须   | Assistant 设置       | 在没有 skill space 目录契约时，不展示导入入口或假目录；明确说明只支持已配置 skills。 | 页面出现“Configured skills only”说明；代码中没有 skill space 导入调用。                                           |
| FR-3 | 用户原始请求 / FACT-4 / FACT-5          | 必须   | Scheduled & activity | 按当前会话的 `agentId` 跨页过滤任务；读取有上限；过滤后不宣称全量总数。              | 两页 fixture 中同助手任务全部出现；`CountAvailable` 为 false；超过上限返回明确错误。                              |
| FR-4 | 用户原始请求 / FACT-6 / FACT-7          | 必须   | 应用启动与主题切换   | 挂载时读取原生主题；`themeChanged` 事件继续实时更新。                                | 挂载测试中 `Host.Theme` 返回 Dark 后 React 主题为 Dark；事件测试仍通过。                                          |
| FR-5 | 用户原始请求 / FACT-8                   | 必须   | Spaces 清理          | 删除 `Pages.tsx` 与对应测试；保留 `workspace.flush()`。                              | `rg DesktopPages` 为空；关闭流程测试通过。                                                                        |
| FR-6 | 用户原始请求 / FACT-9                   | 必须   | 视觉烟雾             | 生成 chat 与 settings 两个 sidebar PNG，空文件或尺寸错误时测试失败。                 | `TestSidebarPreviewImages` 输出两张大于 1000 字节的 PNG。                                                         |
| FR-7 | 用户原始请求 / FACT-11                  | 必须   | 侧边栏               | 助手选择是可筛选 Combobox；New conversation 位于其下；会话列表只显示当前助手的会话。 | 侧边栏测试找到 Assistant、New conversation、CONVERSATIONS；没有 SPACES。                                          |
| FR-8 | 用户原始请求 / FACT-12                  | 必须   | Knowledge            | 知识库只针对当前选中的会话，不提供全局会话下拉。                                     | Knowledge 页显示当前会话作用域；没有 session selector。                                                           |

### 5.2 非功能约束

| ID    | 来源                           | 优先级 | 触发/适用范围    | 必须行为或约束                                                                    | 可观察结果/度量                                                                                    |
| ----- | ------------------------------ | ------ | ---------------- | --------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| NFR-1 | FACT-4 / 用户原始请求          | 必须   | 定时任务跨页读取 | 单次加载最多 20 页、1000 个匹配任务；超过时返回 `boundedResult`，不展示部分结果。 | 单元测试覆盖两页成功与上限错误；30 秒上下文超时。                                                  |
| NFR-2 | FACT-7 / 用户原始请求          | 必须   | 主题读取         | 主题拉取失败不能阻塞首屏；失败时保留本地或系统主题。                              | 主题读取 Promise 的 rejection 被吞掉，不产生错误横幅。                                             |
| NFR-3 | `desktop/README.md` 的凭据合同 | 必须   | 助手配置与主题   | 不把 AK/SK、Runtime key、JWT 或 skill 内容写入 shell state、日志或 UI。           | 代码审查与现有 secret gate 通过。                                                                  |
| NFR-4 | 用户原始请求                   | 必须   | 全部改动         | 不新增运行时依赖；现有测试、类型检查和构建必须通过。                              | `go test ./...`、`npx tsc --noEmit`、`npx vitest run`、`npm run build:desktop:frontend` 全部通过。 |
