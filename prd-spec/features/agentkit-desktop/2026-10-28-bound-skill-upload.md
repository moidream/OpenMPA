# Runtime 绑定技能空间与上传

日期：2026-10-28

## 交付行为

资源页使用紧凑的分类导航、空间选择、搜索、刷新及上传工具栏。空间显示真实名称和 ID，技能显示名称、描述及返回的版本；分页仅在有多页时显示。搜索过滤当前页。资源无需先选择会话，供 Runtime 的 Worker 使用。

通过 Runtime Agent Card 的资源拓扑中 agent → mounts → configured skill-space 识别绑定 ID，再调用 GetSkillSpace 读取真实名称、项目与描述。技能优先读取 ListSkillsBySkillSpace；当关系目录为空或接口不存在时，检查同一绑定 ID 的 ListSkillsBySpaceId / GetSkillInfo，避免把旧目录技能误显示为空。鉴权及权限错误直接展示，不降级为空列表。

上传目标只能是当前 Runtime 拓扑中已挂载的空间。支持 ZIP 或单个 SKILL.md；后者由宿主包装为 ZIP。宿主验证大小、路径、文件类型及 YAML 元数据，拒绝覆盖同名技能。ZIP 与解压总大小均最多 20 MiB、有效文件最多 100 个，需在根目录或单层包装目录包含 UTF-8 SKILL.md，name 和 description 规则与 Studio 一致。

上传复用登录云账户的 AK/SK / STS，使用官方 TOS Go SDK；按账户、项目、区域派生私有桶及内容寻址对象。查询桶所有权、必要时建桶并上传，再调用 CreateSkill → ListSkillVersions → PublishSkillToSkillSpace。需要账户拥有相应 TOS 和 AgentKit 权限。存储地址、凭据不返回前端。

发布成功后刷新当前空间。初次上传期间禁止重复提交；处理阶段可关闭弹窗继续轮询。发布响应丢失后仅查询关联关系确认结果，不重复发起发布；CreateSkill 结果未知时阻止自动重建。

## 范围与限制

不创建或变更 Runtime 的 Skill Space 绑定，不调整 Main 配置。上传进度记录保存在宿主进程内，重启应用后无法恢复本次操作记录；创建结果未知且未得到技能 ID 时，需要在云端核实，客户端不会猜测成功。失败上传产生的私有 TOS 对象不会自动删除。

## 验证

- 544 个前端测试通过，包含绑定目标、上传发布、过大文件及结果未知时禁止重复提交。
- Go 全量 race 测试与 vet 通过，包含空间真实身份、旧目录兼容、包校验、目标约束、存储签名与发布响应丢失后的只读恢复。
- TypeScript、ESLint、格式、设计令牌检查及桌面前端构建通过。
- 本地浏览器使用实际 React 组件和模拟数据检查列表、搜索及上传弹窗；未向真实云空间上传或发布技能。当前用户 Runtime 的实际绑定与目录数据仍需连接后验证。
