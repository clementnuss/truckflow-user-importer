package payrexx

import (
	"encoding/csv"
	"io"
	"os"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/spkg/bom"
)

// Convert the CSV string as internal date
func (date *DateTime) UnmarshalCSV(csv string) (err error) {
	date.Time, err = time.Parse("2006-01-02 15:04:05", csv)
	return err
}

type CSVTransaction struct {
	Id           string   `csv:"#"`
	FirstName    string   `csv:"First name"`
	LastName     string   `csv:"Last Name"`
	Date         DateTime `csv:"Date and time"`
	Status       string   `csv:"Status"`
	Number       int      `csv:"Number"`
	StreeAndNo   string   `csv:"Street & No."`
	ZIPCode      string   `csv:"Zip code"`
	City         string   `csv:"City"`
	Country      string   `csv:"Country"`
	Telephone    string   `csv:"Telephone"`
	Email        string   `csv:"Email address"`
	Company      string   `csv:"entreprise"`
	PlateNumbers string   `csv:"numeros_de_plaques"`
	ClientNumber string   `csv:"numero_client_optionnel"`
}

func ParseCSV(file *os.File) ([]CSVTransaction, error) {
	transactions := []CSVTransaction{}
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		r := csv.NewReader(bom.NewReader(in))
		r.Comma = ';'
		return r
	})

	err := gocsv.UnmarshalFile(file, &transactions)
	if err != nil {
		return nil, err
	}

	return transactions, nil
}
