package cyberbiz

import (
	"context"
	"fmt"
	"iter"
	"net/http"
)

// SmartCollectionCreateRequest is the body of CreateSmart. Title, Handle and
// Published are required; rules are added afterwards with CreateSmartRule.
type SmartCollectionCreateRequest struct {
	Title     string  `json:"title"`
	Handle    string  `json:"handle"` // storefront path /collections/{handle}
	Published bool    `json:"published"`
	BodyHTML  *string `json:"body_html,omitzero"`
}

// SmartCollectionUpdateRequest is the body of UpdateSmart; every field is
// optional and unset fields are left unchanged.
type SmartCollectionUpdateRequest struct {
	Title     *string `json:"title,omitzero"`
	Handle    *string `json:"handle,omitzero"`
	Published *bool   `json:"published,omitzero"`
	BodyHTML  *string `json:"body_html,omitzero"`
}

// SmartCollectionRuleRequest is the body of CreateSmartRule and
// UpdateSmartRule: products whose Column stands in Relation to Condition
// belong to the collection. All three are required on create; on update an
// unset field is left unchanged.
type SmartCollectionRuleRequest struct {
	Column    SmartRuleColumn   `json:"column,omitzero"`
	Relation  SmartRuleRelation `json:"relation,omitzero"`
	Condition string            `json:"condition,omitzero"` // the reference value, always a string
}

// ListSmart returns one page of smart collections with their rules and
// matching products (GET /v1/smart_collections).
func (s *CollectionsService) ListSmart(ctx context.Context, opts *CollectionListOptions) (*Page[SmartCollection], error) {
	return collectionsList[SmartCollection](ctx, s.client, "v1/smart_collections", opts)
}

// AllSmart walks every page of smart collections (GET /v1/smart_collections).
func (s *CollectionsService) AllSmart(ctx context.Context, opts *CollectionListOptions) iter.Seq2[SmartCollection, error] {
	return collectionsAll[SmartCollection](ctx, s.client, "v1/smart_collections", opts)
}

// ListSmartV2 returns one page of smart collections matched by title or
// handle, with rules and products (GET /v2/smart_collections).
func (s *CollectionsService) ListSmartV2(ctx context.Context, opts *CollectionSearchOptions) (*Page[SmartCollection], error) {
	return collectionsList[SmartCollection](ctx, s.client, "v2/smart_collections", opts)
}

// AllSmartV2 walks every page of the v2 smart collection search
// (GET /v2/smart_collections).
func (s *CollectionsService) AllSmartV2(ctx context.Context, opts *CollectionSearchOptions) iter.Seq2[SmartCollection, error] {
	return collectionsAll[SmartCollection](ctx, s.client, "v2/smart_collections", opts)
}

// GetSmart returns one smart collection (GET /v1/smart_collections/{id}).
func (s *CollectionsService) GetSmart(ctx context.Context, id int64) (*SmartCollection, *Response, error) {
	return collectionsGet(ctx, s.client, fmt.Sprintf("v1/smart_collections/%d", id),
		func(c *SmartCollection) { c.ID = id })
}

// CreateSmart creates a smart collection (POST /v1/smart_collections).
func (s *CollectionsService) CreateSmart(ctx context.Context, req *SmartCollectionCreateRequest) (*SmartCollection, *Response, error) {
	return collectionsWrite[SmartCollection](ctx, s.client, http.MethodPost, "v1/smart_collections", req)
}

// UpdateSmart changes a smart collection (PUT /v1/smart_collections/{id}).
func (s *CollectionsService) UpdateSmart(ctx context.Context, id int64, req *SmartCollectionUpdateRequest) (*SmartCollection, *Response, error) {
	return collectionsWrite[SmartCollection](ctx, s.client, http.MethodPut, fmt.Sprintf("v1/smart_collections/%d", id), req)
}

// DeleteSmart removes a smart collection (DELETE /v1/smart_collections/{id}).
func (s *CollectionsService) DeleteSmart(ctx context.Context, id int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/smart_collections/%d", id), nil)
}

// CreateSmartRule adds a rule to a smart collection and returns the updated
// collection (POST /v1/smart_collections/{id}/rules).
func (s *CollectionsService) CreateSmartRule(ctx context.Context, id int64, req *SmartCollectionRuleRequest) (*SmartCollection, *Response, error) {
	return collectionsWrite[SmartCollection](ctx, s.client, http.MethodPost,
		fmt.Sprintf("v1/smart_collections/%d/rules", id), req)
}

// UpdateSmartRule changes a rule of a smart collection and returns the
// updated collection (PUT /v1/smart_collections/{id}/rules/{rule_id}).
func (s *CollectionsService) UpdateSmartRule(ctx context.Context, id, ruleID int64, req *SmartCollectionRuleRequest) (*SmartCollection, *Response, error) {
	return collectionsWrite[SmartCollection](ctx, s.client, http.MethodPut,
		fmt.Sprintf("v1/smart_collections/%d/rules/%d", id, ruleID), req)
}

// DeleteSmartRule removes a rule from a smart collection and returns the
// updated collection (DELETE /v1/smart_collections/{id}/rules/{rule_id}).
func (s *CollectionsService) DeleteSmartRule(ctx context.Context, id, ruleID int64) (*SmartCollection, *Response, error) {
	var out SmartCollection
	resp, err := s.client.delete(ctx, fmt.Sprintf("v1/smart_collections/%d/rules/%d", id, ruleID), &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}
