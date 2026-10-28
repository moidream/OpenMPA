# OpenMPA 全模块设计体系与轨迹完善

## 需求依据与执行边界

用户要求继续完成统一 token 体系、全部软件模块的视觉一致性、轨迹可视化和核心流程验证，验证无问题后准出。沿用已批准的 ZCode 语义体系、OpenMPA teal 和本地 React 组件方案。参考 Codex 的阅读列与稳定输入区、ZCode 的语义层级、deepseek-harness 的渐进执行展示，不引入其私有业务依赖。

工作目录为 `react-composer-codex-style`，保留已有业务改动。主 checkout、其他工作树、已安装应用与远端服务不因本任务被重置、删除或替换。准出是有证据的发布就绪判定，不隐含 commit、push 或部署。密码仅用于授权本地登录，不进入日志、截图输入状态或代码。

## 当前证据与缺口

- `design/README.md` 仍称 v1 为唯一正本，与根 DESIGN.md 冲突。
- 原生 `host/rail.go` 使用固定浅色 RGB，无法由 React 暗色 token 保证一致性。
- token 检查只扫描六个文件且按文本行判定，不能证明全部页面一致性或状态对比度。
- 现有 fixture 的 Settings/Knowledge 为手写预览，不是生产页面，其展示不能代替真实业务验收。
- TracePanel 显式工具定位可能被原来的 turn 选择覆盖；切换 turn 加载期间可能仍展示旧数据。
- 普通会话 Trace 的 404 限制来自历史后端证据，必须重新验证，不把任何 404 一概断言为采集策略。

## 完整验收矩阵

| ID | 需求 | 权威证据 | 当前状态 |
|---|---|---|---|
| DS-01 | 一份语义源定义 Light/Dark、排版、间距、圆角、控件、动效、层级及状态色 | schema、生成器、漂移与颜色对比度检查 | 已实现并通过正反例；生产样式迁移继续 |
| DS-02 | MyGo 壳与 React 使用同源配色与尺寸，支持系统主题与手动切换 | Go 生成文件、原生像素测试、真实窗口 | 生成链与 Light/Dark 切换像素测试通过；新版完整窗口待补 |
| DS-03 | 生产 Account/Runtime/Assistant/Authentication/Deployment/Theme/Language 页面风格一致 | 真实组件浏览器矩阵及本地桌面验收 | 未全量验证 |
| DS-04 | Channels、Schedules、Knowledge 的列表、表单、空态、失败态一致 | 真实组件交互与截图 | 未全量验证 |
| DS-05 | Chat/Markdown/Composer/执行过程/Trace 语义统一且保留真实顺序与隐私边界 | 投影测试、交互与真实会话 | 已有部分证据 |
| DS-06 | 轨迹支持跨 turn 定位、树/时间轴、筛选、详情、键盘关闭与焦点恢复 | 行为回归与浏览器交互 | 待完善 |
| DS-07 | 轨迹缺失、权限、错误、运行中、截断有准确反馈，不伪造统计或因果 | 故障注入、后端真实返回 | 待完善 |
| DS-08 | 全页面 Light/Dark、320/768/1280、200% 缩放、键盘导航、减少动态效果 | 生产组件自动矩阵，溢出及可访问性结果 | 未全量验证 |
| DS-09 | 本地登录、Runtime 选择、对话及核心知识/计划/渠道流程正常 | 授权桌面真实操作；写操作仅用隔离测试对象 | 待验证 |
| DS-10 | 所有适用构建、类型、测试、设计与格式检查通过，两轮复核，无未决必要项 | 当前源绑定报告与结果 | 待最终准出 |

## 实现顺序与架构

先完成 DS-01/02，token 生成器输出 CSS 与 Go。原生适配器消费生成的类型化 Palette，禁止复制 RGB；React 延续本地 UI 组件。随后迁移生产样式并用生产组件替代手写预览做验证，最后完善 Trace 交互与真实桌面验收。

保留既有 v2 路径避免迁移破坏，但将其明确为当前项目级契约。v1 保持历史文件。新增状态颜色拥有配套前景色，检查每个 Light/Dark 键集合、尺寸类型、引用和文本对比度。设计检查要覆盖实际加载样式及 Native consumer，不能以扫描少量新文件声称全量达标。

Trace 读取绑定精确 epoch/session/invocation。显式定位优先于自动跟随；用户切换历史 turn 后仍能继续浏览；切换身份或 turn 时旧详情立刻失效；迟到的请求不得回写。404 只能证明数据不可用，不能凭状态码猜测具体后端配置。

## 执行检查点

- 前置约束：用户已要求自主执行。规范与真实缺口先核对，随后小步修复。
- 独立工作：token/原生适配、生产页面、轨迹、验收按模块分片。
- 共享状态：同一脏工作树由当前 agent 串行修改，不启用子代理，不覆盖未拥有改动。
- 最小完整单元：同源 token 输出必须同时验证 CSS 和 Go；轨迹状态变更必须配正反例。

## 回滚、验证与准出

本轮没有持久化数据模型迁移。回滚使用任务拥有文件的差异，不 reset 整个工作树。先跑每个单元负控与聚焦测试，稳定批次再跑全量 `npm test -- --maxWorkers=2`、typecheck、lint、Desktop build、Go tests/vet/build、design:check 及实际桌面矩阵。任何缺失的真实业务证据保留未验证，不判定完整准出。

## 2026-10-28 验证进度与未准出项

- 完成 JSON Schema、CSS/Go 同源生成、状态独立前景色、对比度校验、原生高对比度/字体缩放适配和原生像素测试。生产 Desktop 主样式直接使用语义值，CSS 检查改为 AST 并校验引用。
- 修复 Trace 跨 turn 工具定位、加载期间旧详情混显、迟到响应、实时刷新、Escape/焦点恢复和树形键盘操作。修复 token 总量重复累加、截断轨迹虚构计数、缺失结束时间无限累加。相关负控和回归通过。
- 生产组件在模拟 MyGo 传输下完成 12 页面 × 2 主题 × 3 宽度矩阵，72 个场景无根文档溢出、无页面异常，主按钮颜色匹配 token。真实错误/空态、200% 缩放和新版原生完整窗口尚未全部验证。
- Node 24 下全量 Vitest 为 90 文件 / 459 测试通过。typecheck、lint、Desktop build、design:check、DESIGN lint、相关 Go tests/vet 通过。最终全量 Go/race、全仓格式与完整准出审计尚待收尾。
- 在安装应用数据库的私有快照上通过真实 Go Vault/Platform 服务完成解锁、登录；实时目录返回 36 个 Runtime；按保存的 ID 显式绑定成功；已有会话列表 2 项、选中历史 32 条事件且完整。未提交新 run 或修改远端资源。
- 同一真实验证中，渠道返回 `authRequired`，计划列表返回 HTTP 422，知识列表返回 HTTP 500，普通 turn Trace 返回 HTTP 404。证据在私有 `.openmpa/design-live-read-report.ndjson`。这些结果不等于核心流程通过。
- 已读取参考后端当前源码：计划接口使用独立的 scheduled-task identity；Trace 路由只检索 admin_debug session。它们帮助定位，但不证明部署端版本一致。需继续核对部署契约与真实错误，不能凭 HTTP 状态猜测全部根因。

下一单元：完成未覆盖状态与原生窗口验收，核对真实 Runtime 契约和身份前提，定位计划 422 与知识 500；保持真实 Trace 404 边界，未满足完整矩阵前不准出。

## 2026-10-28 聚焦修复

- Assistant Settings 从普通文档流 section 改为 UI kernel 的原生 `dialog`。Chat 和 Settings 复用同一组件；弹窗不改变 Chat 尺寸，Escape/遮罩/关闭按钮恢复原焦点。
- 弹窗宽屏为双栏，760px 以下单栏；共享 36px 字段、token 间距、状态色、按钮和唯一 teal focus ring。Compound model search 禁止内外双 focus ring。
- 下一 Turn 模型选择按 Runtime execution-config revision 更新，仅运行中锁定。旧普通 Session 首次明确选择时按需初始化；普通新对话不再创建 admin-debug 会话。
- Stop 状态按 exact invocation 重新核验，覆盖任务恰好结束、终态迟到、abort false 仍 active 和有界超时。
