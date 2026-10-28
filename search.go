package kintone

import (
	"context"
	"iter"
	"time"
)

// SearchRequest is the request of [Client.Search]. Query is required, and the
// other fields filter the results. The conditions in Query are all combined,
// and a query with only NOT conditions returns no hits.
type SearchRequest struct {
	Query           []SearchQuery `json:"query"`
	Types           []SearchType  `json:"types,omitzero"`
	Scopes          []SearchScope `json:"scopes,omitzero"`
	ExcludeScopes   []SearchScope `json:"excludeScopes,omitzero"`
	CreatedAfter    time.Time     `json:"createdAfter,omitzero"`
	CreatedBefore   time.Time     `json:"createdBefore,omitzero"`
	Creators        []string      `json:"creators,omitzero"`
	IncludeSynonyms bool          `json:"includeSynonyms,omitzero"`
	Sort            SearchSort    `json:"sort,omitzero"`
	Limit           int64         `json:"limit,omitzero"`
	PageToken       string        `json:"pageToken,omitzero"`
}

// SearchQuery is a condition of a search.
type SearchQuery struct {
	Operator SearchOperator `json:"operator"`
	Keywords []string       `json:"keywords"`
}

// SearchOperator is the operator of a search condition.
type SearchOperator string

// Search operators.
const (
	SearchOperatorAnd SearchOperator = "AND"
	SearchOperatorOr  SearchOperator = "OR"
	SearchOperatorNot SearchOperator = "NOT"
)

// SearchType is the type of a search hit.
type SearchType string

// Search types.
const (
	SearchTypeRecord         SearchType = "RECORD"
	SearchTypeRecordComment  SearchType = "RECORD_COMMENT"
	SearchTypeSpace          SearchType = "SPACE"
	SearchTypeThread         SearchType = "THREAD"
	SearchTypeThreadComment  SearchType = "THREAD_COMMENT"
	SearchTypePeopleComment  SearchType = "PEOPLE_COMMENT"
	SearchTypeMessageComment SearchType = "MESSAGE_COMMENT"
	SearchTypeAttachment     SearchType = "ATTACHMENT"
)

// SearchScope is a scope of a search. IDs are app or space IDs for the app or
// space scope, and Codes are login names for the people or message scope.
type SearchScope struct {
	Scope SearchScopeType `json:"scope"`
	IDs   []int64         `json:"ids,omitzero"`
	Codes []string        `json:"codes,omitzero"`
}

// SearchScopeType is the type of a search scope.
type SearchScopeType string

// Search scope types.
const (
	SearchScopeTypeApp     SearchScopeType = "APP"
	SearchScopeTypeSpace   SearchScopeType = "SPACE"
	SearchScopeTypePeople  SearchScopeType = "PEOPLE"
	SearchScopeTypeMessage SearchScopeType = "MESSAGE"
)

// SearchSort is the sort of search hits.
// The ascending order is not allowed when sorting by relevance.
type SearchSort struct {
	By    SearchSortBy `json:"by,omitzero"`
	Order SortOrder    `json:"order,omitzero"`
}

// SearchSortBy is the target of a sort of search hits.
type SearchSortBy string

// Search sort targets.
const (
	SearchSortByRelevance SearchSortBy = "RELEVANCE"
	SearchSortByCreatedAt SearchSortBy = "CREATED_AT"
)

// SearchResponse is the response of [Client.Search].
// NextPageToken is empty on the last page.
type SearchResponse struct {
	Hits          []SearchHit `json:"hits"`
	NextPageToken string      `json:"nextPageToken"`
}

// SearchHit is a hit of a search. The field of the type is set, and for an
// attachment, the field of the content the file is attached to is also set.
// The creation time and the creator of a record are empty when the user has
// no permission to view them.
type SearchHit struct {
	Type           SearchType               `json:"type"`
	URL            string                   `json:"url"`
	Snippets       []string                 `json:"snippets"`
	Record         *SearchHitRecord         `json:"record"`
	RecordComment  *SearchHitRecordComment  `json:"recordComment"`
	Space          *SearchHitSpace          `json:"space"`
	Thread         *SearchHitThread         `json:"thread"`
	ThreadComment  *SearchHitThreadComment  `json:"threadComment"`
	PeopleComment  *SearchHitPeopleComment  `json:"peopleComment"`
	MessageComment *SearchHitMessageComment `json:"messageComment"`
	Attachment     *SearchHitAttachment     `json:"attachment"`
}

// SearchHitRecord is a record of a search hit.
// SpaceID and SpaceName are empty for an app not in a space.
type SearchHitRecord struct {
	AppID         int64          `json:"appId,string"`
	AppName       string         `json:"appName"`
	RecordID      int64          `json:"recordId,string"`
	RecordTitle   string         `json:"recordTitle"`
	CreatedAt     time.Time      `json:"createdAt"`
	Creator       Entity         `json:"creator"`
	MatchedFields []MatchedField `json:"matchedFields"`
	SpaceID       int64          `json:"spaceId,string"`
	SpaceName     string         `json:"spaceName"`
}

// MatchedField is a field matched by a search.
type MatchedField struct {
	Code  string `json:"code"`
	Label string `json:"label"`
}

// SearchHitRecordComment is a record comment of a search hit.
// SpaceID and SpaceName are empty for an app not in a space.
type SearchHitRecordComment struct {
	AppID       int64     `json:"appId,string"`
	AppName     string    `json:"appName"`
	RecordID    int64     `json:"recordId,string"`
	RecordTitle string    `json:"recordTitle"`
	CommentID   int64     `json:"commentId,string"`
	CreatedAt   time.Time `json:"createdAt"`
	Creator     Entity    `json:"creator"`
	SpaceID     int64     `json:"spaceId,string"`
	SpaceName   string    `json:"spaceName"`
}

// SearchHitSpace is a space of a search hit.
type SearchHitSpace struct {
	SpaceID   int64     `json:"spaceId,string"`
	SpaceName string    `json:"spaceName"`
	CreatedAt time.Time `json:"createdAt"`
	Creator   Entity    `json:"creator"`
}

// SearchHitThread is a thread of a search hit.
type SearchHitThread struct {
	ThreadID   int64     `json:"threadId,string"`
	ThreadName string    `json:"threadName"`
	SpaceID    int64     `json:"spaceId,string"`
	SpaceName  string    `json:"spaceName"`
	CreatedAt  time.Time `json:"createdAt"`
	Creator    Entity    `json:"creator"`
}

// SearchHitThreadComment is a thread comment of a search hit.
// ReplyID is set for a reply to the comment.
type SearchHitThreadComment struct {
	CommentID  int64     `json:"commentId,string"`
	ReplyID    int64     `json:"replyId,string"`
	SpaceID    int64     `json:"spaceId,string"`
	SpaceName  string    `json:"spaceName"`
	ThreadID   int64     `json:"threadId,string"`
	ThreadName string    `json:"threadName"`
	CreatedAt  time.Time `json:"createdAt"`
	Creator    Entity    `json:"creator"`
}

// SearchHitPeopleComment is a comment on the People of a search hit.
// ReplyID is set for a reply to the comment.
type SearchHitPeopleComment struct {
	CommentID int64     `json:"commentId,string"`
	ReplyID   int64     `json:"replyId,string"`
	Owner     Entity    `json:"owner"`
	CreatedAt time.Time `json:"createdAt"`
	Creator   Entity    `json:"creator"`
}

// SearchHitMessageComment is a comment on a message of a search hit.
type SearchHitMessageComment struct {
	CommentID int64     `json:"commentId,string"`
	CreatedAt time.Time `json:"createdAt"`
	Creator   Entity    `json:"creator"`
	Recipient Entity    `json:"recipient"`
}

// SearchHitAttachment is an attachment of a search hit.
// AttachedTo is the type of the content the file is attached to.
type SearchHitAttachment struct {
	AttachedTo SearchType `json:"attachedTo"`
	FileKey    string     `json:"fileKey"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"createdAt"`
	Creator    Entity     `json:"creator"`
}

// Search searches the records, comments, spaces, threads, and attachments
// across kintone. The result contains only the data the user can view. This
// API is in the API Lab, and needs to be enabled in the update options.
func (c *Client) Search(ctx context.Context, req SearchRequest) (*SearchResponse, error) {
	result, err := c.post[SearchResponse](ctx, "search.json", req)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// AllSearchHits iterates over the search hits, using [Client.Search] with the
// page tokens. Limit of the request is the maximum number of hits in each
// request.
func (c *Client) AllSearchHits(ctx context.Context, req SearchRequest) iter.Seq2[SearchHit, error] {
	return func(yield func(SearchHit, error) bool) {
		for {
			res, err := c.Search(ctx, req)
			if err != nil {
				yield(SearchHit{}, err)
				return
			}
			for _, hit := range res.Hits {
				if !yield(hit, nil) {
					return
				}
			}
			if res.NextPageToken == "" {
				return
			}
			req.PageToken = res.NextPageToken
		}
	}
}
