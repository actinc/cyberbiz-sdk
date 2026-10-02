package cyberbiz

import (
	"errors"
	"net/http"
	"testing"
)

func TestCustomFieldsGoldenErrorShape(t *testing.T) {
	// The recorded GET /v1/custom_fields reply is the platform rejecting the
	// custom_field_type parameter; it must surface as an APIError even on 200.
	c := goldenServer(t)
	_, _, err := c.CustomFields.List(testCtx, CustomFieldOwnerOrder)
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("got %v", err)
	}
	if len(apiErr.Messages) != 1 || apiErr.Messages[0] != "custom_field_type 無效值" {
		t.Errorf("messages = %q", apiErr.Messages)
	}
}

func TestCustomFieldsGoldenUnknownNameIsNotFound(t *testing.T) {
	c := goldenServer(t)
	_, _, err := c.CustomFields.Get(testCtx, CustomFieldOwnerCustomer, "foo")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v", err)
	}
}

func TestCustomFieldTypesGoldenList(t *testing.T) {
	var out []CustomFieldType
	decodeGolden(t, "v1/GET_v1_custom_field_types.json", &out)
	if out == nil || len(out) != 0 {
		t.Errorf("types = %#v", out)
	}
	c := goldenServer(t)
	page, err := c.CustomFields.ListTypes(testCtx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 0 || page.Pagination.Page != 1 {
		t.Errorf("page = %+v", page)
	}
}

func TestCustomFieldSettingDecodesPostmanShape(t *testing.T) {
	c, _ := New("tok")
	var out []CustomFieldSetting
	body := `[{"component":"text","label":"會員編號","name":"member_no","required":true,"visible_when_create":true,"visible_when_update":false}]`
	if err := c.decode([]byte(body), &out); err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || out[0].Name != "member_no" || !out[0].Required || out[0].VisibleWhenUpdate {
		t.Errorf("fields = %+v", out)
	}
}

func TestCustomFieldsRequests(t *testing.T) {
	c, call := discountsSpyClient(t, 200, `[]`)
	if _, _, err := c.CustomFields.List(testCtx, CustomFieldOwnerCustomer); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "GET", "/v1/custom_fields")
	if call.Query.Get("custom_field_type") != "customer" {
		t.Errorf("query = %v", call.Query)
	}

	if _, err := c.CustomFields.ListTypes(testCtx, &ListOptions{Page: 3}); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, call, "GET", "/v1/custom_field_types")
	if call.Query.Get("page") != "3" {
		t.Errorf("query = %v", call.Query)
	}

	one, callOne := discountsSpyClient(t, 200, `{"component":"text","label":"L","name":"member no","required":false}`)
	f, _, err := one.CustomFields.Get(testCtx, CustomFieldOwnerCustomer, "member no")
	if err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, callOne, "GET", "/v1/custom_fields/member no")
	if callOne.Query.Get("custom_field_type") != "customer" || f.Label != "L" {
		t.Errorf("query = %v field = %+v", callOne.Query, f)
	}

	ty, callTy := discountsSpyClient(t, 201, `{"model":"Customer","name":"職業"}`)
	got, _, err := ty.CustomFields.CreateType(testCtx, &CustomFieldTypeCreateRequest{Model: CustomFieldTypeModelCustomer, Name: "職業"})
	if err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, callTy, "POST", "/v1/custom_field_types")
	discountsAssertJSONBody(t, callTy, `{"model":"Customer","name":"職業"}`)
	if got.Model != CustomFieldTypeModelCustomer {
		t.Errorf("type = %+v", got)
	}

	if _, _, err := ty.CustomFields.GetType(testCtx, 12); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, callTy, "GET", "/v1/custom_field_types/12")

	if _, err := ty.CustomFields.DeleteType(testCtx, 12); err != nil {
		t.Fatal(err)
	}
	discountsAssertCall(t, callTy, "DELETE", "/v1/custom_field_types/12")
}

func TestCustomFieldsGetTypeSetsID(t *testing.T) {
	c, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"model":"Customer","name":"x"}`))
	})
	ty, _, err := c.CustomFields.GetType(testCtx, 44)
	if err != nil {
		t.Fatal(err)
	}
	if ty.ID != 44 {
		t.Errorf("id = %d", ty.ID)
	}
}
