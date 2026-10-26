# 删除钥匙串账户恢复功能 Hotfix

状态：已接受并替代旧稳定签名方案；日期：2026-10-26；类型：身份/秘密存储 Bug 修复，L3。

## 问题与最终决策

旧版本把 AK/SK 等凭据写入 macOS 登录钥匙串。ad-hoc 签名重编译后 CDHash 变化，旧条目的 ACL 仍绑定旧进程身份，导致用户反复输入系统密码仍无法保存或恢复。为单台开发机建立私有 Root、独立钥匙串和稳定本地签名虽可技术绕过，但不能让其他用户可靠使用，也不具备发布资格。

因此产品不再提供“保存到钥匙串”“已保存账户”“恢复账户”能力，也不建立、分发或依赖机器专属签名链。AK、SK、Security Token、JWT、Runtime Key 和渠道秘密只存在于当前进程内存；退出应用后必须重新输入。非敏感设置由 `openmpa.db` 持久化，详见《本地配置数据库与多 Region 设计》。

## 需求

- 删除 Keychain checkbox、已保存账户列表和恢复入口。
- 删除 `SavedProfiles`、`Restore`、Remember 等 typed IPC/DTO 与生产装配。
- 登录失败时保留 AK/SK/Security Token 草稿，允许显式显示、编辑和重试；只有登录成功后才清空。
- 离开设置页、登出、关闭应用和身份作用域切换仍清理内存秘密。
- 不读取、不迁移、不删除历史钥匙串项或旧 `saved-logins.json`；避免触发旧 ACL 授权框。
- SQLite 只保存语言、Region、Project、Runtime ID 和窗口状态，禁止保存任何秘密。

## 验收

| ID | Given | When | Then | 验证 |
|---|---|---|---|---|
| AC-1 | 登录返回 401/403 | 用户回到表单 | 原输入仍可显示、修改和重试 | MyGo headless test |
| AC-2 | 登录成功 | 请求完成 | 凭据输入清空，秘密未写入 SQLite | Go/SQLite test |
| AC-3 | 生成 typed client | 检查服务接口 | 不存在 SavedProfiles、Restore、Remember 或 Keychain 字段 | binding/typecheck |
| AC-4 | 启动、登录、退出应用 | 观察系统 | 不出现钥匙串授权框，不访问历史项 | 本机 smoke |
| AC-5 | 切换页面或身份作用域 | 检查内存状态 | AK/SK/Token/JWT/渠道凭据被清理 | native UI regression |

## 兼容与回滚

历史钥匙串项保持原状，OpenMPA 不再枚举或读取。回滚到旧版本可能重新触发旧 ACL，因此不作为推荐恢复路径。未来若重新提出跨重启保存秘密，必须独立设计正式发布签名或用户主密码加密 vault，不能恢复本机专属证书方案。
