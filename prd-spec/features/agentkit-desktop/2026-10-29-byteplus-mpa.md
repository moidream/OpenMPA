# Studio MPA BytePlus support in OpenMPA

## Scope

Expose BytePlus in account settings and carry the selected provider through the
native cloud client and bundled Studio MPA lifecycle. Existing requests and saved
credentials that omit the provider retain Volcengine behavior. No SDK version
upgrade is required: the pinned veadk-python 1.1.19 managed BP modules match the
local Studio reference implementation.

## Behavior

- Account platform selection: Volcengine (Beijing/Shanghai) or BytePlus (Singapore,
  ap-southeast-1). Switching the login selection clears entered credentials.
- The provider travels with the in-memory credential, encrypted vault payload,
  signed-in state and private lifecycle-engine initialization. Requests reject
  mismatched provider/region pairs before sending credentials.
- BP STS and ModelArk control plane use open.byteplusapi.com. AgentKit and
  Function Service use their regional byteplusapi.com hosts. Model catalog data
  uses ark.ap-southeast.bytepluses.com and TOS uses tos-ap-southeast-1.bytepluses.com.
- Runtime directory, binding, image updates, tool and skill management use the
  provider-aware cloud client. Existing runtime HTTP APIs remain unchanged.
- Studio creation routes receive provider=byteplus and retain it in creation
  name scope, journal payload and runner. Defaults cover BP Runtime/Worker images,
  IAM, networking/gateway, database workspaces, OpenViking, model configuration,
  and observability. Deletion's RuntimeCloud receives the same provider.
- Model availability and Responses metadata checks retain their existing logic,
  using the selected platform's catalog and the original credential snapshot.
- BP local scopes use byteplus:<account>/<identity>/<principal>. Domestic scopes
  are unchanged, preserving runtime preferences, cache keys, creation journals
  and ownership tags. BP and domestic accounts cannot share those local records.
- Old encrypted vault payloads without provider still unlock as Volcengine;
  newly saved BP payloads restore BytePlus. No plaintext secret metadata added.

## Existing limits retained

Only Studio's supported regions are offered. Shanghai runtime management remains
available; Studio currently has no managed creation profile for Shanghai, so the
existing unconfigured response remains. No fallback to another cloud is allowed.
Desktop process environment isolation is retained; Studio server environment
credential/endpoint overrides are not imported from the host shell.

## Verification

- Native contract tests cover BP identity/signature/host selection, model catalog
  activation and permission behavior, TOS region filtering, webshell routing,
  cross-provider rejection, vault restore, and legacy credential compatibility.
- Python tests exercise the actual private subprocess and pinned Studio creation
  routes for both providers, including BP payload/name scope, deletion factory,
  images, database/memory and Worker environment defaults.
- Packaged engine tests extract the distributable archive and load both domestic
  and BP profiles through the Go-to-Python private bridge.
- React tests cover login platform switching, credential clearing, region
  filtering, and BP Runtime directory when the first account snapshot is delayed.
- Browser fixture script: scripts/verify-desktop-byteplus.mjs. It uses synthetic
  credentials and a mocked native bridge, never cloud credentials or writes.

Live BP account creation/deletion is not verified by these offline tests.

Verification results (2026-10-29): relevant native packages passed; both packaged
engine smoke tests passed (65.5 s). Python lifecycle/diagnostic suites passed 95
existing tests, and the provider suite passed all 6 cases after correcting the
Shanghai expectation to the existing unsupported-template behavior. The frontend
lifecycle/provider suites passed 58 tests; runtime directory, page service and
runtime configuration regression tests also passed. Typecheck, ESLint, design
checks, desktop frontend build and diff whitespace checks passed. Browser checks
passed all 7 journeys, with English and Chinese screenshots under
`.openmpa/byteplus-verification/2026-10-29T13-50-24.649Z/`.
