package tachyon

import (
	"bytes"
	"database/sql"
	"fmt"
	"go/build"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path"
	"sort"
)

const FixturePath = "testdata"

// Fixture represents a test Fixture.
type Fixture struct {
	Table   string            `yaml:"table"`
	Fields  []string          `yaml:"fields"`
	Records map[string]Record `yaml:"data"`
}

func NewFixture(name string) (*Fixture, error) {
	var dataDir = ""

	// Test if the parent directory is a go package, and if it is, look for
	// fixtures there.
	_, err := build.Import("./", "../", build.IgnoreVendor)
	if err == nil {
		dataDir = "../"
	}

	raw, err := ioutil.ReadFile(path.Join(dataDir, FixturePath, name+".yml"))
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

func ReadFixtures(names ...string) (FixtureList, error) {
	list := FixtureList{}
	for _, name := range names {
		f, err := NewFixture(name)
		if err != nil {
			return nil, err
		}
		list = append(list, f)
	}

	return list, nil
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

func (f *Fixture) Clean(db *sql.DB) error {
	_, err := db.Exec(f.deleteStr())
	return err
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

func (f *Fixture) deleteStr() string {
	buf := bytes.NewBuffer([]byte{})
	fmt.Fprintf(buf, "DELETE FROM %v;", f.Table)
	return string(buf.Bytes())
}
