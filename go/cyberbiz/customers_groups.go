package cyberbiz

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/actinc/cyberbiz-sdk/go/internal/query"
)

// CustomerGroupType distinguishes the platform's ORFM segments from
// merchant-defined ones.
type CustomerGroupType string

// Known CustomerGroupType values (group_type of a customer group).
const (
	CustomerGroupTypeDefault CustomerGroupType = "default" // ORFM preset segment
	CustomerGroupTypeCustom  CustomerGroupType = "custom"  // merchant-defined segment
)

// CustomerGroup is a customer segment (GET /v1/customers/customer_groups).
type CustomerGroup struct {
	ID        int64             `json:"id"`
	Name      string            `json:"name"`
	Term      string            `json:"term"`   // free-text search term of the segment
	Query     string            `json:"query"`  // the segment's filter rules, platform-encoded
	Amount    int               `json:"amount"` // customers in the segment as of the last recount
	GroupType CustomerGroupType `json:"group_type"`
	CreatedAt Time              `json:"created_at"`
	UpdatedAt Time              `json:"updated_at"`
}

// CustomerGroupListOptions filter GET /v1/customers/customer_groups.
type CustomerGroupListOptions struct {
	ListOptions
	IDs       []int64           `url:"customer_group_ids,omitempty,comma"`
	Name      string            `url:"name,omitempty"`
	GroupType CustomerGroupType `url:"group_type,omitempty"`
}

// CustomerGroupJob identifies a background job started on customer groups.
type CustomerGroupJob struct {
	JobID string `json:"job_id"`
}

// CustomerGroupMember is one customer of a segment as reported by the
// filter job.
type CustomerGroupMember struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	UID   string `json:"uid"`
}

// CustomerGroupStatus is the reply of the filter job status check. While the
// job is still running the platform returns a status object (Success,
// JobID, Message); once finished it returns the members instead, in which
// case Done is true and Members holds the page.
type CustomerGroupStatus struct {
	Success string                `json:"success"` // platform job state, e.g. "QUEUED"
	JobID   string                `json:"job_id"`
	Message string                `json:"message"`
	Done    bool                  `json:"-"`
	Members []CustomerGroupMember `json:"-"`
}

// ListGroups returns one page of customer groups
// (GET /v1/customers/customer_groups).
func (s *CustomersService) ListGroups(ctx context.Context, opts *CustomerGroupListOptions) (*Page[CustomerGroup], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[CustomerGroup](ctx, s.client, "v1/customers/customer_groups", q)
}

// ScheduleGroupFilter starts the job that computes a group's members; poll
// it with CheckGroupStatus
// (POST /v1/customers/customer_groups/schedule_customer_filter_worker).
func (s *CustomersService) ScheduleGroupFilter(ctx context.Context, groupID int64) (*CustomerGroupJob, *Response, error) {
	body := struct {
		CustomerGroupID int64 `json:"customer_group_id"`
	}{CustomerGroupID: groupID}
	var out CustomerGroupJob
	resp, err := s.client.post(ctx, "v1/customers/customer_groups/schedule_customer_filter_worker", body, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// UpdateGroupAmounts starts the job that recounts every group's members.
// While a recount is already running the platform answers 200 with
// {"success":false,"errors":[...]} instead of a job id; that is returned as
// an *APIError carrying the messages
// (PUT /v1/customers/customer_groups/update_customer_group_amounts).
func (s *CustomersService) UpdateGroupAmounts(ctx context.Context) (*CustomerGroupJob, *Response, error) {
	const path = "v1/customers/customer_groups/update_customer_group_amounts"
	var out struct {
		JobID   string   `json:"job_id"`
		Success *bool    `json:"success"`
		Errors  []string `json:"errors"`
	}
	resp, err := s.client.put(ctx, path, nil, &out)
	if err != nil {
		return nil, resp, err
	}
	if out.JobID == "" && out.Success != nil && !*out.Success {
		return nil, resp, &APIError{StatusCode: resp.StatusCode, Method: http.MethodPut, Path: path,
			RequestID: resp.RequestID, Messages: out.Errors, Body: resp.Body}
	}
	return &CustomerGroupJob{JobID: out.JobID}, resp, nil
}

// CheckGroupStatus polls a filter job. Until the job finishes the result
// carries the job state; afterwards Done is set and Members holds one page
// of the group's customers, paged by opts
// (GET /v1/customers/customer_groups/check_status/{job_id}).
func (s *CustomersService) CheckGroupStatus(ctx context.Context, jobID string, opts *ListOptions) (*CustomerGroupStatus, *Response, error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, nil, err
	}
	path := "v1/customers/customer_groups/check_status/" + url.PathEscape(jobID)
	resp, err := s.client.get(ctx, path, q, nil)
	if err != nil {
		return nil, resp, err
	}
	out, err := customersDecodeGroupStatus(s.client, resp.Body)
	if err != nil {
		return nil, resp, fmt.Errorf("cyberbiz: GET %s: decoding response: %w", path, err)
	}
	return out, resp, nil
}

// customersDecodeGroupStatus handles the two shapes the status endpoint returns.
func customersDecodeGroupStatus(c *Client, body []byte) (*CustomerGroupStatus, error) {
	var out CustomerGroupStatus
	if bytes.HasPrefix(bytes.TrimSpace(body), []byte("[")) {
		out.Done = true
		out.Members = []CustomerGroupMember{}
		if err := c.decode(body, &out.Members); err != nil {
			return nil, err
		}
		return &out, nil
	}
	if err := c.decode(body, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
