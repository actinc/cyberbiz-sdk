package cyberbiz

import (
	"context"
	"fmt"
)

// VIPGroupsService exposes VIP groups, their levels, and the draft workflow
// (/v1/vip_groups). Changes to groups and levels land in a draft set that
// ApplyDraft publishes on a chosen date.
type VIPGroupsService struct {
	client *Client
}

// VIPGroupCreateRequest is the body of Create.
type VIPGroupCreateRequest struct {
	Name string `json:"name"`
	// CustomerTags is the comma-separated list of customer tags that place a
	// customer in the group.
	CustomerTags   string  `json:"customer_tags"`
	DescriptionURL *string `json:"description_url,omitzero"`
	Position       *int    `json:"position,omitzero"`
}

// VIPGroupLevelRequest is the body of CreateLevel and UpdateLevel. Every
// field is optional on update; the platform requires Name and ValidityDays
// on create. Conditions left nil are not set.
type VIPGroupLevelRequest struct {
	Name *string `json:"name,omitzero"`
	// ValidityDays and UpgradeValidityDays accept 30, 180, 360, 365, 720,
	// 730 or 36500.
	ValidityDays                             *int                `json:"validity_days,omitzero"`
	UpgradeValidityDays                      *int                `json:"upgrade_validity_days,omitzero"`
	UpgradeConditionTotalSpent               *Money              `json:"upgrade_condition_total_spent,omitzero"`
	UpgradeConditionTotalSpentInValidityDays *Money              `json:"upgrade_condition_total_spent_in_validity_days,omitzero"`
	RenewalConditionTotalSpent               *Money              `json:"renewal_condition_total_spent,omitzero"`
	RenewalConditionTotalSpentInValidityDays *Money              `json:"renewal_condition_total_spent_in_validity_days,omitzero"`
	OrderDiscountEnabled                     *bool               `json:"order_discount_enabled,omitzero"`
	OrderDiscountValue                       *float64            `json:"order_discount_value,omitzero"`
	OrderDiscountSetting                     *VIPRestrictSetting `json:"order_discount_setting,omitzero"`
	FreeShippingEnabled                      *bool               `json:"free_shipping_enabled,omitzero"`
	FreeShippingThreshold                    *Money              `json:"free_shipping_threshold,omitzero"`
	FreeShippingSetting                      *VIPRestrictSetting `json:"free_shipping_setting,omitzero"`
	BonusPointEnabled                        *bool               `json:"bonus_point_enabled,omitzero"`
	BonusPointThreshold                      *Money              `json:"bonus_point_threshold,omitzero"`
	BonusPointValue                          *Money              `json:"bonus_point_value,omitzero"`
	BonusPointExpiryDays                     *int                `json:"bonus_point_expiry_days,omitzero"`
	UpgradeGiftEnabled                       *bool               `json:"upgrade_gift_enabled,omitzero"`
	UpgradeGiftSetting                       *VIPGiftSetting     `json:"upgrade_gift_setting,omitzero"`
	BirthGiftEnabled                         *bool               `json:"birth_gift_enabled,omitzero"`
	BirthGiftName                            *string             `json:"birth_gift_name,omitzero"`
	BirthGiftBeforeDays                      *int                `json:"birth_gift_before_days,omitzero"`
	BirthGiftSetting                         *VIPGiftSetting     `json:"birth_gift_setting,omitzero"`
}

// List returns the current and draft VIP groups (GET /v1/vip_groups).
func (s *VIPGroupsService) List(ctx context.Context) (*VIPGroupsOverview, *Response, error) {
	var out VIPGroupsOverview
	resp, err := s.client.get(ctx, "v1/vip_groups", nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// Get returns one VIP group (GET /v1/vip_groups/{id}).
func (s *VIPGroupsService) Get(ctx context.Context, id int64) (*VIPGroup, *Response, error) {
	var out VIPGroup
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/vip_groups/%d", id), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// Create adds a VIP group to the draft set (POST /v1/vip_groups).
func (s *VIPGroupsService) Create(ctx context.Context, req *VIPGroupCreateRequest) (*VIPGroup, *Response, error) {
	var out VIPGroup
	resp, err := s.client.post(ctx, "v1/vip_groups", req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// Delete removes a VIP group (DELETE /v1/vip_groups/{id}).
func (s *VIPGroupsService) Delete(ctx context.Context, id int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/vip_groups/%d", id), nil)
}

// ApplyDraft schedules the draft groups to replace the current ones on
// startDate (POST /v1/vip_groups/apply_draft).
func (s *VIPGroupsService) ApplyDraft(ctx context.Context, startDate Date) (*Response, error) {
	body := struct {
		StartDate Date `json:"start_date"`
	}{StartDate: startDate}
	return s.client.post(ctx, "v1/vip_groups/apply_draft", body, nil)
}

// CancelDraft discards the draft groups (POST /v1/vip_groups/cancel_draft).
func (s *VIPGroupsService) CancelDraft(ctx context.Context) (*Response, error) {
	return s.client.post(ctx, "v1/vip_groups/cancel_draft", nil, nil)
}

// CreateLevel adds a level to a group (POST /v1/vip_groups/{id}/vip_group_levels).
func (s *VIPGroupsService) CreateLevel(ctx context.Context, groupID int64, req *VIPGroupLevelRequest) (*VIPGroupLevel, *Response, error) {
	var out VIPGroupLevel
	resp, err := s.client.post(ctx, fmt.Sprintf("v1/vip_groups/%d/vip_group_levels", groupID), req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// UpdateLevel changes a level (PUT /v1/vip_groups/{id}/vip_group_levels/{level_id}).
func (s *VIPGroupsService) UpdateLevel(ctx context.Context, groupID, levelID int64, req *VIPGroupLevelRequest) (*VIPGroupLevel, *Response, error) {
	var out VIPGroupLevel
	resp, err := s.client.put(ctx, fmt.Sprintf("v1/vip_groups/%d/vip_group_levels/%d", groupID, levelID), req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// DeleteLevel removes a level (DELETE /v1/vip_groups/{id}/vip_group_levels/{level_id}).
func (s *VIPGroupsService) DeleteLevel(ctx context.Context, groupID, levelID int64) (*Response, error) {
	return s.client.delete(ctx, fmt.Sprintf("v1/vip_groups/%d/vip_group_levels/%d", groupID, levelID), nil)
}
