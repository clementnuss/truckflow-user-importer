package truckflow

type Tiers struct {
	Type         string `json:"TiersType"`
	Code         string `json:"TiersCode"`
	Label        string `json:"Label"`
	Active       bool   `json:"IsActive"`
	Address      string `json:"Address1"`
	ZIPCode      string `json:"Address2"`
	City         string `json:"Address3"`
	Telephone    string `json:"Address4"`
	ProductCodes string `json:"ProductCodes"`
}

type Import struct {
	Version string  `json:"version"`
	Items   []Tiers `json:"Items"`
}
