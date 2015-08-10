package server

import (
	"fmt"
	"regexp"
	"database/sql"
	"encoding/json"
)

type Table struct {
	Db *sql.DB `json:"-"`
	Type string `json:"type"`
	Name string `json:"name"`
	DocField string `json:"docfield"`
	PubKey string `json:"pubkey,omitempty"`
}

func NewTable(name string) *Table {
	return &Table{Type: "table", Name: name, DocField: "doc"}
}

func (t *Table) validIdentifier(identifier string) (error) {
	match, err := regexp.MatchString("^[a-z0-9_-]+$", identifier)
	if err != nil {
		return err
	}
	if !match {
		return fmt.Errorf("invalid name: %s", identifier)
	}
	return nil
}

func (t *Table) Create() error {
	if err := t.validIdentifier(t.Name); err != nil {
		return err
	}

	if err := t.validIdentifier(t.DocField); err != nil {
		return err
	}

	if t.Db == nil {
		return fmt.Errorf("database connection required")
	}

	q := `
CREATE TABLE %s
(
  id text DEFAULT uuid_generate_v4(),
  rev text DEFAULT uuid_generate_v4(),
  %s jsonb NOT NULL DEFAULT '{}',
  CONSTRAINT pk_%s PRIMARY KEY (id)
)
WITH (
  OIDS=FALSE
);
`
	tx, err := t.Db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(fmt.Sprintf(q, t.Name, t.DocField, t.Name))
	if err != nil {
		return err
	}

	id := fmt.Sprintf("table/%s", t.Name)
	b, err := json.Marshal(t)
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO divan (id, doc) VALUES ($1, $2)", id, string(b))
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (t *Table) Delete() error {
	if err := t.validIdentifier(t.Name); err != nil {
		return err
	}

	id := fmt.Sprintf("table/%s", t.Name)

	tx, err := t.Db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(fmt.Sprintf("DROP TABLE %s", t.Name))
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM divan WHERE id=$1", id)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return err
}

func TableList(db *sql.DB) (map[string]Table, error) {
	rows, err := db.Query("SELECT id, rev, doc FROM divan WHERE doc->>'type' = 'table'")
	if err != nil && err.Error() == `pq: relation "divan" does not exist` {
		return nil, fmt.Errorf("divan table missing! Trying running with -b flag to bootstrap the environment or create the table manually.")
	}
	if err != nil {
		return nil, err
	}

	tables := make(map[string]Table)
	for rows.Next() {
		doc := new(Doc)
		if err := rows.Scan(&doc.Id, &doc.Rev, &doc.Doc); err != nil {
			return nil, err
		}
		t := new(Table)
		t.Db = db
		err := doc.JSON(t)
		if err != nil {
			return nil, err
		}
		tables[t.Name] = *t
	}

	return tables, nil
}