# 原生壳层与设计复核

范围：用户暂停整体目标后，单独授权顶部对齐、保留快捷栏、侧栏原生化及设计优化。Runtime/Worker源码不变；本次不执行云端业务测试。使用项目局部安装的emilkowalski/skills（apple-design、emil-design-eng）。

## 设计复核

| Before | After | Why |
| --- | --- | --- |
| 标题栏下另有当前位置行 | 44px原生顶部行统一位置与操作按钮 | 减少重复层级，按钮和位置垂直中心对齐 |
| 全部目录挤在顶部 | 48px快捷栏与240px业务侧栏纵向分组 | 保留原产品使用习惯，内容区有独立空间 |
| WebView导航按钮 | MyGo ui原生壳层和虚拟目录 | 导航不依赖DOM布局和页面重绘 |
| 长标题和大量行缺明确边界 | 单行省略、提示、最多1000项安全投影、虚拟列表 | 保留按钮可达性并限制绘制工作 |
| 深色主题导致选中图标白色 | 原生壳层浅色主题，图标显式深灰 | 保证浅色背景下对比度 |
| 纯静态按钮反馈 | 120ms颜色反馈、按下即时响应 | 清楚反馈且避免高频装饰动画 |

## 两轮自审

第一轮：类型、生成绑定、编译、布局位置、固定动作、safe DTO及原UI入口检查。WebView保留聊天/富文本/其它内容页面，未宣称所有设置和日程表单已原生化。原生按钮是MyGo ui绘制，交通灯为AppKit原生控件。

第二轮：检查Session/Space目录成员、草稿保护、选中状态、重复ID、迟到ShellState、按钮与原生窗口事件、MyGo ui虚拟列表Key、系统ReduceMotion和频繁导航。发现List初始化后Key错误，已改为Context.Key；发现迁移后的验收不能依赖DOM侧栏或仅MutationObserver，已改为真实原生控件与resize事件。没有恢复后端开发或删除工作树。

仍需交付整体目标：其它适合原生的内容表单按设置→渠道→日程→知识/产物渐进迁移，复用现有Go服务和OpenAPI；聊天和富文本WebView继续保留。该后续工作属于暂停的整体目标，未完成项不能冒充当前原生壳层完成。

## 后续原生迁移切片

| 切片 | 原生界面 | 必须保留的契约与验收 |
| --- | --- | --- |
| N1 账户与Runtime设置 | 账号、安全输入、Runtime目录、资源与image表单 | 复用Platform/Connection/Deployment/RuntimeConfig；原operation回读，账户/epoch隔离 |
| N2 渠道管理 | 账号列表、配置、权限与诊断表单 | 现有Channels typed服务；原绑定回读、凭据只入不出 |
| N3 定时任务 | 列表、编辑、启停、历史和运行详情 | 现有Schedule；创建/执行/删除精确ID，退出客户端后调度 |
| N4 知识与产物 | 目录、导入/搜索表单、引用/文件保存 | 现有Knowledge/Artifact，原生选文件；长文本预览可保留WebView |
| N5 内容交互 | 适合的工具栏、原生prompt/contextmenu | 保留会话、pending、取消及草稿保护，聊天与富文本保留WebView |

用户目标暂停期间，这些切片作为当前方案中的待实现项；本次只验证原生壳层与设计交互，不触发云端测试。任何原生迁移均不修改Runtime/Worker源码和镜像构建流程。

## 当前验证结果

- pass：npm全量75文件382测试；typecheck、lint、desktop frontend build。
- pass：Go全包test；host/ipc race；GOWORK=off host test与Go vet；MyGo0.6.4绑定生成、macOS arm64 app/dmg候选构建。
- pass：原生MyGo headless测试布局（44px顶部、48px快捷栏、240px业务侧栏）、长中英文标题对齐与无重叠、六快捷动作、Session/Space精确ID与创建动作、侧栏折叠后快捷栏保留。前端测试验证原生导航投影、单调版本、精确目录成员与会话选择保留。
- pass：真实ad-hoc签名NSBundle run20261025T164535.885617Z中原生AppKit交通灯、WebView无网页导航、原生辅助功能控件press→NavigationRequested→前端→ShellState、折叠/展开后的1232/992px内容宽度、原生Space选择、中文富文本/Markdown编辑保存及持久重开。辅助功能输入与处理待绘制的AppKit帧为验收方法；未称物理鼠标输入验证。失败的旧尝试全部保留，不当通过证据。
- blocked：该run导出保存阶段系统检测desktopLocked=true，code=blockedDesktopLocked；不绕过锁屏，不称导出/取消已通过当前最终构建。

整体用户目标仍paused。本轮只交付UI调整候选与证据，无云业务调用、无Runtime/Worker代码或云image变更，无提交/推送/发布。app/dmg为本地预览候选，不是全模块准出；其它内容表单的原生迁移待后续切片。
