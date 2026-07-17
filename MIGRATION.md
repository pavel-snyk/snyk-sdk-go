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
