package server

import (
	"fmt"
	"strings"
	"database/sql"
	"encoding/json"
)

var ErrDocumentUpdateConflict = fmt.Errorf("Document update conflict.")

type Doc struct {
	Db *sql.DB `json:"-"`
	Id string
	Rev string
	Doc []byte `json:"-"`
}

func (d *Doc) Post(t Table, j map[string]interface{}) error {
	var fields []string
	var values []interface{}
	valueMarks := []string{"$1", "$2", "$3"}
	id, idExists := j["_id"]
	if idExists && id != nil && id != "" {
		fields = append(fields, "id")
		values = append(values, id)
	}
	rev, revExists := j["_rev"]
	if !revExists && rev != nil && rev != "" {
		fields = append(fields, "rev")
		values = append(values, rev)
	}
	delete(j, "_id")
	delete(j, "_rev")
	b, err := json.Marshal(j)
	if err != nil {
		return err
	}

	fields = append(fields, t.DocField)
	values = append(values, string(b))

	qFields := strings.Join(fields, ", ")
	qValues := strings.Join(valueMarks[:len(fields)], ", ")
	q := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id, rev", t.Name, qFields, qValues)
	return d.Db.QueryRow(q, values...).Scan(&d.Id, &d.Rev)
}

func (d *Doc) Get(t Table, id string) error {
	q := fmt.Sprintf("SELECT id, rev, %s FROM %s WHERE id=$1", t.DocField, t.Name)
	return d.Db.QueryRow(q, id).Scan(&d.Id, &d.Rev, &d.Doc)
}

func (d *Doc) Put(t Table, id string, j map[string]interface{}) error {
	err := d.Head(t, id)
	if err == sql.ErrNoRows {
		j["_id"] = id
		return d.Post(t, j)
	}
	if err != nil {
		return err
	}

	if j["_rev"] != d.Rev {
		return ErrDocumentUpdateConflict
	}

	delete(j, "_id")
	delete(j, "_rev")
	b, err := json.Marshal(j)
	if err != nil {
		return err
	}
	
	q := fmt.Sprintf("UPDATE %s SET %s=$1, rev=uuid_generate_v4() WHERE id=$2 RETURNING id, rev", t.Name, t.DocField)
	return d.Db.QueryRow(q, string(b), id).Scan(&d.Id, &d.Rev)
}

func (d *Doc) Delete(t Table, id string, rev string) error {
	q := fmt.Sprintf("DELETE FROM %s WHERE id=$1 and rev=$2", t.Name)
	return d.Db.QueryRow(q, id, rev).Scan(&d.Rev)
}

func (d *Doc) Head(t Table, id string) error {
	d.Id = id
	q := fmt.Sprintf("SELECT rev FROM %s WHERE id=$1", t.Name)
	return d.Db.QueryRow(q, id).Scan(&d.Rev)
}

func (d *Doc) String() (string, error) {
	var j map[string]interface{}
	if err := d.JSON(&j); err != nil {
		return "", err
	}
	j["_id"] = d.Id
	j["_rev"] = d.Rev
	b, err := json.Marshal(j)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (d *Doc) JSON(v interface{}) error {
	return json.Unmarshal(d.Doc, v)
}