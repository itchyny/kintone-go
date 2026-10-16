# kintone-go

A Go client for the [kintone](https://kintone.cybozu.co.jp/) REST API.

## Installation

```sh
go get github.com/itchyny/kintone-go
```

## Usage

```go
import "github.com/itchyny/kintone-go"

client, err := kintone.NewClient(
	"https://example.cybozu.com",
	kintone.WithAPIToken(os.Getenv("KINTONE_API_TOKEN")),
)
if err != nil {
	return err
}
```

### Authentication

| Option | Description |
| --- | --- |
| `WithAPIToken(tokens...)` | API tokens. Specify multiple tokens to access multiple apps, such as on lookup fields. |
| `WithPassword(username, password)` | Login name and password. |
| `WithOAuthToken(token)` | OAuth access token. |
| `WithBasicAuth(username, password)` | Basic authentication, required when the domain enables it. |

### Other options

| Option | Description |
| --- | --- |
| `WithGuestSpace(id)` | Access apps in the guest space. |
| `WithHTTPClient(httpClient)` | Use the HTTP client for requests. |
| `WithUserAgent(userAgent)` | Set the User-Agent header. |

### Errors

Error responses from kintone are returned as `*kintone.Error`,
which has the HTTP status code, the error code, the error ID and the message.

```go
if kintone.IsErrorCode(err, "GAIA_RE01") {
	// the record is not found
}
```

## Bug Tracker

Report bugs at [Issues - itchyny/kintone-go - GitHub](https://github.com/itchyny/kintone-go/issues).

## Author

itchyny (<https://github.com/itchyny>)

## License

This software is released under the MIT License, see [LICENSE](LICENSE).
