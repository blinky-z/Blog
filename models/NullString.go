package models

import (
	"database/sql"
	"encoding/json"
	"reflect"
)

// NullString - wraps built-in sql.NullString
// We need this struct to use it in models.Comment since field ParentID can be null, but we can use it somewhere else
// in future
// Declaring own type, we can use custom marshal and unmarshal json functions
// Now json object will store only null or string value, not the whole sql.NullString struct
type NullString sql.NullString

// Value - helpful function to get value of NullString type field
func (ns *NullString) Value() interface{} {
	if ns.Valid {
		return ns.String
	}
	return sql.NullString{}
}

// Scan - function to scan value from sql row's field
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

// MarshalJSON - custom marshal func for NullString
// Now we store in json representation not the whole sql.NullString struct but only "bull" or string value, if row's field
// is not null
func (ns *NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}

// UnmarshalJSON - custom unmarshal func for NullString
// We also need to define custom unmarshal func to parse json null value or json string into NullString type
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
