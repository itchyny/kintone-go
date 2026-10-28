package kintone

import (
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestClientSearch(t *testing.T) {
	testCases := []struct {
		name     string
		req      SearchRequest
		wantBody string
	}{
		{
			name: "all parameters",
			req: SearchRequest{
				Query: []SearchQuery{
					{Operator: SearchOperatorAnd, Keywords: []string{"estimate", "approval"}},
					{Operator: SearchOperatorNot, Keywords: []string{"draft"}},
				},
				Types:           []SearchType{SearchTypeRecord, SearchTypeAttachment},
				Scopes:          []SearchScope{{Scope: SearchScopeTypeApp, IDs: []int64{1, 2}}},
				ExcludeScopes:   []SearchScope{{Scope: SearchScopeTypePeople, Codes: []string{"alice"}}},
				CreatedAfter:    time.Date(2026, 10, 1, 0, 0, 0, 0, time.FixedZone("", 9*60*60)),
				CreatedBefore:   time.Date(2026, 10, 31, 23, 59, 59, 0, time.UTC),
				Creators:        []string{"bob"},
				IncludeSynonyms: true,
				Sort:            SearchSort{By: SearchSortByCreatedAt, Order: SortOrderAsc},
				Limit:           10,
				PageToken:       "token1",
			},
			wantBody: `{
				"query": [
					{"operator": "AND", "keywords": ["estimate", "approval"]},
					{"operator": "NOT", "keywords": ["draft"]}
				],
				"types": ["RECORD", "ATTACHMENT"],
				"scopes": [{"scope": "APP", "ids": [1, 2]}],
				"excludeScopes": [{"scope": "PEOPLE", "codes": ["alice"]}],
				"createdAfter": "2026-10-01T00:00:00+09:00",
				"createdBefore": "2026-10-31T23:59:59Z",
				"creators": ["bob"],
				"includeSynonyms": true,
				"sort": {"by": "CREATED_AT", "order": "ASC"},
				"limit": 10,
				"pageToken": "token1"
			}`,
		},
		{
			name: "only required parameters",
			req: SearchRequest{
				Query: []SearchQuery{{Operator: SearchOperatorOr, Keywords: []string{"estimate"}}},
			},
			wantBody: `{"query": [{"operator": "OR", "keywords": ["estimate"]}]}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ts, got := newTestServer(t, http.StatusOK, `{
				"hits": [
					{
						"type": "RECORD",
						"url": "https://example.cybozu.com/k/1/show#record=2",
						"snippets": ["estimate approval"],
						"record": {
							"appId": "1",
							"appName": "Estimates",
							"recordId": "2",
							"recordTitle": "Estimate 2",
							"createdAt": "2026-10-16T01:23:45.678Z",
							"creator": {"code": "bob", "name": "Bob"},
							"matchedFields": [{"code": "title", "label": "Title"}],
							"spaceId": null,
							"spaceName": null
						}
					},
					{
						"type": "RECORD_COMMENT",
						"url": "https://example.cybozu.com/k/1/show#record=2&comment=3",
						"snippets": [],
						"recordComment": {
							"appId": "1",
							"appName": "Estimates",
							"recordId": "2",
							"recordTitle": "Estimate 2",
							"commentId": "3",
							"createdAt": null,
							"creator": null,
							"spaceId": "4",
							"spaceName": "Sales"
						}
					},
					{
						"type": "SPACE",
						"url": "https://example.cybozu.com/k/#/space/4",
						"snippets": [],
						"space": {
							"spaceId": "4",
							"spaceName": "Sales",
							"createdAt": "2026-10-16T01:23:45.678Z",
							"creator": {"code": "bob", "name": "Bob"}
						}
					},
					{
						"type": "THREAD",
						"url": "https://example.cybozu.com/k/#/space/4/thread/5",
						"snippets": [],
						"thread": {
							"threadId": "5",
							"threadName": "Estimates",
							"spaceId": "4",
							"spaceName": "Sales",
							"createdAt": "2026-10-16T01:23:45.678Z",
							"creator": {"code": "bob", "name": "Bob"}
						}
					},
					{
						"type": "THREAD_COMMENT",
						"url": "https://example.cybozu.com/k/#/space/4/thread/5/6/7",
						"snippets": [],
						"threadComment": {
							"commentId": "6",
							"replyId": "7",
							"spaceId": "4",
							"spaceName": "Sales",
							"threadId": "5",
							"threadName": "Estimates",
							"createdAt": "2026-10-16T01:23:45.678Z",
							"creator": {"code": "bob", "name": "Bob"}
						}
					},
					{
						"type": "PEOPLE_COMMENT",
						"url": "https://example.cybozu.com/k/#/people/user/alice/8",
						"snippets": [],
						"peopleComment": {
							"commentId": "8",
							"replyId": null,
							"owner": {"code": "alice", "name": "Alice"},
							"createdAt": "2026-10-16T01:23:45.678Z",
							"creator": {"code": "bob", "name": "Bob"}
						}
					},
					{
						"type": "MESSAGE_COMMENT",
						"url": "https://example.cybozu.com/k/#/message/9",
						"snippets": [],
						"messageComment": {
							"commentId": "9",
							"createdAt": "2026-10-16T01:23:45.678Z",
							"creator": {"code": "bob", "name": "Bob"},
							"recipient": {"code": "alice", "name": "Alice"}
						}
					},
					{
						"type": "ATTACHMENT",
						"url": "https://example.cybozu.com/k/1/show#record=2",
						"snippets": [],
						"record": {
							"appId": "1",
							"appName": "Estimates",
							"recordId": "2",
							"recordTitle": "Estimate 2",
							"createdAt": "2026-10-16T01:23:45.678Z",
							"creator": {"code": "bob", "name": "Bob"},
							"matchedFields": [],
							"spaceId": null,
							"spaceName": null
						},
						"attachment": {
							"attachedTo": "RECORD",
							"fileKey": "file1",
							"name": "estimate.pdf",
							"createdAt": "2026-10-16T01:23:45.678Z",
							"creator": {"code": "bob", "name": "Bob"}
						}
					}
				],
				"nextPageToken": "token2"
			}`)
			c := newTestClient(t, ts)
			res, err := c.Search(t.Context(), tc.req)
			if err != nil {
				t.Fatalf("Search returned error: %v", err)
			}
			assertRequest(t, got, http.MethodPost, "/k/v1/search.json", tc.wantBody)
			createdAt := time.Date(2026, 10, 16, 1, 23, 45, 678000000, time.UTC)
			bob := Entity{Code: "bob", Name: "Bob"}
			alice := Entity{Code: "alice", Name: "Alice"}
			want := &SearchResponse{
				Hits: []SearchHit{
					{
						Type:     SearchTypeRecord,
						URL:      "https://example.cybozu.com/k/1/show#record=2",
						Snippets: []string{"estimate approval"},
						Record: &SearchHitRecord{
							AppID:         1,
							AppName:       "Estimates",
							RecordID:      2,
							RecordTitle:   "Estimate 2",
							CreatedAt:     createdAt,
							Creator:       bob,
							MatchedFields: []MatchedField{{Code: "title", Label: "Title"}},
						},
					},
					{
						Type:     SearchTypeRecordComment,
						URL:      "https://example.cybozu.com/k/1/show#record=2&comment=3",
						Snippets: []string{},
						RecordComment: &SearchHitRecordComment{
							AppID:       1,
							AppName:     "Estimates",
							RecordID:    2,
							RecordTitle: "Estimate 2",
							CommentID:   3,
							SpaceID:     4,
							SpaceName:   "Sales",
						},
					},
					{
						Type:     SearchTypeSpace,
						URL:      "https://example.cybozu.com/k/#/space/4",
						Snippets: []string{},
						Space: &SearchHitSpace{
							SpaceID:   4,
							SpaceName: "Sales",
							CreatedAt: createdAt,
							Creator:   bob,
						},
					},
					{
						Type:     SearchTypeThread,
						URL:      "https://example.cybozu.com/k/#/space/4/thread/5",
						Snippets: []string{},
						Thread: &SearchHitThread{
							ThreadID:   5,
							ThreadName: "Estimates",
							SpaceID:    4,
							SpaceName:  "Sales",
							CreatedAt:  createdAt,
							Creator:    bob,
						},
					},
					{
						Type:     SearchTypeThreadComment,
						URL:      "https://example.cybozu.com/k/#/space/4/thread/5/6/7",
						Snippets: []string{},
						ThreadComment: &SearchHitThreadComment{
							CommentID:  6,
							ReplyID:    7,
							SpaceID:    4,
							SpaceName:  "Sales",
							ThreadID:   5,
							ThreadName: "Estimates",
							CreatedAt:  createdAt,
							Creator:    bob,
						},
					},
					{
						Type:     SearchTypePeopleComment,
						URL:      "https://example.cybozu.com/k/#/people/user/alice/8",
						Snippets: []string{},
						PeopleComment: &SearchHitPeopleComment{
							CommentID: 8,
							Owner:     alice,
							CreatedAt: createdAt,
							Creator:   bob,
						},
					},
					{
						Type:     SearchTypeMessageComment,
						URL:      "https://example.cybozu.com/k/#/message/9",
						Snippets: []string{},
						MessageComment: &SearchHitMessageComment{
							CommentID: 9,
							CreatedAt: createdAt,
							Creator:   bob,
							Recipient: alice,
						},
					},
					{
						Type:     SearchTypeAttachment,
						URL:      "https://example.cybozu.com/k/1/show#record=2",
						Snippets: []string{},
						Record: &SearchHitRecord{
							AppID:         1,
							AppName:       "Estimates",
							RecordID:      2,
							RecordTitle:   "Estimate 2",
							CreatedAt:     createdAt,
							Creator:       bob,
							MatchedFields: []MatchedField{},
						},
						Attachment: &SearchHitAttachment{
							AttachedTo: SearchTypeRecord,
							FileKey:    "file1",
							Name:       "estimate.pdf",
							CreatedAt:  createdAt,
							Creator:    bob,
						},
					},
				},
				NextPageToken: "token2",
			}
			if !reflect.DeepEqual(res, want) {
				t.Errorf("response: got %#v, want %#v", res, want)
			}
		})
	}
}

func TestClientSearchGuestSpace(t *testing.T) {
	ts, got := newTestServer(t, http.StatusOK, `{"hits": [], "nextPageToken": null}`)
	c := newTestClient(t, ts, WithGuestSpace(3))
	res, err := c.Search(t.Context(), SearchRequest{
		Query: []SearchQuery{{Operator: SearchOperatorAnd, Keywords: []string{"estimate"}}},
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	assertRequest(t, got, http.MethodPost, "/k/guest/3/v1/search.json",
		`{"query": [{"operator": "AND", "keywords": ["estimate"]}]}`)
	want := &SearchResponse{Hits: []SearchHit{}}
	if !reflect.DeepEqual(res, want) {
		t.Errorf("response: got %#v, want %#v", res, want)
	}
}

func TestClientAllSearchHits(t *testing.T) {
	ts, requests := newSequenceServer(t,
		response{http.StatusOK, `{"hits": [
			{"type": "SPACE", "space": {"spaceId": "1"}},
			{"type": "SPACE", "space": {"spaceId": "2"}}
		], "nextPageToken": "token1"}`},
		response{http.StatusOK, `{"hits": [], "nextPageToken": "token2"}`},
		response{http.StatusOK, `{"hits": [
			{"type": "SPACE", "space": {"spaceId": "3"}}
		], "nextPageToken": null}`},
	)
	c := newTestClient(t, ts)
	var ids []int64
	for hit, err := range c.AllSearchHits(t.Context(), SearchRequest{
		Query: []SearchQuery{{Operator: SearchOperatorAnd, Keywords: []string{"sales"}}},
		Limit: 2,
	}) {
		if err != nil {
			t.Fatalf("AllSearchHits returned error: %v", err)
		}
		ids = append(ids, hit.Space.SpaceID)
	}
	reqs := requests()
	wantBodies := []string{
		`{"query": [{"operator": "AND", "keywords": ["sales"]}], "limit": 2}`,
		`{"query": [{"operator": "AND", "keywords": ["sales"]}], "limit": 2, "pageToken": "token1"}`,
		`{"query": [{"operator": "AND", "keywords": ["sales"]}], "limit": 2, "pageToken": "token2"}`,
	}
	if len(reqs) != len(wantBodies) {
		t.Fatalf("number of requests: got %d, want %d", len(reqs), len(wantBodies))
	}
	for i, body := range wantBodies {
		assertRequest(t, &reqs[i], http.MethodPost, "/k/v1/search.json", body)
	}
	if want := []int64{1, 2, 3}; !reflect.DeepEqual(ids, want) {
		t.Errorf("space IDs: got %v, want %v", ids, want)
	}
}

func TestClientAllSearchHitsBreak(t *testing.T) {
	ts, requests := newSequenceServer(t,
		response{http.StatusOK, `{"hits": [
			{"type": "SPACE", "space": {"spaceId": "1"}},
			{"type": "SPACE", "space": {"spaceId": "2"}}
		], "nextPageToken": "token1"}`},
	)
	c := newTestClient(t, ts)
	for _, err := range c.AllSearchHits(t.Context(), SearchRequest{
		Query: []SearchQuery{{Operator: SearchOperatorAnd, Keywords: []string{"sales"}}},
	}) {
		if err != nil {
			t.Fatalf("AllSearchHits returned error: %v", err)
		}
		break
	}
	if reqs := requests(); len(reqs) != 1 {
		t.Errorf("number of requests: got %d, want %d", len(reqs), 1)
	}
}

func TestClientAllSearchHitsError(t *testing.T) {
	ts, _ := newSequenceServer(t,
		response{http.StatusOK, `{"hits": [{"type": "SPACE", "space": {"spaceId": "1"}}], "nextPageToken": "token1"}`},
		response{http.StatusBadRequest, `{"code": "CB_VA01", "id": "abc123", "message": "Invalid page token."}`},
	)
	c := newTestClient(t, ts)
	var count int
	var err error
	for _, e := range c.AllSearchHits(t.Context(), SearchRequest{
		Query: []SearchQuery{{Operator: SearchOperatorAnd, Keywords: []string{"sales"}}},
	}) {
		if e != nil {
			err = e
			continue
		}
		count++
	}
	if count != 1 {
		t.Errorf("number of hits: got %d, want %d", count, 1)
	}
	if !IsErrorCode(err, "CB_VA01") {
		t.Errorf("AllSearchHits returned unexpected error: %v", err)
	}
}
