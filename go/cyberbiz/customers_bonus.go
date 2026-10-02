package cyberbiz

import (
	"context"
	"fmt"
)

// CustomerBonusPoint is one grant of bonus points to a customer
// (GET /v1/customers/{id}/bonus_points).
type CustomerBonusPoint struct {
	ID               int64  `json:"id"`
	Title            string `json:"title"`
	Points           Money  `json:"points"`
	UnusedPoints     Money  `json:"unused_points"`
	ConsumptionPrice Money  `json:"consumption_price"` // the order amount that earned the points
	Deadline         Time   `json:"deadline"`          // zero when the points never expire
	CustomerID       int64  `json:"customer_id"`
	Source           string `json:"source"`   // platform text, e.g. "外部紅利"
	OrderID          int64  `json:"order_id"` // zero when not tied to an order
}

// CustomerBonusPointCreateRequest is the body of
// POST /v1/customers/{id}/bonus_points. A zero Deadline grants points that
// never expire.
type CustomerBonusPointCreateRequest struct {
	Title            string `json:"title"`
	Points           Money  `json:"points"`
	ConsumptionPrice *Money `json:"consumption_price,omitzero"`
	Deadline         Time   `json:"-"`
}

// customersBonusPointCreateBody is the wire form: the platform takes the literal "0"
// for "no deadline".
type customersBonusPointCreateBody struct {
	Title            string `json:"title"`
	Points           Money  `json:"points"`
	ConsumptionPrice *Money `json:"consumption_price,omitzero"`
	Deadline         string `json:"deadline"`
}

func (r *CustomerBonusPointCreateRequest) wire() customersBonusPointCreateBody {
	deadline := "0"
	if !r.Deadline.IsZero() {
		deadline = r.Deadline.String()
	}
	return customersBonusPointCreateBody{
		Title:            r.Title,
		Points:           r.Points,
		ConsumptionPrice: r.ConsumptionPrice,
		Deadline:         deadline,
	}
}

// CustomerBonusPointUpdateRequest is the body of
// PUT /v1/customers/{id}/bonus_points/{bonus_point_id}. BonusPointsDiff
// adjusts the balance by a signed amount; Points and UnusedPoints set it.
type CustomerBonusPointUpdateRequest struct {
	Title           *string `json:"title,omitzero"`
	BonusPointsDiff *Money  `json:"bonus_points_diff,omitzero"`
	Deadline        Time    `json:"deadline,omitzero"`
	Points          *Money  `json:"points,omitzero"`
	UnusedPoints    *Money  `json:"unused_points,omitzero"`
}

// ListBonusPoints returns every bonus point grant of the customer; the
// endpoint is not paginated (GET /v1/customers/{id}/bonus_points).
func (s *CustomersService) ListBonusPoints(ctx context.Context, id int64) ([]CustomerBonusPoint, *Response, error) {
	out := []CustomerBonusPoint{}
	resp, err := s.client.get(ctx, fmt.Sprintf("v1/customers/%d/bonus_points", id), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return out, resp, nil
}

// CreateBonusPoints grants bonus points to the customer
// (POST /v1/customers/{id}/bonus_points).
func (s *CustomersService) CreateBonusPoints(ctx context.Context, id int64, req *CustomerBonusPointCreateRequest) (*CustomerBonusPoint, *Response, error) {
	var out CustomerBonusPoint
	resp, err := s.client.post(ctx, fmt.Sprintf("v1/customers/%d/bonus_points", id), req.wire(), &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// UpdateBonusPoints changes one bonus point grant
// (PUT /v1/customers/{id}/bonus_points/{bonus_point_id}).
func (s *CustomersService) UpdateBonusPoints(ctx context.Context, customerID, bonusPointID int64, req *CustomerBonusPointUpdateRequest) (*CustomerBonusPoint, *Response, error) {
	var out CustomerBonusPoint
	path := fmt.Sprintf("v1/customers/%d/bonus_points/%d", customerID, bonusPointID)
	resp, err := s.client.put(ctx, path, req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// DeleteBonusPoints removes one bonus point grant
// (DELETE /v1/customers/{id}/bonus_points/{bonus_point_id}).
func (s *CustomersService) DeleteBonusPoints(ctx context.Context, customerID, bonusPointID int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/customers/%d/bonus_points/%d", customerID, bonusPointID), nil)
}
