> 范围修订：本模块只复用已部署OpenAPI，不执行任何后端源码/迁移/镜像开发。当前准入与交付合同以[现有API客户端合同](../2026-10-25-existing-api-client-spec.md)为准。本文旧后端扩展与对应准出要求已失效，客户端部分按当前合同核对。

# M01 桌面宿主 spec

状态：已接受，实施中。日期：2026-10-25。类型：功能。风险：L3，承载可信前端与宿主接口。首次设计，无 issue/MR。需求来源：[PRD](../2026-10-25-prd.md) FR-1、FR-11、NFR-1、NFR-2、NFR-4、NFR-5。事实：证据（私有研究，不随版本发布） FACT-2/FACT-3/FACT-13/FACT-14。

## 1. 结果与范围

当前React工作台在MyGo系统WebView中运行，保留CSS、组件和导航策略。运行不需要Node/Bun/Python服务；开发仍可使用现有Web模板检查页面。此模块只提供宿主、生成绑定和生命周期，连接API由M02负责。

前置P0确认Go工具链、MyGo固定revision、Node构建工具及系统WebView。首批默认macOS arm64，最终各平台单独准出；不能用交叉编译通过代表GUI已验证。

## 2. 构建与依赖方向

新增桌面Go消费模块、入口、mygo配置、frontend-only dev/build脚本和生成客户端位置；这些是拟新增结构，不是现有文件。开发可用local replace或workspace指向用户指定MyGo checkout；发布固定经过验证的版本/revision，不依赖用户机器绝对路径。

- 保留`vite.config.ts` React插件与dark-theme构建逻辑，frontendDist对应现有`dist/client`。
- frontend-only开发只启动Vite；现有`npm run dev`会同时启动Node server，不直接作为desktop devCommand。
- 主窗口加载相对`/`，正式静态产物嵌入binary。SPA/hash导航均保留，不把`/api`相对fetch当成后端。
- 生成TypeScript绑定仅引用安全服务，生成物与Go服务版本匹配；CI校验再生成无差异。
- 生产不包含OpenDots后端、browser进程或第三方Intelligence运行配置；Web路径在渐进阶段仍可运行。

MyGo消费使用CGO_ENABLED=0。需要更改MyGo框架时另立跨仓库spec并遵循其AGENTS的二进制大小/多平台规则，不把客户端适配自动升级为框架修改授权。

## 3. 窗口、菜单与关闭流程

初始采用系统标题栏与单主窗口，恢复尺寸和布局。导航维持Space→页面、Specialist→会话、任务和设置；原生菜单提供普通编辑角色、保存与退出。应用内快捷键由单一分发负责，避免菜单与编辑器重复触发。

宿主先定义可异步的closePending/allow/block关闭协议，P1用无脏稿和受控回调验证，不依赖M04存储。P3接入M04后询问dirty/saving/conflict草稿；flush成功或恢复草稿落盘后才完成；失败提供继续编辑、导出副本、明确放弃。页面beforeunload只作兜底，宿主等待异步服务不阻塞主线程。完整草稿关闭验收属于P3，避免M01与M04准入依赖成环。

首次关闭主窗口建议实际退出应用；常驻tray、登录启动和全局快捷键不属于首批，后续按产品确认。退出断开订阅和下载，不取消后端run或日程；有运行时提示“远程任务继续运行”且与真正取消操作区分。

## 4. 信任与原生能力

MyGo默认trusted包含file/about/custom schemes；应用仅加载自有构建前端与明确dev URL，不把下载HTML、登录页面或Computer外部内容装入绑定窗口。

应用和网页两侧导航均校验，外链交给系统浏览器。禁止追加任意TrustedOrigins或通用Shell执行接口。原生文件操作需CallerWindow关联与用户选择；取消对话框返回无授权结果，不在其后保存默认目录。页面上下文取消不保证对话框关闭，拿到迟到选择结果后也要检查窗口/epoch是否仍有效。

静态页面生产开启合适的资源策略；调试工具只在开发使用。渲染Markdown默认不执行HTML/脚本；外部资源不携带连接凭据。

## 5. 验收与验证

| 对应验收 | 场景 | 可判定结果 | 证明方式 |
|---|---|---|---|
| AC-1/AC-11 | frontend-only构建、实际启动安装包 | 工作台完整加载；不存在Node/Bun服务依赖 | MyGo窗口截图、进程清单、干净环境启动 |
| AC-14 | 中文IME、富文本粘贴、Cmd/Ctrl+S、窄窗口 | 不误提交、保存仅触发一次、焦点与导航可达 | 实际WKWebView交互；其他平台重复 |
| AC-12 | 外部http/file/about导航、恶意iframe调用绑定 | 非自有内容无Go调用和凭据访问 | 生产构建负向操作及安全错误回执 |
| AC-13 | P1受控异步关闭回调；P3接入dirty/conflict草稿 | P1能阻止/允许关闭；P3真实草稿保存/恢复或明确放弃 | 原生关闭协议及P3存储故障注入 |
| AC-15 | desktop路径停用 | 旧Web命令保持可运行 | Web构建与基本导航 |

未来验证命令：现有`npm run typecheck/lint/test/build`按改动执行；frontend-only脚本落地后单独调用。Go消费模块用`CGO_ENABLED=0 go build ./...`与`go test ./...`，再MyGo CLI的generate/dev/build。本轮未运行这些命令，工具链不足不能绕过版本检查。

## 6. 失败与回退

缺WebView或前端加载失败显示可定位错误，不启动Node替代后端。实际编辑不兼容时保留Markdown源码模式并修复必要组件；无法满足必需行为时停止P1，重评DEC-1。绑定生成失败/漂移阻止发布。桌面路径可独立停用，生成物和应用数据保留，不删除任何worktree。
