# 桌面品牌与对齐验收

验证范围：当前工作区的 B1–B3，2026-10-26，macOS arm64 / MyGo 0.6.4。全项目准出仍遵循原有验收台账，本次仅完成桌面外观修复及本地预览。

## 两轮复核

| Before | After | Why |
| --- | --- | --- |
| Divider 的 AlignSelf(Stretch)，实际 Y=0 | AlignSelf(Center)，实际 Y=13.5、H=16 | 与 32px 折叠按钮同轴；44px 标题栏的 1px 底边使内部中心为 21.5px |
| 首页使用字体字符 O· | 同源、嵌入的单色 SVG，仍为原生 36px 按钮 | 高分屏清晰，保留键盘、可访问名称与 home 操作 |
| Bundle 没有自定义图标 | 顶层 icon 指向 1024px PNG，MyGo 生成 AppIcon.icns | Dock/Finder 使用与原生标志一致的几何图形 |

第一轮：确认 SVG 解析、Go 编译、按钮行为、标题栏布局、PNG 尺寸/透明度、Info.plist 声明和图标打包。无阻断项。

第二轮：检查长位置/中文标题在 640/760/1200px 的展开和折叠布局、真实交通灯 72px 左侧预留、分割线相邻间距、16/32/64/128px 辨识、品牌节点与业务状态的区分、现有助手插画保留及应用包标识兼容性。无阻断项。

## 证据

| 检查 | 结果 |
| --- | --- |
| 修改前 `TestTitlebarDividerIsCentered` | fail（预期）：Y=0，稳定复现截图问题 |
| 修改后六组原生交通灯预留布局 | pass：居中且不覆盖相邻控件 |
| `GOWORK=off GOTOOLCHAIN=go1.27.3 CGO_ENABLED=0 go test ./internal/host ./internal/nativeui -count=1` | pass |
| 同环境 `go vet ./internal/host ./internal/nativeui` | pass |
| 标准 MyGo `build -platform darwin/arm64 .` | pass：生成 app 与 DMG，绑定生成检查通过 |
| 图标资源 | pass：1024px sRGB PNG、有透明外边距；ICNS 包含 32/64/128/256/512/1024px |
| `codesign --verify --deep --strict` | pass：本地 ad-hoc 签名 |
| `git diff --check` 和源文件 gofmt | pass |
| 原生布局视觉 | pass：MyGo headless 渲染；该截图不伪装为实际应用窗口截图 |
| 实际 app 图标 | pass：NSRunningApplication 返回的系统图标已是新品牌 |
| 旧预览结束、新包启动 | pass：原 PID 56501 已退出；新 app finishedLaunching=true，唯一运行实例 |
| 实际桌面前台显示 | blocked：CGSession 实时返回 screenLocked=true；新应用已启动，解锁后可查看 |

私有证据：`.openmpa/evidence/desktop-brand-candidate.json`、`brand-launch.json`、`brand-native-layout.png`、`brand-running-app-icon.png`。品牌预览见 `desktop/branding/preview.png`；可编辑资源及重新生成命令见 `desktop/branding/README.md`。未执行提交、发布或云端配置修改。
