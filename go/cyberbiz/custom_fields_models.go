package cyberbiz

// CustomFieldOwner is the resource a custom field definition belongs to; it
// is the custom_field_type query parameter of /v1/custom_fields. The test
// shop accepts only customer; order and product return 422.
type CustomFieldOwner string

// Known CustomFieldOwner values (custom_field_type of /v1/custom_fields).
const (
	CustomFieldOwnerCustomer CustomFieldOwner = "customer" // fields on customer profiles
	CustomFieldOwnerOrder    CustomFieldOwner = "order"    // fields on orders
	CustomFieldOwnerProduct  CustomFieldOwner = "product"  // fields on products
)

// CustomFieldSetting is a custom field definition from the shop settings
// (GET /v1/custom_fields). No successful Golden File exists for it (the test
// shop rejects the query); the shape follows the swagger and Postman
// examples.
type CustomFieldSetting struct {
	// Component is the input kind of the field, e.g. a text box or select.
	Component string `json:"component"`
	// Label is the display name; Name is the key used in custom field values.
	Label    string `json:"label"`
	Name     string `json:"name"`
	Required bool   `json:"required"`
	// VisibleWhenCreate and VisibleWhenUpdate control when the field is shown
	// or editable.
	VisibleWhenCreate bool `json:"visible_when_create"`
	VisibleWhenUpdate bool `json:"visible_when_update"`
}

// CustomFieldTypeModel is the resource model a custom field type attaches to.
type CustomFieldTypeModel string

// CustomFieldTypeModelCustomer is the only model the platform accepts.
const CustomFieldTypeModelCustomer CustomFieldTypeModel = "Customer"

// CustomFieldType is a custom field type (GET /v1/custom_field_types). The
// detail response omits id; GetType fills it in.
type CustomFieldType struct {
	ID    int64                `json:"id"`
	Model CustomFieldTypeModel `json:"model"`
	Name  string               `json:"name"`
}
