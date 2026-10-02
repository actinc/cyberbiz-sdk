package gendocs

import (
	"encoding/json"
	"strings"
	"testing"
)

const miniSwagger = `{
  "swagger": "2.0",
  "info": {"title": "t", "version": "1", "description": "<p>hmac legacy</p>"},
  "tags": [{"name": "orders", "description": "訂單相關 API"}, {"name": "order_returns", "description": "Translation missing: x"}],
  "paths": {
    "/v1/orders": {
      "get": {
        "description": "取得所有訂單資料",
        "operationId": "getV1Orders",
        "tags": ["orders"],
        "parameters": [
          {"in": "query", "name": "page", "type": "integer", "default": 1},
          {"in": "query", "name": "per_page", "type": "integer", "default": 50},
          {"in": "query", "name": "statuses", "type": "string", "description": "訂單狀態<br> 「已開啟」: open"}
        ],
        "responses": {"200": {"description": "ok", "schema": {"type": "array", "items": {"$ref": "#/definitions/Cyberbiz_Entities_V1_OrderEntity"}}}}
      }
    },
    "/v1/orders/{order_id}": {
      "put": {
        "description": "更新",
        "operationId": "putV1OrdersOrderId",
        "tags": ["orders"],
        "parameters": [
          {"in": "path", "name": "order_id", "type": "integer", "required": true},
          {"in": "formData", "name": "note", "type": "string", "description": "備註"},
          {"in": "formData", "name": "billing_address[city]", "type": "string"},
          {"in": "formData", "name": "tags[]", "type": "array", "items": {"type": "string"}},
          {"in": "formData", "name": "extra_infos[products][][name]", "type": "string"}
        ],
        "responses": {"200": {"description": "ok", "schema": {"$ref": "#/definitions/Cyberbiz_Entities_V1_OrderEntity"}}}
      }
    },
    "/v1/assets/pictures": {
      "post": {
        "description": "upload", "operationId": "postV1AssetsPictures", "tags": ["orders"],
        "parameters": [{"in": "formData", "name": "picture", "type": "file", "required": true}],
        "responses": {"201": {"description": "created"}}
      }
    }
  },
  "definitions": {
    "Cyberbiz_Entities_V1_OrderEntity": {
      "type": "object",
      "properties": {
        "id": {"type": "integer", "description": "訂單ID"},
        "order_number": {"type": "string"},
        "total_price": {"type": "integer"},
        "third_party_discount": {"type": "Interger"},
        "created_at": {"type": "string"},
        "exchange_histories": {"$ref": "#/definitions/Cyberbiz_Entities_V1_Order_ExchangeHistoriesEntity"},
        "photo": {"type": "file"}
      }
    },
    "Cyberbiz_Entities_V1_Order_ExchangeHistoriesEntity": {"type": "object", "properties": {"price": {"type": "number", "format": "float"}}}
  }
}`

func miniInputs(t *testing.T) *inputs {
	t.Helper()
	var sw swaggerDoc
	if err := json.Unmarshal([]byte(miniSwagger), &sw); err != nil {
		t.Fatal(err)
	}
	return &inputs{
		swagger: &sw,
		observed: &observedDoc{Endpoints: map[string]map[string]*observedField{
			"/v1/orders": {
				"$":                      {Types: []string{"array"}},
				"$[]":                    {Types: []string{"object"}},
				"$[].order_number":       {Types: []string{"integer"}},
				"$[].exchange_histories": {Types: []string{"array"}},
				"$[].created_at":         {Types: []string{"string", "null"}, Formats: []string{"YYYY-MM-DD HH:MM:SS"}},
			},
		}},
		golden:    &goldenSet{},
		postmanV1: postmanIndex{},
		webhooks:  &webhookSource{Sections: map[string][]whField{}},
	}
}

func TestComponentName(t *testing.T) {
	cases := map[string]string{
		"Cyberbiz_Entities_V1_OrderEntity":                      "Order",
		"Cyberbiz_Entities_V1_Order_LineItemsEntity":            "OrderLineItems",
		"AppStore_Api_Entities_OrderCustomerCancelReasonEntity": "OrderCustomerCancelReason",
		"postV1Customers": "PostV1CustomersBody",
	}
	for in, want := range cases {
		if got := componentName(in); got != want {
			t.Errorf("componentName(%s) = %s, want %s", in, got, want)
		}
	}
}

func TestNestFormField(t *testing.T) {
	root := &Schema{Type: "object", Properties: NewOMap()}
	nestFormField(root, "billing_address[city]", &Schema{Type: "string"}, false)
	nestFormField(root, "tags[]", &Schema{Type: "string"}, false)
	nestFormField(root, "extra_infos[products][][name]", &Schema{Type: "string"}, true)
	if root.Prop("billing_address").Prop("city").BaseType() != "string" {
		t.Error("bracket path must nest")
	}
	if tags := root.Prop("tags"); tags.BaseType() != "array" || tags.Items.BaseType() != "string" {
		t.Errorf("tags[] must be an array of strings, got %+v", tags)
	}
	products := root.Prop("extra_infos").Prop("products")
	if products.BaseType() != "array" || products.Items.Prop("name") == nil {
		t.Errorf("[] in the middle must be an array of objects, got %+v", products)
	}
	if len(root.Required) != 0 {
		t.Error("nested required fields must not mark the root as required")
	}
}

func TestBuildV1MiniPipeline(t *testing.T) {
	in := miniInputs(t)
	log := newReport()
	doc, err := buildV1(in, localeEN, log)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(doc.Info.Description, "hmac") {
		t.Error("legacy HMAC description must not be copied")
	}
	if !strings.Contains(doc.Info.Description, "## Release Notes") || doc.Info.Version != DocVersion {
		t.Error("info must carry the version and release notes")
	}
	for _, tag := range doc.Tags {
		if strings.Contains(tag.Description, "Translation missing") {
			t.Error("tag description not fixed")
		}
	}
	order := doc.Components.Schemas.Get2("Order")
	if order.Prop("third_party_discount").RefName() != "Money" {
		t.Errorf("Interger typo -> integer -> Money, got %+v", order.Prop("third_party_discount"))
	}
	if order.Prop("order_number").BaseType() != "integer" {
		t.Error("order_number must be integer")
	}
	if !order.Prop("created_at").IsNullable() || order.Prop("created_at").RefName() != "Timestamp" {
		t.Errorf("created_at: %+v", order.Prop("created_at"))
	}
	list := doc.Paths.Get2Path("/v1/orders").Get
	if list.Parameters[0].Ref != "#/components/parameters/page" {
		t.Error("page must reference the shared parameter")
	}
	if statuses := list.Parameters[2]; strings.Contains(statuses.Description, "<br>") || statuses.Example == nil {
		t.Errorf("statuses: %+v", statuses)
	}
	resp, _ := list.Responses.Get("200")
	if resp.(*Response).Headers == nil || !resp.(*Response).Headers.Has("X-Total") {
		t.Error("list response must declare pagination headers")
	}
	for _, code := range []string{"401", "403", "422", "429"} {
		if !list.Responses.Has(code) {
			t.Errorf("missing shared %s response", code)
		}
	}
	if list.Responses.Has("404") {
		t.Error("404 must only be declared for paths with parameters")
	}
	media, _ := resp.(*Response).Content.Get("application/json")
	if media.(*MediaType).Example == nil {
		t.Error("response example missing")
	}
	put := doc.Paths.Get2Path("/v1/orders/{order_id}").Put
	if put.RequestBody == nil || !put.RequestBody.Content.Has("application/json") {
		t.Fatal("formData must become a JSON request body")
	}
	body, _ := put.RequestBody.Content.Get("application/json")
	if body.(*MediaType).Schema.Prop("billing_address").Prop("city") == nil || body.(*MediaType).Example == nil {
		t.Error("nested body schema or example missing")
	}
	if !put.Responses.Has("404") {
		t.Error("404 must be declared for paths with parameters")
	}
	upload := doc.Paths.Get2Path("/v1/assets/pictures").Post
	if !upload.RequestBody.Content.Has("multipart/form-data") || upload.RequestBody.Content.Has("application/json") {
		t.Error("file uploads must stay multipart")
	}
	y, err := MarshalYAMLDocument(toTree(doc))
	if err != nil {
		t.Fatal(err)
	}
	if err := validateOpenAPI(y); err != nil {
		t.Fatalf("generated document is invalid: %v\n%s", err, y)
	}
	if err := scanOutput("mini", y); err != nil {
		t.Fatal(err)
	}
	y2, _ := MarshalYAMLDocument(toTree(doc))
	if string(y) != string(y2) {
		t.Error("marshalling must be deterministic")
	}
}

func TestValidateOpenAPIRejectsBrokenDocument(t *testing.T) {
	broken := []byte("openapi: 3.1.0\ninfo:\n  title: x\npaths:\n  /a:\n    get:\n      responses:\n        '200':\n          content:\n            application/json:\n              schema:\n                $ref: '#/components/schemas/Missing'\n")
	if err := validateOpenAPI(broken); err == nil {
		t.Fatal("expected a validation error for a missing info.version and a dangling $ref")
	}
}

func TestBuildV2MiniPipeline(t *testing.T) {
	in := miniInputs(t)
	log := newReport()
	doc, err := buildV2(in, localeEN, log)
	if err != nil {
		t.Fatal(err)
	}
	if countOperations(doc) != len(v2Endpoints) {
		t.Errorf("v2 has %d operations, want %d", countOperations(doc), len(v2Endpoints))
	}
	coupon := doc.Components.Schemas.Get2("CustomerV2")
	if coupon.Prop("vip_info") == nil {
		t.Error("CustomerV2 must document vip_info")
	}
	y, err := MarshalYAMLDocument(toTree(doc))
	if err != nil {
		t.Fatal(err)
	}
	if err := validateOpenAPI(y); err != nil {
		t.Fatalf("v2 document invalid: %v", err)
	}
	if err := scanOutput("v2", y); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(y), "`cod` (cash on delivery)") {
		t.Error("documented order status enums must appear in the v2 document")
	}
}
