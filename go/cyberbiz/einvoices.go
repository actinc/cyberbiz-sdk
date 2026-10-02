package cyberbiz

import (
	"context"
	"fmt"
	"net/url"
)

// EinvoicesService exposes electronic invoices (/v1/einvoices) and offline
// invoice submission (/v1/offline_einvoices).
type EinvoicesService struct {
	client *Client
}

// EinvoiceUpdateRequest is the body of UpdateByOrder. InvoiceType is
// required; everything else is optional.
type EinvoiceUpdateRequest struct {
	InvoiceType   InvoiceType   `json:"invoice_type"`
	Title         *string       `json:"title,omitzero"`
	CompanyNo     *string       `json:"company_no,omitzero"`
	InvoiceNo     *string       `json:"invoice_no,omitzero"`
	InvoiceStatus InvoiceStatus `json:"invoice_status,omitzero"`
	InvoiceAt     Time          `json:"invoice_at,omitzero"`
	InvalidAt     Time          `json:"invalid_at,omitzero"`
	RandomNum     *string       `json:"random_num,omitzero"`
	LoveCode      *string       `json:"love_code,omitzero"`
	PhoneBarcode  *string       `json:"phone_barcode,omitzero"`
	NaturePerson  *string       `json:"nature_person,omitzero"`
}

// OfflineEinvoiceRequest is the body of CreateOffline.
type OfflineEinvoiceRequest struct {
	CustomerID      int64           `json:"customer_id"`
	OfflineEinvoice OfflineEinvoice `json:"offline_einvoice"`
}

// Get returns one invoice by its id (GET /v1/einvoices/{id}).
func (s *EinvoicesService) Get(ctx context.Context, id int64) (*Einvoice, *Response, error) {
	var out Einvoice
	resp, err := s.client.getOne(ctx, fmt.Sprintf("v1/einvoices/%d", id), nil, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// GetByNumber returns one invoice by its invoice number
// (GET /v1/einvoices/get_einvoice).
func (s *EinvoicesService) GetByNumber(ctx context.Context, invoiceNo string) (*Einvoice, *Response, error) {
	var out Einvoice
	q := url.Values{"invoice_no": {invoiceNo}}
	resp, err := s.client.getOne(ctx, "v1/einvoices/get_einvoice", q, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// UpdateByOrder writes the invoice data of an order. It is a custom feature
// CYBERBIZ enables per shop (PUT /v1/einvoices/{order_id}).
func (s *EinvoicesService) UpdateByOrder(ctx context.Context, orderID int64, req *EinvoiceUpdateRequest) (*Einvoice, *Response, error) {
	var out Einvoice
	resp, err := s.client.put(ctx, fmt.Sprintf("v1/einvoices/%d", orderID), req, &out)
	if err != nil {
		return nil, resp, err
	}
	return &out, resp, nil
}

// CreateOffline submits an invoice issued outside the shop so the customer
// earns bonus points for it (POST /v1/offline_einvoices).
func (s *EinvoicesService) CreateOffline(ctx context.Context, req *OfflineEinvoiceRequest) (*Response, error) {
	return s.client.post(ctx, "v1/offline_einvoices", req, nil)
}
