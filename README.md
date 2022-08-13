# Snyk SDK for Go

snyk-sdk-go is the (un)official Snyk SDK for the Go programming language.

## Installation

```sh
# X.Y.Z is the version you need
go get github.com/pavel-snyk/snyk-sdk-go@vX.Y.Z

# for non Go modules usage or latest version
go get github.com/pavel-snyk/snyk-sdk-go
```

## Usage

```go
import "github.com/pavel-snyk/snyk-sdk-go"
```

Create a new Snyk client, then use the exposed services to access different
parts of the Snyk API.

### Authentication

To use the SDK, you must get your API token from Snyk. You can find your token
in your General Account Settings on https://snyk.io/account/ after you register
with Snyk and log in. See [Authentication for API](https://docs.snyk.io/snyk-api-info/authentication-for-api).

```go
client := snyk.NewClient("your-api-token")
```
