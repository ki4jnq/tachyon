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

func (fl FixtureList) Clean(db *sql.DB) error {
	for _, f := range fl {
		if err := f.Clean(db); err != nil {
			return err
		}
	}
	return nil
}
