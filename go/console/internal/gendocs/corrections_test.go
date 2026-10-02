package gendocs

import (
	"reflect"
	"testing"
)

func TestMarkNullable(t *testing.T) {
	s := markNullable(&Schema{Type: "string"})
	if !reflect.DeepEqual(s.Type, []string{"string", "null"}) {
		t.Fatalf("string: got %v", s.Type)
	}
	if markNullable(s) != s || len(s.Type.([]string)) != 2 {
		t.Fatal("marking twice must be a no-op")
	}
	r := markNullable(&Schema{Ref: schemaRef("Order"), Description: "d"})
	if r.Ref != "" || len(r.AnyOf) != 2 || r.AnyOf[0].Ref != schemaRef("Order") || r.AnyOf[1].Type != "null" || r.Description != "d" {
		t.Fatalf("ref: got %+v", r)
	}
	if !r.IsNullable() || r.RefName() != "Order" {
		t.Fatal("anyOf ref must report nullable and its ref name")
	}
}

func TestIsMoneyKey(t *testing.T) {
	yes := []string{"price", "total_price", "subtotal_price", "cost", "bonus_remain", "order_price_threshold", "discount", "amount", "balance_after", "points", "consumption_price", "signup_bonus"}
	no := []string{"quantity", "position", "discount_name", "discount_type", "bonus_point_expiry_days", "price_discount_enabled", "weight", "sell_weight", "order_discount_value", "coupon_value", "id", "product_ids"}
	for _, k := range yes {
		if !isMoneyKey(k) {
			t.Errorf("%s should be money", k)
		}
	}
	for _, k := range no {
		if isMoneyKey(k) {
			t.Errorf("%s should not be money", k)
		}
	}
}

func miniDoc() *Document {
	order := &Schema{Type: "object", Properties: NewOMap()}
	order.SetProp("id", &Schema{Type: "integer"})
	order.SetProp("order_number", &Schema{Type: "string"})
	order.SetProp("total_price", &Schema{Type: "integer"})
	order.SetProp("created_at", &Schema{Type: "string"})
	order.SetProp("note", &Schema{Type: "string"})
	order.SetProp("exchange_histories", &Schema{Ref: schemaRef("Exchange")})
	order.SetProp("financial_status", &Schema{Type: "string"})
	exchange := &Schema{Type: "object", Properties: NewOMap()}
	exchange.SetProp("price", &Schema{Type: "number"})
	doc := &Document{
		OpenAPI: "3.1.0",
		Info:    &Info{Title: "mini", Version: "0.0.1"},
		Paths:   NewOMap(),
		Components: &Components{Schemas: NewOMap().
			Set("Order", order).
			Set("Exchange", exchange)},
	}
	op := &Operation{Tags: []string{"orders"}, Summary: "list", OperationID: "getOrders", Responses: NewOMap()}
	op.Parameters = []*Parameter{{Name: "page", In: "query", Schema: &Schema{Type: "integer"}}}
	op.Responses.Set("200", &Response{Description: "ok", Content: NewOMap().Set("application/json",
		&MediaType{Schema: &Schema{Type: "array", Items: &Schema{Ref: schemaRef("Order")}}})})
	doc.Paths.Set("/v1/orders", &PathItem{Get: op})
	doc.Tags = []*Tag{{Name: "orders"}}
	return doc
}

func TestApplyNameRules(t *testing.T) {
	doc := miniDoc()
	log := newReport()
	applyNameRules(doc, log)
	order := doc.Components.Schemas.Get2("Order")
	if order.Prop("order_number").BaseType() != "integer" {
		t.Error("order_number must become integer")
	}
	if order.Prop("total_price").RefName() != "Money" {
		t.Error("total_price must reference Money")
	}
	if order.Prop("created_at").RefName() != "Timestamp" {
		t.Error("created_at must reference Timestamp")
	}
	if order.Prop("note").BaseType() != "string" {
		t.Error("note must stay a string")
	}
	for _, id := range []string{"order-number", "money", "timestamp-suffix"} {
		if log.counts[id] == 0 {
			t.Errorf("rule %s not counted", id)
		}
	}
}

func TestApplyObserved(t *testing.T) {
	doc := miniDoc()
	obs := &observedDoc{Endpoints: map[string]map[string]*observedField{
		"/v1/orders": {
			"$":                              {Types: []string{"array"}},
			"$[]":                            {Types: []string{"object"}},
			"$[].id":                         {Types: []string{"integer"}},
			"$[].order_number":               {Types: []string{"integer"}},
			"$[].note":                       {Types: []string{"null", "string"}},
			"$[].exchange_histories":         {Types: []string{"array"}},
			"$[].exchange_histories[]":       {Types: []string{"object"}},
			"$[].exchange_histories[].price": {Types: []string{"number"}},
			"$[].created_at":                 {Types: []string{"string"}, Formats: []string{"YYYY-MM-DD HH:MM:SS"}},
			"$[].shipping_vendor":            {Types: []string{"object"}},
			"$[].shipping_vendor.name":       {Types: []string{"string", "null"}},
			"$[].shipping_vendor.type":       {Types: []string{"string"}},
		},
	}}
	log := newReport()
	applyObserved(doc, obs, log)
	order := doc.Components.Schemas.Get2("Order")
	if got := order.Prop("order_number").BaseType(); got != "integer" {
		t.Errorf("observed type: order_number = %s", got)
	}
	if !order.Prop("note").IsNullable() {
		t.Error("note observed null must be nullable")
	}
	if eh := order.Prop("exchange_histories"); eh.BaseType() != "array" || eh.Items.Ref != schemaRef("Exchange") {
		t.Errorf("exchange_histories must become an array of Exchange, got %+v", eh)
	}
	if order.Prop("created_at").RefName() != "Timestamp" {
		t.Error("observed timestamp format must reference Timestamp")
	}
	sv := order.Prop("shipping_vendor")
	if sv == nil || !sv.XObserved || sv.Prop("name") == nil || !sv.Prop("name").IsNullable() || sv.Prop("type").BaseType() != "string" {
		t.Errorf("shipping_vendor must be added from observation, got %+v", sv)
	}
	for _, id := range []string{"observed-type", "observed-null", "observed-array", "observed-format", "observed-field"} {
		if log.counts[id] == 0 {
			t.Errorf("rule %s not counted", id)
		}
	}
}

func TestApplyObservedRootEnvelope(t *testing.T) {
	doc := miniDoc()
	obs := &observedDoc{Endpoints: map[string]map[string]*observedField{
		"/v1/orders": {
			"$":              {Types: []string{"object"}},
			"$.current":      {Types: []string{"array"}},
			"$.current[]":    {Types: []string{"object"}},
			"$.current[].id": {Types: []string{"integer"}},
			"$.draft":        {Types: []string{"array"}},
		},
	}}
	applyObserved(doc, obs, newReport())
	media := successMedia(doc.Paths.Get2Path("/v1/orders").Get)
	if media.Schema.BaseType() != "object" || media.Schema.Prop("current") == nil {
		t.Fatalf("array response observed as object must become an object, got %+v", media.Schema)
	}
	if media.Schema.Prop("current").Items.Ref != schemaRef("Order") {
		t.Errorf("items should reuse the Order ref, got %+v", media.Schema.Prop("current").Items)
	}
}

// Get2Path is a test helper returning a path item.
func (m *OMap) Get2Path(p string) *PathItem {
	v, _ := m.Get(p)
	return v.(*PathItem)
}

func TestApplyEnumDocsNullable(t *testing.T) {
	doc := miniDoc()
	order := doc.Components.Schemas.Get2("Order")
	order.SetProp("financial_status", markNullable(&Schema{Type: "string"}))
	applyEnumDocs(doc, newReport())
	e := order.Prop("financial_status").Enum
	if len(e) == 0 || e[len(e)-1] != nil {
		t.Fatalf("nullable enum must end with null, got %v", e)
	}
	if order.Prop("financial_status").Description == "" {
		t.Error("enum description missing")
	}
}

func TestCorrectionsTableIsUnique(t *testing.T) {
	seen := map[string]bool{}
	for _, r := range correctionsTable {
		if seen[r.ID] {
			t.Errorf("duplicate rule %s", r.ID)
		}
		seen[r.ID] = true
		if r.Description == "" {
			t.Errorf("rule %s has no description", r.ID)
		}
	}
}
