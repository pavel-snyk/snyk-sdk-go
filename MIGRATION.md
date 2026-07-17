# Migrating to the v2 technical preview

The v2 API is being redesigned before stabilization. The changes below are
deliberately source-breaking and establish conventions for the stable SDK.

## Services are concrete

`Client` exposes concrete service pointers such as `*ProjectsService` instead
of SDK-owned interfaces such as `ProjectsServiceAPI`. This lets services gain
new operations without breaking every external implementation of an interface
owned by the SDK.

Consumers that need substitution or mocking should define the smallest
interface their own code consumes:

```go
type ProjectGetter interface {
	Get(
		context.Context,
		string,
		string,
	) (*snyk.Project, *snyk.Response, error)
}

type Handler struct {
	projects ProjectGetter
}

func NewHandler(projects ProjectGetter) *Handler {
	return &Handler{projects: projects}
}
```

Production code can pass `client.Projects`; tests can pass a consumer-owned
fake. Testability therefore lives at the consumer's dependency boundary rather
than by replacing fields inside the SDK client.

## Project is an SDK read model

`Project` is now a flat, SDK-owned read model rather than a public mirror of
the REST JSON:API resource. `ProjectAttributes` has been removed. Code that
previously read fields through `project.Attributes` must use the corresponding
fields directly on `Project`.

The SDK's JSON field names for `Project` are also distinct from the REST wire
representation and form part of the v2 public contract.

## Project iteration

`ProjectsService.All` accepts `*ProjectListOptions`, so the same filters used
for one page can be preserved across every page. The options are deeply
snapshotted when `All` is called.

`ProjectsService.List` no longer supplies an SDK default limit of 100 when its
options are nil. It now omits the limit and uses the endpoint's server-defined
default, like other REST list operations.

Iteration is forward-only: an ending-before cursor is rejected. Each page is
converted atomically, so a malformed page yields none of its Projects while
Projects from previously completed pages remain yielded. The returned sequence
may be iterated multiple times sequentially, but concurrent or overlapping
iteration is not supported.
