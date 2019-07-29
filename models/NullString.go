package models

import (
	"database/sql"
	"encoding/json"
	"reflect"
)

// NullString - wraps built-in sql.NullString
// We need this struct to store either null or value (string) depending on database value
// Also, declaring own type, we can use custom marshal and unmarshal json functions
// So now json object will store only null or string value, not the whole sql.NullString struct
type NullString sql.NullString

// Value - helpful function to get value of NullString type field
func (ns *NullString) Value() interface{} {
	if ns.Valid {
		return ns.String
	}
	return sql.NullString{}
}

// Scan - function to scan value from sql row field
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

// MarshalJSON - custom marshal func (override)
// this function will store either "null" or value but not the whole NullString struct
func (ns *NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}

// UnmarshalJSON - custom unmarshal func (override)
// we need to define custom unmarshal func to parse JSON "null" or value into NullString type
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
