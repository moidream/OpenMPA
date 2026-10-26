# Runtime 选择恢复与设置 Tab 联动 Hotfix

## 1. 文档状态

- 状态：已接受；日期：2026-10-26；类型：Bug 修复；风险：L2。
- Change ID：`runtime-selection-tab-linkage-hotfix`。
- 来源：用户报告已设置 Runtime 后其它设置 Tab 仍要求连接、已选 Runtime 不再展示、名称或 ID 查找需要下拉筛选、当前 Tab 缺少明显标记。
- 本文补充 [本地配置数据库与多 Region 设计](2026-10-26-local-config-region-spec.md)，并替代其第 4 节“登录后只预选、不恢复连接”的行为；无凭据时不自动发云请求的安全边界继续有效。

## 2. 问题、影响与证据

当前原生设置页存在三份未统一的状态：SQLite 保存的上次成功 Runtime、页面内 `selected/preferredRuntime`、宿主 `Connection.Status`。成功绑定或身份变更会递增 epoch，原生控制器随后清空页面目录和候选选择；登录后又只恢复 SQLite ID，不恢复实际连接。因此 Runtime 页可能显示 0 项或丢失已选项，而身份认证、部署 Tab 正确地根据实时连接判定为未连接。普通按钮实现的设置导航也没有原生 Tab 的选中语义。

| ID     | 类型       | 结论/假设                                                                                                                 | 精确证据                                                                                                                                    | 设计影响                                                                                             | 验证/关闭方式                      |
| ------ | ---------- | ------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------- | ---------------------------------- |
| FACT-1 | 已验证事实 | 成功绑定会改变 connection epoch，`complete` 随后调用 `invalidate`                                                         | `desktop/internal/nativeui/controller.go:192-220`；`desktop/internal/ipc/services.go:167-211`                                               | 修复后必须在 epoch 变化后从实时连接重新同步显示状态                                                  | 原生状态机测试                     |
| FACT-2 | 已验证事实 | `invalidate` 只保留 `preferredRuntime`，清空目录、`selected` 和其它非秘密展示状态                                         | `desktop/internal/nativeui/controller.go:118-135`                                                                                           | 成功绑定后的组合框和目录会丢失                                                                       | 缺陷态测试先红、修复后绿           |
| FACT-3 | 已验证事实 | SQLite 已事务保存 Region、Project、成功绑定 Runtime ID，但旧记录没有绑定时 principal scope；目录变化会清空旧 Runtime      | `desktop/internal/configstore/store.go` 的 `Snapshot/SetDirectory/SetRuntime`；`desktop/internal/dto/types.go` 的 `Principal/PlatformState` | 不新增表结构；新记录增加 scope 与 tuple 指纹，旧记录以当前账户完整目录 membership 作为一次性恢复权威 | configstore 测试 + 跨账户/混跑负控 |
| FACT-4 | 已验证事实 | 当前搜索是独立文本框加列表，且搜索条件会反向使已选 Runtime 无效                                                           | `desktop/internal/nativeui/settings.go:117-129,367-420`                                                                                     | 搜索与选择应合并为一个原生 Combobox，选择有效性不依赖瞬时搜索文本                                    | 组合框交互测试                     |
| FACT-5 | 已验证事实 | 设置导航使用普通按钮，未声明 tablist/tab/selected                                                                         | `desktop/internal/nativeui/settings.go:146-164`；MyGo `ui/tabs.go:3-40`                                                                     | 改用 MyGo 原生 Tabs                                                                                  | headless 可访问性/选中态测试       |
| FACT-6 | 已验证事实 | 身份认证和部署页只接受实时 `Connection.Status.Connected`                                                                  | `desktop/internal/nativeui/settings.go:431-437,487-491`                                                                                     | SQLite 偏好不能冒充连接；恢复必须实际调用既有 BindRuntime                                            | 跨 Tab 状态测试                    |
| FACT-7 | 已验证事实 | 现场账户 `2100131095` 在 `cn-beijing/default` 返回完整空目录，而授权账户 `2112682748` 可只读读取 `r-yexsqsrpxcxfapor9fvl` | 用户 2026-10-26 截图；`TestLiveSkillCatalog` 只读运行通过                                                                                   | “0 个 Runtime”是当前账户不可见，不是搜索失败；UI 必须给出账户级原因                                  | 当前账户空目录 + 授权账户只读探针  |

复现链路：加载目录并选择 Runtime → 成功 `BindRuntime` → connection epoch 改变 → `invalidate` 清空页面目录和候选 → 切换设置 Tab 或重绘 → 页面状态与实时连接/SQLite 偏好分裂。重启或重新登录时，偏好存在但连接尚未绑定，依赖连接的 Tab 继续阻断。

### GateGap

`GG-1 planned`：默认 Go 测试覆盖目录 membership、Region 失效和基本 Tab 切换，但没有覆盖“绑定导致 epoch 更新后仍展示同一 Runtime”“登录后按当前目录安全恢复”“Tab 的原生 selected 语义”。稳定承载点为 `desktop/internal/nativeui/settings_test.go`，负控为旧实现清空选择或恢复后仍未连接，默认 admission 为 `go test ./...`。

`GG-2 planned`：审查发现自发 `ResetRuntime` 推进 connection epoch 后，下一帧 `observe` 可能把它误判为外部失效并取消同一次 Region 持久化；既有 Region 测试将 Config 设为 nil，未覆盖该调度。稳定检查为 `TestSelfInitiatedRegionResetDoesNotInvalidateItsOwnPersistenceWork`，缺陷态 generation 被推进，修复态控制器同步记录自身 epoch，默认 admission 同为 `go test ./...`。

`GG-3 planned`：自动使用模型首轮实现后，跨 Go/React 测试未覆盖“未连接 newConversation 返回导航层后再次回调原生”的循环，也未覆盖唯一 Runtime Bind 失败后 MyGo Combobox 同值不再触发 Changed。稳定检查为 `desktop-navigation.test.tsx` 的 disconnected navigation 用例与 `TestFailedRuntimeBindAllowsSameChoiceRetry`；缺陷态分别重复调用原生创建、无法重选同项，修复态导航 Settings/Runtime 且同项可重试。

## 3. 目标、非目标与修复范围

目标：当前账户已设置的 Runtime 自动使用；切换时目录自动加载、选择后立即生效；其它账户或未归属当前目录的历史 Runtime 完全隐藏；当前账户未设置时始终展示可见筛选入口；所有设置 Tab 读取同一实时连接；Region/Project 控件对齐；未连接的新建助手直接进入 Runtime 设置。

非目标：不修改 Runtime/Worker、镜像、AgentKit OpenAPI、SQLite schema、Runtime Key 保管方式或业务身份规则；不把仅存在于 SQLite 的 Runtime ID直接视为已连接；不在未认证时发云请求；不改变 Region/Project 切换必须失效旧连接的规则。

范围内：`nativeui` 设置状态与渲染、既有 SQLite 偏好读取、现有 Platform List/Bind 调用、原生回归和本地应用验证。WebView 对话、Session API、渠道/任务业务合同不变，只消费修复后的同一宿主连接。

## 4. 需求与场景

| ID    | 来源                                 | 优先级 | 触发/适用范围                               | 必须行为或约束                                                                                                                  | 可观察结果/度量                                                 |
| ----- | ------------------------------------ | ------ | ------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------- |
| FR-1  | 用户原始请求、FACT-1、FACT-2         | 必须   | Runtime 成功绑定及后续 epoch 重绘           | 从实时连接同步当前 Runtime，保留其目录选择展示                                                                                  | Runtime Tab 显示当前名称/ID；其它 Tab 不再误报未连接            |
| FR-2  | 用户原始请求、FACT-3、FACT-6         | 必须   | 已认证账户存在保存的 Region/Project/Runtime | 自动读取当前账户完整目录；新记录还需 scope 匹配，旧记录以当前目录 membership 验证；匹配后立即 BindRuntime                       | 恢复后所有 Tab 读取 Connected；当前账户不可见时不绑定并显示原因 |
| FR-3  | 用户原始请求、FACT-4                 | 必须   | Runtime 设置页                              | 始终显示名称/ID 可搜索下拉；目录自动加载；选中候选立即切换，不再要求刷新和连接按钮                                              | 键盘或鼠标选中后直接成为当前 Runtime                            |
| FR-4  | 用户原始请求、FACT-5                 | 必须   | 五个设置 Tab                                | 使用原生 Tabs 并暴露 selected 状态                                                                                              | 当前 Tab 有强调标记，键盘方向键可切换                           |
| FR-5  | 用户原始请求、FACT-7                 | 必须   | 历史 Runtime 不属于当前账户完整目录         | 不展示该 Runtime ID、不作为默认值、不尝试 Bind；显示“当前账户尚未选择 Runtime”和带可见标题的筛选下拉                            | 页面不存在跨账户 Runtime 标识，当前账户始终有选择入口           |
| FR-6  | 用户原始请求                         | 必须   | 未连接时点击新建助手                        | 不打开失败弹窗；导航到 Settings/Runtime                                                                                         | Runtime 设置页成为唯一恢复入口                                  |
| NFR-1 | FACT-3、FACT-6、既有本地配置安全边界 | 必须   | 自动恢复                                    | 未认证不请求；登录 Region、设置 Region/Project 和完整目录 membership 必须匹配；有 scope 的新记录还需 scope 匹配；失败不伪造连接 | 当前账户不可见、Region/Project/目录不匹配时 Bind 调用数为 0     |
| NFR-2 | GG-1                                 | 必须   | 默认本地 admission                          | 回归进入既有 Go 全量测试                                                                                                        | 缺陷态负控失败，修复态 `go test ./...` 通过                     |

主路径：用户登录或解锁账户后，客户端自动加载当前账户目录；目录中存在该账户已设置 Runtime，则自动绑定并在所有设置 Tab 显示相同 Runtime。切换路径：用户在 Combobox 输入名称或 ID，选中候选后立即切换并持久化。失败路径：目录不完整、ID 不存在或 Bind 失败时不冒充连接；不属于当前账户目录的历史 ID 完全隐藏，页面回到“当前账户尚未选择 Runtime”的筛选入口。未连接时点击新建助手直接进入该恢复界面。

## 5. 修复方案

### 5.1 状态权威与同步机制

| ID    | 决策                                                                                                                                             | 依据           | 影响范围                             | 重新评估条件                          |
| ----- | ------------------------------------------------------------------------------------------------------------------------------------------------ | -------------- | ------------------------------------ | ------------------------------------- |
| DEC-1 | `Connection.Status` 是“当前正在使用 Runtime”的唯一权威；SQLite 仅表示“上次成功连接偏好”                                                          | FACT-3、FACT-6 | 所有设置 Tab 与宿主通知              | 远端提供持久连接会话合同              |
| DEC-2 | 成功 Bind 时将非敏感 principal scope、Runtime tuple 与同事务 tuple 指纹保存；新记录恢复需 scope+目录匹配，旧记录允许由当前账户完整目录一次性接管 | FR-2、NFR-1    | 本地设置、登录/解锁/进入设置后的恢复 | AgentKit 提供可验证的账户绑定资源引用 |
| DEC-3 | 候选选择采用 MyGo `ui.Combobox`，显示唯一标签 `name · runtimeId`；目录自动加载，选择后立即 Bind，不保留刷新/连接主按钮                           | FR-3           | Runtime 设置表单                     | 远端切换需要额外破坏性确认            |
| DEC-4 | 设置导航采用 MyGo `ui.Tabs`，内部稳定映射到现有字符串 tab key                                                                                    | FR-4           | 设置导航与键盘操作                   | MyGo Tabs 无法满足平台可访问性        |

状态不变量：

1. `connectedRuntime = Connection.Status.RuntimeID`，不得由 SQLite 或输入字段推导。
2. `preferredRuntime` 和 `preferredPrincipalScope` 只在成功 Bind 且 SQLite `SetRuntime` 成功后更新。
3. `selectedRuntime` 必须属于当前完整目录；过滤输入不能使已选 membership 失效。
4. Region 选择或 Project 首次编辑立即断开连接、清目录并清除旧 Runtime 偏好；Region 立即、Project 防抖后自动加载新目录，半输入值不写入持久配置。
5. epoch 失效继续清理秘密、确认框和跨 scope 数据，但随后从实时连接同步安全显示字段。

恢复顺序：读取 SQLite 偏好 → 确认账户已认证且登录 Region 匹配 → 自动 ListRuntimes(region, project) → 要求 `Complete=true` 且精确匹配 Runtime ID/Region/Project → 新记录额外核对 principal scope，旧记录以本次目录 membership 接管 → BindRuntime → SetRuntime → 从 Connection.Status 同步选择与共享状态 → 通知 WebView。并发或迟到结果继续由 controller generation、connection epoch 和 platform epoch 拒绝。

### 5.2 UI 设计

- Tabs 使用 MyGo 原生下划线强调、粗体选中文字和 focus ring，与现有 macOS 浅色表面一致，不再自绘一组无状态按钮。
- Runtime 区顶部只展示已成功绑定的“当前 Runtime”；未绑定时显示当前账户尚未选择，不投影其它账户的历史 ID。
- Combobox 始终可定位，取代独立搜索框与滚动列表；目录自动加载，选项包含名称和 ID，选中不同候选立即切换。
- 切换区是次级操作，不再显示整行刷新/连接按钮。
- Region 和 Project 使用相同可见标签、控件高度及顶部对齐；Project 编辑防抖后自动加载。
- 未连接的新建助手不创建 Modal，交由导航层进入 Settings/Runtime。

### 5.3 备选方案

仅恢复 SQLite 选中样式但不 Bind 的方案仍会让依赖连接的 Tab 阻断，拒绝。每个 Tab 独立重连会产生重复请求和竞争，拒绝。刷新后再连接的管理后台流程经用户实机验收拒绝；桌面端采用目录自动加载和选择即切换。

## 6. 边界、失败与恢复

| 场景                           | 处理                                                                      | 状态变化                                     | 信号                            | 恢复                           |
| ------------------------------ | ------------------------------------------------------------------------- | -------------------------------------------- | ------------------------------- | ------------------------------ |
| 未登录但 SQLite 有 Runtime     | 不 List、不 Bind                                                          | 仅展示上次偏好 ID                            | 账户页仍要求登录                | 登录后进入验证恢复             |
| 目录不完整或保存 ID 不存在     | 不 Bind，不向 UI 投影原偏好                                               | 当前连接保持未连接                           | 当前账户尚未选择 + 可见筛选下拉 | 从当前账户自动目录选择 Runtime |
| 自动 Bind 失败                 | 保留目录和候选，连接保持失败关闭                                          | 不写新偏好                                   | 安全错误码/HTTP 状态            | 用户重试或改选                 |
| 成功 Bind 后 epoch 改变        | 清理秘密和 scope 数据，从实时连接重建安全展示                             | 所有 Tab 看到同一 Runtime                    | connectionChanged               | 无需用户重复选择               |
| 用户切换 Region 或编辑 Project | 立即断开并清除旧目录/选择/偏好；Region 立即、Project 防抖后自动加载新目录 | 所有 Tab 立即未连接，半输入 Project 不持久化 | 新目录状态                      | 自动目录完成后选择新 Runtime   |
| 恢复请求迟到                   | generation/epoch 不匹配则丢弃                                             | 不覆盖新状态                                 | 无陈旧 UI 更新                  | 使用当前页面状态重试           |

安全边界不变：AK/SK、Token、Runtime Key、JWT 和渠道秘密不进入 Combobox、SQLite、日志或 WebView；自动恢复只使用非敏感资源 ID，真正凭据仍由 Go Cloud/Connection 层持有。恶意 SQLite Runtime ID 必须通过现有 ID 校验和当前账户目录 membership，不能直接构造 endpoint 或跨 Region 绑定。

## 7. 发布、验收与追踪

发布检查、验收标准、风险与追踪矩阵见[发布与验证 companion](2026-10-26-runtime-selection-tab-linkage-verification.md)。
