# Changelog
## [v0.1.0](https://github.com/itchyny/kintone-go/compare/6b46947..v0.1.0) (2026-10-20)
* initial implementation of the kintone REST API client
* authenticate with API tokens, password, OAuth access token, and Basic authentication,
  with options of guest spaces, HTTP client, User-Agent, and Accept-Language
* implement app APIs to get and add apps, deploy the settings, and wait for the
  deployments (`WaitForDeploy`)
* implement APIs of the app settings: form fields, layout, views, graphs, general
  settings, process management, customization, notifications, permissions, actions,
  plugins, categories, and admin notes
* implement record APIs to get, add, update, upsert, and delete records, with typed field
  values and `Quote` for query values
* implement cursor APIs, and iterators over all records by a cursor (`AllRecords`) and by
  record IDs (`AllRecordsByID`)
* implement bulk request with typed results
* implement `AddAllRecords`, `UpdateAllRecords`, `UpsertAllRecords`, and
  `DeleteAllRecords` beyond the per-request limits
* implement APIs of record comments, statuses, assignees, and record permissions
* implement file APIs to upload and download files
* implement space APIs of spaces, members, threads, and guests
* implement plugin APIs and API information
* report errors with the code, the message, the details, and the response header
* implement `IsErrorCode` to check the code of an error
