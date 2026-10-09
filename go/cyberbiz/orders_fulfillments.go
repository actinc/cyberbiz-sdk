package cyberbiz

import (
	"context"
	"fmt"
	"net/http"

	"github.com/actinc/cyberbiz-sdk/go/internal/query"
)

// Fulfillment is one shipment of an order. CVSShippingType is only set by
// the v2 CVS shipping endpoint.
type Fulfillment struct {
	ID              int64             `json:"id"`
	TrackingCompany TrackingCompany   `json:"tracking_company"`
	TrackingNumber  string            `json:"tracking_number"`
	FulfilledAt     Time              `json:"fulfilled_at"`
	ReceivedAt      Time              `json:"received_at"`
	Status          FulfillmentStatus `json:"status"`
	LineItems       []LineItem        `json:"line_items"`
	TrackingURL     string            `json:"tracking_url"`      // Uber Direct and Pandago only
	CVSShippingType CVSShippingType   `json:"cvs_shipping_type"` // label type for PrintCVSShippingLabels
}

// TrackingCompany is a carrier code accepted by the fulfillment endpoints.
type TrackingCompany string

// Known TrackingCompany codes (tracking_company of a fulfillment).
const (
	TrackingCompanyEzcat                 TrackingCompany = "ezcat"                    // Black Cat (黑貓宅急便)
	TrackingCompanyECan                  TrackingCompany = "e_can"                    // Taiwan Pelican Express (台灣宅配通)
	TrackingCompanyPostserv              TrackingCompany = "postserv"                 // Chunghwa Post (中華郵政)
	TrackingCompanyHCT                   TrackingCompany = "hct"                      // HCT Logistics (新竹物流)
	TrackingCompanyDHL                   TrackingCompany = "DHL"                      // DHL
	TrackingCompanyFedex                 TrackingCompany = "fedex"                    // FedEx
	TrackingCompanyUPS                   TrackingCompany = "UPS"                      // UPS
	TrackingCompanyTNT                   TrackingCompany = "TNT"                      // TNT
	TrackingCompanyTjoin                 TrackingCompany = "tjoin"                    // Kerry TJ Logistics (嘉里大榮)
	TrackingCompanyCFEExpress            TrackingCompany = "cfe_express"              // CFE Express (超峰快遞)
	TrackingCompanyDPEX                  TrackingCompany = "dpex"                     // DPEX (迪比翼)
	TrackingCompanySDExpress             TrackingCompany = "sd_express"               // SD Express (鑫達快遞)
	TrackingCompanyMapleExpress          TrackingCompany = "maple_express"            // Maple Express (便利帶)
	TrackingCompanyEZShip                TrackingCompany = "EZSHIP"                   // EZShip convenience-store pickup
	TrackingCompanySeven                 TrackingCompany = "SEVEN"                    // 7-ELEVEN store pickup
	TrackingCompanyAllpaySeven           TrackingCompany = "Allpay_SEVEN"             // ECPay 7-ELEVEN
	TrackingCompanyEMS                   TrackingCompany = "ems"                      // Chunghwa Post EMS
	TrackingCompanyAllpayFamily          TrackingCompany = "Allpay_FAMILY"            // ECPay FamilyMart
	TrackingCompanyFamily                TrackingCompany = "FAMILY"                   // FamilyMart store pickup
	TrackingCompanyFamilyC2C             TrackingCompany = "FAMILY_C2C"               // FamilyMart store-to-store
	TrackingCompanyHilife                TrackingCompany = "HILIFE"                   // Hi-Life store pickup
	TrackingCompanySevenC2C              TrackingCompany = "SEVEN_C2C"                // 7-ELEVEN store-to-store (交貨便)
	TrackingCompanySF                    TrackingCompany = "SF"                       // SF Express (順豐速運)
	TrackingCompanyFamilyCold            TrackingCompany = "FAMILY_COLD"              // FamilyMart cold chain
	TrackingCompanyShyangYih             TrackingCompany = "SHYANG_YIH"               // Shyang Yih Freight (祥億貨運)
	TrackingCompanyEzcatCVS              TrackingCompany = "ezcat_cvs"                // Black Cat store pickup, room temperature
	TrackingCompanyEzcatCVSRefrigerate   TrackingCompany = "ezcat_cvs_refrigerate"    // Black Cat store pickup, refrigerated
	TrackingCompanyEzcatCVSCold          TrackingCompany = "ezcat_cvs_cold"           // Black Cat store pickup, frozen
	TrackingCompanyECPayB2CSevenCOD      TrackingCompany = "ecpay_b2c_seven_cod"      // ECPay B2C 7-ELEVEN, cash on delivery
	TrackingCompanyECPayB2CSevenPrepaid  TrackingCompany = "ecpay_b2c_seven_prepaid"  // ECPay B2C 7-ELEVEN, prepaid
	TrackingCompanyECPayB2CFamilyCOD     TrackingCompany = "ecpay_b2c_family_cod"     // ECPay B2C FamilyMart, cash on delivery
	TrackingCompanyECPayB2CFamilyPrepaid TrackingCompany = "ecpay_b2c_family_prepaid" // ECPay B2C FamilyMart, prepaid
	TrackingCompanyAmazonFBA             TrackingCompany = "amazon_fba"               // Fulfillment by Amazon
	TrackingCompanyHilifeC2C             TrackingCompany = "HILIFE_C2C"               // Hi-Life store-to-store
	TrackingCompanyHilifeCold            TrackingCompany = "HILIFE_COLD"              // Hi-Life frozen
	TrackingCompanyJapanPost             TrackingCompany = "japan_post"               // Japan Post
	TrackingCompanyLinexHubEZ            TrackingCompany = "linex_hub_ez"             // Linex Logistics (宇迅)
	TrackingCompanyPandago               TrackingCompany = "pandago"                  // pandago
	TrackingCompanyUberDirect            TrackingCompany = "uber_direct"              // Uber Direct
	TrackingCompanyFamilyColdC2C         TrackingCompany = "FAMILY_COLD_C2C"          // FamilyMart frozen store-to-store
	TrackingCompanySevenC2B              TrackingCompany = "SEVEN_C2B"                // 7-ELEVEN return service (退貨便)
	TrackingCompanyNinjaVan              TrackingCompany = "ninja_van"                // Ninja Van
	TrackingCompanyGDex                  TrackingCompany = "g_dex"                    // GDex
	TrackingCompanyLineClearExpress      TrackingCompany = "line_clear_express"       // Line Clear Express
	TrackingCompanyJTExpress             TrackingCompany = "jt_express"               // J&T Express
	TrackingCompanyPosLaju               TrackingCompany = "pos_laju"                 // Pos Laju
	TrackingCompanyLalamove              TrackingCompany = "lalamove"                 // Lalamove
	TrackingCompanyABXExpress            TrackingCompany = "abx_express"              // ABX Express (KEX Express)
	TrackingCompanySkynet                TrackingCompany = "skynet"                   // Skynet Express
	TrackingCompanyCityLink              TrackingCompany = "city_link"                // Citylink
	TrackingCompanyDirectFreightExpress  TrackingCompany = "direct_freight_express"   // Direct Freight Express
	TrackingCompanyAustraliaPost         TrackingCompany = "australia_post"           // Australia Post
	TrackingCompanyCyberbizExpress       TrackingCompany = "cyberbiz_express"         // CYBERBIZ Express
	TrackingCompanyOther                 TrackingCompany = "other"                    // any other carrier
)

// SupportShippingSource is a home-delivery carrier CYBERBIZ can book
// automatically.
type SupportShippingSource string

// Known SupportShippingSource values (source of the support_shipping endpoint).
const (
	SupportShippingEzcat   SupportShippingSource = "ezcat"   // 黑貓
	SupportShippingPelican SupportShippingSource = "pelican" // 宅配通
	SupportShippingSF      SupportShippingSource = "sf"      // 順豐
	SupportShippingHCT     SupportShippingSource = "hct"     // 新竹物流
)

// ShippingTemperature is the temperature band of a shipment.
type ShippingTemperature string

// Known ShippingTemperature values (temperature of a support_shipping booking).
const (
	ShippingTemperatureNormal ShippingTemperature = "normal" // room temperature
	ShippingTemperatureCold   ShippingTemperature = "cold"   // low temperature (refrigerated or frozen)
)

// FridgeOrFrozen is the label marking for a cold shipment.
type FridgeOrFrozen string

// Known FridgeOrFrozen values (fridge_or_frozen of a support_shipping booking);
// not supported by the sf and hct sources.
const (
	FridgeOrFrozenNone   FridgeOrFrozen = "none"   // no cold-chain marking
	FridgeOrFrozenFridge FridgeOrFrozen = "fridge" // mark as refrigerated
	FridgeOrFrozenFrozen FridgeOrFrozen = "frozen" // mark as frozen
)

// CVSMeasurement is the parcel volume class for Hi-Life and FamilyMart
// cold-chain shipments.
type CVSMeasurement string

// Known CVSMeasurement values (measurement of a cvs_shipping booking).
const (
	CVSMeasurementS60  CVSMeasurement = "S60"  // small parcel (up to 60 cm combined)
	CVSMeasurementS105 CVSMeasurement = "S105" // large parcel (up to 105 cm combined)
)

// CVSShippingType is the convenience-store label type used by the v2 CVS
// endpoints.
type CVSShippingType string

// Known CVSShippingType values (cvs_shipping_type of a fulfillment).
const (
	CVSShippingSeven               CVSShippingType = "seven"                 // 7-ELEVEN store pickup
	CVSShippingSevenC2C            CVSShippingType = "seven_c2c"             // 7-ELEVEN store-to-store
	CVSShippingFamily              CVSShippingType = "family"                // FamilyMart store pickup
	CVSShippingFamilyC2C           CVSShippingType = "family_c2c"            // FamilyMart store-to-store
	CVSShippingFamilyCold          CVSShippingType = "family_cold"           // FamilyMart cold chain
	CVSShippingFamilyColdC2C       CVSShippingType = "family_cold_c2c"       // FamilyMart frozen store-to-store
	CVSShippingHilife              CVSShippingType = "hilife"                // Hi-Life store pickup
	CVSShippingHilifeCold          CVSShippingType = "hilife_cold"           // Hi-Life frozen
	CVSShippingHilifeC2C           CVSShippingType = "hilife_c2c"            // Hi-Life store-to-store
	CVSShippingEzcatCVS            CVSShippingType = "ezcat_cvs"             // Black Cat store pickup, room temperature
	CVSShippingEzcatCVSCold        CVSShippingType = "ezcat_cvs_cold"        // Black Cat store pickup, frozen
	CVSShippingEzcatCVSRefrigerate CVSShippingType = "ezcat_cvs_refrigerate" // Black Cat store pickup, refrigerated
)

// SupportShippingPrintType selects the HCT label format.
type SupportShippingPrintType string

// Known SupportShippingPrintType values (print_type for HCT labels).
const (
	SupportShippingPrintNormal  SupportShippingPrintType = "normal"  // plain paper label
	SupportShippingPrintThermal SupportShippingPrintType = "thermal" // thermal printer label
)

// CustomShippingRequest books a shipment with a carrier the merchant
// arranged themselves.
type CustomShippingRequest struct {
	LineItemIDs     IDList          `json:"line_item_ids"`
	TrackingNumber  string          `json:"tracking_number"`
	TrackingCompany TrackingCompany `json:"tracking_company"`
	NotifyCustomer  bool            `json:"notify_customer"`
}

// SupportShippingRequest books a home-delivery shipment through CYBERBIZ.
type SupportShippingRequest struct {
	Email          string                `json:"email"` // receives the shipping label
	LineItemIDs    IDList                `json:"line_item_ids"`
	Source         SupportShippingSource `json:"source"`
	Size           int                   `json:"size"` // 60, 90, 120 or 150 cm
	Temperature    ShippingTemperature   `json:"temperature"`
	FridgeOrFrozen FridgeOrFrozen        `json:"fridge_or_frozen"`
	IsFragile      bool                  `json:"is_fragile"`
	UseTransfer    *bool                 `json:"use_transfer,omitzero"` // deprecated by CYBERBIZ
}

// CVSShippingRequest books a convenience-store shipment. Measurement is
// required for Hi-Life and FamilyMart cold chain; Size (60, 90 or 105) for
// ezcat CVS delivery.
type CVSShippingRequest struct {
	Measurement CVSMeasurement `json:"measurement,omitzero"`
	Size        int            `json:"size,omitzero"`
}

// PartialCVSShippingRequest ships part of a Hi-Life order (custom feature).
type PartialCVSShippingRequest struct {
	Items  []PartialCVSShippingItem `json:"items"`
	Charge *Money                   `json:"charge,omitzero"` // amount to collect on a COD order
}

// PartialCVSShippingItem is one line item quantity to ship.
type PartialCVSShippingItem struct {
	LineItemID int64 `json:"line_item_id"`
	Quantity   int   `json:"quantity"`
}

// FulfillmentInfo is the label data of one CVS fulfillment
// (POST /v1/orders/fulfillment_infos).
type FulfillmentInfo struct {
	ID             int64  `json:"id"`
	IDVerification bool   `json:"id_verification"` // pickup requires an id check
	Receiver       string `json:"receiver"`
	ReceiverPhone  string `json:"receiver_phone"`
	Amount         Money  `json:"amount"`
	TrackingNumber string `json:"tracking_number"`
	StoreName      string `json:"store_name"`
	StoreNo        string `json:"store_no"`
	Barcode        string `json:"barcode"`
	ShopName       string `json:"shop_name"`
	OrderNo        string `json:"order_no"`
	ShopPhone      string `json:"shop_phone"`
	ShopURL        string `json:"shop_url"`
	SupplierName   string `json:"supplier_name"` // Hi-Life only
	ImageURL       string `json:"image_url"`     // FamilyMart only
	PDF            string `json:"pdf"`           // base64 PDF, Hi-Life only
	HTML           string `json:"html"`          // 7-11 C2C only
	PDFJSON        string `json:"pdf_json"`      // Hi-Life cold chain only
}

// SupportShippingBatchRequest books home-delivery shipments for several
// orders at once (POST /v1/orders/fulfillments/support_shipping).
type SupportShippingBatchRequest struct {
	SupportShippings []SupportShippingBatchItem `json:"support_shippings"`
	Email            string                     `json:"email"`
}

// SupportShippingBatchItem is one order in a SupportShippingBatchRequest.
type SupportShippingBatchItem struct {
	OrderID        int64                 `json:"order_id"`
	LineItemIDs    IDList                `json:"line_item_ids"`
	Source         SupportShippingSource `json:"source"`
	Size           int                   `json:"size"`
	Temperature    ShippingTemperature   `json:"temperature"`
	FridgeOrFrozen FridgeOrFrozen        `json:"fridge_or_frozen"`
	IsFragile      bool                  `json:"is_fragile"`
	UseTransfer    *bool                 `json:"use_transfer,omitzero"`
}

// SupportShippingBatchResult is the outcome of a SupportShippingBatchRequest.
type SupportShippingBatchResult struct {
	RequestID string                `json:"request_id"`
	Results   []SupportShippingTask `json:"results"`
}

// SupportShippingTask is the outcome for one order of a batch booking.
type SupportShippingTask struct {
	OrderID           int64             `json:"order_id"`
	FulfillmentStatus FulfillmentStatus `json:"fulfillment_status"`
	// TrackingNumbers lists the order's fulfillments with carrier and number.
	TrackingNumbers []SupportShippingTrackingNumber `json:"tracking_numbers"`
	Message         string                          `json:"message"` // error or warning
	Status          string                          `json:"status"`
	LineItems       string                          `json:"line_items"` // comma-separated ids
}

// SupportShippingTrackingNumber is the booked carrier and number.
type SupportShippingTrackingNumber struct {
	TrackingCompany TrackingCompany `json:"tracking_company"`
	TrackingNumber  string          `json:"tracking_number"`
}

// CVSShippingLabelsRequest selects the CVS labels to print
// (POST /v2/orders/cvs_shipping_labels).
type CVSShippingLabelsRequest struct {
	ShippingType   CVSShippingType `json:"shipping_type"`
	FulfillmentIDs []int64         `json:"fulfillment_ids"`
}

// SupportShippingsRequest books home-delivery labels for several orders
// (POST /v2/orders/fulfillments/support_shippings).
type SupportShippingsRequest struct {
	ShippingType   SupportShippingSource `json:"shipping_type"`
	Size           int                   `json:"size"`
	Temperature    ShippingTemperature   `json:"temperature"`
	FridgeOrFrozen FridgeOrFrozen        `json:"fridge_or_frozen,omitzero"` // not for sf, hct
	IsFragile      bool                  `json:"is_fragile"`
	UseTransfer    *bool                 `json:"use_transfer,omitzero"` // not for sf, hct
	ShippingOrders []ShippingOrder       `json:"shipping_orders"`
}

// ShippingOrder is one order and its line items in a SupportShippingsRequest.
type ShippingOrder struct {
	OrderID     int64   `json:"order_id"`
	LineItemIDs []int64 `json:"line_item_ids"`
}

// SupportShippingsResult is the outcome of a SupportShippingsRequest.
type SupportShippingsResult struct {
	FailedOrders []FailedShippingOrder `json:"failed_orders"`
	Fulfillments []ShippingFulfillment `json:"fulfillments"`
}

// FailedShippingOrder is an order that could not be booked.
type FailedShippingOrder struct {
	OrderID int64  `json:"order_id"`
	Message string `json:"message"`
}

// ShippingFulfillment is a fulfillment created by a batch booking.
// TrackingNumber is empty while the carrier has not yet assigned one.
type ShippingFulfillment struct {
	ID              int64           `json:"id"`
	OrderID         int64           `json:"order_id"`
	TrackingNumber  string          `json:"tracking_number"`
	TrackingCompany TrackingCompany `json:"tracking_company"`
}

// SupportShippingLabelsRequest selects the home-delivery labels to print
// (POST /v2/orders/fulfillments/support_shipping_labels).
type SupportShippingLabelsRequest struct {
	ShippingType SupportShippingSource    `json:"shipping_type"`
	PrintType    SupportShippingPrintType `json:"print_type,omitzero"` // HCT only
	OrderIDs     []int64                  `json:"order_ids"`
}

// ListFulfillments returns one page of an order's fulfillments
// (GET /v1/orders/{id}/fulfillments).
func (s *OrdersService) ListFulfillments(ctx context.Context, orderID int64, opts *ListOptions) (*Page[Fulfillment], error) {
	q, err := query.Values(opts)
	if err != nil {
		return nil, err
	}
	return list[Fulfillment](ctx, s.client, fmt.Sprintf("v1/orders/%d/fulfillments", orderID), q)
}

// GetFulfillment returns one fulfillment
// (GET /v1/orders/{id}/fulfillments/{fulfillment_id}).
func (s *OrdersService) GetFulfillment(ctx context.Context, orderID, fulfillmentID int64) (*Fulfillment, *Response, error) {
	var out Fulfillment
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/orders/%d/fulfillments/%d", orderID, fulfillmentID), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// postFulfillment posts body and decodes the created fulfillment; a 202 or
// null reply yields a nil fulfillment.
func (s *OrdersService) postFulfillment(ctx context.Context, path string, body any) (*Fulfillment, *Response, error) {
	var out Fulfillment
	resp, err := s.client.post(ctx, path, body, &out)
	if err != nil {
		return nil, resp, err
	}
	if resp.StatusCode == http.StatusAccepted || resp.IsNull() || len(resp.Body) == 0 {
		return nil, resp, nil
	}
	return &out, resp, nil
}

// CreateCustomShipping ships line items with a merchant-arranged carrier
// (POST /v1/orders/{id}/fulfillments/custom_shipping).
func (s *OrdersService) CreateCustomShipping(ctx context.Context, orderID int64, req *CustomShippingRequest) (*Fulfillment, *Response, error) {
	return s.postFulfillment(ctx, fmt.Sprintf("v1/orders/%d/fulfillments/custom_shipping", orderID), req)
}

// CreateSupportShipping books a home-delivery shipment through CYBERBIZ. The
// platform may answer 202 while the tracking number is still being issued;
// the fulfillment is then nil and can be read with ListFulfillments later
// (POST /v1/orders/{id}/fulfillments/support_shipping).
func (s *OrdersService) CreateSupportShipping(ctx context.Context, orderID int64, req *SupportShippingRequest) (*Fulfillment, *Response, error) {
	return s.postFulfillment(ctx, fmt.Sprintf("v1/orders/%d/fulfillments/support_shipping", orderID), req)
}

// CreateCVSShippingV1 books a convenience-store shipment through the v1
// endpoint; prefer CreateCVSShipping, whose reply carries the label type
// (POST /v1/orders/{id}/fulfillments/cvs_shipping).
func (s *OrdersService) CreateCVSShippingV1(ctx context.Context, orderID int64, req *CVSShippingRequest) (*Fulfillment, *Response, error) {
	return s.postFulfillment(ctx, fmt.Sprintf("v1/orders/%d/fulfillments/cvs_shipping", orderID), req)
}

// CreatePartialCVSShipping ships part of a Hi-Life order
// (POST /v1/orders/{id}/fulfillments/partial_cvs_shipping).
func (s *OrdersService) CreatePartialCVSShipping(ctx context.Context, orderID int64, req *PartialCVSShippingRequest) (*Fulfillment, *Response, error) {
	return s.postFulfillment(ctx, fmt.Sprintf("v1/orders/%d/fulfillments/partial_cvs_shipping", orderID), req)
}

// ConcludePartialCVSShipping closes partial shipping on a Hi-Life order
// (POST /v1/orders/{id}/fulfillments/partial_cvs_shipping/conclude).
func (s *OrdersService) ConcludePartialCVSShipping(ctx context.Context, orderID int64) (*Order, *Response, error) {
	return s.postOrder(ctx, fmt.Sprintf("v1/orders/%d/fulfillments/partial_cvs_shipping/conclude", orderID), nil)
}

// CreateExpressDeliveryShipping ships line items by express delivery from a
// branch store (POST /v1/orders/{id}/fulfillments/express_delivery_shipping).
func (s *OrdersService) CreateExpressDeliveryShipping(ctx context.Context, orderID int64, lineItemIDs []int64) (*Fulfillment, *Response, error) {
	body := struct {
		LineItemIDs IDList `json:"line_item_ids"`
	}{LineItemIDs: lineItemIDs}
	return s.postFulfillment(ctx, fmt.Sprintf("v1/orders/%d/fulfillments/express_delivery_shipping", orderID), body)
}

// MarkArrived marks a CVS order as arrived at the store
// (POST /v1/orders/{id}/fulfillments/arrived).
func (s *OrdersService) MarkArrived(ctx context.Context, orderID int64) (*Order, *Response, error) {
	return s.postOrder(ctx, fmt.Sprintf("v1/orders/%d/fulfillments/arrived", orderID), nil)
}

// MarkReceived marks an order as received by the customer
// (POST /v1/orders/{id}/fulfillments/received).
func (s *OrdersService) MarkReceived(ctx context.Context, orderID int64) (*Order, *Response, error) {
	return s.postOrder(ctx, fmt.Sprintf("v1/orders/%d/fulfillments/received", orderID), nil)
}

// MarkExpired marks a CVS order as not collected in time
// (POST /v1/orders/{id}/fulfillments/expired).
func (s *OrdersService) MarkExpired(ctx context.Context, orderID int64) (*Order, *Response, error) {
	return s.postOrder(ctx, fmt.Sprintf("v1/orders/%d/fulfillments/expired", orderID), nil)
}

// GetFulfillmentInfos returns the CVS label data of the given fulfillments
// (POST /v1/orders/fulfillment_infos).
func (s *OrdersService) GetFulfillmentInfos(ctx context.Context, fulfillmentIDs []int64) ([]FulfillmentInfo, *Response, error) {
	body := struct {
		FulfillmentIDs []int64 `json:"fulfillment_ids"`
	}{FulfillmentIDs: fulfillmentIDs}
	var out []FulfillmentInfo
	resp, err := s.client.post(ctx, "v1/orders/fulfillment_infos", body, &out)
	if err != nil {
		return nil, resp, err
	}
	return out, resp, nil
}

// CreateSupportShippingBatch books home-delivery shipments for several
// orders through the v1 endpoint (POST /v1/orders/fulfillments/support_shipping).
func (s *OrdersService) CreateSupportShippingBatch(ctx context.Context, req *SupportShippingBatchRequest) (*SupportShippingBatchResult, *Response, error) {
	var out SupportShippingBatchResult
	resp, err := s.client.post(ctx, "v1/orders/fulfillments/support_shipping", req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// CreateCVSShipping books a convenience-store shipment. Only a reply with a
// non-empty CVSShippingType can be printed with PrintCVSShippingLabels
// (POST /v2/orders/{id}/cvs_shipping).
func (s *OrdersService) CreateCVSShipping(ctx context.Context, orderID int64, req *CVSShippingRequest) (*Fulfillment, *Response, error) {
	return s.postFulfillment(ctx, fmt.Sprintf("v2/orders/%d/cvs_shipping", orderID), req)
}

// PrintCVSShippingLabels returns a zip archive of CVS shipping labels in
// Response.Body (POST /v2/orders/cvs_shipping_labels).
func (s *OrdersService) PrintCVSShippingLabels(ctx context.Context, req *CVSShippingLabelsRequest) (*Response, error) {
	return s.client.post(ctx, "v2/orders/cvs_shipping_labels", req, nil)
}

// CreateSupportShippings books home-delivery labels for several orders
// (POST /v2/orders/fulfillments/support_shippings).
func (s *OrdersService) CreateSupportShippings(ctx context.Context, req *SupportShippingsRequest) (*SupportShippingsResult, *Response, error) {
	var out SupportShippingsResult
	resp, err := s.client.post(ctx, "v2/orders/fulfillments/support_shippings", req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// PrintSupportShippingLabels returns a zip archive of home-delivery labels in
// Response.Body (POST /v2/orders/fulfillments/support_shipping_labels).
func (s *OrdersService) PrintSupportShippingLabels(ctx context.Context, req *SupportShippingLabelsRequest) (*Response, error) {
	return s.client.post(ctx, "v2/orders/fulfillments/support_shipping_labels", req, nil)
}
