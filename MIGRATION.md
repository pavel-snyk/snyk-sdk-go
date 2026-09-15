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

## Group is an SDK read model

`Group` is now a flat, SDK-owned read model rather than a public mirror of the
REST JSON:API resource:

```go
type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
```

Code that previously read a name through `group.Attributes.Name` must now use
`group.Name`. The public `Type`, `Attributes`, and `Relationships` fields have
been removed, along with the exported `GroupAttributes` and
`GroupRelationships` transport types. The Get-only slug, timestamps, and tenant
relationship are intentionally not part of this minimal read model.

The JSON representation of `Group` has consequently changed from the REST
JSON:API resource shape to the flat `id` and `name` fields shown above. `List`,
`All`, and `Get` still return `Group` values, but those values now use this
SDK-owned representation.

`Group.String()` now reflects the flattened `Group` representation, so its
diagnostic output has changed accordingly.

## Group list options and iteration

`ListGroupsOptions` has been replaced by the resource-first
`GroupListOptions`. The Groups List endpoint currently supports only the shared
pagination options, so the new type embeds `ListOptions` and does not add
speculative filters.

`GroupsService.List` and `GroupsService.All` now accept `*GroupListOptions`
instead of `*ListOptions`:

```go
groups, response, err := client.Groups.List(ctx, &snyk.GroupListOptions{
	ListOptions: snyk.ListOptions{Limit: 20},
})

groups, iterErr := client.Groups.All(ctx, &snyk.GroupListOptions{
	ListOptions: snyk.ListOptions{StartingAfter: cursor, Limit: 20},
})
```

`All` snapshots its options when it is constructed and rebuilds every request
from that snapshot plus the cursor returned by the preceding page. It does not
adopt the endpoint, version, limit, or other query values from a server link.
Repeated sequential iteration starts from the original cursor. Concurrent or
overlapping iteration remains unsupported.

Group iteration is now explicitly forward-only: an ending-before cursor is
rejected. Each page is converted atomically, so a malformed page yields none of
its Groups while Groups from earlier pages remain yielded.

## Group response validation

Groups List and Get now reject malformed JSON:API resources with an empty ID,
an unexpected resource type, or missing attributes. Missing or null response
`data` is also an error; `data: []` remains a valid empty List result. These
post-decode errors preserve the HTTP `*Response` returned alongside them.

`GroupsService.Get` now reports a missing Group ID as
`group ID: argument is empty`, and the error matches `snyk.ErrEmptyArgument`.
This replaces the previous ad-hoc `failed to get org: id must be supplied`
message.

## Broker deployment metadata updates

`BrokersService.UpdateDeployment` now treats a nil `Metadata` map as an omitted
field, preserving the deployment's existing metadata. Previously, nil was sent
as an empty object and cleared existing metadata. To clear metadata explicitly,
pass a non-nil empty map:

```go
request.Metadata = map[string]string{}
```

Create behavior is unchanged: nil and empty metadata maps are both sent as an
empty object. A populated map is sent as the replacement metadata for both
create and update requests.

The exported `KeyValueMap` helper has been removed. It had no remaining SDK use
after field-presence handling moved to the containing Broker request payloads.
Code that referred to it directly should use `map[string]string` instead.
