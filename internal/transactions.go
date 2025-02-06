package internal

import (
	"encoding/csv"
	"io"
	"os"
	"time"

	"github.com/gocarina/gocsv"
	"github.com/spkg/bom"
)

type DateTime struct {
	time.Time
}

// Convert the internal date as CSV string
func (date *DateTime) MarshalCSV() (string, error) {
	return date.Time.Format("20060201"), nil
}

// You could also use the standard Stringer interface
func (date *DateTime) String() string {
	return date.String() // Redundant, just for example
}

// Convert the CSV string as internal date
func (date *DateTime) UnmarshalCSV(csv string) (err error) {
	date.Time, err = time.Parse("2006-01-02 15:04:05", csv)
	return err
}

type Transaction struct {
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

func ParseCSV(file *os.File) ([]Transaction, error) {
	transactions := []Transaction{}
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
