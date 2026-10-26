# 本地配置数据库与多 Region 设计

状态：已接受；日期：2026-10-26；类型：本地架构与用户功能，L2；基线为当前 Runtime-key/localization hotfix 工作区。

## 1. 结论与证据

OpenMPA 使用一个纯 Go SQLite 数据库统一管理非敏感本地设置；AK/SK、STS Token、JWT 和渠道秘密只保存在当前进程，不进入数据库。MyGo 0.6.4 没有专用 SQLite 模块，但提供 `PathUserData` 和可注入的 `WindowStateStore`，适合由应用自有数据库承载设置和窗口状态。

| ID | 已验证事实 | 证据 | 设计影响 |
|---|---|---|---|
| FACT-1 | MyGo 提供每用户 `PathUserData`，不提供数据库封装 | MyGo `paths.go:21-23,52-103`；源码检索无 SQLite/database API | 数据库由 OpenMPA 拥有，路径由 MyGo 解析 |
| FACT-2 | MyGo 可将窗口状态委托给应用数据库 | MyGo `window.go:195-217`、`window_state_test.go:133-210` | SQLite Store 实现 `WindowStateStore` |
| FACT-3 | AgentKit 当前中国 Runtime 区域为北京和上海 | 火山 AgentKit“服务地址”文档（2026-10-26 复核）；VeADK `agentkit_sandbox_region.py:25` | 下拉框只给出 `cn-beijing`、`cn-shanghai`，北京为首次默认 |
| FACT-4 | 控制面 host 随 Region 变化 | 官方服务地址：`agentkit.<region>.volcengineapi.com`；现 `cloud.New.endpoint` 同规则 | 每次调用以当前请求 Region 选择 host 与签名 Region |
| FACT-5 | AgentKit 2025-10-30 未发现独立 ListRegions action | 官方 API、SDK/VeADK 与控制面 action 检索 | 区域目录来自官方服务地址 catalog；各区 Runtime 由该区 `ListRuntimes` 实时获取 |

## 2. 范围与安全边界

SQLite 持久化：显示语言模式、最近 Region、Project、成功绑定的 Runtime ID、窗口状态和后续非敏感设置。SQLite 不持久化：AK/SK/Security Token、Runtime Key、JWT、渠道凭据、Authorization、完整 endpoint URL、对话正文或工具参数。

删除产品中的“保存到钥匙串/已保存账户/恢复账户”及其 typed IPC。登录失败保留当前内存中的输入并支持显示/修改；退出应用、登出、离开设置的既有秘密清理规则不变。

## 3. 数据与接口

数据库固定为 `<PathUserData>/openmpa.db`，目录 0700、数据库 0600。使用 `database/sql` 与 `modernc.org/sqlite`，无 CGO。单进程单连接，WAL、foreign_keys、busy_timeout 与 FULL synchronous。Schema v1：

- `settings(key PRIMARY KEY, value, updated_at)`：只允许应用内枚举的 key；值有长度和类型校验。
- `window_states(key PRIMARY KEY, x, y, width, height, maximized, full_screen, updated_at)`：实现 MyGo `WindowStateStore`。
- `PRAGMA user_version=1`：未知新版本失败关闭，不自动降级。

设置键：`display.language`、`platform.region`、`platform.project`、`platform.runtime_id`。目录更新与 Runtime 选择使用事务；Region 或 Project 变化时原子清空旧 Runtime ID。

## 4. Region 行为

Region 下拉项：华北2（北京）`cn-beijing`、华东2（上海）`cn-shanghai`；默认北京。选择变化立即清空旧目录、旧 Runtime 选择及连接投影，持久化新 Region，然后向 `https://agentkit.<region>.volcengineapi.com` 调用该区 `ListRuntimes`。STS 身份接口保持官方 `https://sts.volcengineapi.com`；AgentKit 请求签名 Region 与所选 Region 相同。

Project 默认 `default`，可编辑；刷新目录时持久化。Runtime 只有在当前 Region/Project 完整目录中且 `BindRuntime` 成功后才持久化。重启后只预选保存 ID，不在没有凭据时自动连接或发请求。

## 5. 验收

| ID | Given | When | Then | 验证 |
|---|---|---|---|---|
| AC-1 | 空数据目录 | 首次打开 | DB v1，region=cn-beijing、project=default、language=system | SQLite 集成测试 |
| AC-2 | 选择上海 | 刷新 Runtime | 请求 host 为 `agentkit.cn-shanghai.volcengineapi.com`、签名 region=cn-shanghai | HTTP mock |
| AC-3 | 北京已有 Runtime 选择 | 切换上海 | 目录/Runtime/连接失效，DB runtime ID 为空 | 原生 + SQLite 测试 |
| AC-4 | 上海绑定成功 | 重启 | 上海/Project/Runtime 被预选，但不自动读取云端 | 重启测试 |
| AC-5 | 登录返回 401 | 查看表单 | 凭据草稿仍在内存，可显示/编辑/重试，DB 无秘密 | headless + DB 内容扫描 |
| AC-6 | 任意运行 | 生成 typed IPC | 无 SavedProfiles/Restore/Keychain API | 绑定检查 |
| AC-7 | 窗口关闭/重开 | MyGo 保存/读取 | 窗口状态由 SQLite Store 承载 | WindowStateStore 测试 |

## 6. 兼容、迁移与回滚

首次打开数据库时可读取合法的旧 `display-language.json` 作为一次性默认，但不删除旧文件。旧 `saved-logins.json`、`credential-profiles.json` 和钥匙串条目不读取、不迁移、不删除。回滚旧版不会破坏这些历史文件；SQLite 是新增文件。数据库损坏或版本过新应显示本地配置错误，不得回退到明文秘密文件。

风险：Region catalog 未来扩展时需要更新官方 catalog；服务端未提供 ListRegions 前不猜测其它区域。SQLite 不解决秘密跨重启持久化；若未来重新提出该需求，必须单独设计用户主密码加密 vault 或正式平台签名方案。

## 7. GateGap

`GG-1 confirmed_escape`：旧默认测试只验证了 Region 字符串参与请求，没有约束设置控件必须是官方 catalog 下拉，也没有断言上海 host 与签名 Region；因此 `services.ts` 的北京硬编码和原生自由文本输入通过了既有全量测试。

闭环检查：`TestAgentKitRegionCatalogRoutesHostAndSignature` 断言上海域名、上海签名 scope 与非法 Region 负控；`TestRegionCatalogIsAClosedDropdownAndRegionChangeInvalidatesRuntime` 断言下拉 catalog 和跨区清理；`desktop-runtime-directory.test.tsx` 必须通过 `<select>` 切到上海后才加载、绑定上海目录。这些检查进入默认 Go/Node 全量门禁。授权账户的 opt-in `TestLiveAgentKitRegionCatalog` 另对北京、上海逐区执行只读 `ListRuntimes`，但不替代默认离线门禁。
