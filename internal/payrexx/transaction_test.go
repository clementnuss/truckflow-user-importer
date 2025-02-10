package payrexx_test

import (
	"encoding/json"
	"testing"

	"github.com/clementnuss/truckflow-user-importer/internal/payrexx"
	"github.com/stretchr/testify/assert"
)

func TestCorrectWebhookParsing(t *testing.T) {
	sampleTransaction := `
{
  "transaction": {
    "uuid": "b63112e9",
    "time": "2025-01-27 22:08:58",
    "status": "cancelled",
    "invoice": {
      "products": [
        {
          "name": "SomeProduct",
          "description": "SomeProductDescription",
          "price": 2000,
          "quantity": 2,
          "sku": null,
          "vatRate": null
        }
      ],
      "custom_fields": [
        {
          "type": "text",
          "name": "Numéro client (optionnel)",
          "value": "00014"
        },
        {
          "type": "text",
          "name": "Numéros de plaques (séparés par des virgules)",
          "value": "JU12345"
        }
      ]
    },
    "contact": {
      "title": "mister",
      "firstname": "Foo",
      "lastname": "Bar",
      "company": "",
      "street": "NotFar 2",
      "zip": "12345",
      "place": "Away",
      "country": "Suisse",
      "countryISO": "CH",
      "phone": "+41123456789",
      "email": "some@email.ch"
    }
  }
}

`

	formData := struct {
		Transaction payrexx.Transaction `json:"transaction"`
	}{}
	err := json.Unmarshal([]byte(sampleTransaction), &formData)

	assert.NoError(t, err)
	assert.Equal(t, "some@email.ch", formData.Transaction.Contact.Email)
	assert.Equal(t, 2, formData.Transaction.Invoice.Products[0].Quantity)
}
