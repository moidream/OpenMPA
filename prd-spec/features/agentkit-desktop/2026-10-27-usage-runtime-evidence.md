# Usage Runtime Contract Evidence

## Scope And Result

**PASS for the local Runtime smoke scope only.** This is not desktop end-to-end acceptance, release approval, or actual cloud validation.

- Author-mode verification on macOS arm64, Python 3.12.13, pytest 9.1.1; evidence window 2026-10-27 07:07-07:17 UTC.
- Requirement: `2026-10-27-runtime-usage-spec.md`, FR3/FR4/FR5 and MainProfile/identity contracts. Parent owns broader Go/native/TS tests and reviews.
- OpenMPA HEAD: `7cae08782c2c3e359eb4c4af08f33c159bc5c5d8`; Runtime HEAD: `5053fbabd25128d2043926378676c60d9e90e6b6`. Both working trees were already dirty. Results include current uncommitted source, including Runtime MainProfile, not just these commits.
- Runtime root: `/Users/bytedance/Desktop/code/agentkit-mpa-agent` (read-only). Artifacts: `/tmp/openmpa-usage-smoke.r0a8JR`.
- Existing Runtime tests: **56/56 passed**, zero failures/errors/skips. Dedicated harness: **27/27 expected HTTP responses**, comprising 20 HTTP 200, one 401, two 404, three 409, one 412. These are different counting units, not 83 tests.
- Five required contract groups below executed and passed: execution `5/5`, pass `5/5`, local contract coverage `5/5`. Optional integrations are excluded from those denominators.
- No Runtime source edits, dependency installation, cloud writes, messages, real credentials, commit, or application-profile access. Only this report was added to OpenMPA; harness and generated stores/results remain in `/tmp`.

## Evidence Register

All entries are **Observed**, collected by this verification task.

| ID  | Locator                                                                              | Result / Boundary                                                                |
| --- | ------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------- |
| E-1 | `/tmp/openmpa-usage-smoke.r0a8JR/runtime-tests.xml`; command A below                 | 45 passed in 3.58s, 6 warnings; actual Runtime tests, no pytest cache or pyc     |
| E-2 | `/tmp/openmpa-usage-smoke.r0a8JR/runtime-extra.xml`; command B                       | 11 passed in 3.18s, 4 warnings                                                   |
| E-3 | `/tmp/openmpa-usage-smoke.r0a8JR/route_smoke.py` and `route-results.json`; command C | Actual FastAPI handlers/services through TestClient; 27 checked requests, exit 0 |
| E-4 | Runtime and Go source references below                                               | Static contract comparison only; Go adapters were not executed                   |
| E-5 | Toolchain commands below                                                             | Installed Go 1.26.3; desktop requires 1.27.3; offline toolchain selection failed |

E-1 warnings: four dependency `BaseAgentConfig` deprecations and two duplicate OpenAPI operation IDs on retired collaborator routes. E-2: four dependency deprecations. No test failures resulted.

Harness first attempt exited 1 before HTTP checks because its synthetic session ID was not a canonical UUID. Only the temporary fixture was corrected. The successful run used `00000000-0000-4000-8000-000000000001` and a fresh database; this is not a product regression or a hidden failed product test.

## Contract Matrix

Runtime paths below are relative to its read-only root. Go paths are relative to OpenMPA.

| Change / Test                                            | Expected Contract And Observed Result                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | Status / Evidence                                                                                                                                                                                  |
| -------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| C-1 / T-1: AGENTS.md, revision/auth risk                 | `GET /api/v1/studio/agents-md` returned `{content:"",revision:0}`; `POST {content,revision:0}` returned revision 1; stale POST returned 409; wrong Studio key returned 401; final GET retained the saved shared document. Existing tests cover clear, 40,000-character bound, invalid fields, and preserving other config.                                                                                                                                                                                                                                                            | PASS; E-1/E-3. `app/api/v1/studio_metadata.py:17-74` vs `desktop/internal/usage/resources.go:16-64`. Production no-op caveat R-2 below.                                                            |
| C-2 / T-2: MCP, optional-tool projection risk            | `GET /api/v1/sessions/{session_id}/mcp-servers` returned `{sessionId,mcpServers:[...]}` with `id/name/alias/description/gatewayUrl/tools`; tools contain `name/description`. Second server omitted `tools`, matching Go `ToolsKnown=false`. Missing session returned HTTP 200 and an empty array, not 404.                                                                                                                                                                                                                                                                            | PASS; E-1/E-3. `app/api/v1/sessions.py:737-782` vs `desktop/internal/usage/resources.go:85-137`.                                                                                                   |
| C-3 / T-3: Runtime identity, wrong-ID risk               | `GET /api/v1/agents` returned 404 before any agent config, then `mpaInstanceId:"mi-offlinetest"` after local AGENTS.md bootstrap. It is not the platform `r-` Runtime ID. Studio `agent-info` has no identity field.                                                                                                                                                                                                                                                                                                                                                                  | PASS; E-1/E-2/E-3. `app/api/v1/agents.py:82-143` vs `desktop/internal/usage/resource_identity.go:10-28`.                                                                                           |
| C-4 / T-4: A2A directory/binding/aliases, CAS/scope risk | Registry `GET /spaces?page=1/2` returned `items,total,page,page_size:50`; `/agents?space_id=...` added `Alias`, `AliasRevision`, `AliasAvailable` and sanitized `RuntimeConfig.RuntimeId`. `GET/PUT /binding` used `space_id/space_name/region/revision`, nullable unbound fields, revisions 0 -> 1 -> 2 and stale 409. `PUT /aliases/{agent_id}` used `{space_id,alias,revision}` and returned `{alias,revision}`; set/readback/stale-409/clear passed.                                                                                                                              | PASS; E-2/E-3. `app/api/v1/collaborating_agents.py:59-124`, `app/services/a2a_registry.py:64-120`, `app/services/a2a_space.py:25-52,140-166` vs `desktop/internal/usage/collaboration.go:115-302`. |
| C-5 / T-5: MainProfile, replacement/CAS risk             | `GET /api/v1/agents/{mpaInstanceId}/main-profile` initially returned 404 with `detail.code=main_profile_not_found`. Local setup PUT used `Idempotency-Key` + `If-None-Match:*`. Replacement PUT used `Idempotency-Key` + `If-Match:"1"` and returned revision 2, quoted body/header ETag, digest, identity, profile, provenance, author and timestamp. Model/skills/provenance survived instruction update; stale PUT returned 412; unchanged PUT retained revision 2; shared AGENTS.md stayed unchanged. Existing tests exercise authority, replay, preconditions and invalid input. | PASS; E-1/E-3. `app/api/v1/main_profile.py`, `app/schemas/main_profile.py`, `app/services/main_profile.py` vs `desktop/internal/usage/main_profile.go:79-199`.                                     |

## Real And Mocked Layers

- **Real in E-3:** FastAPI routing/validation/serialization/error handling; Studio header comparison; `AgentConfigService`; `SessionService` MCP resolution; `RegistryClient` pagination/sanitization/identity checks; `A2ASpaceService`, `A2ASpaceStore`, `A2AAliasStore`; `MainProfileService` and `SqlMainProfileStore`. A2A/MainProfile/channel dependencies use actual SQLite under a fresh `route-data-*` directory.
- **Mocked in E-3:** AGENTS.md PostgreSQL persistence uses existing `_AgentConfigStore`; Studio metadata/key store returns a fixture credential; `/agents` auth, session user and MainProfile principal are dependency overrides. MCP catalog uses `_Cache/_resource`; session backing is in-memory with no-op mirror and unused `FakeRunner`. Registry **only `request()`** is canned; no SDK/IAM/cloud transport executes.
- **E-1/E-2 vary by existing fixture:** MainProfile API tests use its in-memory store; identity tests use `FakeAgentConfigService`; A2A management-route tests use `AsyncMock` service/auth overrides; A2A business tests use SQLite plus `FakeRegistry`. Therefore their route mocks alone are not whole-stack proof; E-3 supplies the combined service path.
- Test processes used an empty inherited environment, temporary HOME/TMPDIR/cwd, and Runtime's explicit offline fixture settings. E-3 additionally rejects non-loopback socket connect/sendto operations. Full application lifespan, scheduler, model execution, sandbox provisioning and outbound dispatch were not started.
- A2A list responses **do not contain AgentCard**, even when canned upstream data does: the real registry sanitizer strips it. Go correctly permits `Card=nil`; this smoke cannot establish Card detail availability in the UI.

## Commands

Paths are factored into variables below; the expanded argv/environment are those executed. Commands A-C ran with cwd `$T`, outside Runtime, so developer `.env` and config files were not loaded.

```bash
R=/Users/bytedance/Desktop/code/agentkit-mpa-agent
T=/tmp/openmpa-usage-smoke.r0a8JR
P="$R/.venv/bin/python"

# A: exit 0; 45 passed, 6 warnings.
env -i PATH=/usr/bin:/bin:/usr/sbin:/sbin HOME="$T" TMPDIR="$T" PYTHONDONTWRITEBYTECODE=1 PYTEST_DISABLE_PLUGIN_AUTOLOAD=1 PYTHONPATH="$R" "$P" -m pytest -p no:cacheprovider --basetemp="$T/pytest-base" --junitxml="$T/runtime-tests.xml" -q "$R/tests/test_studio_metadata.py" "$R/tests/test_studio_agents_md.py" "$R/tests/test_a2a_registry_collaborators.py" "$R/tests/test_main_profile_api.py" "$R/tests/test_session_mcp_update.py"

# B: exit 0; 11 passed, 4 warnings.
env -i PATH=/usr/bin:/bin:/usr/sbin:/sbin HOME="$T" TMPDIR="$T" PYTHONDONTWRITEBYTECODE=1 PYTEST_DISABLE_PLUGIN_AUTOLOAD=1 PYTHONPATH="$R" "$P" -m pytest -p no:cacheprovider --basetemp="$T/pytest-extra" --junitxml="$T/runtime-extra.xml" -q "$R/tests/test_a2a_business_aliases.py" "$R/tests/test_a2a_management_routes.py" "$R/tests/test_agents.py::test_get_agent_returns_current_agent_identity_only" "$R/tests/test_agents.py::test_get_agent_without_current_config_returns_404" "$R/tests/test_agents.py::test_get_agent_allows_non_admin_jwt_user"

# C: initial fixture error above, then exit 0; status=passed, requests=27.
env -i PATH=/usr/bin:/bin:/usr/sbin:/sbin HOME="$T" TMPDIR="$T" PYTHONDONTWRITEBYTECODE=1 PYTEST_DISABLE_PLUGIN_AUTOLOAD=1 PYTHONPATH="$R" "$P" "$T/route_smoke.py"

# E-5: cwd /Users/bytedance/Desktop/github-code/OpenMPA/desktop
env GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local go version
# exit 0: go version go1.26.3 darwin/arm64
env GOPROXY=off GOSUMDB=off go version
# exit 1: go: golang.org/toolchain@v0.0.1-go1.27.3.darwin-arm64:
# verifying module: checksum database disabled by GOSUMDB=off
```

Counts were cross-checked with `xmllint --xpath` against both JUnit files and `jq` against `route-results.json`. The XML files report 45 and 11 tests with zero failures/errors/skips; JSON contains exactly 27 observations.

## Limits And Handoff

The R-1/R-2/A-1 entries below describe the original verifier handoff. The parent
completed Go-to-ASGI integration and no-op adapter repair in the follow-up section
at the end of this document; physical native-window and cloud checks remain not_run.

- **R-1, blocked/not_run:** Actual Go service -> localhost ASGI integration was not run after E-5. No `runtime_smoke_test.go` was created. Optional native-window/CapturePage smoke was also **not_run**; no screenshots or ignored native harness were created. The native path shares the unmet Go prerequisite, and native work was optional in the request.
- **R-2, source-level adapter caveat:** Production `app/stores/agent_config.py:255-283,470-499` returns the same revision for an unchanged configuration. The existing `_AgentConfigStore` fixture always increments. Go `SaveAgentsMD` at `resources.go:56` rejects equal revisions as `outcomeUnknown`. This was **not** reproduced against PostgreSQL or the Go binary. Native `resourceCanSave` (`desktop/internal/nativeui/resources.go:290-299`) requires a dirty draft, limiting the current UI path; do not infer a confirmed user-facing failure. Parent should decide whether adapter no-op parity needs a focused follow-up.
- **R-3, behavior limits:** MCP missing-session 200/empty cannot prove session existence. `/agents` 404 on an unconfigured Runtime prevents identity discovery through this route. MainProfile 404 remains ambiguous after Go's generic error-code projection. These are observed/source-supported limitations, not cloud findings.
- **R-4, not_run:** Real PostgreSQL AGENTS.md CAS/concurrency; live JWT/gateway/IAM; cloud Skill Space and A2A discovery; actual MCP tool discovery/invocation; A2A dispatch; all channel writes; native UI and TS workflows. Skill/topology validation was outside the requested five Runtime contract groups. Production durability and live permissions are not established.
- **A-1 (parent, before desktop acceptance):** Complete owned Go/native/TS checks and reviews, including R-2 disposition. If exercising the optional integration, supply an already-available compatible toolchain; this task authorizes no installation.
- **V-1 / A-2 (release owner, before any cloud-validation claim):** Live deployment verification is **not_run**, no release or observation window occurred. Use separately authorized test resources for readback/auth/version checks. No remote rollback is required because this task performed no remote mutation.
- Runtime `git status --short` was unchanged from the initial snapshot. Temporary evidence is retained for parent inspection; it is ephemeral, so durable conclusions and command results are included here. Compact English report invariants were checked manually; the canonical Chinese-template validator was not applicable.

## 主 Agent 后续：Go 客户端直连本地 Runtime

已执行并通过。使用原 `/tmp/openmpa-usage-smoke.r0a8JR/route_smoke.py` 的隔离
FastAPI app，以 Uvicorn 在 `127.0.0.1:18864` 监听，实际运行 OpenMPA 的
connection/usage 服务。临时 Go 测试通过 overlay 注入；源码和测试文件均留在 `/tmp`，
未向参考仓库或本仓库添加验证程序。

`TestLocalRuntimeUsageContract` 在 0.06 秒通过，16 个 HTTP 请求覆盖共享
AGENTS.md 读/写/409、MCP 工具明细、真实 mpaInstanceId、A2A 空间/成员/绑定/别名
及冲突、Main Profile 更新保留模型与 Skill、共享文档隔离；消耗过的 Main snapshot
再写在本地被拒绝。证据：`/tmp/openmpa-usage-asgi-go.log`、
`/tmp/openmpa-usage-asgi.log`、`/tmp/openmpa-usage-live-contract_test.go`。

执行命令：

```sh
env -i PATH=/usr/bin:/bin:/usr/sbin:/sbin TMPDIR=/tmp/openmpa-usage-smoke.r0a8JR PYTHONDONTWRITEBYTECODE=1 PYTHONPATH=/Users/bytedance/Desktop/code/agentkit-mpa-agent:/tmp/openmpa-usage-smoke.r0a8JR /Users/bytedance/Desktop/code/agentkit-mpa-agent/.venv/bin/python -m uvicorn route_smoke:app --host 127.0.0.1 --port 18864
GOTOOLCHAIN=go1.27.3 go test -overlay=/tmp/openmpa-usage-overlay.json ./internal/usage -run TestLocalRuntimeUsageContract -count=1 -v
```

当前参考源码新增了 Main Profile 专用 principal dependency，首次启动旧临时 harness
因此在其自检阶段返回缺少 JWT 公钥的 500；Go 首次连接随之失败，未抵达适配路径。
仅更新临时 harness 对该 dependency 的 fixture override 后通过。此处的授权仍为 mock，
Studio key 比较与 Go 请求头注入为真实路径；不以该测试证明 JWT/IAM 或生产网关授权。
所有其他真实/mock 层边界仍如前文；服务仅在本机临时数据库写入，没有云端操作。

R-2 已关闭：adapter 接受与正文一致且 revision 未倒退的无变化保存回执，
新 `TestResourcesAgentsMDNoOpKeepsRevision` 通过。原 verifier 的工具链不可用
仅限其离线环境；主会话使用已可用的 Go1.27.3，无依赖安装。
