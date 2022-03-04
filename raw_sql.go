package godb

import "database/sql"

// RawSQL allows the execution of a custom SQL query.
// Initialize it with the RawSQL method.
//
// WARNING : the arguments will be used 'as is' by the Go sql package, then it
// will not duplicate the placeholders if you use a slice.
//
// Note : the API could have been build without an intermediaite struct.
// But this produce a mode homogeneous API, and allows
// later evolutions without breaking the it.
type RawSQL struct {
	db        *DB
	sql       string
	arguments []interface{}
}

// RawSQL create a RawSQL structure, allowing the executing of a custom sql
// query, and the mapping of its result.
func (db *DB) RawSQL(sql string, args ...interface{}) *RawSQL {
	return &RawSQL{
		db:        db,
		sql:       sql,
		arguments: args,
	}
}

// Do executes the raw query.
// The record argument has to be a pointer to a struct or a slice.
// If the argument is not a slice, a row is expected, and Do returns
// sql.ErrNoRows is none where found.
func (raw *RawSQL) Do(record interface{}) error {
	recordInfo, err := buildRecordDescription(record)
	if err != nil {
		return err
	}

	// the function which will return the pointers according to the given columns
	pointersGetter := func(record interface{}, columns []string) ([]interface{}, error) {
		var pointers []interface{}
		pointers, err := recordInfo.structMapping.GetPointersForColumns(record, columns...)
		return pointers, err
	}

	rowsCount, err := raw.db.doSelectOrWithReturning(raw.sql, raw.arguments, recordInfo, pointersGetter)
	if err != nil {
		return err
	}

	// When a single instance is requested but not found, sql.ErrNoRows is
	// returned like QueryRow in database/sql package.
	if !recordInfo.isSlice && rowsCount == 0 {
		err = sql.ErrNoRows
	}

	return err
}

// DoWithIterator executes the select query and returns an Iterator allowing
// the caller to fetch rows one at a time.
// Warning : it does not use an existing transation to avoid some pitfalls with
// drivers, nor the prepared statement.
func (raw *RawSQL) DoWithIterator() (Iterator, error) {
	return raw.db.doWithIterator(raw.sql, raw.arguments)
}
