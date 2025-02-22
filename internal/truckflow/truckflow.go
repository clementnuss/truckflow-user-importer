package truckflow

type Tiers struct {
	Type          string `json:"TiersType"`
	Code          string `json:"TiersCode"`
	Label         string `json:"Label"`
	Active        bool   `json:"IsActive"`
	Address       string `json:"Address1"`
	ZIPCode       string `json:"Address2"`
	City          string `json:"Address3"`
	Telephone     string `json:"Address4"`
	Email         string `json:"Address5"`
	ContactPerson string `json:"Address6"`
	ProductCodes  string `json:"ProductCodes"`
}

type TiersImport struct {
	Version string  `json:"version"`
	Culture string  `json:"culture"`
	Items   []Tiers `json:"Items"`
}

type Pass struct {
	ParkCode    string `json:"ParkCode"`
	Label       string `json:"Label"`
	FlowType    string `json:"FlowType"`
	Plate       string `json:"Plate"`
	CompanyCode string `json:"CompanyCode"`
	TiersCode   string `json:"TiersCode"`
	ProductCode string `json:"ProductCode"`
}

type PassImport struct {
	Version string `json:"version"`
	Culture string `json:"culture"`
	Items   []Pass `json:"Items"`
}

func NewPass() *Pass {
	p := Pass{}
	p.FlowType = "Réception"
	p.ProductCode = "Dechets verts"

	return &p
}
