# Snyk Go SDK - Agent Instructions

## Philosophy

The SDK is designed to be:

- stable and backward compatible
- predictable across services
- idiomatic Go
- consistent in developer experience

Agents must prioritize **consistency and safety** over local optimizations
or large-scale refactors. When in doubt, match the existing pattern in a
neighboring service file rather than introducing a new one.

## 1. Core SDK Design

1.1. This is a **service-oriented** SDK, not object-oriented.
1.2. Public APIs are explicit and stateless.
1.3. Users always pass IDs and request structs directly — no hidden
lookups, no implicit state on the client beyond configuration.
1.4. Resource structs (`Organization`, `Project`, `BrokerDeployment`, ...)
MUST NOT contain behavior methods. `String()` for debugging is the
only exception. Behavior lives on the `*Service` types.

## 2. Developer Experience (HIGHEST PRIORITY)

The SDK is optimized for developer experience and predictability over
strict backend fidelity.

- Prefer consistency across services over mirroring backend API structure
  1:1. If you just want a 1:1 mirror, generate one from the OpenAPI spec —
  that's not this SDK's job.
- Similar operations SHOULD have similar method shapes and return types.
- Behavior MUST NOT be surprising across services.
- APIs SHOULD be guessable without reading implementation details.

**Tradeoff rule:** backend fidelity vs. SDK consistency → SDK consistency wins.

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
- **API version is not a caller-supplied parameter.** Each service
  hardcodes its own version constant (e.g. `orgsAPIVersion`), bumped
  deliberately as a reviewed code change tied to an SDK release. Do not
  add a public `Version` field to option structs or a global
  version-override `ClientOption`. (Snyk versions per-endpoint, not
  per-service or globally — a single override can't express that
  correctly, so don't add one that pretends to.)
- The SDK surfaces server-provided facts (`SnykRequestID`,
  `ServedAPIVersion`, `RetryAfter`, `Sunset`) as struct fields. It never
  acts on them automatically.

## 4. Conventions

- One file per Snyk resource/concept (`orgs.go`, `brokers.go`, ...),
  even when large. Don't split a single resource's sub-concepts
  (e.g. Broker deployments/connections/credentials) into separate files —
  they're one lifecycle, keep them together. Split only when the same
  resource spans two distinct Snyk API families: `orgs.go` wraps the REST
  API (`/rest/`), `orgs_v1.go` wraps the legacy v1 API (`/v1/`). Use the
  `_v1` suffix to mark v1 API files. Don't create `_v2.go`-style splits
  for REST version bumps — those stay in the same file with an updated
  version constant.
- Every exported service method has a doc comment linking to the
  relevant Snyk API docs page.
- New services follow the checklist in §5.
- `context.Context` is always the first parameter.

## 5. Adding a New Service (checklist)

Follow these steps in order. Each step has a known failure mode if skipped.

1. **Check the API version first.** Look up the target endpoint(s) in
   [Snyk API docs](https://docs.snyk.io/snyk-api/reference) and record
   the current version string (e.g. `2024-10-15`). Don't copy it from a
   neighboring service — Snyk versions per-endpoint, so it may differ.

2. **Create `xservice.go`** (or `xservice_v1.go` for a legacy v1 API
   resource). One file for the whole resource lifecycle.

3. **Declare the version constant** at the top of the file:
   `const xAPIVersion = "<version-from-step-1>"`.
   Build REST request paths with
   `restPath(endpoint, xAPIVersion, opts)`; never expose the version on
   public option structs.

4. **Define `XServiceAPI` interface** with all public methods, each with
   a doc comment linking to the Snyk API docs page for that endpoint.

5. **Implement `XService`** as `type XService service` and add the
   compile-time check: `var _ XServiceAPI = (*XService)(nil)`.

6. **Wire into `Client`** in `client.go` — two places:
   - Add the field to the `Client` struct: `X XServiceAPI`
   - Assign it in `NewClient`: `c.X = (*XService)(&c.common)`

Forgetting this step means the service compiles but is unreachable
from `client.X`.

## 6. Verification

- Run `make test` for the full test suite and coverage report.
- Run `make lint` before considering implementation work complete.
