package cyberbiz

import "context"

// BonusRuleService exposes the shop's bonus point rule (/v1/bonus_rule).
type BonusRuleService struct {
	client *Client
}

// Get returns the bonus point rule (GET /v1/bonus_rule).
func (s *BonusRuleService) Get(ctx context.Context) (*BonusRule, *Response, error) {
	var out BonusRule
	resp, err := s.client.get(ctx, "v1/bonus_rule", nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}
