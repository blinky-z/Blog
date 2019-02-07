package models

import (
	"database/sql"
	"encoding/json"
	"reflect"
)

type NullString sql.NullString

func (ns *NullString) Value() string {
	return ns.String
}

func (ns *NullString) Scan(value interface{}) error {
	var s sql.NullString
	if err := s.Scan(value); err != nil {
		return err
	}

	if reflect.TypeOf(value) == nil {
		*ns = NullString{String: s.String, Valid: false}
	} else {
		*ns = NullString{String: s.String, Valid: true}
	}

	return nil
}

func (ns *NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}

func (ns *NullString) UnmarshalJSON(b []byte) error {
	var x interface{}
	err := json.Unmarshal(b, &x)
	if err != nil {
		return err
	}
	switch s := x.(type) {
	case nil:
		ns.Valid = false
	case string:
		ns.String = s
		ns.Valid = true
	}

	return nil
}
