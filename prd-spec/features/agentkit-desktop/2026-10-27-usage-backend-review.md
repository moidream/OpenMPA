# Usage Backend: Independent Two-Round Review

Date: 2026-10-27. Mode: read-only code review; this evidence document is the only authorized write.

## Result

**最终本地 Gate: pass（主 Agent 修复后复验）。** 独立审查时的两个 medium 发现已修复；
原审查候选为 21 项：2 confirmed、18 rejected、1 unverified（部署版本兼容性）。
以下主体保留修复前的独立审查证据与指纹，最终修复见文末，不将其冒充修复后的独立复审。

| ID    | Severity | Category                     | Location                                                                                                        | Finding                                                                                                                     | Evidence                                                        | Disposition              |
| ----- | -------- | ---------------------------- | --------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------- | ------------------------ |
| UB-01 | medium   | Behavioral/API compatibility | `desktop/internal/usage/collaboration.go:305-312`; `desktop/internal/nativeui/collaboration.go:337-339,423-436` | Registry display-name fallback does not identify the selected member in Runtime mention resolution.                         | E-01, complete E1                                               | open; parent owns repair |
| UB-02 | medium   | API compatibility            | `desktop/internal/usage/resources.go:280-284,301-306`                                                           | Missing/null optional Skill descriptions reject the complete directory, contrary to the selected Runtime consumer contract. | E-02, complete E1; E-03 verifies the current rejection behavior | open; parent owns repair |

## Scope and Authority

- Reviewed all eight `desktop/internal/usage/*.go` files, including their tests.
- Reviewed only the new Studio credential injection/path predicate and header allowlist in `desktop/internal/connection/connection.go`, plus `studio_test.go`. Existing transport and tests were read as supporting evidence, not expanded defect scope.
- Reviewed only the `ListSkillsBySpaceId` allowlist addition in `desktop/internal/cloud/client.go`. Existing signing/transport helpers were read to trace it.
- Reviewed `desktop/internal/host/usage.go`, `desktop/internal/nativeui/collaboration.go`, and `desktop/internal/nativeui/usage.go`. Read collaboration tests, Controller lifecycle, host rail and frontend event consumers as supporting call paths.
- Did not read or review `desktop/internal/nativeui/resources.go` or its tests. That implementation remains owned by the resource UI agent. Package test output involving its test file is execution evidence only.
- Preserved the pre-existing user diff. No production/test edits, commits, external writes, cloud calls, Runtime writes, new probes, or external-repository test execution.
- Governing requirement: [runtime usage spec](2026-10-27-runtime-usage-spec.md), especially lines 15-18, 25-30, 34-45. Read `code-review/SKILL.md` and `references/evidence-validation.md`; complete E1 requires a current contract, admitted trigger, full path, deterministic consequence, counterevidence, and no unresolved assumptions.
- No applicable `AGENTS.md` or local instruction files were found in the reviewed repository/ancestors checked.
- Parent explicitly owns the JSON evidence ledger and fixes. This reviewer did not create a ledger or run its validator. No separate GateGap/admission-escape claim is made.

Reference roots used below:

| Name    | Root                                                | HEAD observed                              |
| ------- | --------------------------------------------------- | ------------------------------------------ |
| Desktop | `/Users/bytedance/Desktop/github-code/OpenMPA`      | `7cae08782c2c3e359eb4c4af08f33c159bc5c5d8` |
| Runtime | `/Users/bytedance/Desktop/code/agentkit-mpa-agent`  | `5053fbabd25128d2043926378676c60d9e90e6b6` |
| VeADK   | `/Users/bytedance/Desktop/github-code/veadk-python` | `2301477104f6f77e76c8182529545950e2d8ae6e` |

Desktop findings refer to the working files, not just HEAD. Runtime API/service/schema files and VeADK `veadk/skills/utils.py` were read locally, read-only; targeted `git diff` checks on the principal contract files were empty. Main Profile source is explicitly **unpublished**, not evidence of deployed endpoint availability.

## Round One: Validated Findings

### E-01 / UB-01: Display Name Is Not a Runtime Address

Contract: FR5/AC5 requires the selected member to become the conversation target. Runtime resolves explicit mentions by business alias first, then Agent Card name, not the registry display `Name`.

Source-to-consumer proof:

1. [Runtime registry projection](file:///Users/bytedance/Desktop/code/agentkit-mpa-agent/app/services/a2a_registry.py#L70-L105) includes `Id`, `Name`, status and safe Runtime metadata, but deliberately excludes `AgentCard`. [Management enrichment](file:///Users/bytedance/Desktop/code/agentkit-mpa-agent/app/services/a2a_space.py#L114-L128) adds aliases without adding a Card. A valid, never-aliased member therefore has `Alias:null`, `AliasRevision:0`, `AliasAvailable:true`, and no Card.
2. [Desktop decoder](file:///Users/bytedance/Desktop/github-code/OpenMPA/desktop/internal/usage/collaboration.go#L170-L201) accepts that legitimate row and produces a nil Card. [Mention construction](file:///Users/bytedance/Desktop/github-code/OpenMPA/desktop/internal/usage/collaboration.go#L304-L331) then falls back to `agent.Name`.
3. [Native eligibility](file:///Users/bytedance/Desktop/github-code/OpenMPA/desktop/internal/nativeui/collaboration.go#L337-L339) and [submission](file:///Users/bytedance/Desktop/github-code/OpenMPA/desktop/internal/nativeui/collaboration.go#L423-L436) require a fresh bound space, running/non-self member and valid task, but do not require an authoritative alias/address. These guards all pass for the valid row above.
4. [Runtime resolver](file:///Users/bytedance/Desktop/code/agentkit-mpa-agent/app/services/a2a_space.py#L148-L174) checks aliases first. Its fallback [directory scan](file:///Users/bytedance/Desktop/code/agentkit-mpa-agent/app/services/a2a_space.py#L192-L232) compares each fetched Card name. [Runtime tools](file:///Users/bytedance/Desktop/code/agentkit-mpa-agent/app/runtime/tools/collaborating_agents.py#L69-L112) resolve the mention and validate dispatch against that resolved ID. They never receive the Desktop selection ID: [UsageIntent](file:///Users/bytedance/Desktop/github-code/OpenMPA/desktop/internal/host/usage.go#L9-L13) contains only epoch, session and text.

Admitted counterexample: current bound space contains running non-self member A, registry `Name="Demo"`, Card `name="Card Name"`, no alias. Selecting A creates `@Demo task`. If no member has alias/Card name `Demo`, Runtime returns `agent_mention_not_found`. If member B has the valid business alias `Demo`, alias precedence resolves B, not selected A. Alias uniqueness is only across aliases, not against registry display names or Card names.

Independent contract corroboration, read but not executed: [Runtime test](file:///Users/bytedance/Desktop/code/agentkit-mpa-agent/tests/test_a2a_aliases.py#L77-L99) explicitly expects registry display `Demo` not to resolve after the Card name changes, and explicitly expects another member's alias to override a matching Card name.

Round-two challenge:

- If A has its own valid alias, current code uses it and this counterexample does not apply.
- A Card-bearing test fixture cannot establish real-list compatibility: the actual Runtime projection strips Cards.
- Equal display/Card names sometimes work but are not a promised invariant.
- The existing Runtime target-mismatch guard cannot recover A's discarded ID; it checks the ID obtained from the already incorrect mention.
- Parent's composer prepend/receipt repairs affect delivery of the text, not which member that text resolves to. No duplicate composer finding is raised.

Consequence: valid selections fail lookup or address the wrong member. This is medium functional/selection correctness, not a claim of observed credential leakage or a reproduced production dispatch.

Surgical recommendation: for this published list contract, require a confirmed member alias before enabling/generating a targeted mention; use the existing explicit alias editor to recover. Do not automatically create aliases. Any future Card/name resolution route must verify the resolved member ID and respect alias precedence; merely obtaining the Card name is insufficient.

Suggested parent regression: use a list response without `AgentCard`, a different registry/Card name, and a different member's colliding alias. Verify no misleading mention can be generated without an authoritative address.

Evidence level: complete E1. Unresolved assumptions within the stated supported input/state: `[]`. Deployment prevalence and actual LLM dispatch are not asserted.

### E-02 / UB-02: Optional Skill Description Rejects the Whole Page

Contract: spec lines 25 and 37-38 select the Runtime consumer shape for `ListSkillsBySpaceId`, not a different managed-Skill schema.

Proof:

1. [VeADK item consumer](file:///Users/bytedance/Desktop/github-code/veadk-python/veadk/skills/utils.py#L281-L296) requires a truthy `Name` but explicitly uses `item.get("Description") or ""`. Both an absent Description and `Description:null` are supported and produce a Skill with empty descriptive text.
2. [VeADK action call](file:///Users/bytedance/Desktop/github-code/veadk-python/veadk/skills/utils.py#L333-L378) sends the same ID-only `ListSkillsBySpaceId` action and passes returned Items to that consumer.
3. Desktop [wire decoding and validation](file:///Users/bytedance/Desktop/github-code/OpenMPA/desktop/internal/usage/resources.go#L280-L306) represents Description as `*string`, then rejects `item.Description == nil` before building or paginating the result.
4. For a mounted space and signed-in current identity, a valid response `{"Result":{"Items":[{"Name":"review"}]}}` reaches that condition and deterministically returns `invalidResponse`. The same happens for `Description:null`. One such item also rejects otherwise valid neighboring items.

Round-two challenge:

- An explicit empty string works; the mismatch concerns absence/null, which the chosen consumer explicitly supports.
- Missing/null `Items`, missing/blank Name, wrong field types and duplicate names remain separate validation concerns. The repair need not weaken those checks or fabricate a successful empty directory.
- [Existing test](file:///Users/bytedance/Desktop/github-code/OpenMPA/desktop/internal/usage/resources_test.go#L352-L367) includes `{"Name":"x"}` among rejected responses. Its passing result confirms the current rejection branch, but its expected failure is not authoritative against the actual consumer contract.
- No assumption is made that a particular live cloud account currently returns such an item. The finding is a source-defined compatibility mismatch, not an observed live incident.

Consequence: a Runtime-compatible Skill list cannot be displayed because optional descriptive metadata is absent.

Surgical recommendation: retain nullable wire decoding, map missing/null Description to `""` when building the safe projection, and keep the remaining identity/type/list validation. Update the existing missing-description test and add the null-description counterpart.

Evidence level: complete E1. Unresolved assumptions within the stated supported response shape: `[]`.

## Other Candidate Decisions

All rows went through `candidate -> validating ->` the recorded disposition. Severity below is the proposed severity, not a confirmed finding.

| ID    | State      | Proposed severity | Claim challenged                                                                                                       | Evidence and disposition reason                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| ----- | ---------- | ----------------- | ---------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| UB-03 | rejected   | medium            | Collaboration 5xx writes remain ordinary errors and permit unsafe replay.                                              | `usage/service.go:33-38` centrally converts every non-GET 5xx to `outcomeUnknown/verifyOnly`, before `collaborationRequestError`. `usage/collaboration_test.go:300-325` exercises both writes and checks one request. `false_positive_refuted`.                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| UB-04 | rejected   | medium            | Lost/malformed success bodies or CAS conflicts become success or auto-retry.                                           | `service.go:40-50`; `resources.go:48-62`; `collaboration.go:232-248,279-301`; `main_profile.go:90-139,142-200` validate receipts. Transport `connection.go:138-142,166-174` disables request replay bodies and marks uncertain writes. Native binding sets `bindingFresh=false` before writing (`collaboration.go:377`), alias sets `needsReload=true` (`:405`). Main Profile consumes the snapshot atomically (`:122-124`). Existing lost-response, conflict and malformed-receipt tests pass. `false_positive_refuted`.                                                                                                                                                                          |
| UB-05 | rejected   | medium            | Missing/null directory fields produce a fabricated successful empty page.                                              | `collaboration.go:78-89,129-136,165-176,211-218` requires Items, total, matching page, positive bounded page size, valid IDs and binding revision/identity. `service.go:45` rejects JSON null/trailing JSON. Optional display metadata does not authorize a mutation. `false_positive_refuted`.                                                                                                                                                                                                                                                                                                                                                                                                    |
| UB-06 | rejected   | medium            | Missing optional alias metadata necessarily invalidates older member lists, or missing revision enables alias writes.  | `collaboration.go:172-183` permits alias-unavailable legacy rows but requires Alias plus AliasRevision when `AliasAvailable` is true. Native alias editor/write guard requires AliasAvailable (`nativeui/collaboration.go:325-332,393`). Tests `collaboration_test.go:148-178` cover both forms. `false_positive_refuted`.                                                                                                                                                                                                                                                                                                                                                                         |
| UB-07 | rejected   | medium            | Missing MCP tool details are rendered as a known empty tool list.                                                      | `resources.go:120-123` derives ToolsKnown from slice presence. Runtime `app/schemas/session.py:19-31` declares optional tools; API `app/api/v1/sessions.py:763-782` excludes None fields. Missing/null tools remain unknown while `[]` is known empty; `resources_test.go:61-84` checks this distinction. `false_positive_refuted`.                                                                                                                                                                                                                                                                                                                                                                |
| UB-08 | rejected   | medium            | Main Profile rejects a valid default model or silently loses model/Skill settings.                                     | Runtime `app/schemas/main_profile.py:49-66` permits null model and bounded Skills. Desktop `main_profile.go:179-239` strictly decodes the complete profile and validates the combined budget; `:104-113,258-280` preserves a private clone of settings/source references. Unknown nested fields are rejected rather than dropped. Tests `main_profile_test.go:20-64,90-132,189-267,269-317,345-372` cover preservation, null model, limits and concurrent snapshot consumption. `false_positive_refuted`.                                                                                                                                                                                          |
| UB-09 | rejected   | medium            | Main Profile 404 authorizes creation, or the service substitutes the selected platform Runtime ID for MPA identity.    | `main_profile.go:77-99,125-128,242-255` requires a read snapshot and uses update-only If-Match; 404 remains unsupported-or-unconfigured. `resource_identity.go:13-26` reads `/api/v1/agents` and validates its returned MPA ID. Runtime `app/api/v1/agents.py:128-139` returns actual current AgentRecord metadata. No inferred `r- == mpaInstanceId` equivalence; unpublished availability remains UB-19. `false_positive_refuted`.                                                                                                                                                                                                                                                               |
| UB-10 | rejected   | medium            | Identity changes or cancellation let old service results commit into the new native page.                              | `connection.go:34-39,107-112,125-126,166-185` ties requests to connection epochs. `service.go:42-44`, resource postchecks and `resources.go:249-292` reject stale results. Controller `controller.go:168-200,206-249` fences generation plus connection/platform epochs before applying results, with a 60-second context. Existing usage identity/cancel tests and Main Profile mid-body tests pass. This rejects stale application, not a claim that sign-out synchronously stops every cloud socket; see UB-21. `false_positive_refuted`.                                                                                                                                                       |
| UB-11 | rejected   | medium            | Space/member changes reuse another scope's alias draft or stale confirmation, including mixed binding/directory reads. | Native `collaboration.go:48-63` keys drafts by space/member; `:166-176` clears member/results on space switch; `:125-134` rereads binding; `:147-164` checks version/epoch and incomplete snapshots; `:373-375,401-403` revalidate confirmations. Controller invalidates the entire state on identity changes. A partial directory failure intentionally leaves explicit unbind available from a fresh binding. Tests `nativeui/collaboration_test.go:157-210,213-258,290-303` cover these paths. `false_positive_refuted`.                                                                                                                                                                        |
| UB-12 | rejected   | medium            | Host delivery alone permits a stale choice in another session after its rail snapshot.                                 | `host/usage.go:25-29` checks epoch/session; `nativeui/usage.go:15-18` uses currentSessions, whose Controller guard checks shell epoch (`controller.go:407-411`). Supporting frontend consumer `DesktopAppState.tsx:279-306` rechecks epoch, selected session and membership at event consumption. No backend session misdelivery path survives that consumer. Composer receipt/unmount handling is parent-owned, not re-reviewed here. `false_positive_refuted`.                                                                                                                                                                                                                                   |
| UB-13 | rejected   | high              | Forged Studio headers, redirects, encoded traversal or alias paths expose the key to an arbitrary route/host.          | `connection.go:117-123,147-164,192-203` validates paths, injects the current key only on fixed Studio paths, excludes caller Studio/Authorization forwarding and refuses redirects. Alias suffix excludes slash, backslash, percent, query and fragment delimiters; service IDs have a stricter ASCII regex. Permitting extra non-routing suffix characters at this transport layer still reaches only the Runtime alias endpoint, whose API rejects invalid IDs (`Runtime app/api/v1/collaborating_agents.py:99-110`). `studio_test.go:10-65` plus existing redirect tests support the bounded claim. No new source-to-external-sink path found. `false_positive_refuted`.                        |
| UB-14 | rejected   | medium            | Skill paging uses unsupported server pagination fields or discovers unmounted spaces.                                  | `resources.go:187-208,260-279,308-314` selects only root mounts, verifies the requested space and sends only SkillSpaceId/InnerTags, then pages Items locally. VeADK `utils.py:344-357` and spec `:37-38` agree. `resources_test.go:120-159` checks signed host/action/body, local page 2 and rejection before cloud access for an unmounted space. `intended_by_contract`.                                                                                                                                                                                                                                                                                                                        |
| UB-15 | rejected   | high              | Cloud allowlist addition broadens creation/write authority.                                                            | The only allowlist diff is `ListSkillsBySpaceId` (`cloud/client.go:150`). Its scoped caller sends the fixed read action, uses existing region/signing rules and projects only name/description. No new write action or arbitrary endpoint was added. `false_positive_refuted`.                                                                                                                                                                                                                                                                                                                                                                                                                     |
| UB-16 | rejected   | medium            | Lack of a new dedicated cloud-package test is itself a defect.                                                         | Coverage absence is not behavioral proof. Moreover `usage/resources_test.go:104-158,352-398` exercises the real cloud Action path with an injected transport, checks signing/contract and errors/epochs. Live compatibility remains UB-19, not an invented confirmed defect. `non_actionable_preference`.                                                                                                                                                                                                                                                                                                                                                                                          |
| UB-17 | rejected   | medium            | Absence of rendered Agent Card detail is independently a client decoder bug.                                           | Runtime deliberately strips Cards (`app/services/a2a_registry.py:70-105`; source test `tests/test_a2a_registry_collaborators.py:137-149`). Desktop presents unavailable (`nativeui/collaboration.go:314-320`). The independently actionable misuse of display name is UB-01; do not add a duplicate missing-Card finding. `duplicate/intended_by_contract`.                                                                                                                                                                                                                                                                                                                                        |
| UB-18 | rejected   | medium            | The initial resource UI fallback test failure or composer findings should be charged to this backend review.           | Initial allowed test run failed only in the explicitly excluded resource UI test. Parent reported concurrent refinement plus separate composer repairs. The identical allowed test command later passed. Neither is attributed to a backend hunk; resource UI tests were not opened. `out_of_scope`.                                                                                                                                                                                                                                                                                                                                                                                               |
| UB-19 | unverified | medium            | Deployed Runtime/Main Profile or live platform Skill permissions differ from the local source contract.                | No external calls authorized/executed. Main Profile is expressly unpublished; API source cannot prove deployment or principal capabilities. Source-level compatibility is reviewed, but live availability/permission behavior is `environment_unavailable`. No repair recommendation or confirmed production mismatch.                                                                                                                                                                                                                                                                                                                                                                             |
| UB-20 | rejected   | medium            | The generic 64-KiB draft limit versus Runtime dispatch's 32,000-byte task limit alone proves broken execution.         | `nativeui/usage.go:10-18` creates an editable draft, not a dispatch request. Runtime `app/runtime/tools/collaborating_agents.py:98-101` validates the eventual tool task, which need not be byte-identical to the draft. No deterministic execution failure follows from this UI limit alone. Do not conflate draft validation with dispatch validation. `non_actionable_preference/insufficient causal proof`.                                                                                                                                                                                                                                                                                    |
| UB-21 | rejected   | medium            | Lack of synchronous platform-epoch socket cancellation proves credential leakage or cross-account result application.  | A platform change can occur between the precheck and Action's own credential snapshot (`resources.go:271-286`; existing `cloud/client.go:154-162`); postchecks reject changed epochs (`resources.go:287-292`) and Controller also fences application. Runtime connection cancellation is wired to context (`resources.go:253-259`); existing cloud call is bounded to 30 seconds (`cloud/client.go:184-185`). No write, cross-account result application, or forbidden external credential sink is established. NFR1 specifies epoch isolation, not atomic precheck/synchronous cloud socket shutdown. No shared cloud refactor recommended. `false_positive_refuted` for the claimed consequence. |

## Round Two: Adversarial Outcome

- Missing/null fields: retained UB-02 for Description specifically; rejected fabricated-empty-page, partial alias-revision and MCP unknown-tool claims.
- 5xx writes, broken success JSON, lost replies, conflicts: inspected the shared helper before individual adapters. Centralized unknown-outcome mapping and read-before-retry UI state refute UB-03/04.
- Epoch changes and cancellation: traced connection context, successful response-body checks, cloud postchecks, Controller generation/epochs and snapshot consumption. No scoped stale-result application was proven.
- Selection scopes: challenged space changes, absent current member, partial directory failure, read/re-read binding changes, stale confirmations and late host events. Scope guards survive; UB-01 remains because it is an address-identity mismatch inside the correct space/session.
- Alias grammar: Desktop `collaboration.go:255-277,313-331` uses NFKC/trim, bounded aliases, control rejection and mention quoting. Runtime `app/stores/a2a_alias.py:37-51` and `app/services/a2a_mentions.py:15-26` support the tested forms. Quotes, leading second mentions, empty tasks, control characters and overlong names are rejected. The confirmed defect concerns name resolution, not quoting.
- Main Profile: challenged null model, missing profile, extra fields, same-revision no-op, modified settings, stale snapshots, concurrent saves and 404. Runtime `app/stores/main_profile.py:68-90` returns the same revision on unchanged profile digest, matching the Desktop no-op rule. No assumption of publication or ID equivalence was used.
- No confirmed finding was withdrawn in round two. Candidate UB-03 was explicitly rejected after tracing centralized 5xx handling. Other rejected decisions are retained above.

## Executed Verification

E-03, run twice from `/Users/bytedance/Desktop/github-code/OpenMPA/desktop`:

```sh
GOTOOLCHAIN=go1.27.3 go test ./internal/usage ./internal/connection ./internal/nativeui
```

| Run                                | Result                                                                                                                                                                        | What it proves                                                                                                                   |
| ---------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| First                              | exit 1; usage and connection passed (cached); nativeui failed `TestNativeResourcesUnknownErrorHasVisibleFallback`, `resources_test.go:554`, `unknown safe error became blank` | Records a transient failure in the concurrently edited, excluded resource UI surface. It does not prove a scoped backend defect. |
| Latest, after parent status update | exit 0; usage, connection and nativeui all `ok` (cached)                                                                                                                      | Existing package suites are green at the observed content. Does not disprove E-01 or the contract mismatch behind E-02.          |

No custom reproduction tests were written. Complete E1 proofs above are the promotion basis; read-only Runtime test-source corroboration is not mislabeled as an executed Runtime test. No test sessions remain running.

Parent-reported, not independently rerun here: all Go tests and vet passed; focused race initially encountered the resource fallback refinement; 45 focused JS tests with `LANG=en_US.UTF-8`, typecheck and lint passed. Parent also reported fixing mention prepend and scoped durable-draft receipt reconciliation. This review does not certify those separate fixes.

## Reviewed Content Fingerprints

SHA-256 values below were captured before the final adversarial closeout and were identical on the closing recheck. They identify the reviewed working-tree content despite concurrent work elsewhere.

```text
cba5d2ea65fb4ba5d4aa62f875555809c28d2603f7820304c302c3380906e8bb  desktop/internal/usage/collaboration_test.go
8e392230a7515c86e2615264ac93921c64fd4b3670d776afe8b9429de5430345  desktop/internal/usage/collaboration.go
1804b774b826edff3fc866fb6115d9d55ab1f3e2e7576a2d3607c84a73173b61  desktop/internal/usage/main_profile_test.go
7a37ce4447f06a42d434639b579228dd2e6b99acc3b8b48df41b32eabae132ae  desktop/internal/usage/main_profile.go
3af94b578647c988ef096c8337c782e61718051d71fe04e22c4379a6b3e73cd9  desktop/internal/usage/resource_identity.go
8f66925235db111c1b8c02412d624a63df8bdb51a9e1f7a1f81c93d8a90b09ba  desktop/internal/usage/resources_test.go
774e4002ee30f8098ee74aa3c7b745c9792786e55bf6c726916e1b445bce3b62  desktop/internal/usage/resources.go
e1cb79320d5a43f1fd4dac2db6d09d5d4b7a882d92b79286794657c4b1f58b4d  desktop/internal/usage/service.go
73e43a2fda29f06481b2bcf36c64c0e9b481c94e2a49117591d7fe6dabfd1490  desktop/internal/connection/connection.go
1691b262d50f6c04f2a50d4cde6dde5f997809c0916f1d36ea17977843e14fe0  desktop/internal/connection/studio_test.go
827d03af95b19ba627ac4de4d317cbc023978584bc5f91a3df6e99a2f31a2e93  desktop/internal/cloud/client.go
8cb308917f142d5e1fcdfc563dd10a9509afd199f7184cfa1b87d9dd487f9456  desktop/internal/host/usage.go
2044094cd25a3aca40a709d52cb51bf4374704ab0b69a22934cda4db4357dedd  desktop/internal/nativeui/collaboration.go
3f9fa81bfd365367a8d07621533ec3ae4c1046274fdfb97bc9f8d6d5084b512d  desktop/internal/nativeui/usage.go
ec4358211812d12d78a0919ccd4b17773070486e5fd408d16120c4140a5d60cd  prd-spec/features/agentkit-desktop/2026-10-27-runtime-usage-spec.md
```

## Handoff

Parent: apply only the two scoped repairs above, retain the rejected-candidate evidence, bind follow-up tests and any ledger entries to the repaired content, and rerun the applicable gates. The sections above report the pre-repair snapshot.

## 修复后复验（主 Agent）

- UB-01 已修复：只用已保存的成员别名生成地址；无别名时显示操作指引并禁用调用。
  `TestCollaborationMentionMatchesRuntimeGrammar`、`TestNativeCollaborationRequiresSavedAliasForSelectedMember`
  通过；Go 直连实际本地 Runtime 的别名写入与引用也通过。
- UB-02 已修复：Description 缺失/null 默认空字符串，错误类型仍拒绝。
  `TestResourcesSkillDescriptionsAreOptional` 与目录失败契约测试通过。
- 另修复 AGENTS.md 无变化保存的同 revision 回执兼容性，
  `TestResourcesAgentsMDNoOpKeepsRevision` 通过。
- 前端两轮审查发现的地址顺序与切换页面后回执清理问题已修复，
  全量 80 文件/428 项测试通过（`LANG=en_US.UTF-8`）。
- 主 Agent 对资源编辑器、取消、epoch、原生页面及调用入口执行了两轮自审，
  页面导航保留草稿、失败写入消耗预条件、重新读取保留草稿与最新基线、跨账户回执拒绝的反例均通过。
  额外独立原生审查因会话中断停止，未声称其完成。
- Go 全包、受影响包 race、vet、typecheck、lint 与构建通过。
  修复台账 `/tmp/openmpa-usage-review-ledger.json` 已经 `validate_review_ledger.py` 校验；
  本轮修复台账 5 confirmed/fixed、2 rejected，无未关闭发现。完整未验证线上范围保留在验收记录。
