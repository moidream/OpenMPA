> 范围修订：本模块只复用已部署OpenAPI，不执行任何后端源码/迁移/镜像开发。当前准入与交付合同以[现有API客户端合同](../2026-10-25-existing-api-client-spec.md)为准。本文旧后端扩展与对应准出要求已失效，客户端部分按当前合同核对。

# M02 连接、身份与传输 spec

状态：已接受，实施中。日期：2026-10-25。类型：功能。风险：L3。草稿按平台接入修订，无issue/MR。需求来源[PRD](../2026-10-25-prd.md) FR-2/FR-12/FR-13及NFR-1/NFR-2/NFR-3；事实含原FACT-4/FACT-5/FACT-13/FACT-14与平台证据（私有研究，不随版本发布）。前置M01/M08。

## 1. 结果与职责

用户通过M08云登录/地域/项目/Runtime选择获得绑定，M02消费其endpoint与入站认证，Go负责业务HTTP及秘密。手填业务地址仅为用户明确的开发诊断入口，不替代平台主流程。原api.ts的sessionStorage token与相对/api fetch退出desktop路径，renderer只消费安全DTO与结果信封。

采用Go net/http以便统一身份头、重定向、幂等、ETag、文件流和取消。MyGo fetch插件可作兼容备选，但不默认暴露通用可带凭据URL代理。loopback BFF不是必需，只有实测无法保留必需组件时再评估。

## 2. 连接配置与秘密

配置包括bindingId、cloudPrincipalScope、region/project/runtimeId、完整normalizedBaseURL、networkType/endpointSource、当前版本、authorizerMode、businessAuthMode、secretRefs与epoch；源自M08，不由renderer自由改host或组合身份。

业务authMode分Runtime JWT、gateway key、legacy JWT和显式local test；云AK/SK/STS只在M08签名平台OpenAPI，不能复制为业务Authorization。ApiKeyName是Key名称，按已核验合同发送Authorization Bearer，不拼任意header。任务legacy身份、Studio管理Key及channelManagementAuth分槽；平台Key存在不证明这些业务可用。

业务authMode细分gatewayAuth/applicationAuth/taskAuth，三者独立，按请求族确定合法组合。只发送该族必需的头，不能全槽广播。

| 网关要求 | 应用/任务要求 | 固定发送方式 | 准入 |
|---|---|---|---|
| Authorization RuntimeKey | 应用key-auth信任网关；任务同模式需另核验 | Authorization Bearer RuntimeKey；subject不由昵称生成 | 目标部署已验证 |
| Authorization RuntimeKey | legacy应用/任务X-Jwt-Token | Authorization RuntimeKey + X-Jwt-Token合法应用用户JWT | 两层合同同时通过 |
| Authorization用户JWT | 应用RuntimePrincipal使用同JWT | 一个Authorization Bearer，issuer/aud/claims/actions均匹配两层 | 双层认可同Token |
| Authorization RuntimeKey | 应用同时要求另一个Authorization JWT | 不拼接、不选其一、不复制到猜测Header | authCompositionUnsupported；需明确可信代理合同 |
| 无网关认证的显式local test | 应用关闭JWT | 显式测试x-user-id等合同头 | 仅用户选择的开发范围 |

Studio/Channel管理Key仅对应管理请求族；taskAuth仍按选定部署legacy claims或已验证新模式处理。平台KeyAuth Location/Name只按官方及目标映射解释，首版Header支持限定为已验证Authorization Bearer。

RuntimeKey来自被用户选中的已授权GetRuntime，Go内部提取/保存；手动补目标Key也为write-only。秘密分槽，Secrets.Set只返回状态，UI/日志不含原值。系统仓库不可用/超过2560字节时明确失败或仅内存，不明文回退。注销取消所选登录/Binding并保留文档，不撤销云资源或后台任务。

云主体、region/runtimeId或业务身份改变建立新scope/epoch。仅同Runtime版本/Key更新时重新Get与业务核验，更新epoch并恢复原会话引用；旧HTTP不污染新状态。IAM UserId/本地Space/昵称不当业务x-user-id；VeADK的Studio owner映射需要其可信认证/代理上下文，不能照搬。

M09管理使用部署明确的X-MPA-Channel-Key或legacy管理JWT；runtime_key owner是可信代理映射，缺来源不启用owner动作。渠道管理授权不替代M03会话权限；扫码binding GET可触发注册，不作为连接测试或只读探针。

## 3. HTTP规范与连接测试

基地址为平台已核验合法HTTPS，无userinfo/query/fragment，保留/runtime/{id}等base prefix。私网显示需VPN/内网，不能自动暴露公网；多候选/矛盾按M08核验。显式local HTTP仅用于独立开发连接，不放宽平台网关。固定路径由Go构造，leading slash resolve不能丢前缀；参数/query单独编码。

认证请求不跟随未知host或身份范围重定向；每跳检查，禁止复制Authorization到不同源。不弱化TLS。GET健康与可见session可证明各自资源存在，不能证明执行、日程或所有动作已授权。

平台Ready不等于业务已连接。测试只读MPA健康/会话/Agent/明确control/config，不调用run/smoke/pause/delete/profile apply探测。不支持MPA合同的通用Runtime标不兼容，不改用/invoke(prompt)或A2A合成会话。每能力标supported/denied/unknown/unavailable并附来源，业务权限后端裁决，M06-E1补握手。

| 结果 | 内部语义 | UI与重试 |
|---|---|---|
| 401 | authRequired | 重新认证，不自动换模式 |
| 403 | denied | 展示动作/资源权限不足，禁止隐式升级身份 |
| 404 | resourceNotVisible或unsupported需契约区分 | 不把所有404当接口不存在 |
| 409/412/428 | conflict/precondition | 提供安全currentState，重读并确认 |
| 410 | cursorExpired等明确code | 交给对应模块恢复 |
| 422/400 | invalidInput | 保留输入，显示安全字段错误 |
| 503 | unavailable | 能力暂不可用；GET可重试，写入不能换键重发 |
| 网络中断/未收响应 | outcomeUnknown（写）或connectionFailure（读） | 写由M03原operation恢复 |

FastAPI `{detail:...}`、schedule `{code,message,requestId,details}`均解析为SafeError。不可读响应按HTTP status分类，不将原HTML、URL、Authorization或完整schema body打印。

SafeError.plane允许platform/gateway/mpa/transport/runtimeIngress，originConfidence为verified/unknown。同Runtime URL的401/403/502不单凭status判断来源，无验证标记使用runtimeIngress/unknownOrigin；仅可信响应header/信封合同才细分gateway/mpa。错误不触发自动换槽或降级，平台HMAC失败与业务入口明确分离。

SafeError经明确结果信封返回：`{ok:true,value}`或`{ok:false,error}`，字段互斥；每个Go绑定方法有对应具体结果DTO。不能直接return Go error承载409/412业务错误，MyGo只传err.Error()，会丢结构化恢复字段。Channel的领域错误用ClientEvent.error，普通IPC失败独立显示transportFailure；renderer不得把两者混为后端拒绝。

## 4. 生命周期与安全

普通短请求接收call context；MyGo绑定方法在goroutine运行，共享client/config加锁。流式订阅用Channel并在close取消HTTP body。UI导航/应用关闭只结束消费；业务取消必须另发经过权限检查的动作。

响应DTO使用字段白名单，尤其`/agents`等配置响应不原样透传。renderer不可设置任意header、service身份、local路径和任意目标URL。请求日志只记录连接摘要、API族、HTTP状态、安全码和operation关联。

## 5. 验收与回退

| 对应验收 | 场景 | 结果 | 证明 |
|---|---|---|---|
| AC-2 | 有效/失效/错误模式/缺动作凭据 | 状态明确；不误用另一槽位 | 伪服务头断言 +测试Runtime读取 |
| AC-12 | 重定向、畸形URL、外部IPC、日志导出 | 凭据与敏感正文不泄漏 | HTTP负向服务、生产WebView验证 |
| AC-13 | 切换连接后旧HTTP迟到 | 旧数据不更新当前状态 | controllable HTTP竞态用例 |
| AC-11 | .dev与production秘密隔离，存储不存在/超限 | 正确隔离和明确失败 | 真系统仓库及失败注入 |
| AC-17/AC-18 | Ready但业务拒绝；STS/Key/JWT分槽；同网关双Runtime | 双状态准确、prefix完整、秘密不串发 | M08/M02合同及真实测试网关 |
| AC-18 | 合法双层auth组合、双Authorization冲突、同status不同/未知来源 | 按请求族发头，冲突拒绝，来源不猜测 | Go认证/错误fixture与测试网关 |

工具链/测试连接由P0解决，真实token写入测试由用户已确认的实施验证范围执行。本轮只完成设计。失败可停用连接模块保留本地编辑，回到旧Web路径时不自动迁入其sessionStorage凭据。
