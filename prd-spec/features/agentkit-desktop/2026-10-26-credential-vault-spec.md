# 本地加密凭据库（AK/SK 记忆）

状态：已实现，待实机确认；日期：2026-10-26；类型：安全存储 + 用户便利功能，L3。

## 1. 目标与非目标

目标：

- 用户登录成功后，可用自选口令把当前 AK/SK/Security Token 加密保存在本机。
- 后续启动只需输入该口令即可解密并自动重新登录，不再重复输入 AK/SK。
- 口令、AK/SK、Security Token 均不以明文落盘，不进入日志、IPC 响应或错误消息。
- 保留“仅本次会话”模式；保存是显式选择，默认不保存。

非目标：

- 不恢复 macOS Keychain 方案，不依赖代码签名或机器绑定。
- 不提供口令找回；忘记口令只能删除后重新输入 AK/SK。
- 不保存 Runtime Key、JWT、渠道密钥或对话内容。
- 不做跨设备同步。

## 2. 安全设计

| 项 | 选择 |
|---|---|
| 口令派生 | PBKDF2-HMAC-SHA256（Go 标准库 `crypto/pbkdf2`） |
| 迭代次数 | 600,000（记录内保存，读取时校验下限 100,000、上限 2,000,000） |
| 盐 | 16 字节 `crypto/rand` |
| 加密 | AES-256-GCM，12 字节随机 nonce |
| 附加数据 | `openmpa-vault-v1 \0 accountId \0 region`，防止密文/元数据互换 |
| 明文载荷 | `{accessKey, secretKey, securityToken, expiresAt}` 严格 JSON |
| 存储 | SQLite `credential_vault` 单行表，BLOB 列；数据库 0600、目录 0700 |
| 口令约束 | 8–1024 字节，拒绝全空白；不写入任何字段 |
| 失败处理 | 统一 `vaultWrongPassphrase`，不区分长度/格式/解密细节；不自动清空凭据库 |

威胁模型：能读取本地数据库的攻击者可离线暴力尝试口令；PBKDF2 成本只能提高尝试代价，弱口令仍可被攻破。产品文案必须提示“口令丢失无法找回”。

## 3. 数据与接口

### 3.1 SQLite v2

```sql
CREATE TABLE credential_vault (
  id INTEGER PRIMARY KEY CHECK (id = 1),
  version INTEGER NOT NULL,
  kdf TEXT NOT NULL,
  iterations INTEGER NOT NULL,
  salt BLOB NOT NULL,
  nonce BLOB NOT NULL,
  ciphertext BLOB NOT NULL,
  account_id TEXT NOT NULL,
  region TEXT NOT NULL,
  updated_at INTEGER NOT NULL
) STRICT;
```

迁移：v1 → 创建该表并 `PRAGMA user_version = 2`；v0 直接建 v2。未知更高版本失败关闭。

### 3.2 服务接口（Go）

`internal/vault`：

- `Seal(passphrase, payload, accountID, region) (Record, error)`
- `Open(passphrase, Record) (Payload, error)`
- `Service.Status(ctx)`：只返回 `Exists/AccountID/Region/UpdatedAt`。
- `Service.Save(ctx, passphrase, payload, accountID, region)`。
- `Service.Unlock(ctx, passphrase)`：返回 payload 与记录元数据。
- `Service.Forget(ctx)`。

`ipc.Platform`：

- `VaultStatus()`：无网络、无秘密。
- `SaveVault({passphrase})`：要求已登录；加密当前内存凭据。
- `UnlockVault({passphrase})`：解密后用记录中的 Region 调用 `SignIn`，校验返回账号与记录一致；不一致立即 `SignOut`。
- `ForgetVault()`：删除记录，不改变当前登录状态。

### 3.3 内存清理

- 口令以 `[]byte` 传递，派生密钥用后清零；IPC 请求处理完清空。
- 解密后的 payload 在 `SignIn` 后清零；失败同样清零。
- 登出清空 `cloud` 内存凭据，但保留密文库。
- 离开设置页/切换身份时清空口令输入草稿。

## 4. UI

原生设置 Account 页新增“Saved credentials”区块：

- 已登录且无记录：口令 + 确认口令 + “Save credentials”；说明“口令不会保存，丢失后需重新输入 AK/SK”。
- 有记录且未登录：显示账号/Region/更新时间；口令 + “Unlock and sign in”；“Forget saved credentials”。
- 有记录且已登录：显示已启用；“Forget saved credentials”。
- 口令输入默认隐藏，可显式显示；提交中禁用；成功后清空输入。
- 错误只显示统一文案，不回显口令或密钥。

## 5. 验收

| ID | Given | When | Then | 验证 |
|---|---|---|---|---|
| AC-1 | 已登录 | 输入口令并保存 | SQLite 单行写入；数据库文件字节中不含 AK/SK 明文 | 集成测试 + 文件扫描 |
| AC-2 | 已保存 | 重启后输入正确口令 | 自动登录成功，账号与记录一致，输入清空 | Go 测试 + 实机 |
| AC-3 | 已保存 | 输入错误口令 | 返回 `vaultWrongPassphrase`，未登录，记录保留 | Go 测试 |
| AC-4 | 已保存 | 篡改密文/AAD/元数据 | 解密失败，不登录 | Go 测试 |
| AC-5 | 已登录且已保存 | 登出 | 内存凭据清空，记录仍在，可再次解锁 | Go 测试 |
| AC-6 | 已保存 | Forget | 记录删除，状态 exists=false | Go 测试 |
| AC-7 | v1 数据库 | 打开应用 | 迁移到 v2，原设置与窗口状态保留 | SQLite 测试 |
| AC-8 | 任意状态 | 检查 IPC/日志 | 响应与日志不含口令、AK/SK、Token、密文 | 类型与日志检查 |
| AC-9 | 口令 <8 字节或全空白 | 保存/解锁 | `invalidInput`，不写库 | Go 测试 |

## 6. 风险与回滚

- 口令强度不足：文案明确提示；PBKDF2 600k 提高离线成本，但不能保证弱口令安全。
- 数据库损坏/版本过新：失败关闭，不退回明文文件。
- 回滚：删除 `credential_vault` 表与 vault 服务即可；不会留下明文；已保存密文在旧版本中不可读，但不影响登录（重新输入 AK/SK）。

## 7. 实施与验证

实现文件：

- `desktop/internal/vault/vault.go`：PBKDF2-HMAC-SHA256 + AES-256-GCM、AAD、记录校验、清零。
- `desktop/internal/configstore/store.go`：schema v2、`credential_vault` 单行表、Save/Load/Delete、v1 迁移。
- `desktop/internal/cloud/client.go`：仅进程内 `Credential()` 访问器，不跨 IPC。
- `desktop/internal/ipc/services.go`：`VaultStatus/SaveVault/UnlockVault/ForgetVault`。
- `desktop/internal/nativeui/settings.go`、`controller.go`：Account 页 “Saved credentials” 区块、口令显隐、解锁/保存/删除、离开设置清空口令。
- `desktop/locales/*`、`desktop/internal/localization/ui.go`：中英文文案。

验证结果：

| 检查 | 结果 |
|---|---|
| vault 单元测试 | 通过：保存/解锁/删除、错误口令、篡改密文、AAD 换账号、弱 KDF、弱口令、记录无明文 |
| configstore 测试 | 通过：v1→v2 迁移保留设置、单行替换、非法记录拒绝、0600 权限 |
| ipc 测试 | 通过：保存要求登录、错误口令、账号不匹配、Forget、状态无秘密、数据库与 WAL 无 AK/SK 明文 |
| nativeui 测试 | 通过：保存/解锁/Forget 入口、未登录提示、身份失效后重新加载密文库、离开设置清空口令 |
| `go test ./...` | 通过 |
| `npm test -- --run` | 78 文件 / 392 项通过 |
| `npm run typecheck`、`npm run lint` | 通过 |

未覆盖：真实重启后的解锁与 STS 重新登录、实际 600k 迭代在目标机器上的解锁耗时、实机口令输入与显隐交互。这些需要用户在本机确认。
