package tachyon

import (
	"database/sql"
)

type FixtureList []*Fixture

func (fl FixtureList) Load(db *sql.DB) error {
	for _, f := range fl {
		if err := f.Load(db); err != nil {
			return err
		}
	}
	return nil
}

func (fl FixtureList) LoadTx(tx *sql.Tx) error {
	for _, f := range fl {
		if err := f.LoadTx(tx); err != nil {
			return err
		}
	}
	return nil
}

// Clean deletes all records in effected tables in the reverse order that they
// were loaded. This helps ensure that any child records are deleted before
// the parent records they depend on are.
func (fl FixtureList) Clean(db *sql.DB) error {
	for i := len(fl) - 1; i >= 0; i-- {
		f := fl[i]
		if err := f.Clean(db); err != nil {
			return err
		}
	}
	return nil
}
