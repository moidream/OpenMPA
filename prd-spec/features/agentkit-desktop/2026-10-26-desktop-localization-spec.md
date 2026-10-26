# 桌面中英文配置 Spec

## 1. 状态与范围

- 状态：已接受
- 日期：2026-10-26
- 类型：用户可见功能，L2
- Change ID：`desktop-zh-en-localization`
- 关联：用户补充“设置页面支持多语言配置，中文、英文”，并进一步确认默认跟随系统；与 [Runtime Key Session Hotfix](2026-10-26-runtime-key-session-hotfix.md) 同一 dev-loop 实施，互不改变身份边界。
- 基线：`af85b28c021e162946a236d4a63d0f1c3dfbe839`。

结论：使用 MyGo 0.6.4 的单一 App Locale 与生成式 catalog，设置页提供“跟随系统 / 中文 / English”三项，首次启动跟随系统，用户显式选择持久化；原生 UI、系统菜单和 WebView 同步切换。系统首选语言不是中文或英文时回退 catalog source 简体中文。首个切片覆盖桌面壳层、设置、创建助手及聊天/Space 主要路径，不允许出现“设置已切换但同屏关键操作仍为另一语言”的伪多语言状态。

## 2. 事实与用户场景

| ID | 类型 | 结论/假设 | 精确证据 | 设计影响 | 验证/关闭方式 |
|---|---|---|---|---|---|
| FACT-1 | 已验证事实 | MyGo 的 `App.SetLocale` 同步原生窗口、菜单和页面，并持久化选择 | MyGo 0.6.4 `locale.go:119-187`；`mygo-runtime` 的 `onLocaleChange` | 使用框架 locale，不建立两套状态 | Go/页面联动测试 |
| FACT-2 | 已验证事实 | MyGo catalog 可从同一 JSON 生成 typed Go 与 TypeScript accessors | MyGo 0.6.4 `cmd/mygo/i18n_test.go:82-119` | catalog 是唯一翻译源 | `mygo i18n check`、生成一致性 |
| FACT-3 | 已验证事实 | 当前 `mygo.json` 未配置 i18n，原生与 WebView 文案以英文硬编码 | `desktop/mygo.json`；nativeui/host 与 `src/client/desktop` 当前实现 | 需要接入 catalog 并迁移可见文案 | 静态 coverage + UI 测试 |
| FACT-4 | 已验证事实 | 语言选择属于应用偏好，不含凭据或云端配置 | 用户原始请求；现有本地设置边界 | 本地持久化，不调用 Runtime OpenAPI | 网络调用负控 |

主场景：Given 首次安装，When 打开 OpenMPA，Then 原生壳层和桌面内容匹配 macOS 在中文/英文中的首选语言；When 在设置页选择 English，Then 当前窗口立即切换且重启后保持 English；When 切回跟随系统，Then 清除显式覆盖。边界场景：系统首选语言不是已支持语言时回退简体中文；偏好文件损坏或含不支持值时回到跟随系统且不影响账户/Runtime 状态。

## 3. 目标、非目标与需求

目标：首次启动跟随系统；中文与英文可即时切换并重启保持；可恢复跟随系统；原生与 WebView 使用同一 locale；翻译缺失可由生成检查发现。

非目标：本期不支持繁体中文或第三种语言；不翻译用户输入、Runtime/Session 名称、服务端错误正文和技术标识；不把语言写入云端、Session 或 Runtime 配置。

| ID | 来源 | 优先级 | 触发/适用范围 | 必须行为或约束 | 可观察结果/度量 |
|---|---|---|---|---|---|
| FR-1 | 用户原始请求 | 必须 | 首次启动且无偏好 | 跟随系统，在 `zh-Hans|en` 中匹配；其它语言回退 zh-Hans | Settings、壳层与主内容使用解析后的同一语言 |
| FR-2 | 用户原始请求 | 必须 | 设置页语言配置 | 提供跟随系统、中文和 English，切换立即生效 | 当前窗口无需重启即变更 |
| FR-3 | 用户原始请求、FACT-1 | 必须 | 重启应用 | 恢复最后一次显式选择 | locale 与上次一致 |
| FR-4 | FACT-1、FACT-2 | 必须 | 语言变化 | 原生与 WebView 读取同一 MyGo locale | 同一帧关键导航和页面标题语言一致 |
| NFR-1 | FACT-2 | 必须 | catalog 和生成代码 | 两种 catalog 参数集合一致、无缺失 key | `mygo i18n check` 通过 |
| NFR-2 | FACT-4 | 必须 | 设置语言 | 不触发账户、Runtime 或业务 API，不保存秘密 | 仅本地偏好文件与 UI locale 变化 |
| NFR-3 | 现有客户端合同 | 必须 | 用户/远端动态文本 | 不翻译、不拼接或重解释身份与资源 ID | 名称和 ID byte-for-byte 保持 |

## 4. 方案与取舍

| ID | 决策 | 依据 | 影响范围 | 重新评估条件 |
|---|---|---|---|---|
| DEC-1 | 使用 MyGo `i18n` catalog，source=`zh-Hans`，另有 `en` | FACT-1、FACT-2 | Go、TS、系统控件 | MyGo catalog 无法覆盖新增前端框架 |
| DEC-2 | 本地 Localization 服务持久化 `system|zh-Hans|en`；system 调用 `App.SetLocale("")` | FR-1 至 FR-3 | 本地 UserData 与 typed IPC | 产品新增其它 locale 模式 |
| DEC-3 | WebView 订阅 `mygo:locale`，React 根状态变化后重新读取生成 messages | FACT-1 | desktop React 树 | WebView 被替换为全原生 |

备选 A 是 Go 和 TypeScript 各维护字典；实现快但容易漂移，拒绝。备选 B 是只设置 `<html lang>`；不会翻译界面，拒绝。备选 C 是只暴露中文/英文、没有“跟随系统”项；无法恢复默认策略，拒绝。

状态流：启动时 `Localization.Load` 读取 0600 JSON；文件不存在或无效时使用 `system`；服务对 system 调用 `App.SetLocale("")`，对显式值调用对应 tag。设置页调用 `Localization.Set({mode})`，只接受 `system|zh-Hans|en`，原子写入后设置 MyGo locale。MyGo 触发原生重绘和页面 locale 事件；React locale provider 递增状态并从生成 catalog 获取当前 messages。语言变化不改变 Connection epoch、ShellState 目录、草稿或表单普通字段；安全输入仍按既有页面切换规则处理。

## 5. 失败、兼容、安全与回滚

| 场景 | 行为 | 信号 | 恢复 |
|---|---|---|---|
| 偏好文件不存在 | 使用 system，不强制写文件 | Status 返回 system 与解析后的 locale | 用户可显式选择 |
| JSON 损坏/未知值 | 不执行未知 locale，返回 system 并保留安全错误 | 本地安全错误码 | 覆盖为有效选择 |
| 写入失败 | 当前选择不变，不显示伪成功 | 设置页错误 | 修复本地目录后重试 |
| catalog 缺 key/参数漂移 | 构建前检查失败 | `mygo i18n check` | 补齐 catalog 后再构建 |
| 页面切换期间改语言 | MyGo/React 各自按最新 locale 重绘，不重发远端操作 | locale 事件 | 无业务恢复动作 |

偏好文件不含凭据；路径不进入 WebView。Language IPC 固定枚举，不接受路径或任意 locale。用户/服务端内容不进入翻译 key，避免改变远端语义。旧安装没有偏好文件时迁移为中文；回滚旧版不会删除账户和 workspace 数据，只会忽略 locale 偏好文件。

## 6. 实施与验收

1. L1：catalog、Localization 服务、跟随系统与显式持久化单元测试；生成 Go/TS accessors。
2. L2：设置页语言入口，壳层/设置/创建会话与主要 WebView 路径迁移；动态名称保持原值。
3. L3：i18n check、生成一致性、Go/Node 全量与构建；安装后系统默认、中文/英文切换和英文重启复验。

| ID | Given | When | Then | 验证方法/命令 | 环境 |
|---|---|---|---|---|---|
| AC-1 | 空 UserData，系统语言分别为中文/英文 | 启动 Localization | mode 为 system，locale 分别解析为 zh-Hans/en | Go 单元测试 | 本地 |
| AC-2 | 当前 zh-Hans | 设置 English | 原生 Settings/侧栏和 WebView 标题立即为英文 | Go headless + React 测试 | 本地 |
| AC-3 | 已保存 en | 重建服务/重启 App | locale 仍为 en | Go 持久化测试 + 安装复验 | 本地/macOS |
| AC-4 | 两种 catalog | 执行生成和检查 | key/参数完整且生成文件无差异 | `mygo i18n check` + `mygo generate` | 本地 |
| AC-5 | 任一语言 | 显示 Runtime/Session/Space 动态名称 | 动态值不变 | 组件测试 | 本地 |
| AC-6 | 语言切换 | 观察 HTTP mock | 账户/Runtime/Session 请求数不增加 | Localization 单元/集成负控 | 本地 |

| 需求 | 方案位置/DEC | 验收标准 | 验证方法 | 覆盖状态 |
|---|---|---|---|---|
| FR-1 | 第 4 节 / DEC-1、DEC-2 | AC-1 | Go 单元 + 安装复验 | 已覆盖 |
| FR-2 | 第 4 节 / DEC-2、DEC-3 | AC-2 | headless + React | 已覆盖 |
| FR-3 | 第 4、5 节 / DEC-2 | AC-3 | 持久化 + 重启 | 已覆盖 |
| FR-4 | 第 4 节 / DEC-1、DEC-3 | AC-2、AC-4 | 跨渲染层测试 | 已覆盖 |
| NFR-1 | 第 4 节 / DEC-1 | AC-4 | i18n admission | 已覆盖 |
| NFR-2 | 第 4、5 节 / DEC-2 | AC-6 | 网络负控 | 已覆盖 |
| NFR-3 | 第 4、5 节 | AC-5 | 动态值测试 | 已覆盖 |

## 7. 风险

| ID | 触发条件 | 概率/影响 | 缓解措施 | 监测信号 | 兜底/回滚 |
|---|---|---|---|---|---|
| RISK-1 | 部分硬编码文案遗漏 | 中/中 | 关键路径清单、catalog 检查、双语截图 | 同屏语言混杂 | 补齐 key，不回退状态服务 |
| RISK-2 | 中文长度改变布局 | 中/中 | 复用弹性布局、640/760/1200px headless | 截断/重叠 | 调整布局，不缩小点击区 |
| RISK-3 | locale 切换误清表单或触发业务请求 | 低/高 | locale 独立于 identity generation；AC-6 | API 调用或草稿变化 | 回滚 UI 订阅，保留偏好服务 |

无阻塞待确认项。本期明确 catalog 只有 `zh-Hans` 和 `en`，默认模式为跟随系统。
