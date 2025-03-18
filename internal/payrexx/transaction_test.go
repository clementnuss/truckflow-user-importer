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
          "name": "Badge Ajoverts",
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
          "value": "JU12345, Ju54321"
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

  tr := formData.Transaction 
  tr.SanitizeFields()

  assert.Equal(t, []string{"JU12345", "JU54321"}, tr.Plates)

	assert.NoError(t, err)
	assert.Equal(t, "some@email.ch", tr.Contact.Email)
	assert.Equal(t, 2, tr.Invoice.Products[0].Quantity)
}

func TestMultiPlate(t *testing.T) {
	sampleTransaction := `
{
  "transaction": {
    "uuid": "b63112e9",
    "time": "2025-01-27 22:08:58",
    "status": "cancelled",
    "invoice": {
      "products": [
        {
          "name": "Badge Ajoverts",
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
          "value": "JU12345, Ju2,  ju3,; ju4!,"
        }
      ]
    }
  }
}

`

	formData := struct {
		Transaction payrexx.Transaction `json:"transaction"`
	}{}
	err := json.Unmarshal([]byte(sampleTransaction), &formData)

  tr := formData.Transaction 
  tr.SanitizeFields()

  assert.Equal(t, []string{"JU12345", "JU2.JU3.JU4"}, tr.Plates)

	assert.NoError(t, err)
}

func TestLongLastPlate(t *testing.T) {
	sampleTransaction := `
{
  "transaction": {
    "uuid": "b63112e9",
    "time": "2025-01-27 22:08:58",
    "status": "cancelled",
    "invoice": {
      "products": [
        {
          "name": "Badge Ajoverts",
          "description": "SomeProductDescription",
          "price": 2000,
          "quantity": 1,
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
          "value": "JU12345,JU12345,JU12345,JU12345"
        }
      ]
    }
  }
}

`

	formData := struct {
		Transaction payrexx.Transaction `json:"transaction"`
	}{}
	err := json.Unmarshal([]byte(sampleTransaction), &formData)

  tr := formData.Transaction 
  tr.SanitizeFields()

  assert.Equal(t, []string{"JU12345.JU12345.JU12345.JU1..."}, tr.Plates)

	assert.NoError(t, err)
}
