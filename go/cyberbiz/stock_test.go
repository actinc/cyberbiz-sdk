package cyberbiz

import (
	"encoding/json/v2"
	"errors"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

// stockRecordedRequest records one request received by stockRecordingClient. The prefix keeps the
// helper names distinct from those of other service test files.
type stockRecordedRequest struct {
	Method string
	Path   string
	Query  url.Values
	Body   map[string]any // nil when the request had no body
}

// stockRecordingClient returns a client whose server records the request and answers
// with status and body.
func stockRecordingClient(t *testing.T, status int, body string) (*Client, *stockRecordedRequest) {
	t.Helper()
	call := &stockRecordedRequest{}
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		call.Method, call.Path, call.Query = r.Method, r.URL.Path, r.URL.Query()
		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading body: %v", err)
		}
		if len(data) > 0 {
			if err := json.Unmarshal(data, &call.Body); err != nil {
				t.Errorf("body is not a JSON object: %v", err)
			}
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}, WithMaxRetries(0))
	return c, call
}

// stockShapeCase is one request-shape expectation.
type stockShapeCase struct {
	name   string
	call   func(c *Client) error
	method string
	path   string
	query  string         // wanted r.URL.RawQuery after canonical encoding
	body   map[string]any // wanted keys and values; nil means no body
	absent []string       // keys that must be omitted from the body
	reply  string         // server reply, "{}" when empty
	status int            // server status, 200 when zero
}

func stockRunShapes(t *testing.T, cases []stockShapeCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reply, status := tc.reply, tc.status
			if reply == "" {
				reply = "{}"
			}
			if status == 0 {
				status = http.StatusOK
			}
			c, got := stockRecordingClient(t, status, reply)
			if err := tc.call(c); err != nil {
				t.Fatal(err)
			}
			if got.Method != tc.method || got.Path != tc.path {
				t.Errorf("got %s %s, want %s %s", got.Method, got.Path, tc.method, tc.path)
			}
			if q := got.Query.Encode(); q != tc.query {
				t.Errorf("query = %q, want %q", q, tc.query)
			}
			if tc.body == nil && got.Body != nil {
				t.Errorf("unexpected body %v", got.Body)
			}
			for k, want := range tc.body {
				if v, ok := got.Body[k]; !ok || !stockJSONEqual(v, want) {
					t.Errorf("body[%q] = %v (%T), want %v (%T)", k, v, v, want, want)
				}
			}
			for _, k := range tc.absent {
				if _, ok := got.Body[k]; ok {
					t.Errorf("body[%q] should be omitted", k)
				}
			}
		})
	}
}

// stockJSONEqual compares a decoded JSON value with an expectation; both
// sides use the encoding/json/v2 "any" representation.
func stockJSONEqual(got, want any) bool { return reflect.DeepEqual(got, want) }

func stockPtr[T any](v T) *T { return &v }

func TestStockReceiptGolden(t *testing.T) {
	var page []StockReceipt
	decodeGolden(t, "v1/GET_v1_stock_receipts.json", &page)
	if len(page) != 2 {
		t.Fatalf("len = %d", len(page))
	}
	r := page[0]
	if r.ID != 979332 || r.PosShopID != 1446 || r.SourcePosShopID != 0 || r.StockInvoiceID != 370648 {
		t.Errorf("ids = %+v", r)
	}
	if r.Status != StockStatusDone || r.SourcePosShopName != "官網" {
		t.Errorf("status = %q, source = %q", r.Status, r.SourcePosShopName)
	}
	if len(r.Items) != 1 || r.Items[0].SKU != "SKU-8102e6" || r.Items[0].Quantity != 3 || r.Items[0].QC != "" {
		t.Errorf("items = %+v", r.Items)
	}
	if len(r.CheckLogs) != 1 || r.CheckLogs[0].CreatedAt.String() != "2026-02-25 10:02:05" || r.CheckLogs[0].Items[0].Quantity != 0 {
		t.Errorf("check_logs = %+v", r.CheckLogs)
	}
	if r.CreatedAt.String() != "2025-12-19 15:59:28" || r.Comment != "" {
		t.Errorf("created_at = %v, comment = %q", r.CreatedAt, r.Comment)
	}

	var one StockReceipt
	decodeGolden(t, "v1/GET_v1_stock_receipts_{id}.json", &one)
	if one.ID != 979332 || one.Barcode == "" {
		t.Errorf("detail = %+v", one)
	}
	var arrived StockReceipt
	decodeGolden(t, "v1/GET_v1_stock_receipts_{id}_arrival.json", &arrived)
	if arrived.Status != StockStatusProcessing {
		t.Errorf("arrival status = %q", arrived.Status)
	}
}

func TestStockRequisitionGolden(t *testing.T) {
	var page []StockRequisition
	decodeGolden(t, "v1/GET_v1_stock_requisitions.json", &page)
	if len(page) != 2 || page[0].ID != 87314 || page[0].Status != StockStatusDone {
		t.Fatalf("page = %+v", page)
	}
	if page[0].SourceShopName != "商家 1" || page[0].PosShopName != "官網" {
		t.Errorf("names = %q / %q", page[0].SourceShopName, page[0].PosShopName)
	}
	var one StockRequisition
	decodeGolden(t, "v1/GET_v1_stock_requisitions_{id}.json", &one)
	if len(one.Items) != 4 || one.Items[3].SKU != "SKU-b17d99" || one.Items[3].Quantity != 1 {
		t.Errorf("items = %+v", one.Items)
	}
	if one.CreatedAt.String() != "2025-12-19 16:06:16" || one.Comment != "" {
		t.Errorf("created_at = %v, comment = %q", one.CreatedAt, one.Comment)
	}
}

func TestStockAdjustmentGolden(t *testing.T) {
	var page []StockAdjustment
	decodeGolden(t, "v1/GET_v1_stock_adjustments.json", &page)
	if len(page) != 2 || page[0].ID != 9336548 || len(page[0].Items) != 7 {
		t.Fatalf("page = %+v", page)
	}
	item := page[0].Items[2]
	if item.SKU != "SKU-7b9ac9" || item.Quantity != 999 || item.Type != StockAdjustmentTypeSurplus {
		t.Errorf("item = %+v", item)
	}
	if item.Note != "" || item.QC != "" || item.CreatedAt.String() != "2023-09-08 17:58:48" {
		t.Errorf("item nullable fields = %+v", item)
	}
	var one StockAdjustment
	decodeGolden(t, "v1/GET_v1_stock_adjustments_{id}.json", &one)
	if one.ID != 9336548 || len(one.Items) != 7 {
		t.Errorf("detail = %+v", one)
	}
	var items []StockAdjustmentItem
	decodeGolden(t, "v1/GET_v1_stock_adjustments_items.json", &items)
	if len(items) != 2 || items[0].SKU != "SKU-7b5fc5" || items[0].Type != StockAdjustmentTypeSurplus {
		t.Errorf("items = %+v", items)
	}
}

func TestStockInvoiceGolden(t *testing.T) {
	var page []StockInvoice
	decodeGolden(t, "v1/GET_v1_stock_invoices.json", &page)
	if len(page) != 2 {
		t.Fatalf("len = %d", len(page))
	}
	inv := page[0]
	if inv.ID != 370648 || inv.PosShopID != 0 || inv.TargetPosShopID != 1446 || inv.StockReceiptID != 979332 {
		t.Errorf("ids = %+v", inv)
	}
	if inv.Status != StockStatusDone || inv.TargetPosShopName != "商家 1" || inv.Comment != "" {
		t.Errorf("fields = %+v", inv)
	}
	if len(inv.Items) != 1 || inv.Items[0].Quantity != 3 || inv.CreatedAt.String() != "2025-12-19 15:59:28" {
		t.Errorf("items = %+v, created_at = %v", inv.Items, inv.CreatedAt)
	}
	var one StockInvoice
	decodeGolden(t, "v1/GET_v1_stock_invoices_{id}.json", &one)
	if one.ID != 370648 {
		t.Errorf("detail = %+v", one)
	}
}

func TestStockActionErrorsGolden(t *testing.T) {
	cases := []struct {
		golden string
		status int
		want   error
		msg    string
		call   func(c *Client) error
	}{
		{"errors/GET_v1_stock_receipts_{id}_approve.json", 403, ErrForbidden, "Receipt can not be approved.",
			func(c *Client) error { _, _, err := c.Stock.ApproveReceipt(testCtx, 1); return err }},
		{"errors/GET_v1_stock_receipts_{id}_reject.json", 403, ErrForbidden, "Receipt can not be rejected.",
			func(c *Client) error { _, _, err := c.Stock.RejectReceipt(testCtx, 1); return err }},
		{"errors/GET_v1_stock_receipts_{id}_cancel.json", 403, ErrForbidden, "Receipt can not be canceled.",
			func(c *Client) error { _, _, err := c.Stock.CancelReceipt(testCtx, 1); return err }},
		{"errors/GET_v1_stock_requisitions_{id}_approve.json", 403, ErrForbidden, "Requisition is not pending.",
			func(c *Client) error { _, _, err := c.Stock.ApproveRequisition(testCtx, 1); return err }},
		{"errors/GET_v1_stock_requisitions_{id}_reject.json", 403, ErrForbidden, "Requisition is not pending.",
			func(c *Client) error { _, _, err := c.Stock.RejectRequisition(testCtx, 1); return err }},
		{"errors/GET_v1_stock_invoices_{id}_confirm.json", 403, ErrForbidden, "Invoice is not ready.",
			func(c *Client) error { _, _, err := c.Stock.ConfirmInvoice(testCtx, 1); return err }},
		{"errors/GET_v1_stock_invoices_{id}_cancel.json", 403, ErrForbidden, "Invoice is not ready.",
			func(c *Client) error { _, _, err := c.Stock.CancelInvoice(testCtx, 1); return err }},
		{"errors/GET_v1_inventory_sync_groups.json", 422, ErrValidation, "無「庫存同步群組」功能，請聯絡您的開店顧問",
			func(c *Client) error { _, err := c.Stock.ListInventorySyncGroups(testCtx, nil); return err }},
		{"errors/GET_v1_inventory_sync_groups_{id}.json", 422, ErrValidation, "無「庫存同步群組」功能，請聯絡您的開店顧問",
			func(c *Client) error { _, _, err := c.Stock.GetInventorySyncGroup(testCtx, 1); return err }},
	}
	for _, tc := range cases {
		t.Run(tc.golden, func(t *testing.T) {
			c, _ := stockRecordingClient(t, tc.status, string(readGolden(t, tc.golden)))
			err := tc.call(c)
			var apiErr *APIError
			if !errors.Is(err, tc.want) || !errors.As(err, &apiErr) {
				t.Fatalf("got %v", err)
			}
			if len(apiErr.Messages) != 1 || apiErr.Messages[0] != tc.msg {
				t.Errorf("messages = %q", apiErr.Messages)
			}
		})
	}
}

func TestStockGoldenRoundTrip(t *testing.T) {
	c := goldenServer(t)
	receipts, err := c.Stock.ListReceipts(testCtx, &StockReceiptListOptions{PosShopID: stockPtr[int64](1446)})
	if err != nil || len(receipts.Items) != 2 || receipts.Pagination.Total != 1 {
		t.Fatalf("ListReceipts: %v %+v", err, receipts)
	}
	if r, _, err := c.Stock.GetReceipt(testCtx, 979332); err != nil || r.ID != 979332 {
		t.Errorf("GetReceipt: %v %+v", err, r)
	}
	if r, _, err := c.Stock.ArriveReceipt(testCtx, 979332); err != nil || r.Status != StockStatusProcessing {
		t.Errorf("ArriveReceipt: %v %+v", err, r)
	}
	if p, err := c.Stock.ListRequisitions(testCtx, nil); err != nil || p.Items[0].ID != 87314 {
		t.Errorf("ListRequisitions: %v", err)
	}
	if r, _, err := c.Stock.GetRequisition(testCtx, 87314); err != nil || len(r.Items) != 4 {
		t.Errorf("GetRequisition: %v", err)
	}
	if p, err := c.Stock.ListAdjustments(testCtx, nil); err != nil || p.Items[0].ID != 9336548 {
		t.Errorf("ListAdjustments: %v", err)
	}
	if a, _, err := c.Stock.GetAdjustment(testCtx, 9336548); err != nil || len(a.Items) != 7 {
		t.Errorf("GetAdjustment: %v", err)
	}
	if p, err := c.Stock.ListAdjustmentItems(testCtx, nil); err != nil || p.Items[0].SKU != "SKU-7b5fc5" {
		t.Errorf("ListAdjustmentItems: %v", err)
	}
	if p, err := c.Stock.ListInvoices(testCtx, nil); err != nil || p.Items[0].ID != 370648 {
		t.Errorf("ListInvoices: %v", err)
	}
	if inv, _, err := c.Stock.GetInvoice(testCtx, 370648); err != nil || inv.StockReceiptID != 979332 {
		t.Errorf("GetInvoice: %v", err)
	}
	var n int
	for _, err := range c.Stock.AllReceipts(testCtx, nil) {
		if err != nil {
			t.Fatal(err)
		}
		n++
	}
	if n != 2 {
		t.Errorf("AllReceipts yielded %d", n)
	}
}

func TestStockRequestShapes(t *testing.T) {
	items := []StockLineItemInput{{SKU: "sku-1", Quantity: 2}}
	wantItems := []any{map[string]any{"sku": "sku-1", "quantity": 2.0}}
	stockRunShapes(t, []stockShapeCase{
		{name: "ListReceipts filters", method: "GET", path: "/v1/stock_receipts", reply: "[]",
			query: "end_date=2026-02-28&page=2&per_page=10&pos_shop_id=0&source_pos_shop_id=-1&start_date=2026-01-01&status=ready",
			call: func(c *Client) error {
				_, err := c.Stock.ListReceipts(testCtx, &StockReceiptListOptions{
					ListOptions: ListOptions{Page: 2, PerPage: 10}, PosShopID: stockPtr[int64](0), SourcePosShopID: stockPtr[int64](-1),
					Status: StockStatusReady, StartDate: NewDate(2026, 1, 1), EndDate: NewDate(2026, 2, 28)})
				return err
			}},
		{name: "ListReceipts nil options", method: "GET", path: "/v1/stock_receipts", reply: "[]",
			call: func(c *Client) error { _, err := c.Stock.ListReceipts(testCtx, nil); return err }},
		{name: "CreateReceipt", method: "POST", path: "/v1/stock_receipts",
			body:   map[string]any{"pos_shop_id": 0.0, "items": wantItems, "est_date": "2026-03-01"},
			absent: []string{"comment"},
			call: func(c *Client) error {
				_, _, err := c.Stock.CreateReceipt(testCtx, &StockReceiptCreateRequest{Items: items, EstDate: NewDate(2026, 3, 1)})
				return err
			}},
		{name: "CreateReceipt explicit empty comment", method: "POST", path: "/v1/stock_receipts",
			body:   map[string]any{"pos_shop_id": 1446.0, "comment": ""},
			absent: []string{"est_date"},
			call: func(c *Client) error {
				_, _, err := c.Stock.CreateReceipt(testCtx, &StockReceiptCreateRequest{PosShopID: 1446, Items: items, Comment: stockPtr("")})
				return err
			}},
		{name: "GetReceipt", method: "GET", path: "/v1/stock_receipts/5",
			call: func(c *Client) error { _, _, err := c.Stock.GetReceipt(testCtx, 5); return err }},
		{name: "ApproveReceipt", method: "GET", path: "/v1/stock_receipts/5/approve",
			call: func(c *Client) error { _, _, err := c.Stock.ApproveReceipt(testCtx, 5); return err }},
		{name: "RejectReceipt", method: "GET", path: "/v1/stock_receipts/5/reject",
			call: func(c *Client) error { _, _, err := c.Stock.RejectReceipt(testCtx, 5); return err }},
		{name: "CancelReceipt", method: "GET", path: "/v1/stock_receipts/5/cancel",
			call: func(c *Client) error { _, _, err := c.Stock.CancelReceipt(testCtx, 5); return err }},
		{name: "ArriveReceipt", method: "GET", path: "/v1/stock_receipts/5/arrival",
			call: func(c *Client) error { _, _, err := c.Stock.ArriveReceipt(testCtx, 5); return err }},
		{name: "CheckReceipt", method: "POST", path: "/v1/stock_receipts/5/check",
			body: map[string]any{"items": wantItems},
			call: func(c *Client) error {
				_, _, err := c.Stock.CheckReceipt(testCtx, 5, &StockReceiptCheckRequest{Items: items})
				return err
			}},
		{name: "ListRequisitions filters", method: "GET", path: "/v1/stock_requisitions", reply: "[]",
			query: "pos_shop_id=3&source_pos_shop_id=0&status=pending",
			call: func(c *Client) error {
				_, err := c.Stock.ListRequisitions(testCtx, &StockRequisitionListOptions{
					PosShopID: stockPtr[int64](3), SourcePosShopID: stockPtr[int64](0), Status: StockStatusPending})
				return err
			}},
		{name: "GetRequisition", method: "GET", path: "/v1/stock_requisitions/7",
			call: func(c *Client) error { _, _, err := c.Stock.GetRequisition(testCtx, 7); return err }},
		{name: "CreateRequisition", method: "POST", path: "/v1/stock_requisitions",
			body:   map[string]any{"pos_shop_id": 3.0, "source_pos_shop_id": 0.0, "items": wantItems},
			absent: []string{"comment"},
			call: func(c *Client) error {
				_, _, err := c.Stock.CreateRequisition(testCtx, &StockRequisitionCreateRequest{PosShopID: 3, Items: items})
				return err
			}},
		{name: "ApproveRequisition", method: "GET", path: "/v1/stock_requisitions/7/approve",
			call: func(c *Client) error { _, _, err := c.Stock.ApproveRequisition(testCtx, 7); return err }},
		{name: "RejectRequisition", method: "GET", path: "/v1/stock_requisitions/7/reject",
			call: func(c *Client) error { _, _, err := c.Stock.RejectRequisition(testCtx, 7); return err }},
		{name: "ListAdjustments", method: "GET", path: "/v1/stock_adjustments", reply: "[]", query: "pos_shop_id=0",
			call: func(c *Client) error {
				_, err := c.Stock.ListAdjustments(testCtx, &StockAdjustmentListOptions{PosShopID: stockPtr[int64](0)})
				return err
			}},
		{name: "GetAdjustment", method: "GET", path: "/v1/stock_adjustments/9",
			call: func(c *Client) error { _, _, err := c.Stock.GetAdjustment(testCtx, 9); return err }},
		{name: "CreateAdjustment", method: "POST", path: "/v1/stock_adjustments",
			body: map[string]any{"pos_shop_id": 0.0, "items": []any{map[string]any{"sku": "s", "quantity": -1.0, "type": "loss"}}},
			call: func(c *Client) error {
				_, _, err := c.Stock.CreateAdjustment(testCtx, &StockAdjustmentCreateRequest{
					Items: []StockAdjustmentItemInput{{SKU: "s", Quantity: -1, Type: StockAdjustmentTypeLoss}}})
				return err
			}},
		{name: "ListAdjustmentItems filters", method: "GET", path: "/v1/stock_adjustments/items", reply: "[]",
			query: "end_date=2026-01-31&start_date=2026-01-01&type=return",
			call: func(c *Client) error {
				_, err := c.Stock.ListAdjustmentItems(testCtx, &StockAdjustmentItemListOptions{
					StartDate: NewDate(2026, 1, 1), EndDate: NewDate(2026, 1, 31), Type: StockAdjustmentTypeReturn})
				return err
			}},
		{name: "ListInvoices filters", method: "GET", path: "/v1/stock_invoices", reply: "[]",
			query: "status=done&target_pos_shop_id=-1",
			call: func(c *Client) error {
				_, err := c.Stock.ListInvoices(testCtx, &StockInvoiceListOptions{TargetPosShopID: stockPtr[int64](-1), Status: StockStatusDone})
				return err
			}},
		{name: "GetInvoice", method: "GET", path: "/v1/stock_invoices/11",
			call: func(c *Client) error { _, _, err := c.Stock.GetInvoice(testCtx, 11); return err }},
		{name: "CreateInvoice third party", method: "POST", path: "/v1/stock_invoices",
			body:   map[string]any{"pos_shop_id": 0.0, "target_type": "third_party", "third_party": "ACME", "items": wantItems},
			absent: []string{"target_pos_shop_id", "est_date", "comment"},
			call: func(c *Client) error {
				_, _, err := c.Stock.CreateInvoice(testCtx, &StockInvoiceCreateRequest{
					TargetType: StockInvoiceTargetThirdParty, ThirdParty: stockPtr("ACME"), Items: items})
				return err
			}},
		{name: "CreateInvoice to EC", method: "POST", path: "/v1/stock_invoices",
			body: map[string]any{"pos_shop_id": 1446.0, "target_type": "ec", "target_pos_shop_id": 0.0},
			call: func(c *Client) error {
				_, _, err := c.Stock.CreateInvoice(testCtx, &StockInvoiceCreateRequest{
					PosShopID: 1446, TargetType: StockInvoiceTargetEC, TargetPosShopID: stockPtr[int64](0), Items: items})
				return err
			}},
		{name: "ConfirmInvoice", method: "GET", path: "/v1/stock_invoices/11/confirm",
			call: func(c *Client) error { _, _, err := c.Stock.ConfirmInvoice(testCtx, 11); return err }},
		{name: "CancelInvoice", method: "GET", path: "/v1/stock_invoices/11/cancel",
			call: func(c *Client) error { _, _, err := c.Stock.CancelInvoice(testCtx, 11); return err }},
	})
}

func TestInventorySyncGroupRequestShapes(t *testing.T) {
	group := `{"title":"T","inventory_quantity":5,"inventory_policy":"deny","variants":[{"id":1,"inventory_sync_group_id":2,"product_variant_id":3,"created_at":"2026-01-01 00:00:00","updated_at":null}]}`
	stockRunShapes(t, []stockShapeCase{
		{name: "ListInventorySyncGroups", method: "GET", path: "/v1/inventory_sync_groups", reply: "[" + group + "]",
			query: "per_page=5",
			call: func(c *Client) error {
				p, err := c.Stock.ListInventorySyncGroups(testCtx, &ListOptions{PerPage: 5})
				if err == nil && (len(p.Items) != 1 || p.Items[0].InventoryPolicy != InventorySyncPolicyDeny) {
					t.Errorf("page = %+v", p.Items)
				}
				return err
			}},
		{name: "GetInventorySyncGroup fills id", method: "GET", path: "/v1/inventory_sync_groups/2", reply: group,
			call: func(c *Client) error {
				g, _, err := c.Stock.GetInventorySyncGroup(testCtx, 2)
				if err == nil && (g.ID != 2 || g.Variants[0].ProductVariantID != 3 || !g.Variants[0].UpdatedAt.IsZero()) {
					t.Errorf("group = %+v", g)
				}
				return err
			}},
		{name: "CreateInventorySyncGroup", method: "POST", path: "/v1/inventory_sync_groups", reply: group,
			body: map[string]any{"title": "T", "inventory_quantity": 5.0, "inventory_policy": "continue"},
			call: func(c *Client) error {
				_, _, err := c.Stock.CreateInventorySyncGroup(testCtx, &InventorySyncGroupCreateRequest{
					Title: "T", InventoryQuantity: 5, InventoryPolicy: InventorySyncPolicyContinue})
				return err
			}},
		{name: "UpdateInventorySyncGroup", method: "PUT", path: "/v1/inventory_sync_groups/2", reply: group,
			body: map[string]any{"inventory_quantity": 0.0}, absent: []string{"title", "inventory_policy"},
			call: func(c *Client) error {
				g, _, err := c.Stock.UpdateInventorySyncGroup(testCtx, 2, &InventorySyncGroupUpdateRequest{InventoryQuantity: stockPtr(0)})
				if err == nil && g.ID != 2 {
					t.Errorf("id = %d", g.ID)
				}
				return err
			}},
		{name: "DeleteInventorySyncGroup", method: "DELETE", path: "/v1/inventory_sync_groups/2", reply: group,
			call: func(c *Client) error { _, err := c.Stock.DeleteInventorySyncGroup(testCtx, 2); return err }},
		{name: "AddInventorySyncGroupVariants", method: "POST", path: "/v1/inventory_sync_groups/2/variants", reply: group,
			body: map[string]any{"variant_ids": "10,11"},
			call: func(c *Client) error {
				_, _, err := c.Stock.AddInventorySyncGroupVariants(testCtx, 2, []int64{10, 11})
				return err
			}},
		{name: "RemoveInventorySyncGroupVariants", method: "DELETE", path: "/v1/inventory_sync_groups/2/variants",
			reply: group, query: "variant_ids=10%2C11",
			call: func(c *Client) error {
				_, _, err := c.Stock.RemoveInventorySyncGroupVariants(testCtx, 2, []int64{10, 11})
				return err
			}},
	})
}
