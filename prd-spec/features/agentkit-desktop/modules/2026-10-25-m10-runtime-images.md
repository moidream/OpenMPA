> 范围修订：本模块只复用已部署OpenAPI，不执行任何后端源码/迁移/镜像开发。当前准入与交付合同以[现有API客户端合同](../2026-10-25-existing-api-client-spec.md)为准。本文旧后端扩展与对应准出要求已失效，客户端部分按当前合同核对。

公开版本已将验收账户、资源与回执标识替换为示例或占位符；实际对象仅保存在本机私有证据中。示例不是可用部署。
# M10 Runtime选择与MPA/Worker镜像更新

状态：已接受，实施中。日期：2026-10-25。类型：功能。风险：L3，测试/远程部署变更。来源：用户本轮明确补充设置页Runtime列表筛选切换、MPA与Worker镜像修改更新。补充原先只读平台范围，不能仍按旧非目标删除用户新增能力。

## 1. 结果与接口

设置页按当前云账户、region/project及veadk:agent-type=mpa筛选ListRuntimes，选择后GetRuntime确认并绑定。切Runtime递增epoch，旧事件不能更新新视图，原执行不自动取消。实际绑定当前生效endpoint，根路径或平台返回的/runtime前缀均保留原意。

新增设置页“部署镜像”：显示选中Runtime ID/current version/ArtifactUrl及其GetRuntime.ToolId关联Worker的GetTool.ImageUrl。不得按名称猜Tool、不得由本地会话scope派生资源ID。测试Runtime实际ToolId=t-example-codex-worker，与AGENTKIT_TOOL_ID环境元信息一致；优先正式ToolId合同，矛盾时拒绝写。

平台统一POST AgentKit OpenAPI，version2025-10-30，service agentkit；Runtime用GetRuntime/UpdateRuntime，Worker用GetTool/UpdateTool。SDK契约证据为本机agentkit SDK runtime/types.py中的UpdateRuntimeRequest与tools/types.py中的UpdateToolRequest，以及veadk/integrations/mpa/mpa_tool.py GetTool/UpdateTool使用。实际字段分别RuntimeId/ArtifactUrl/ReleaseEnable与ToolId/ImageUrl；它们不是MPA业务REST，也不是启动新Bot。

## 2. 写入与一致性

| 操作 | 前置/输入 | 输出与验收 |
|---|---|---|
| 镜像读取 | 授权当前云scope/Runtime；GetRuntime→GetTool | 安全image/id/version/status及review指纹，无Envs/Key/TOS/secret |
| Runtime更新 | 用户选择Runtime并确认目标image、当前版本、review指纹；重新Get核验当前配置 | UpdateRuntime仅写镜像及明确release字段；保留所有无关配置；记录accepted，再轮询当前正式版本/镜像/Ready和业务健康 |
| Worker更新 | 核验Runtime仍关联同ToolId，原image/status/updatedAt指纹，目标image确认 | UpdateTool仅ImageUrl及ToolId；回读GetTool，当前会话不声称已换worker，新的沙箱才按实际更新策略验证 |
| 恢复未知结果 | 原operation ID、目标ID、旧/目标镜像及安全回执 | 不盲重发Update；Get回读判断待更新/已生效/失败/不确定；既有版本不作为自动回滚授权 |

OpenAPI目前未证实支持CAS expectedVersion参数。客户端preflight+readback可以拒绝已观察到的冲突，但不能消除GET与Update之间外部操作竞态；UI和合同不得声明原子CAS。更新只传必需字段，不将Get响应全量复制（含未知字段/秘密）。服务端要求完整受保护配置时先冻结保留合同，不通过删除Envs回避。

镜像输入按OCI registry/repository:tag或@sha256格式验证，不接受URL userinfo/query/空白及任意命令。runtime与worker独立操作，不能把两个更新当云端原子事务；分别显示目标、接受、生效和失败。源镜像标识与版本保留用于人工选择回退，客户端不自动覆盖外部新版本。

## 3. 权限、安全与验证

云Update权限独立于List/Get和MPA聊天权限。云拒绝返回明确平台error，不能换业务Key重试。Go固定方法Deployment.ReadImages/UpdateRuntimeImage/UpdateWorkerImage/ReadUpdate，React不输入任意Action/URL/Envs。凭据与完整云响应留Go，safe DTO及日志只含资源ID/操作状态/镜像引用与安全错误。

| Case | Given/When | Then | 验证 |
|---|---|---|---|
| CASE-IMAGE-READ | 当前账户多MPA，选择/切换读取关联Tool | 正确目录及绑定，旧epoch不串写，Tool来源可追溯 | 平台fixture及真实Get/List/GetTool |
| CASE-IMAGE-UPDATE | 合法测试资源与确认镜像 | 固定Update动作及最小payload，接受与生效分开；未选资源不写 | httptest payload/权限/冲突/未知结果，测试Runtime受控更新与回读 |
| CASE-IMAGE-WORKER | Tool关联改变或旧指纹已改变 | 写前拒绝，保留表单；既有沙箱不误报使用新镜像 | 关联/并发fixture及真实新执行版本检查 |
| CASE-IMAGE-SECRET | 云响应包含Secrets/Envs/TOS | 页面/日志无秘密 | 负向DTO与生产bundle检查 |

实施先RED接口/负控→Go固定平台服务→typed设置页→真实测试资源读取和必要受控更新→新Runtime/Worker运行回读。当前镜像更新尚未执行，未准出。

固定测试验证入口为`desktop/cmd/verify-deployment`，使用与设置页相同的Platform/Deployment服务。默认只读；云写需显式`--apply --kind runtime|worker --target-image --operation-id --state`且固定账户1234567890、上述测试Runtime和正式ToolId，镜像限定同测试registry。Runtime正式发布字段需另显式`--release`。`--read-operation`只回读原持久账本，不重发Update；命令exit0仅说明本次读取/提交成功，不能据此认定Ready或业务健康。没有凭据、未固定testscope或无持久账本拒绝写入。
