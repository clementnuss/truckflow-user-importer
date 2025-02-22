package payrexx

import (
	"log/slog"
	"regexp"
	"strings"
	"time"
)

type Transaction struct {
	Uuid    string   `json:"uuid"`
	Time    DateTime `json:"time"`
	Status  string   `json:"status"`
	Invoice Invoice  `json:"invoice"`
	Contact Contact  `json:"contact"`
	Plates  []string
}

type DateTime struct {
	time.Time
}

func (date *DateTime) UnmarshalJSON(data []byte) (err error) {
	d := string(data)
	d = strings.Trim(d, "\"")
	date.Time, err = time.Parse("2006-01-02 15:04:05", d)
	return err
}

type Invoice struct {
	Products     []Product     `json:"products"`
	CustomFields []CustomField `json:"custom_fields"`
}

type Product struct {
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
}

type CustomField struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ClientType int

const (
	Individual = iota + 1
	Company
)

type Contact struct {
	Title       string `json:"title"`
	FirstName   string `json:"firstname"`
	LastName    string `json:"lastname"`
	StreetAndNo string `json:"street"`
	ZIPCode     string `json:"zip"`
	City        string `json:"place"`
	Country     string `json:"country"`
	Telephone   string `json:"phone"`
	Email       string `json:"email"`
	Company     string `json:"company"`
	ClientType
}

func (tr *Transaction) SanitizeFields() {
	tr.Contact.FirstName = strings.TrimSpace(tr.Contact.FirstName)
	tr.Contact.LastName = strings.TrimSpace(tr.Contact.LastName)
	tr.Contact.StreetAndNo = strings.TrimSpace(tr.Contact.StreetAndNo)
	tr.Contact.ZIPCode = strings.TrimSpace(tr.Contact.ZIPCode)
	tr.Contact.City = strings.TrimSpace(tr.Contact.City)
	tr.Contact.Country = strings.TrimSpace(tr.Contact.Country)
	tr.Contact.Telephone = strings.TrimSpace(tr.Contact.Telephone)
	tr.Contact.Email = strings.TrimSpace(tr.Contact.Email)
	tr.Contact.Company = strings.TrimSpace(tr.Contact.Company)

	platesQty := tr.Invoice.Products[0].Quantity
	var platesStr string
	for _, f := range tr.Invoice.CustomFields {
		switch {
		case strings.Contains(f.Name, "Numéros de plaques"):
			platesStr = strings.ToUpper(strings.TrimSpace(f.Value))

		case strings.Contains(f.Name, "Entreprise"):
			tr.Contact.Company = strings.TrimSpace(f.Value)

		case f.Name == "Type de client:":
			switch f.Value {
			case "entreprise":
				tr.Contact.ClientType = Company
			case "particulier":
				tr.Contact.ClientType = Individual
			default:
				slog.Error("unknown client type", "field_value", f.Value)
			}
		}
	}

	reg, _ := regexp.Compile(`\s+`) // compile
	platesStr = reg.ReplaceAllString(platesStr, "")

	plates := strings.Split(platesStr, ",")
	for i := range platesQty {
		if i < len(plates) {
			tr.Plates = append(tr.Plates, plates[i])
		} else {
			tr.Plates = append(tr.Plates, "N/D")
		}
	}
}
