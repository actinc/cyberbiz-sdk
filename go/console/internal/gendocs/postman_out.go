package gendocs

import (
	"crypto/sha1"
	"fmt"
	"strings"
)

// buildPostman renders an OpenAPI document as a Postman v2.1 collection so
// the three artefacts (OpenAPI, Postman, SDK) describe the same requests.
// Requests are grouped in one folder per tag; every request carries the
// example body and one saved example response per documented 2xx status.
func buildPostman(doc *Document, name string, loc *locale) *OMap {
	col := NewOMap()
	info := NewOMap().
		Set("_postman_id", stableUUID(name)).
		Set("name", name).
		Set("description", postmanDescription(doc, loc)).
		Set("version", DocVersion).
		Set("schema", "https://schema.getpostman.com/json/collection/v2.1.0/collection.json")
	col.Set("info", info)
	col.Set("auth", NewOMap().Set("type", "bearer").Set("bearer", []any{
		NewOMap().Set("key", "token").Set("value", "{{CyberbizToken}}").Set("type", "string"),
	}))
	col.Set("variable", []any{
		NewOMap().Set("key", "baseUrl").Set("value", baseURL),
		NewOMap().Set("key", "CyberbizToken").Set("value", "").Set("description", loc.T(postmanTokenNote)),
	})
	folders := NewOMap()
	for _, t := range doc.Tags {
		folders.Set(t.Name, NewOMap().Set("name", t.Name).Set("description", t.Description).Set("item", []any{}))
	}
	for _, path := range doc.Paths.Keys() {
		item, _ := doc.Paths.Get(path)
		for _, m := range httpMethods {
			op := item.(*PathItem).Operation(m)
			if op == nil {
				continue
			}
			tag := "other"
			if len(op.Tags) > 0 {
				tag = op.Tags[0]
			}
			if !folders.Has(tag) {
				folders.Set(tag, NewOMap().Set("name", tag).Set("item", []any{}))
			}
			f, _ := folders.Get(tag)
			folder := f.(*OMap)
			items, _ := folder.Get("item")
			folder.Set("item", append(items.([]any), postmanItem(doc, strings.ToUpper(m), path, op)))
		}
	}
	var items []any
	for _, k := range folders.Keys() {
		f, _ := folders.Get(k)
		if list, _ := f.(*OMap).Get("item"); len(list.([]any)) > 0 {
			items = append(items, f)
		}
	}
	col.Set("item", items)
	return col
}

func postmanDescription(doc *Document, loc *locale) string {
	return doc.Info.Title + "\n\n" + loc.T(postmanGeneratedNote) + "\n\n" +
		loc.T(postmanEnvNote) + "\n\n" + releaseNotesMarkdown(loc)
}

const (
	postmanTokenNote     = "The Shop's API Token (never commit a real one)."
	postmanGeneratedNote = "Generated from the CYBERBIZ API documentation and checked against real API responses."
	postmanEnvNote       = "Set `CyberbizToken` in an environment (never commit *.postman_environment.json)."
)

func versionTag(doc *Document) string {
	if strings.Contains(doc.Info.Title, "v2") {
		return "v2"
	}
	return "v1"
}

// postmanItem renders one operation as a request with saved responses.
func postmanItem(doc *Document, method, path string, op *Operation) *OMap {
	req := postmanRequestObject(doc, method, path, op)
	item := NewOMap().
		Set("name", method+" "+path).
		Set("request", req)
	var responses []any
	for _, code := range op.Responses.Keys() {
		r, _ := op.Responses.Get(code)
		resp := resolveResponse(doc, r.(*Response))
		example, ok := responseExample(doc, resp)
		if !ok {
			continue
		}
		body, _ := EncodeJSONIndent(example, "  ")
		responses = append(responses, NewOMap().
			Set("name", code+" "+resp.Description).
			Set("originalRequest", req.Clone()).
			Set("status", statusText(code)).
			Set("code", atoiCode(code)).
			Set("_postman_previewlanguage", "json").
			Set("header", []any{NewOMap().Set("key", "Content-Type").Set("value", "application/json")}).
			Set("cookie", []any{}).
			Set("body", strings.TrimRight(string(body), "\n")))
	}
	item.Set("response", responses)
	return item
}

func postmanRequestObject(doc *Document, method, path string, op *Operation) *OMap {
	req := NewOMap().Set("method", method)
	headers := []any{NewOMap().Set("key", "Accept").Set("value", "application/json")}
	var query, variables []any
	var rawQuery []string
	for _, p := range op.Parameters {
		param := resolveParameter(doc, p)
		if param == nil {
			continue
		}
		val := exampleString(param.Example)
		switch param.In {
		case "query":
			query = append(query, NewOMap().Set("key", param.Name).Set("value", val).Set("description", param.Description))
			rawQuery = append(rawQuery, param.Name+"="+val)
		case "path":
			variables = append(variables, NewOMap().Set("key", param.Name).Set("value", val).Set("description", param.Description))
		}
	}
	pmPath := strings.ReplaceAll(strings.ReplaceAll(path, "{", ":"), "}", "")
	raw := "{{baseUrl}}" + pmPath
	if len(rawQuery) > 0 {
		raw += "?" + strings.Join(rawQuery, "&")
	}
	url := NewOMap().Set("raw", raw).Set("host", []any{"{{baseUrl}}"}).Set("path", splitPath(pmPath))
	if len(query) > 0 {
		url.Set("query", query)
	}
	if len(variables) > 0 {
		url.Set("variable", variables)
	}
	if op.RequestBody != nil {
		mt, media := firstMedia(op.RequestBody.Content)
		if media != nil && media.Example != nil {
			body, _ := EncodeJSONIndent(media.Example, "  ")
			if strings.HasPrefix(mt, "multipart") {
				req.Set("body", multipartBody(media.Example))
			} else {
				headers = append(headers, NewOMap().Set("key", "Content-Type").Set("value", "application/json"))
				req.Set("body", NewOMap().
					Set("mode", "raw").
					Set("raw", strings.TrimRight(string(body), "\n")).
					Set("options", NewOMap().Set("raw", NewOMap().Set("language", "json"))))
			}
		}
	}
	req.Set("header", headers)
	req.Set("url", url)
	req.Set("description", op.Description)
	return req
}

func multipartBody(example any) *OMap {
	var fields []any
	if om, ok := example.(*OMap); ok {
		for _, k := range om.Keys() {
			v, _ := om.Get(k)
			kind := "text"
			if s, ok := v.(string); ok && strings.Contains(s, "binary") {
				kind = "file"
			}
			fields = append(fields, NewOMap().Set("key", k).Set("value", exampleString(v)).Set("type", kind))
		}
	}
	return NewOMap().Set("mode", "formdata").Set("formdata", fields)
}

func firstMedia(content *OMap) (string, *MediaType) {
	if content == nil {
		return "", nil
	}
	for _, k := range content.Keys() {
		v, _ := content.Get(k)
		return k, v.(*MediaType)
	}
	return "", nil
}

func splitPath(p string) []any {
	var out []any
	for _, s := range strings.Split(strings.Trim(p, "/"), "/") {
		out = append(out, s)
	}
	return out
}

func exampleString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case []any:
		parts := make([]string, len(t))
		for i, e := range t {
			parts[i] = exampleString(e)
		}
		return strings.Join(parts, ",")
	}
	return fmt.Sprint(v)
}

// resolveParameter follows a $ref to components.parameters.
func resolveParameter(doc *Document, p *Parameter) *Parameter {
	if p.Ref == "" {
		return p
	}
	name := strings.TrimPrefix(p.Ref, "#/components/parameters/")
	v, ok := doc.Components.Parameters.Get(name)
	if !ok {
		return nil
	}
	return v.(*Parameter)
}

// resolveResponse follows a $ref to components.responses.
func resolveResponse(doc *Document, r *Response) *Response {
	if r.Ref == "" {
		return r
	}
	name := strings.TrimPrefix(r.Ref, "#/components/responses/")
	v, ok := doc.Components.Responses.Get(name)
	if !ok {
		return r
	}
	return v.(*Response)
}

// responseExample returns the sample body of a response (inline example, or
// the first named example, following $refs into components.examples).
func responseExample(doc *Document, r *Response) (any, bool) {
	_, media := firstMedia(r.Content)
	if media == nil {
		return nil, false
	}
	if media.Example != nil {
		return media.Example, true
	}
	if media.Examples != nil {
		for _, k := range media.Examples.Keys() {
			v, _ := media.Examples.Get(k)
			ex := v.(*Example)
			if ex.Ref != "" {
				name := strings.TrimPrefix(ex.Ref, "#/components/examples/")
				if c, ok := doc.Components.Examples.Get(name); ok {
					ex = c.(*Example)
				}
			}
			return ex.Value, true
		}
	}
	return nil, false
}

func statusText(code string) string {
	switch code {
	case "200":
		return "OK"
	case "201":
		return "Created"
	case "202":
		return "Accepted"
	case "204":
		return "No Content"
	case "206":
		return "Partial Content"
	case "401":
		return "Unauthorized"
	case "403":
		return "Forbidden"
	case "404":
		return "Not Found"
	case "422":
		return "Unprocessable Entity"
	case "429":
		return "Too Many Requests"
	}
	return code
}

// stableUUID derives a fixed UUID-shaped id from a name so re-runs do not
// change the collection id.
func stableUUID(name string) string {
	sum := sha1.Sum([]byte("cyberbiz-sdk-go/" + name))
	h := fmt.Sprintf("%x", sum[:16])
	return fmt.Sprintf("%s-%s-4%s-8%s-%s", h[0:8], h[8:12], h[13:16], h[17:20], h[20:32])
}
