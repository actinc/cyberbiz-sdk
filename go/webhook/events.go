package webhook

import (
	"errors"
	"fmt"
	"strings"
)

// EventType is the kind of Shop change an Inbound announces, as sent in the
// X-Cyberbiz-Event header. It is named resource/action, e.g. "orders/paid".
// A value CYBERBIZ adds later still parses; it just will not match any
// constant, and [EventType.Known] reports false for it.
type EventType string

// Every event CYBERBIZ documents for App webhooks.
const (
	// Customers.
	EventCustomersCreate    EventType = "customers/create"
	EventCustomersUpdate    EventType = "customers/update"
	EventUIDProvidersCreate EventType = "uid_providers/create"
	EventUIDProvidersUpdate EventType = "uid_providers/update"

	// Bonus points.
	EventBonusPointsCreate  EventType = "bonus_points/create"
	EventBonusPointsUpdate  EventType = "bonus_points/update"
	EventBonusPointsDestroy EventType = "bonus_points/destroy"
	EventCommentBonusCreate EventType = "comment_bonus/create"

	// Orders.
	EventOrdersCreate                EventType = "orders/create"
	EventOrdersPaid                  EventType = "orders/paid"
	EventOrdersPreparing             EventType = "orders/preparing"
	EventOrdersFulfilled             EventType = "orders/fulfilled"
	EventOrdersReceived              EventType = "orders/received"
	EventOrdersArrived               EventType = "orders/arrived"
	EventOrdersExpired               EventType = "orders/expired"
	EventOrdersCancelled             EventType = "orders/cancelled"
	EventOrdersReturned              EventType = "orders/returned"
	EventOrdersPartialReturn         EventType = "orders/partial_return"
	EventOrdersRefunded              EventType = "orders/refunded"
	EventOrdersPartialRefunded       EventType = "orders/partial_refunded"
	EventOrdersClosed                EventType = "orders/closed"
	EventOrdersRequestReturn         EventType = "orders/request_return"
	EventOrdersOpened                EventType = "orders/opened"
	EventExpressDeliveryOrdersUpdate EventType = "express_delivery_orders/update"

	// Products and variants.
	EventProductsCreate EventType = "products/create"
	EventProductsUpdate EventType = "products/update"
	EventProductsDelete EventType = "products/delete"
	EventVariantsCreate EventType = "variants/create"
	EventVariantsUpdate EventType = "variants/update"
	EventVariantsDelete EventType = "variants/delete"

	// Coupons.
	EventCouponsCreate  EventType = "coupons/create"
	EventCouponsUpdate  EventType = "coupons/update"
	EventCouponsDestroy EventType = "coupons/destroy"

	// VIP levels.
	EventCustomerVIPLevelUpdate EventType = "customer_vip_level/update"

	// Apps.
	EventAppsUninstall EventType = "apps/uninstall"
)

// Resource names, the part of an [EventType] before the slash.
const (
	ResourceCustomers             = "customers"
	ResourceUIDProviders          = "uid_providers"
	ResourceBonusPoints           = "bonus_points"
	ResourceCommentBonus          = "comment_bonus"
	ResourceOrders                = "orders"
	ResourceExpressDeliveryOrders = "express_delivery_orders"
	ResourceProducts              = "products"
	ResourceVariants              = "variants"
	ResourceCoupons               = "coupons"
	ResourceCustomerVIPLevel      = "customer_vip_level"
	ResourceApps                  = "apps"
)

// AllEvents lists every documented event, in documentation order.
var AllEvents = []EventType{
	EventCustomersCreate, EventCustomersUpdate,
	EventUIDProvidersCreate, EventUIDProvidersUpdate,
	EventBonusPointsCreate, EventBonusPointsUpdate, EventBonusPointsDestroy,
	EventCommentBonusCreate,
	EventOrdersCreate, EventOrdersPaid, EventOrdersPreparing,
	EventOrdersFulfilled, EventOrdersReceived, EventOrdersArrived,
	EventOrdersExpired, EventOrdersCancelled, EventOrdersReturned,
	EventOrdersPartialReturn, EventOrdersRefunded, EventOrdersPartialRefunded,
	EventOrdersClosed, EventOrdersRequestReturn, EventOrdersOpened,
	EventExpressDeliveryOrdersUpdate,
	EventProductsCreate, EventProductsUpdate, EventProductsDelete,
	EventVariantsCreate, EventVariantsUpdate, EventVariantsDelete,
	EventCouponsCreate, EventCouponsUpdate, EventCouponsDestroy,
	EventCustomerVIPLevelUpdate,
	EventAppsUninstall,
}

// String returns the wire form, e.g. "orders/paid".
func (t EventType) String() string { return string(t) }

// Resource returns the part before the slash, e.g. "orders".
func (t EventType) Resource() string {
	resource, _, _ := strings.Cut(string(t), "/")
	return resource
}

// Action returns the part after the slash, e.g. "paid". It is empty when
// the type has no slash.
func (t EventType) Action() string {
	_, action, _ := strings.Cut(string(t), "/")
	return action
}

// Known reports whether t is one of the documented events.
func (t EventType) Known() bool {
	for _, e := range AllEvents {
		if e == t {
			return true
		}
	}
	return false
}

// ErrWrongEvent is returned by a typed payload accessor such as
// [Event.Order] when the event's resource does not carry that payload.
var ErrWrongEvent = errors.New("webhook: event does not carry this payload")

// Resource returns the event's resource name, e.g. "orders" for orders/paid.
func (e *Event) Resource() string { return e.Type.Resource() }

// Action returns the event's action name, e.g. "paid" for orders/paid.
func (e *Event) Action() string { return e.Type.Action() }

// Order decodes the payload of an orders/* or express_delivery_orders/*
// event.
func (e *Event) Order() (*OrderPayload, error) {
	return decodeFor[OrderPayload](e, ResourceOrders, ResourceExpressDeliveryOrders)
}

// Customer decodes the payload of a customers/* event.
func (e *Event) Customer() (*CustomerPayload, error) {
	return decodeFor[CustomerPayload](e, ResourceCustomers)
}

// BonusPoint decodes the payload of a bonus_points/* or comment_bonus/*
// event.
func (e *Event) BonusPoint() (*BonusPointPayload, error) {
	return decodeFor[BonusPointPayload](e, ResourceBonusPoints, ResourceCommentBonus)
}

// UIDProvider decodes the payload of a uid_providers/* event.
func (e *Event) UIDProvider() (*UIDProviderPayload, error) {
	return decodeFor[UIDProviderPayload](e, ResourceUIDProviders)
}

// Product decodes the payload of a products/* event.
func (e *Event) Product() (*ProductPayload, error) {
	return decodeFor[ProductPayload](e, ResourceProducts)
}

// Variant decodes the payload of a variants/* event.
func (e *Event) Variant() (*VariantPayload, error) {
	return decodeFor[VariantPayload](e, ResourceVariants)
}

// Coupon decodes the payload of a coupons/* event.
func (e *Event) Coupon() (*CouponPayload, error) {
	return decodeFor[CouponPayload](e, ResourceCoupons)
}

// VIPLevel decodes the payload of a customer_vip_level/* event.
func (e *Event) VIPLevel() (*VIPLevelPayload, error) {
	return decodeFor[VIPLevelPayload](e, ResourceCustomerVIPLevel)
}

// App decodes the payload of an apps/* event.
func (e *Event) App() (*AppPayload, error) {
	return decodeFor[AppPayload](e, ResourceApps)
}

// decodeFor decodes e.Raw into a T after checking that the event's resource
// is one of resources.
func decodeFor[T any](e *Event, resources ...string) (*T, error) {
	resource := e.Resource()
	ok := false
	for _, r := range resources {
		if r == resource {
			ok = true
			break
		}
	}
	if !ok {
		return nil, fmt.Errorf("%w: %s is not a %s event", ErrWrongEvent, e.Type, strings.Join(resources, " or "))
	}
	var out T
	if err := e.Decode(&out); err != nil {
		return nil, fmt.Errorf("webhook: decoding %s payload: %w", e.Type, err)
	}
	return &out, nil
}
