package catalog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleSpec = `
openapi: 3.0.3
info: {title: t, version: "1"}
tags:
  - name: Orders
    description: Order operations
paths:
  /v1/orders/{id}:
    parameters:
      - name: id
        in: path
        required: true
        schema: {type: integer}
        example: 123
    get:
      operationId: getOrder
      tags: [Orders]
      summary: Get an order
      parameters:
        - $ref: '#/components/parameters/Include'
      responses:
        "200":
          description: OK
          content:
            application/json:
              example: {"id": 123}
    put:
      tags: [Orders]
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Order'
            examples:
              basic:
                value: {"order": {"note": "hi"}}
      responses:
        "422":
          description: Validation failed
components:
  parameters:
    Include:
      name: include
      in: query
      schema:
        type: string
        enum: [items, customer]
  schemas:
    Order:
      type: object
      properties:
        note: {type: string}
        parent:
          $ref: '#/components/schemas/Order'
`

func TestParse(t *testing.T) {
	cat, err := Parse([]byte(sampleSpec))
	require.NoError(t, err)
	require.Len(t, cat.Operations, 2)
	require.Len(t, cat.Tags, 1)
	assert.Equal(t, "Order operations", cat.Tags[0].Description)

	get := cat.Operations[0]
	assert.Equal(t, "getOrder", get.ID)
	assert.Equal(t, "GET", get.Method)
	assert.Equal(t, "/v1/orders/{id}", get.Path)
	assert.Equal(t, "Orders", get.Tag)
	require.Len(t, get.PathParams, 1)
	assert.Equal(t, "integer", get.PathParams[0].Type)
	assert.Equal(t, 123, get.PathParams[0].Example)
	require.Len(t, get.QueryParams, 1)
	assert.Equal(t, "include", get.QueryParams[0].Name)
	assert.Equal(t, []any{"items", "customer"}, get.QueryParams[0].Enum)
	require.Len(t, get.Responses, 1)
	assert.Equal(t, "200", get.Responses[0].Status)
	assert.Equal(t, map[string]any{"id": 123}, get.Responses[0].Example)
	assert.Nil(t, get.RequestBody)

	put := cat.Operations[1]
	assert.Equal(t, "put_v1_orders_id", put.ID)
	require.NotNil(t, put.RequestBody)
	assert.Equal(t, "application/json", put.RequestBody.ContentType)
	assert.Equal(t, map[string]any{"order": map[string]any{"note": "hi"}}, put.RequestBody.Example)
	schema := put.RequestBody.Schema.(map[string]any)
	assert.Equal(t, "object", schema["type"])
}

func TestLoadEmbedded(t *testing.T) {
	cat, err := Load()
	require.NoError(t, err)
	assert.NotEmpty(t, cat.Operations)
	var found bool
	for _, op := range cat.Operations {
		if op.Method == "GET" && op.Path == "/shop" {
			found = true
		}
	}
	assert.True(t, found, "embedded spec must contain GET /shop")
}
