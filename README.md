# kintone-go

A Go client for the [kintone](https://kintone.cybozu.co.jp/) REST API.

This is an unofficial client.
The official SDK for kintone is [kintone/js-sdk](https://github.com/kintone/js-sdk),
which provides the REST API client for JavaScript as [`@kintone/rest-api-client`](https://github.com/kintone/js-sdk/tree/main/packages/rest-api-client).

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

<details>
<summary>Example</summary>

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

</details>

When writing records, the field type can be omitted.

<details>
<summary>Example</summary>

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

</details>

`AddRecords`, `UpdateRecords` and `UpsertRecords` process multiple records in a request.
`UpdateRecord`, `UpdateRecords` and `UpsertRecords` can also specify a record by
the value of a field with `UpdateKey` instead of `ID`.

### Cursor

`AllRecords` iterates over all the records matching the query using a cursor.
The cursor is deleted when the iteration stops early.

<details>
<summary>Example</summary>

```go
for record, err := range client.AllRecords(ctx, kintone.CreateCursorRequest{
	App:   app,
	Query: `status in ("Open") order by $id asc`,
}) {
	if err != nil {
		return err
	}
	fmt.Println(record["$id"].Value)
}
```

</details>

When kintone returns a warning in the `X-Cybozu-Warning` response header, `AllRecords` yields a `*kintone.WarningError` after all the records.
`GetRecords` and `CreateCursor` return the warning in `Warning` of the response.

`CreateCursor`, `GetRecordsByCursor` and `DeleteCursor` are also available to use the cursor directly.

`AllRecordsByID` iterates over all the records in ascending order of the record ID without a cursor,
using the condition on `$id` to get the next records.

<details>
<summary>Example</summary>

```go
for record, err := range client.AllRecordsByID(ctx, kintone.AllRecordsByIDRequest{
	App:       app,
	Condition: `status in ("Open")`,
}) {
	if err != nil {
		return err
	}
	fmt.Println(record["$id"].Value)
}
```

</details>

### Bulk request

`NewBulkRequest` sends multiple record operations in one request.
The results are available from the values returned by the methods after `Send`.
If a request in the bulk request fails, `Send` returns a `*kintone.BulkError` with the index of the request.
Other errors are returned as is, and `kintone.IsErrorCode` works for both.

<details>
<summary>Example</summary>

```go
bulk := client.NewBulkRequest()
added := bulk.AddRecord(app, kintone.Record{"title": {Value: "Hello"}})
bulk.UpdateStatus(app, kintone.StatusUpdate{ID: id, Action: "Submit"})
if err := bulk.Send(ctx); err != nil {
	return err
}
fmt.Println(added.Value().ID)
```

</details>

`AddAllRecords`, `UpdateAllRecords`, `UpsertAllRecords` and `DeleteAllRecords` process any number of records
by splitting them into bulk requests, which are sent one after another.
If a bulk request fails, they return a `*kintone.AllRecordsError` with the number of the records processed before it,
and the methods other than `DeleteAllRecords` also return the results of those records.

<details>
<summary>Example</summary>

```go
added, err := client.AddAllRecords(ctx, app, records)
if err != nil {
	var e *kintone.AllRecordsError
	if errors.As(err, &e) {
		fmt.Printf("%d records added before the error\n", e.Processed)
	}
	return err
}
fmt.Println(len(added))
```

</details>

### Apps

App settings are changed in the preview environment, and deployed to the production environment by `DeployApps`.
The getters of the settings get the production settings, or the preview settings with `Preview: true`.

<details>
<summary>Example</summary>

```go
added, err := client.AddApp(ctx, kintone.AddAppRequest{Name: "Deals"})
if err != nil {
	return err
}
_, err = client.AddFormFields(ctx, added.App, map[string]kintone.FieldProperty{
	"title": {Type: kintone.FieldTypeSingleLineText, Code: "title", Label: "Title", Required: new(true)},
}, 0)
if err != nil {
	return err
}

settings, err := client.GetGeneralSettings(ctx, kintone.GetAppSettingsRequest{App: added.App, Preview: true})
if err != nil {
	return err
}
_, err = client.UpdateGeneralSettings(ctx, added.App, kintone.GeneralSettings{
	Description:    new("Deals and activities."),
	EnableComments: new(false),
}, settings.Revision)
if err != nil {
	return err
}

if err := client.DeployApps(ctx, []kintone.AppDeployment{{App: added.App}}); err != nil {
	return err
}
err = client.WaitForDeploy(ctx, []int64{added.App})
```

</details>

### Spaces

`AddSpaceFromTemplate` adds a space from a space template, and `AddSpace` adds a public single thread space by the name.

<details>
<summary>Example</summary>

```go
space, err := client.GetSpace(ctx, spaceID)
if err != nil {
	return err
}

thread, err := client.AddThread(ctx, space.ID, "Weekly report")
if err != nil {
	return err
}
_, err = client.AddThreadComment(ctx, space.ID, thread, kintone.ThreadComment{
	Text:     "Please check the report.",
	Mentions: []kintone.Mention{{Code: "alice", Type: kintone.EntityTypeUser}},
})
if err != nil {
	return err
}

members, err := client.GetSpaceMembers(ctx, space.ID)
if err != nil {
	return err
}
members = slices.DeleteFunc(members, func(m kintone.SpaceMember) bool {
	return m.IsImplicit
})
members = append(members, kintone.SpaceMember{
	Entity: kintone.FieldEntity{Type: kintone.EntityTypeUser, Code: "bob"},
})
err = client.UpdateSpaceMembers(ctx, space.ID, members)
```

</details>

To access the apps and the spaces in a guest space, create the client with `kintone.WithGuestSpace`.

### Plugins

<details>
<summary>Example</summary>

```go
f, err := os.Open("plugin.zip")
if err != nil {
	return err
}
defer f.Close()
fileKey, err := client.UploadFile(ctx, "plugin.zip", f)
if err != nil {
	return err
}
plugin, err := client.AddPlugin(ctx, fileKey)
if err != nil {
	return err
}
if _, err := client.AddAppPlugins(ctx, app, []string{plugin.ID}, 0); err != nil {
	return err
}
err = client.DeployApps(ctx, []kintone.AppDeployment{{App: app}})
```

</details>

### Files

`UploadFile` uploads a file and returns the file key, which can be set to an attachment field.
`DownloadFile` downloads the file of the file key in an attachment field.

<details>
<summary>Example</summary>

```go
f, err := os.Open("report.pdf")
if err != nil {
	return err
}
defer f.Close()
fileKey, err := client.UploadFile(ctx, "report.pdf", f)
if err != nil {
	return err
}
_, err = client.AddRecord(ctx, app, kintone.Record{
	"attachments": {Value: []kintone.FileInfo{{FileKey: fileKey}}},
})
if err != nil {
	return err
}

record, err := client.GetRecord(ctx, app, id)
if err != nil {
	return err
}
file := record["attachments"].Value.([]kintone.FileInfo)[0]
body, _, err := client.DownloadFile(ctx, file.FileKey)
if err != nil {
	return err
}
defer body.Close()
out, err := os.Create(file.Name)
if err != nil {
	return err
}
defer out.Close()
_, err = io.Copy(out, body)
```

</details>

### API info

`GetAPIs` and `GetAPISchema` get the list of the kintone REST APIs and their JSON Schemas.

<details>
<summary>Example</summary>

```go
apis, err := client.GetAPIs(ctx)
if err != nil {
	return err
}
for id := range apis.APIs {
	schema, err := client.GetAPISchema(ctx, id)
	if err != nil {
		return err
	}
	fmt.Println(id, schema.HTTPMethod, schema.Path)
}
```

</details>

### Errors

Error responses from kintone are returned as `*kintone.Error`,
which has the HTTP status code, the error code, the error ID, the message and the response header.

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
