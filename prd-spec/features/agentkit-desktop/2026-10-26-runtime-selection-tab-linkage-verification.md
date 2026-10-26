# Runtime 选择恢复与设置 Tab 联动 Hotfix：发布与验证

状态：已接受；日期：2026-10-26；关联主设计：[Runtime 选择恢复与设置 Tab 联动 Hotfix](2026-10-26-runtime-selection-tab-linkage-hotfix.md)。

## 7. 兼容、回滚与发布检查

SQLite 表 schema 不变，新增 `platform.principal_scope` 和 `platform.runtime_binding` 非敏感设置键，与 Region/Project/Runtime ID 同事务写入；后者是 tuple 指纹，不是凭据。目录变化同时清除 scope、指纹与 Runtime ID。旧数据库可直接读取；缺少 scope/指纹的历史记录只有在当前账户完整目录精确包含该 ID 时才自动接管。新旧应用混跑时旧版改写导致 tuple 与指纹不匹配，记录重新按旧记录规则由当前目录验证，远端无 schema 或数据差异。

回滚可恢复旧原生控件和仅预选行为；SQLite 与远端数据无需回滚。若自动恢复导致错误 Runtime 或重复绑定，停止准出并回滚恢复触发，保留 Combobox/Tabs 和实时连接同步修复。

发布前检查：聚焦状态机测试、`go test ./...`、`go vet ./...`、Node 全量门禁、前端构建、MyGo app 构建和本机五 Tab 目视验证。日志仅观察错误码、Runtime ID 和 Region/Project，不输出 endpoint query、userinfo 或凭据。

## 8. 实施与验收

实施顺序：补 GG-1 RED → 增加 Runtime label/membership/同步 helper → 接入安全恢复状态机 → 替换 Combobox/Tabs → 局部回归 → 全量与构建 → 本机验证。每一步保持可回退；恢复状态机未通过边界测试时不得启用自动 Bind。

| ID    | Given                                                         | When                                   | Then                                                            | 验证方法/命令                                       | 环境          |
| ----- | ------------------------------------------------------------- | -------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------- | ------------- |
| AC-1  | 目录中已选 Runtime 成功 Bind                                  | epoch 更新并切换五个 Tab               | 选择、当前连接和各 Tab Runtime 一致                             | `go test ./internal/nativeui -run RuntimeSelection` | 本地 headless |
| AC-2  | 已认证且 SQLite 保存 ID 在完整当前目录，包括缺 scope 的旧记录 | 进入设置或登录恢复                     | 自动 Bind 一次，所有 Tab Connected                              | 同上，mock cloud/connection                         | 本地 headless |
| AC-3  | 未认证、目录不完整或 ID 不匹配                                | 尝试恢复                               | List/Bind 调用符合 NFR-1，绝不冒充连接                          | 同上负控                                            | 本地 headless |
| AC-4  | 目录包含不同名称和 ID                                         | 输入名称片段或 ID 片段并选择           | Combobox 仅显示匹配项且立即切换；无刷新/连接按钮                | MyGo tester 交互测试                                | 本地 headless |
| AC-5  | 五个设置 Tab                                                  | 点击及键盘切换                         | 仅当前 Tab 为 selected 且有原生强调标记                         | MyGo tester/accessibility 测试                      | 本地 headless |
| AC-6  | 修复后的工作区                                                | 执行默认 admission 与构建              | Go/Node/构建全部通过                                            | 仓库原生命令                                        | 本地          |
| AC-7  | 最新安装包                                                    | 登录并检查 Runtime、身份认证、部署 Tab | 无重复连接提示，视觉与交互一致                                  | 本机应用目视 + 安全日志                             | macOS         |
| AC-8  | 历史 Runtime 不在当前账户完整目录                             | 打开 Runtime 页                        | 不显示历史 ID；显示当前账户未选择状态、可见下拉及仅当前账户候选 | MyGo tester + 实机                                  | 本地 + macOS  |
| AC-9  | 未连接 Runtime                                                | 点击新建助手                           | 不打开创建弹窗，进入 Settings/Runtime                           | 原生 action 测试                                    | 本地 headless |
| AC-10 | 唯一 Runtime 的自动或手动 Bind 失败                           | 再次选择同一 Runtime                   | 产生新的 Bind 尝试，不因 Combobox 同值而无操作                  | `TestFailedRuntimeBindAllowsSameChoiceRetry`        | 本地 headless |

## 9. 风险与追踪

| ID     | 触发条件                                | 概率/影响 | 缓解措施                                                                          | 监测信号            | 兜底/回滚                |
| ------ | --------------------------------------- | --------- | --------------------------------------------------------------------------------- | ------------------- | ------------------------ |
| RISK-1 | 缺 scope 的旧保存 ID 在另一账户恰好复用 | 极低/高   | 当前账户完整目录 membership + BindRuntime 控制面权威校验；新绑定立即写 scope/指纹 | 账户 scope/绑定错误 | 清除默认并要求切换账户   |
| RISK-2 | 登录、目录加载、用户手动切换并发        | 中/中     | generation/epoch fencing，busy 时禁用控件                                         | 迟到结果或重复 Bind | 丢弃旧结果，保留当前连接 |
| RISK-3 | Combobox 标签与 ID 映射不唯一           | 低/中     | 标签始终包含唯一 Runtime ID，反向映射按完整标签                                   | 选中 ID 不符合展示  | 禁用连接并刷新目录       |

无阻塞待确认项；自动使用/即时切换模型已由用户在实机截图复核后确认。自动恢复仅在既有凭据已进入当前进程且目录验证通过后发生，不扩大秘密持久化范围。

| 需求                 | 方案位置/DEC | 验收标准                  | 验证方法                           | 覆盖状态 |
| -------------------- | ------------ | ------------------------- | ---------------------------------- | -------- |
| FR-1                 | 5.1 / DEC-1  | AC-1、AC-7                | 原生状态机 + 本机验证              | 已覆盖   |
| FR-2                 | 5.1 / DEC-2  | AC-2、AC-3                | mock cloud/connection              | 已覆盖   |
| FR-3                 | 5.2 / DEC-3  | AC-4                      | MyGo tester                        | 已覆盖   |
| FR-4                 | 5.2 / DEC-4  | AC-5                      | MyGo tester/accessibility          | 已覆盖   |
| FR-5                 | 5.2 / FACT-7 | AC-8                      | MyGo tester + 实机                 | 已覆盖   |
| FR-6                 | 5.2          | AC-9                      | 原生 action 测试                   | 已覆盖   |
| NFR-1                | 5.1、6       | AC-2、AC-3                | 负控调用计数                       | 已覆盖   |
| NFR-2 / GG-1 至 GG-3 | 2、5.1、8    | AC-1 至 AC-6、AC-9、AC-10 | RED→GREEN + 默认 Go/Node admission | 已覆盖   |
