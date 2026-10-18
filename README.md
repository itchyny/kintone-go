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

### Records

A record is a map from the field code to the field.
When reading records, the field value is decoded into the Go type of the field type,
such as `string`, `[]string`, `kintone.Entity`, `[]kintone.Entity`,
`[]kintone.FileInfo` and `[]kintone.SubtableRow`.

```go
record, err := client.GetRecord(ctx, app, id)
if err != nil {
	return err
}
title := record["title"].Value.(string)

res, err := client.GetRecords(ctx, kintone.GetRecordsRequest{
	App:    app,
	Fields: []string{"$id", "title"},
	Query:  `status in ("Open") order by $id asc limit 100`,
})
if err != nil {
	return err
}
for _, record := range res.Records {
	fmt.Println(record["$id"].Value, record["title"].Value)
}
```

When writing records, the field type can be omitted.

```go
added, err := client.AddRecord(ctx, app, kintone.Record{
	"title": {Value: "Hello"},
	"users": {Value: []kintone.Entity{{Code: "alice"}}},
})
if err != nil {
	return err
}

revision, err := client.UpdateRecord(ctx, app, kintone.RecordUpdate{
	ID:       added.ID,
	Record:   kintone.Record{"title": {Value: "Hello, world"}},
	Revision: added.Revision,
})
if err != nil {
	return err
}

err = client.DeleteRecords(ctx, app, []int64{added.ID}, []int64{revision})
```

`AddRecords`, `UpdateRecords` and `UpsertRecords` process multiple records in a request.
`UpdateRecord`, `UpdateRecords` and `UpsertRecords` can also specify a record by
the value of a field with `UpdateKey` instead of `ID`.

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
