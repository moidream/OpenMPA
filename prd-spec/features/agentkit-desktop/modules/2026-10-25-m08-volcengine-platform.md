> 范围修订：本模块只复用已部署OpenAPI，不执行任何后端源码/迁移/镜像开发。当前准入与交付合同以[现有API客户端合同](../2026-10-25-existing-api-client-spec.md)为准。本文旧后端扩展与对应准出要求已失效，客户端部分按当前合同核对。

# M08 火山AgentKit平台接入 spec

状态：已接受，实施中。日期：2026-10-25。类型：功能。风险：L3，云账号身份和Runtime秘密获取。新增模块补充未接受草稿。需求来源：[PRD](../2026-10-25-prd.md) FR-12/FR-13、FR-2、NFR-2/NFR-3/NFR-5；事实：平台证据（私有研究，不随版本发布） FACT-16–FACT-27。前置M01/P0；M02消费本模块绑定。

## 1. 结果与非目标

桌面通过火山AgentKit平台OpenAPI发现已部署的agentkit-mpa-agent，固定账号/地域/项目/Runtime绑定，然后通过平台返回的入站网关连接其业务API。默认入口由手填Base URL改为平台资源选择。

消费既有Runtime；用户追加授权的Runtime/关联Worker镜像更新按M10执行，不创建或删除云Runtime，不自动创建SSO/用户池/角色，不执行沙箱命令。模型/知识/记忆/会话资源的云ID可显示安全关联摘要，但不能把平台对象CRUD当MPA业务数据CRUD。普通个人文档仍由M04保存。

## 2. Go组件与认证槽

新增CloudCredentialProvider、AgentKitOpenAPIClient、RuntimeCatalogService、RuntimeBindingResolver；这些是拟新增消费组件，不声称已有Go API。复用可信官方Go签名组件，在P0核验版本后锁定，不移植Python SDK进生产运行时。若无匹配生成客户端，使用官方签名器+明确类型JSON合同，避免手写加密算法。

| 槽位 | 凭据/输入 | 可用范围 | 不可替代 |
|---|---|---|---|
| 云平台 | AK/SK，临时STS另含SecurityToken与到期时间 | AgentKit OpenAPI、可授权只读账号核验 | 不作为Runtime Bearer/TIP |
| 云SSO | 管理员发放的可信发现地址、public client元信息，Go持有refresh材料 | 登录并换短期STS；业务JWT用途需单独验证 | 不把普通火山控制台Cookie作为API认证 |
| Runtime KeyAuth | 当前目标GetRuntime的已授权KeyAuth.ApiKey，或用户明确提供的目标Key | 当前Runtime入站网关，实际Header映射须核验 | 不作为平台AK/SK或业务用户身份 |
| Runtime JWT | 该Runtime authorizer认可的aud/client与claims的Token | 当前Runtime入站与应用RuntimePrincipal | 云OIDC ID Token不自动符合应用claims |
| 业务用户/任务身份 | 服务端认可的subject映射或legacy任务凭据 | 会话/日程实际owner范围 | 不把Cloud IAM UserId、昵称或本地Space拼成subject |
| A2A TIP | 工作负载/代理单独授予的TIP | 已验证A2A任务通道 | 不用key-auth或REST turnId替代 |

Renderer只提交明确登录配置，不接收SK/STS/refresh/RuntimeKey原值；普通DTO仅返回hasCredential、expiresAt、认证模式与脱敏scope。Secrets超限处理沿M02，不将多类秘密拼入一个无界JSON值。云STS建议仅内存使用，refresh材料放系统仓库；凭据刷新串行化，临近到期时只读请求可刷新后重试，业务未知提交仍由M03恢复，不能换键。

## 3. 登录路径与授权

推荐条件性SSO：企业已提供发现地址/public client/OIDC与STS role信任时，Go校验HTTPS/issuer/discovery允许host，生成state/nonce/PKCE，调用系统浏览器登录，经受控loopback回调交换并验证Token，再换STS。回调listener仅用于本次登录，随机state单次消费、超时关闭；不为业务API额外起BFF。具体端点/字段由目标发现文档与SDK源码P0核验，不虚构OAuth API。

STS只证明云IAM权限，不能自动证明会话subject或管理员权限。将用户Token当Runtime JWT前必须匹配Runtime authorizer与应用RuntimePrincipal claims/actions；不满足则显示业务认证未完成，而不是降级Key或伪造用户头。

云主体核验使用SDK已实现的STS GetCallerIdentity，service=sts、Version=2018-01-01、host=sts.volcengineapi.com、POST空body签名，读取AccountId/IdentityType/UserId等实际返回字段。SSO换STS是另一操作AssumeRoleWithOIDC，令牌在表单body、非AK/SK签名；不将它列为普通资源读取。具体见平台证据FACT-27。

cloudPrincipalScope明确含accountId、identityType、cloudPrincipalId、可验证roleRef、loginActorRef和identitySource。账号/云主体取GetCallerIdentity；loginActorRef仅来自已验证OIDC issuer+subject，不能昵称代替。角色会话TRN不当稳定用户ID：只有证据支持时归一到角色，再保留OIDC actor区分共享角色用户。无法得到稳定主体时使用独立credentialProfileId namespace，标unknown，禁止自动合并会话。STS刷新同已验证actor/账号保持namespace、刷新epoch；切账号或业务subject改变新namespace，凭据指纹仅诊断。

个人/研发兼容路径为用户明确输入受限AK/SK或完整临时STS三元组，使用官方签名只读核验账号和资源。云长期Key不得内置在包、页面或配置模板。无SSO部署条件时继续提供这一明确路径，不由桌面自动sso-setup创建云资源。

面向普通最终用户且不希望持有云管理权限时，可后续接可信服务端代理：代理持IAM角色STS，按已登录用户筛选Runtime并代表其连接。代理不是首版隐式部署；需要独立资源授权合同且仍使用火山OpenAPI。更详细用户定位见OPEN-7。

## 4. OpenAPI与资源目录

平台调用固定HTTPS、POST，query Action/Version=2025-10-30，请求含UTC X-Date、X-Content-Sha256及临时凭据的X-Security-Token；签名credential scope使用明确region和service=agentkit，不能据此新增字面Region/Service header。使用官方签名器，fixture验证最终wire与canonical结果。官方地域地址为agentkit.cn-beijing.volcengineapi.com及agentkit.cn-shanghai.volcengineapi.com；SDK旧默认open.volcengineapi.com仅作验证后兼容profile，不自动切host/地域。

默认列表请求ProjectName和显式MPA TagFilters，再MaxResults/NextToken分页；分类依据veadk:agent-type=mpa。veadk:managed不是所有用户自部署MPA的必需标签，不强制其为true排除用户目标。名字或镜像命中不能认证分类。

无标签Runtime可通过用户明确选定的可见ID进入“待验证兼容”流程，安全读取其MPA业务合同确认；不自动改标签、不冒充标准MPA目录。Tag过滤不受平台版本支持时，使用有界分页读取并按显式返回Tags本地分类，展示分页完整性；权限失败不能当空列表。需要Tag服务补查时单独声明只读权限和来源，不为获取标签申请云写权限。

GetRuntime读取选中资源；结构化NotFound、访问拒绝、缺认证、平台不可用分别表示。参考VeADK在某些服务角色下List可见而Get返回NotFound；如采用连接专用fallback，只能复核精确ID的可见List记录及同scope认证，状态标为“元信息读取受限”，不能当作完整管理权限或认定资源删除。多条精确ID匹配/分页不完整时不建立绑定。

版本页面只读GetRuntimeVersion/ListRuntimeVersions，始终区分API Version、CurrentVersionNumber、VersionNumber=0 Latest及MPA profileRevision。首版连接当前生效Endpoint，不自动增加平台调试header、切Latest、发布或回滚。

平台响应先识别ResponseMetadata.Error/Result；返回秘密在Go立即分离，云列表缓存只存白名单摘要。API Key、Envs.Value、TOS credentials、FailedLogFileUrl签名参数、原始状态日志均不暴露。平台读取也不能把HTTP200+缺Result误报成功。

## 5. RuntimeBinding合同

建议RuntimeBinding字段：schemaVersion、bindingId、cloudPrincipalScope、region、projectName、runtimeId、gatewayMode、currentVersionNumber、networkType、endpointSource、规范化完整endpoint、authorizerMode、credentialRef、businessPrincipalScope、业务protocol、observedAt、epoch。Go内部可持完整endpoint，UI日志只显示脱敏host/path摘要与安全resource ID。

多个public endpoint不可按数组首项任意选择，显示安全候选并让用户选或按已验证部署策略决定；默认使用public HTTPS。只有private时显示需VPN/内网访问条件，不自动改网关/暴露公网。顶层Endpoint只在network配置缺失且可证明映射时使用，字段矛盾进入需要重新核验状态。不能自动把官方样例HTTP域名当可发秘密地址；HTTPS可达性由P0核验，不探测跨协议认证重定向。

共享网关的/runtime/{id}base prefix必须保留：拼接/api/v1/sessions、/a2a/jsonrpc时不得使用会丢base path的leading slash URL resolve。binding隔离至少account/principal scope + region + runtimeId + project/scope；两个Runtime共用hostname仍是两个连接。平台API Version与网关Path彼此独立。

KeyAuth的ApiKeyLocation/ApiKeyName为平台认证配置，Location指传递位置，Name语义按目标合同核验；不能仅从显示名称猜Header名称。已核验SDK0.8.9/官方调用例子使用Authorization Bearer，首版只支持该已验证映射；自定义传递形式无明确合同则authorizerUnsupported。网关auth与应用auth分开，按M02允许组合表构建请求，不能将所有秘密头全发。

JWT认证没有可从GetRuntime取得的用户Token。按明确登录合同获得，校验issuer/aud/client与应用claims兼容；KeyAuth缺Key则让用户补合法材料或平台管理员处理，不能读Runtime Envs拼秘密。资源读权限与业务动作分开。

## 6. 业务连接和版本变化

用户明确连接Runtime后，M08生成绑定交M02。M02检查合法认证的MPA业务GET、配置和事件合同；通过后M03运行同一binding下会话。官方/invoke(prompt)和SDK invoke可用于理解通用调用，不作为MPA会话CRUD/SSE/profile/幂等替代路径，也不创造平台InvokeRuntime Action。

平台Ready但缺Endpoint/authorizer、访问不可达或业务不兼容分别展示；入站同URL的401/403没有可信来源标记时只显示runtimeIngress/unknownOrigin，不猜来自网关或应用。只有明确响应信封/验证header合同才细分来源。平台Key获取不消除任务legacy身份缺口，E1仍负责业务身份/精确控制。

Runtime版本/endpoint/Key变更时重新Get当前元信息并校验绑定，epoch递增；旧HTTP/stream结果不得进入新scope。仅版本更新而身份/Runtime ID未变时保留会话来源引用，再从权威历史恢复。旧Key401可只读刷新当前binding，不能下载历史版本Key绕过撤销。云STS过期只影响管理刷新，不自动删除已授权的业务历史或取消远程执行。

## 7. 可观测性、失败与回退

| 场景 | UI与行为 | 恢复 |
|---|---|---|
| SignatureDoesNotMatch/InvalidTimestamp | 云认证/签名时钟错误，不切业务凭据 | 检查时钟、Region/Service/固定签名配置 |
| InvalidSecretToken | 平台登录过期 | 正常刷新STS或重新登录，不重发未知业务操作 |
| LackPolicy/LackProjectPolicy | 当前账号/项目读取被拒绝，不显示空目录 | 管理员分配既定只读权限 |
| InvalidActionOrVersion | 协议版本不支持 | 显示准确Action/Version，P0重新核验 |
| FlowLimitExceeded/网络错误 | 保留上一安全摘要、显示刷新失败 | 有界只读退避；NextToken/查询不变 |
| 分页未完成/重复token | 显示“部分结果”，停止循环 | 用户继续/收窄筛选，不宣称完整 |
| Ready但业务不可用 | 平台状态与业务状态分开展示 | 控制台/诊断；禁止自动部署/smoke |
| 资源被移项目/删除/变认证 | binding失效或待复验，保留文档/来源引用 | 同资源重新核验；不按同名新Runtime替换 |

日志只保存plane、action、region、binding摘要、状态/安全码、当前版本与本地operation关联；不保存原始云响应、签名、Token或环境变量。控制台外链由受控host和安全region/runtime/project生成，系统浏览器打开，无自动点击创建/在线测试。

回退可停用某平台能力而保留本地Page导出；用户明确配置的受限直连用于开发诊断，不能作为平台OpenAPI主流程失败后的静默替代。云资源、SSO与账号状态不由桌面回滚修改。

## 8. 验收与追踪

| 对应验收 | Given/When | Then及验证方法 | 环境 |
|---|---|---|---|
| AC-16 | 合法测试云凭据，按地域/项目筛选与分页 | 返回相同真实MPA集合，MaxResults/NextToken正确；无秘密UI字段 | 签名fixture/官方API测试账号 |
| AC-16 | 无权限、token过期、错误签名、HTTP200平台Error | 不显示假空或成功；安全错误保留plane/action | Go伪OpenAPI与授权只读云核验 |
| AC-17 | 两Runtime同网关host不同prefix；用户选择其中一个 | 会话访问绑定正确Runtime，路径完整，不串缓存 | Go合同测试+真实MyGo/测试Runtime |
| AC-17 | Ready但无公网/缺Key/JWT不兼容/业务拒绝 | 平台与业务状态分开，未自动换身份或发布 | 可控响应+已配置测试环境 |
| AC-18 | 已配置SSO或完整STS；重新登录/刷新/切账号 | 原子更新凭据与epoch，无越权和秘密泄漏 | auth合同测试；SSO条件具备时实机 |
| AC-18 | 日志导出/列表/详情/版本查看 | AK/SK/STS/RuntimeKey/Envs/TOS秘密不回传 | payload负控+实际生成DTO检查 |
| AC-18 | 网关Key+应用legacy JWT组合；双Token争用Authorization；同status未知来源 | 合法头映射通过，争用无代理合同拒绝；未知来源不自动换凭据 | auth组合与响应fixture；测试网关 |
| AC-19 | 当前正式版本与Latest不同，打开版本/控制台 | 默认绑定正式运行资源，未执行发布或切流；外链无token | 类型fixture+真实平台只读对照 |

FR-12→§3/§4/§5→AC-16/AC-17；FR-13→§3/§7→AC-18/AC-19；FR-2/NFR-2/NFR-3→认证槽/Binding→AC-17/AC-18。覆盖为文档覆盖，所有实施测试待执行。

P0核验签名器/SSO合同/真实地域和Runtime，P1实现宿主与平台读取，P2在所选binding上实现业务会话。SSO不可用时受限AK/SK/STS路径仍需用户选择，不缩减“平台OpenAPI发现”要求。发布前各平台真实登录/秘密仓库检查；当前仅方案设计。
