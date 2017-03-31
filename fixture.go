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

type Record map[string]string

// Returns a list of of the Record's data sorted the same as f.Fields. It will
// return an interface{} type to correspond to Stmt.Exec's expectations.
func (r Record) orderedData(f *Fixture) []interface{} {
	list := make([]interface{}, 0, len(f.Fields))

	for _, fName := range f.Fields {
		if val, ok := r[fName]; ok {
			list = append(list, val)
		} else {
			list = append(list, "")
		}
	}
	return list
}

// Fixture represents a test Fixture.
type Fixture struct {
	Table   string
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

	f := &Fixture{Table: name}
	err = yaml.Unmarshal(raw, &f)
	if err != nil {
		return nil, err
	}

	f.sortFields()
	return f, nil
}

// Load inserts f's data into db inside of a transaction. If an error is
// encountered, the transaction is rolled back.
// TODO: Should use db.BeginTx as db.Begin is deprecated.
func (f *Fixture) Load(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	err = f.LoadTx(tx)
	if err != nil {
		tx.Commit()
	} else {
		tx.Rollback()
	}
	return err
}

// LoadTx inserts f's data within the transaction, but does not commit or
// rollback. The calling code must commit it manually.
func (f *Fixture) LoadTx(tx *sql.Tx) error {
	stmt, err := tx.Prepare(f.insertStr())
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, record := range f.Records {
		_, err := stmt.Exec(record.orderedData(f)...)
		if err != nil {
			return err
		}
	}
	return nil
}

// FindTag finds the record in f that has the name "tag". The second return
// value will be false if the record cannot be found, and true otherwise.
func (f *Fixture) FindTag(tag string) (Record, bool) {
	r, ok := f.Records[tag]
	return r, ok
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
