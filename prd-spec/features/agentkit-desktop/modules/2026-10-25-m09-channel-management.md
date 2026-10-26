> 范围修订：本模块只复用已部署OpenAPI，不执行任何后端源码/迁移/镜像开发。当前准入与交付合同以[现有API客户端合同](../2026-10-25-existing-api-client-spec.md)为准。本文旧后端扩展与对应准出要求已失效，客户端部分按当前合同核对。

# M09 多渠道添加、配置与设置 spec

状态：已接受，实施中。日期：2026-10-25。类型：功能。风险：L3，外部账号注册、凭据与身份。补充未接受草稿，无 issue/MR。来源：[PRD](../2026-10-25-prd.md) FR-15、NFR-2/NFR-3/NFR-5；事实：Bot/渠道证据（私有研究，不随版本发布） FACT-28–FACT-35。前置M01/M08/M02及合法渠道管理身份；已有会话查看依赖M03，不依赖本地页面。

## 1. 交付结果与能力矩阵

桌面设置提供“渠道与 Bot”页，按当前Runtime管理实际渠道账号，添加后由远程MPA持续接收消息。应用退出不影响远程渠道。飞书多Bot、独立模板及群权限是首批能力；钉钉/企微按当前单配置接入；已有配置时明确显示“替换当前配置”，新账号会清除该平台其他账号/许可行，不显示“新增第二账号”。没有权限或旧部署缺接口时保持明确受限状态。

| 渠道 | 添加 | 同Runtime多账号 | 设置 | 与SessionBot的关系 |
|---|---|---|---|---|
| 飞书/Lark | 扫码或appId/appSecret；shared/private；domain选择 | 已声明支持 | enabled、账号Profile、群权限、诊断、私有用户授权 | 账号默认模板；服务端入站绑定多个独立Session上下文 |
| 钉钉 | 扫码或clientId/clientSecret | 当前不声明支持 | 已核验配置、诊断、按账号删除 | 展示已核验关联；不假设同等Profile/群管理API |
| 企微 | botId/secret手动绑定 | 当前不声明支持 | 已核验配置、诊断、按账号删除 | 不假设扫码、独立Profile或任意会话合流 |

capabilities只表明服务族，动作仍需服务端授权并核验路由。使用现有provider，不把Slack连接或ARK Managed Agent配置移植到此Runtime。需要钉钉/企微多个账号时在E5冻结能力扩展，不通过重复upsert冒充多Bot支持。

## 2. 页面与用户流程

入口沿用Settings & setup，并在Specialist/SessionBot详情提供“所属渠道”引用。顶部显示Runtime、管理身份摘要和能力状态；账号列表显示channel/domain/appId、Gateway bot ID、enabled、shared/private、owner摘要、绑定/诊断状态。身份或账号ID不能用显示名称替代。

添加流程：选择支持渠道 → 选扫码/手动与合法选项 → 明确发起绑定 → 查看二维码/待确认状态 → 服务端注册并回读账号 → 显示结果及可配置项。手动凭据只发送Go主进程，由Go传给指定Runtime业务API；UI不回显秘密，成功后清除表单。服务端当前持久化渠道秘密，本方案只承诺桌面/IPC输出边界，不声称远端已加密。二维码只作为绑定授权材料，禁止进入日志、快照和历史。保留bindingId等安全恢复引用；重启后仅在同身份继续已有操作。

飞书诊断必须携带appId，多账号未选择时可能409；deliveryHealth=not_observed不显示“连接成功”，bindingReady常量不能作为Gateway就绪证明。

配置流程：读账号/模板 → 编辑 → 显示变更影响 → 用户保存 → 按相应并发合同写入 → 回读服务端结果。模板变更影响新Session，旧Session保持pin；不能显示“所有Bot会话均已更新”。群scope改变影响后续路由，可能产生新的隔离会话，旧历史不合并。

启停、解绑账号、撤销私有用户授权、重置私有用户使用独立动作和说明。停用不是取消当前计算，解绑不是删除云Runtime/会话/记忆。删除先核验精确channel/appId和当前归属，提交后回读；历史引用保留可见性状态，不承诺已物理清除。私有用户重置有slot generation变化，旧上下文不能直接作为新用户授权。

## 3. 身份与安全契约

M02增加channelManagementAuth槽：runtime_key配置使用X-MPA-Channel-Key，legacy管理使用合法X-Jwt-Token；外层网关Authorization仍独立。渠道管理Key与Studio Key不能按名字推断等价，按部署合同引用秘密。Go仅绑定固定动作方法，不暴露通用header/HTTP代理。

runtime_key验证共享管理Key，owner动作还需要可信MPA内部用户映射。x-user-id不能由云IAM UserId、昵称、渠道sender或任意输入生成；合法来源为已核验登录/管理代理转发合同。若没有可信owner来源，只开放当前部署允许且不依赖owner的动作，owner Profile/私有绑定明确受限。普通最终用户采用服务端身份代理的路线需E1/E5先提供，不把共享管理Key当个人身份验证。

列表虽已有安全投影，Go仍以白名单DTO返回；SecretStr序列化可能产出原值，错误/原始扫码响应一律不原样跨IPC。appSecret/clientSecret/WeCom secret、用户JWT、Runtime Key留Go与合法服务端。诊断仅返回安全code/阶段/Gateway状态，删除URL query、正文和秘密；未知原始错误不能拼接为可见详情。

当前授权是Runtime级渠道admin：列表、启停和解绑并非统一逐账号owner隔离；Profile/owner/private-user路由另校验owner，binding操作校验所属actor，会话按业务主体独立授权。UI过滤不能提供服务端隔离。普通最终用户若仅能管理自己账号，需要E1/E5追加列表/启停/解绑的逐资源授权合同；首批管理员模式不能声称这一限制已实现。

## 4. 现有接口与并发语义

完整路由见证据HTTP表（私有研究，不随版本发布）。内部IPC建议固定 `Channel.list/capabilities/createBinding/readBinding/retryBinding/saveConfig/setEnabled/permissions/profile/diagnostics/remove` 方法族，输入含connectionId/epoch/channel/appId/bindingId等明确标识，输出使用M02安全Result信封。这些是建议客户端方法名，不是新增后端路径。

| 写入类型 | 输入/默认值 | 一致性与失败处理 |
|---|---|---|
| 飞书Profile PUT | 完整agentsMd/model/skills/mcp + expectedRevision | 409保留草稿并重读，不自动覆盖；null继承pin时默认，[]允许所有可用，不能制作“禁用全部”开关 |
| Profile表单 | agentsMd≤32000；skills/mcp≤200；model≤256且无/或: | 按后端Python字符口径验证，不能套用M04的UTF-16容量口径；不支持role/toolPolicy/memory输入 |
| upsert配置/权限 | 使用当前schema和精确账号引用 | 无已证实CAS，保存前回读展示diff仍不能消除竞态；受控管理员模式提交后回读，不能宣称多客户端冲突保护 |
| enabled PATCH | 飞书appId + enabled | 不把空config重新upsert为启停；其他provider无同等动作时不显示开关 |
| Session Bot升级 | 已绑定飞书账号的sid，无active invocation | 当前接口无expectedRevision/幂等请求，Bot快照变化不推进executionConfigVersion；不自动升级/重发；未知结果只读核验，统一版本保护与定版升级需要E5 |
| 扫码状态/重试 | 服务端bindingId，同owner和provider | GET可推进注册；只有已明确发起的绑定才允许继续；retry是显式动作，409 uncertain先核验Gateway |
| 解绑账号 | 精确channel/appId及合法Runtime管理主体 | 不批量扫删；结果未知不盲重复执行；回读并保持来源引用 |
| 私有用户撤销/重置 | owner校验，先GET slot.version；DELETE query expected_version传用户确认时原值 | 409保持旧slot并重读/重新确认，不自动换最新版本；成功回读新slot/授权状态；不是会话删除 |

## 5. 状态、失败与恢复

客户端显示adding/waiting/registering/registered/failed/unknown等归一化状态，实际来源保留服务端state/code，不把HTTP200直接称注册成功。Binding GET不是只读诊断；它继承用户本次发起绑定的操作意图，连接测试和权限探针不得调用它。

| 场景 | 处理与状态 | 可见信号 | 恢复 |
|---|---|---|---|
| 缺管理Key或可信owner | 不换身份、不假造owner | authentication/owner required | 完成合法认证或E1身份代理 |
| 注册409 uncertain/网络结果未知 | 保留原bindingId，不新建/盲重试 | 注册结果待核实 | 同操作查账号/Gateway；核验后显式retry |
| Profile409/运行中upgrade409 | 保留表单与旧快照 | 配置冲突/执行中 | 重读模板或等空闲后新确认 |
| 配置写失败/并发改动 | 不报已保存 | 保存失败或回读不符 | 回读差异，用户决定下一步 |
| 身份/Runtime切换 | 旧响应丢弃；旧远程操作可能继续 | 新scope重新加载 | 回到原身份核验原binding |
| 部分渠道不可用 | 按provider显示依赖/拒绝 | 已配置与正在接收分开 | 后端修复后诊断回读，不取消其他Bot |
| 删除后历史不可见 | 保留失效来源引用 | 已解绑；历史当前不可访问 | 只用有权业务接口，不借管理Key伪造用户会话读取 |

## 6. 验收、验证与阶段停止点

| 对应验收 | 场景 | 通过结果/证据 |
|---|---|---|
| AC-22 | 三渠道手动、飞书/钉钉扫码、Lark domain、飞书双账号 | 按真实能力完成绑定并回读；同Runtime双飞书账号分别路由，企微无虚构扫码 |
| AC-22 | 飞书enabled/群权限/配置/解绑，应用退出 | 服务端设置生效；退出后渠道仍由远端运行；解绑不宣称purge |
| AC-23 | 模板CAS、null/[]、旧Session、运行中升级及私有slot版本冲突 | 409保留草稿/旧slot；继承/全部正确；撤销/重置用原expected_version，不自动重试最新版本 |
| AC-24 | uncertain/旧binding/重启/切身份/恶意原始错误 | 无盲重复注册；跨owner拒绝；秘密不在IPC/日志/快照 |
| AC-13/AC-21 | 相同用户消息发给两个Bot、不同群scope/私有slot | session引用及profile不串写；旧slot不可当新用户上下文 |

先用Go httptest覆盖请求/安全DTO/未知结果，使用前端状态用例覆盖能力与表单；再于真实MyGo和获准非生产Runtime逐渠道验证绑定与回读。真实扫码注册有副作用，需在实施测试范围中明确；本轮没有触发。消息与渠道可用性由测试用户实际发送、服务端session/回执对照证明，不能仅靠按钮或capabilities200。

P2a交付现有渠道管理与已有有权SessionBot入口，作为首批核心验收；P2b按E5补任意桌面SessionBot创建/绑定，不把M09已存在功能推迟到P5。旧部署或身份不满足时可停在明确受限的渠道设置页，但完整首版验收仍未完成。回退客户端只停管理UI，不删除或恢复远端账号/模板；远端回退必须另行核验当前配置和副作用。
