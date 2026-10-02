package gendocs

import (
	"fmt"
	"strings"
)

const (
	baseURL       = "https://app-store-api.cyberbiz.io"
	rateLimit     = 5
	maxPerPage    = 50
	defaultPage   = 1
	defaultOffset = 0
)

// paginationHeaders are the headers every paginated list response carries.
var paginationHeaders = []struct{ Name, Description string }{
	{"X-Page", "The page that was returned (echoes `page`)."},
	{"X-Per-Page", "Records per page (echoes `per_page`)."},
	{"X-Offset", "Records skipped before paging (echoes `offset`)."},
	{"X-Total", "Total number of records matching the query, regardless of paging."},
	{"X-Total-Pages", "Total pages: CEILING((X-Total - X-Offset) / X-Per-Page)."},
	{"X-Next-Page", "Number of the next page; at most X-Total-Pages. Empty on the last page."},
	{"X-Prev-Page", "Number of the previous page; empty on the first page."},
}

// errorStatuses are the shared error responses declared on every operation.
var errorStatuses = []struct {
	Code, Name, Description, ExampleName string
	Example                              any
}{
	{"401", "Unauthorized", "The token is missing, invalid, expired, lacks the scope, or the feature is not licensed for this shop (CYBERBIZ answers 401 for both).", "unauthorized",
		obj("error", []any{"無權使用該 API"})},
	{"403", "Forbidden", "The action is not allowed in the resource's current state, or a required plugin is missing. The body is `{\"error\": \"...\"}` or `{\"messages\": \"...\"}`.", "forbidden",
		obj("error", "Must have pos_shop_coupon plugin")},
	{"404", "NotFound", "The resource does not exist. Note that some lookups answer `200` with a `null` body or `206` with `[]` instead of `404`.", "notFound",
		obj("error", []any{"無此資源"})},
	{"422", "UnprocessableEntity", "Validation failed, or the feature is not enabled for the shop. The body is `{\"error\": [...]}` or `{\"message\": \"...\"}`.", "unprocessableEntity",
		obj("error", []any{"custom_field_type 無效值"})},
	{"429", "TooManyRequests", "More than 5 requests per second were sent. Back off and retry; the body shape is not documented.", "tooManyRequests",
		obj("error", []any{"Rate limit exceeded"})},
}

// obj builds a one-key sample object.
func obj(k string, v any) *OMap {
	return NewOMap().Set(k, v)
}

// addSharedComponents installs the schemas, parameters, headers, responses,
// examples and security scheme every generated document shares.
func addSharedComponents(doc *Document, log *report) {
	c := doc.Components
	if c == nil {
		c = &Components{}
		doc.Components = c
	}
	if c.Schemas == nil {
		c.Schemas = NewOMap()
	}
	for name, s := range sharedSchemas() {
		c.Schemas.Set(name, s)
	}
	c.Schemas.SortKeys()

	c.SecuritySchemes = NewOMap().Set("bearerAuth", &SecurityScheme{
		Type:         "http",
		Scheme:       "bearer",
		BearerFormat: "JWT",
		Description:  bearerNote,
	})
	doc.Security = []map[string][]string{{"bearerAuth": {}}}
	doc.Servers = []*Server{{URL: baseURL, Description: serverNote}}
	log.count("legacy-auth")

	c.Parameters = NewOMap().
		Set("page", &Parameter{Name: "page", In: "query", Description: pageNote, Schema: &Schema{Type: "integer", Default: defaultPage, Minimum: f64(1)}, Example: 1}).
		Set("per_page", &Parameter{Name: "per_page", In: "query", Description: fmt.Sprintf(perPageNote, maxPerPage), Schema: &Schema{Type: "integer", Default: maxPerPage, Minimum: f64(1), Maximum: f64(maxPerPage)}, Example: maxPerPage}).
		Set("offset", &Parameter{Name: "offset", In: "query", Description: offsetNote, Schema: &Schema{Type: "integer", Default: defaultOffset, Minimum: f64(0)}, Example: 0})

	c.Headers = NewOMap()
	for _, h := range paginationHeaders {
		c.Headers.Set(h.Name, &Header{Description: h.Description, Schema: &Schema{Type: "string"}, Example: headerExample(h.Name)})
	}

	c.Examples = NewOMap()
	c.Responses = NewOMap()
	for _, e := range errorStatuses {
		c.Examples.Set(e.ExampleName, &Example{Summary: e.Name, Value: e.Example})
		media := &MediaType{Schema: &Schema{Ref: schemaRef("Error")}, Examples: NewOMap().Set(e.ExampleName, &Example{Ref: "#/components/examples/" + e.ExampleName})}
		c.Responses.Set(e.Name, &Response{Description: e.Description, Content: NewOMap().Set("application/json", media)})
	}
}

func headerExample(name string) string {
	switch name {
	case "X-Page", "X-Next-Page":
		return map[string]string{"X-Page": "1", "X-Next-Page": "2"}[name]
	case "X-Per-Page":
		return "50"
	case "X-Total":
		return "92"
	case "X-Total-Pages":
		return "2"
	}
	return ""
}

func f64(v float64) *float64 { return &v }

// sharedSchemas are the value types the whole API shares.
func sharedSchemas() map[string]*Schema {
	return map[string]*Schema{
		"Timestamp": {
			Type:        "string",
			Description: "A local date-time in Asia/Taipei written as `YYYY-MM-DD HH:MM:SS` with no zone designator (occasionally `YYYY-MM-DD HH:MM:SS +0800` or `YYYY-MM-DD HH:MM+0800` on v2 endpoints). Parse it in Asia/Taipei; never as UTC.",
			Pattern:     `^\d{4}-\d{2}-\d{2} \d{2}:\d{2}(:\d{2})?( ?\+0800)?$`,
			Examples:    []any{"2026-09-01 10:00:00"},
		},
		"Date": {
			Type:        "string",
			Description: "A calendar date written as `YYYY-MM-DD`.",
			Pattern:     `^\d{4}-\d{2}-\d{2}$`,
			Examples:    []any{"1990-01-01"},
		},
		"Money": {
			Type:        "number",
			Description: "A currency amount in the shop's currency (TWD) serialised as a JSON number, usually with one decimal (`9999.0`). Integers appear for some fields; treat every money field as a decimal.",
			Examples:    []any{9999.0},
		},
		"Error": {
			Type:        "object",
			Description: "The error envelope. CYBERBIZ uses four shapes: `{\"error\": [\"...\"]}`, `{\"error\": \"...\"}`, `{\"message\": \"...\"}` and `{\"messages\": \"...\"}`; messages are in Traditional Chinese or English.",
			Properties: NewOMap().
				Set("error", &Schema{Description: "One or more error messages.", AnyOf: []*Schema{{Type: "array", Items: &Schema{Type: "string"}}, {Type: "string"}}}).
				Set("message", &Schema{Type: "string", Description: "A single error message (422 on some endpoints)."}).
				Set("messages", &Schema{Type: "string", Description: "A single error message (403 on some endpoints)."}),
		},
	}
}

// applySharedParameters replaces inline page / per_page / offset query
// parameters with references and adds pagination headers to list responses.
func applySharedParameters(doc *Document, log *report) {
	for _, path := range doc.Paths.Keys() {
		item, _ := doc.Paths.Get(path)
		for _, m := range httpMethods {
			op := item.(*PathItem).Operation(m)
			if op == nil {
				continue
			}
			paginated := false
			for i, p := range op.Parameters {
				if p.In != "query" {
					continue
				}
				switch p.Name {
				case "page", "per_page", "offset":
					op.Parameters[i] = &Parameter{Ref: "#/components/parameters/" + p.Name}
					paginated = true
					log.count("pagination-params")
				}
			}
			if paginated {
				addPaginationHeaders(op)
			}
		}
	}
}

func addPaginationHeaders(op *Operation) {
	for _, code := range op.Responses.Keys() {
		if !strings.HasPrefix(code, "2") {
			continue
		}
		r, _ := op.Responses.Get(code)
		resp := r.(*Response)
		if resp.Headers == nil {
			resp.Headers = NewOMap()
		}
		for _, h := range paginationHeaders {
			resp.Headers.Set(h.Name, &Header{Ref: "#/components/headers/" + h.Name})
		}
	}
}

// applyErrorResponses adds the shared error responses to every operation.
// 404 is only declared for paths that address one resource.
func applyErrorResponses(doc *Document, golden *goldenSet, log *report) {
	for _, path := range doc.Paths.Keys() {
		item, _ := doc.Paths.Get(path)
		for _, m := range httpMethods {
			op := item.(*PathItem).Operation(m)
			if op == nil {
				continue
			}
			recorded := golden.errorsFor(strings.ToUpper(m), path)
			for _, e := range errorStatuses {
				if e.Code == "404" && !strings.Contains(path, "{") {
					continue
				}
				if op.Responses.Has(e.Code) {
					continue
				}
				if gf, ok := recorded[atoiCode(e.Code)]; ok {
					op.Responses.Set(e.Code, recordedErrorResponse(e.Description, gf))
				} else {
					op.Responses.Set(e.Code, &Response{Ref: "#/components/responses/" + e.Name})
				}
				log.count("error-responses")
			}
		}
	}
}

// recordedErrorResponse documents an error with the message actually
// recorded for this endpoint (the messages carry no personal data).
func recordedErrorResponse(desc string, gf *goldenFile) *Response {
	media := &MediaType{Schema: &Schema{Ref: schemaRef("Error")}, Example: gf.Body}
	return &Response{Description: desc + recordedNote, Content: NewOMap().Set("application/json", media)}
}

// Fragments appended to composed descriptions; localized through
// locale.fragments because the whole string is never a table key.
const (
	recordedNote   = " (message recorded for this endpoint)"
	nullNote       = " Answers `null` when the resource does not exist."
	includeNote    = " Only present when `include_params` requests it."
	scopeTemplate  = "Scope: `%s`."
	serverNote     = "Production (the same host for every shop and every API version)."
	bearerNote     = "The Shop's API Token, sent as `Authorization: Bearer <token>`. One token authorises one Shop and carries the scopes the app was granted."
	pageNote       = "Page number, 1-based."
	perPageNote    = "Records per page; default and maximum %d."
	offsetNote     = "Records to skip before paging. The first returned record is `offset + (page - 1) * per_page + 1`."
	successNote    = "Success"
	unresolvedNote = "unresolved reference "
)

func atoiCode(code string) int {
	n := 0
	for _, r := range code {
		n = n*10 + int(r-'0')
	}
	return n
}

// infoDescription is the markdown shown at the top of each OpenAPI file.
func infoDescription(intro string, loc *locale) string {
	return strings.TrimSpace(loc.T(intro)) + "\n\n" + fmt.Sprintf(loc.T(conventionsTemplate), baseURL, rateLimit, maxPerPage) + "\n\n" + releaseNotesMarkdown(loc)
}

// conventionsTemplate is formatted with baseURL, rateLimit and maxPerPage.
const conventionsTemplate = `## Conventions

- **Base URL**: ` + "`%s`" + ` for every shop and every API version. Versions are path prefixes (` + "`/v1`, `/v2`" + `); the app endpoints ` + "`/shop`" + ` and ` + "`/settings`" + ` have no prefix.
- **Authentication**: ` + "`Authorization: Bearer <API Token>`" + ` on every request. The token is a JWT issued to the app for one Shop and carries its scopes (` + "`read_orders`, `write_products`" + `, ...). ` + "`read_*`" + ` scopes grant the GET operations of a resource group; ` + "`write_*`" + ` scopes grant POST/PUT/DELETE.
- **Rate limit**: at most %d requests per second per token. Expect ` + "`429`" + ` beyond that.
- **Pagination**: list endpoints take ` + "`page`" + ` (1-based), ` + "`per_page`" + ` (default and maximum %d) and ` + "`offset`" + `; the returned slice starts at record ` + "`offset + (page - 1) * per_page + 1`" + `. The response carries ` + "`X-Page`, `X-Per-Page`, `X-Offset`, `X-Total`, `X-Total-Pages`, `X-Next-Page` and `X-Prev-Page`" + `.
- **Timestamps** are local Asia/Taipei strings ` + "`YYYY-MM-DD HH:MM:SS`" + ` without a zone (schema ` + "`Timestamp`" + `); dates are ` + "`YYYY-MM-DD`" + ` (schema ` + "`Date`" + `).
- **Money** fields are JSON numbers (floats such as ` + "`9999.0`" + `) even where the swagger says integer (schema ` + "`Money`" + `).
- **Nullability**: no field is guaranteed non-null. Fields observed as ` + "`null`" + ` in live responses are marked nullable; treat any other field as possibly null too.
- **Request bodies** are sent as JSON. The original swagger documents most writes as form data; the platform accepts ` + "`application/json`" + `.
- **Errors** use one of ` + "`{\"error\": [\"...\"]}`, `{\"error\": \"...\"}`, `{\"message\": \"...\"}`, `{\"messages\": \"...\"}`" + `. ` + "`401`" + ` means the token is invalid **or** the feature is not licensed for the shop. Some lookups answer ` + "`200`" + ` with a bare ` + "`null`" + ` or ` + "`206`" + ` with ` + "`[]`" + ` when nothing matches.
- **Samples** in this document are synthetic: their shape comes from recorded responses, their values are fake.`
