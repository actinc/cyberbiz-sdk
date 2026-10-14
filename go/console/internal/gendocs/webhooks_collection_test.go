package gendocs

import "testing"

func TestCollectionSectionsFromTableAndBody(t *testing.T) {
	desc := "## stock_invoices/created\n\n### Payload\n\n| 欄位 | 型別 | 說明 |\n|---|---|---|\n" +
		"| `id` | Integer | ID |\n| `items` | StockLineItemEntity | 出倉項目 |\n| `items.sku` | String | SKU |\n" +
		"| `primary` | Grape::API::Boolean | 主要 |\n\n### 來源\n"
	fields := payloadTable(desc)
	if len(fields) != 4 || fields[1].Name != "items" {
		t.Fatalf("fields = %+v", fields)
	}
	src := &webhookSource{Sections: map[string][]whField{}}
	coll := map[string]collectionEvent{"stock_invoices/created": {Fields: fields, Body: map[string]any{"items": map[string]any{"sku": "x"}}}}
	addCollectionSections(src, coll, []webhookEvent{{Code: "stock_invoices/created", Payload: "Stock Invoice"}})
	top := src.fields("Stock Invoice")
	if len(top) != 3 || top[1].Type != "[Object]" || top[2].Type != "Boolean" {
		t.Errorf("Stock Invoice = %+v", top)
	}
	if src.Links["Stock Invoice/items"] != "Stock Invoice items" || len(src.fields("Stock Invoice items")) != 1 {
		t.Errorf("links = %v, items = %+v", src.Links, src.fields("Stock Invoice items"))
	}
}

func TestFieldsFromBodyWhenNoTable(t *testing.T) {
	body := map[string]any{"job_id": "sidekiq jid", "result": map[string]any{"sended": float64(0)}}
	got := fieldsFromBody(body, "")
	if len(got) != 3 || got[0].Name != "job_id" || got[0].Desc != "sidekiq jid" || got[2].Name != "result.sended" || got[2].Type != "Integer" {
		t.Errorf("fields = %+v", got)
	}
}
