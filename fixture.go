package tachyon

import (
	"bytes"
	"database/sql"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"sort"
)

const FixturePath = "testdata/fixtures"

// Fixture represents a test Fixture.
type Fixture struct {
	Table   string            `yaml:"table"`
	Fields  []string          `yaml:"fields"`
	Records map[string]Record `yaml:"data"`
}

func NewFixture(name string) (*Fixture, error) {
	base, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	raw, err := ioutil.ReadFile(path.Join(base, FixturePath, name+".yml"))
	if err != nil {
		return nil, err
	}

	f := &Fixture{}
	err = yaml.Unmarshal(raw, &f)
	if err != nil {
		return nil, err
	}

	f.sortFields()
	return f, nil
}

// Load inserts f's data into db inside of a transaction. If an error is
// encountered, the transaction is rolled back.
func (f *Fixture) Load(db *sql.DB) error {
	return f.loadRecords(f.recordList(), db)
}

func (f *Fixture) LoadTag(tag string, db *sql.DB) error {
	record, ok := f.Records[tag]
	if !ok {
		return fmt.Errorf("Failed to find tag %v", tag)
	}

	return f.loadRecords(&[]Record{record}, db)
}

// LoadTx inserts f's data within the transaction, but does not commit or
// rollback. The calling code must commit it manually.
func (f *Fixture) LoadTx(tx *sql.Tx) error {
	return f.loadRecordsTx(f.recordList(), tx)
}

// FindTag finds the record in f that has the name "tag". The second return
// value will be false if the record cannot be found, and true otherwise.
func (f *Fixture) FindTag(tag string) (Record, bool) {
	r, ok := f.Records[tag]
	return r, ok
}

// recordList Returns a flat list of all records in this fixture.
func (f *Fixture) recordList() *[]Record {
	records := make([]Record, 0, len(f.Records))
	for _, r := range f.Records {
		records = append(records, r)
	}
	return &records
}

func (f *Fixture) loadRecords(records *[]Record, db *sql.DB) error {
	// TODO: Should use db.BeginTx as db.Begin is deprecated.
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	err = f.loadRecordsTx(records, tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (f *Fixture) loadRecordsTx(records *[]Record, tx *sql.Tx) error {
	stmt, err := tx.Prepare(f.insertStr())
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, record := range *records {
		_, err := stmt.Exec(record.orderedData(f)...)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Fixture) sortFields() {
	sort.SliceStable(f.Fields, func(i, j int) bool {
		return f.Fields[i] < f.Fields[j]
	})
}

func (f *Fixture) insertStr() string {
	bufH := bytes.NewBuffer([]byte{})
	bufT := bytes.NewBuffer([]byte{})

	fmt.Fprintf(bufH, "INSERT INTO %v (", f.Table)
	fmt.Fprint(bufT, " VALUES (")

	for i, field := range f.Fields {
		fmt.Fprint(bufH, field)
		fmt.Fprintf(bufT, "$%v", i+1)
		if i < len(f.Fields)-1 {
			fmt.Fprint(bufH, ", ")
			fmt.Fprint(bufT, ", ")
		}
	}
	fmt.Fprint(bufH, ")")
	fmt.Fprint(bufT, ");")

	return string(append(bufH.Bytes(), bufT.Bytes()...))
}
