# Snyk Go SDK - Agent Instructions

## Philosophy

The SDK is designed to:

- preserve stable public contracts
- make deliberate technical-preview or major-version breaking changes explicit
  and provide migration guidance
- provide predictable behavior across services
- remain idiomatic Go
- provide a consistent developer experience

Agents must prioritize semantic correctness, consistency, and safety over local
optimizations or large-scale refactors. Follow documented conventions and the
closest applicable paved-path implementation. Do not reproduce a legacy pattern
merely because it exists nearby.

## 1. Core SDK Design

1.1. This is a **service-oriented** SDK, not object-oriented.
1.2. Public APIs are explicit and stateless.
1.3. Users always pass IDs and request structs directly — no hidden
lookups, no implicit state on the client beyond configuration.
1.4. Public models must not own service operations, perform network requests,
or depend on client state. Service behavior lives on the `*Service` types.

## 2. Public API Design

The SDK is optimized for predictable, idiomatic public APIs, subject to
semantic correctness.

- Prefer consistency across services over mirroring backend API structure
  1:1. If you just want a 1:1 mirror, generate one from the OpenAPI spec —
  that's not this SDK's job.
- Do not misrepresent backend behavior or discard information promised by the
  public API. An SDK-owned model may intentionally omit wire fields, but it
  must accurately represent the fields and behavior it exposes.
- Similar operations SHOULD have similar method shapes and return types.
- Behavior MUST NOT be surprising across services.
- APIs SHOULD be guessable without reading implementation details.

When multiple semantically correct designs exist, prefer the design most
consistent with the SDK.

This is a guiding principle for new/changed APIs, not a mandate to
retrofit existing public APIs without a real reason.

## 3. Settled Decisions (do not relitigate)

These rules are not defaults or guidelines — they are closed. If you
disagree, raise it in a separate discussion first; do not work around
them in a PR.

- **No built-in retry/backoff.** The SDK never sleeps or resends a
  request on the caller's behalf. It parses and exposes `Retry-After`
  (`ErrorResponse.RetryAfter`) and `Sunset` (`Response.Sunset`) as data.
  Callers compose their own retry via `WithHTTPClient` + a custom
  `http.RoundTripper`. Do not add a retry `ClientOption`.
- **API version is not a caller-supplied parameter.** Verify the version of
  every endpoint individually. Each endpoint's version is selected by an
  internal constant; endpoints may share a resource- or endpoint-group constant
  only when they intentionally use the same version. Version changes are
  deliberate, reviewed code changes tied to an SDK release. Do not add a public
  `Version` field to option structs or a global version-override `ClientOption`.
- The SDK surfaces server-provided facts (`SnykRequestID`,
  `ServedAPIVersion`, `RetryAfter`, `Sunset`) as struct fields. It never
  acts on them automatically.
- **Keep each resource lifecycle in one service file**, including its secondary
  endpoint groups. Split only when the same resource spans distinct API
  families, such as REST and legacy v1; use the `_v1.go` suffix for the legacy
  implementation. Do not split files for REST version changes.

## 4. Conventions

- Every exported service method has a doc comment linking to the
  relevant Snyk API docs page.
- New services follow the checklist in §5.
- `context.Context` is always the first parameter.
- Expose services from `Client` as concrete pointers. Do not define SDK-owned
  service interfaces. Consumers define narrow interfaces containing only the
  SDK operations they use, including for mocking.
- Name new operation-specific public types with the resource first so related
  APIs group in autocomplete, for example `ProjectListOptions`,
  `ProjectCreateRequest`, and `ProjectUpdateRequest`.
- Representation methods such as `String`, `MarshalJSON`, or `UnmarshalJSON`
  may live on public models when they support the model's documented
  representation.

### 4.1 Test Fixtures

Store large or realistic response fixtures in `snyk/testdata`, normally
using lowercase snake_case names of the form:

```text
<service>_<method>_<scenario>.json
```

For example:

```text
projects_get_success.json
projects_get_expanded_target.json
projects_all_page_2.json
```

Always include a scenario. Use `success` for the ordinary happy path and
a more specific scenario when it communicates more, such as
`expanded_target` or `not_found`.

Keep small, test-specific response bodies inline when that is clearer.
Use unmistakably synthetic fixture data, such as repeated-digit UUIDs and
clearly fictional resource names. Never copy customer or production data.
Load fixtures through the shared
`loadFixture(t *testing.T, fixtureName string) []byte` test helper.

## 5. Adding a New Service (checklist)

Follow these steps in order. Each step has a known failure mode if skipped.

1. **Check every endpoint's API version first.** Look up each target endpoint in
   [Snyk API docs](https://docs.snyk.io/snyk-api/reference) and record
   its current version string (e.g. `2024-10-15`). Don't copy it from a
   neighboring service or endpoint — Snyk versions per-endpoint, so it may
   differ.

2. **Create `<resource>.go`** (or `<resource>_v1.go` for a legacy v1 API
   resource), following the one-resource-file decision in §3.

3. **Declare internal version constants** at the top of the file. Use one
   appropriately named constant for each endpoint or group of endpoints that
   shares the same verified version. Build REST request paths with
   `restPath(endpoint, versionConstant, opts)`; never expose the version on
   public option structs.

4. **Implement `XService`** as `type XService service`. Add each public
   operation directly to this concrete service, with a doc comment linking to
   the Snyk API docs page for that endpoint.

5. **Wire into `Client`** in `client.go` — two places:

   - Add the field to the `Client` struct: `X *XService`
   - Assign it in `NewClient`: `c.X = (*XService)(&c.common)`

   Forgetting this step means the service compiles but is unreachable from
   `client.X`.

6. **Add operation-appropriate tests.** As applicable, cover argument
   validation, HTTP method/path/version/query/body construction, structural
   response validation, transport-to-public-model mapping, pagination, and
   service-specific behavior. Test shared transport and error contracts at the
   shared client layer rather than duplicating them for every operation.

## 6. Verification

- Run `make test` for the full test suite and coverage report.
- Run `make lint` before considering implementation work complete.
- Run `git diff --check` for every change.
- Run `go test -race ./...` for changes involving iterators, shared state,
  concurrency, or request lifecycle behavior. It is not required for
  documentation-only changes.

## 7. Service File Organization

Keep service files readable from top to bottom: service declarations and
primary public types first, followed by private transport types, primary
service methods, and related mapping and helper functions.

Keep methods such as `String` close to the public model they belong to.

Keep public and private types, methods, and helpers used only by a secondary
endpoint group near that group. Treat this as a readability guideline, not a
reason to mechanically reorder existing files or create unrelated churn.
