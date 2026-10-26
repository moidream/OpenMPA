# OpenMPA MyGo 桌面客户端

状态：按最新用户范围重新实施。只修改当前OpenMPA仓库，agentkit-mpa-agent与mpa-codex-worker只读参考。

[现有OpenAPI客户端合同](2026-10-25-existing-api-client-spec.md)是当前正本，[方案复核](2026-10-25-client-restart-review.md)记录准入，[验收台账](2026-10-25-implementation-baseline.md)记录当前证据。

交付范围包括原生MyGo宿主、当前UI、本地页面、连续对话、Session、渠道管理、Runtime设置、定时任务和知识库管理。所有后端扩展及后端镜像开发撤销；设置页仅通过现有OpenAPI承接用户主动配置。

按S0源码回退、S1平台宿主、S2会话配置、S3渠道、S4日程知识、S5原生性能与打包推进dev-loop。各片完整实现和真实验证后再准出；旧模块文档仅保留可复用客户端细节，矛盾内容以当前合同为准。

恢复后的原生内容工作：

- [原生内容实施契约](2026-10-26-native-content-spec.md)
- [两轮复核](2026-10-26-native-content-review.md)
- [当前验收与候选交付](2026-10-26-native-content-acceptance.md)
- [桌面品牌与标题栏对齐](2026-10-26-desktop-brand-spec.md)
- [桌面品牌与对齐验收](2026-10-26-desktop-brand-review.md)
- [Runtime Key 创建 Session Hotfix](2026-10-26-runtime-key-session-hotfix.md)
- [桌面中英文配置](2026-10-26-desktop-localization-spec.md)
- [Runtime Key Hotfix 与本地化评审](2026-10-26-runtime-key-localization-review.md)
- [删除钥匙串账户恢复功能 Hotfix](2026-10-26-adhoc-keychain-hotfix.md)
- [本地配置数据库与多 Region 设计](2026-10-26-local-config-region-spec.md)
- [对话页视觉对齐与重复输出修复](2026-10-26-chat-visual-alignment-spec.md)
- [本地加密凭据库（AK/SK 记忆）](2026-10-26-credential-vault-spec.md)
- [Runtime 选择恢复与设置 Tab 联动 Hotfix](2026-10-26-runtime-selection-tab-linkage-hotfix.md)
- [Runtime 选择恢复与设置 Tab 联动验证](2026-10-26-runtime-selection-tab-linkage-verification.md)

当前releaseReady=false：核心对话/Session/原生日程及退出后调度通过，知识、专用三渠道、锁屏文件和镜像真实写入验收仍待补齐。所有制品为本地候选。
