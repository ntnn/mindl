package sum

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
)

type DB struct {
	path    string
	entries map[string]Entry
}

// entryLength is used when reading records to validate the number of fields.
const entryLength = 3

type Entry struct {
	Key     string
	Sum     string
	Comment string
}

func (e Entry) Equal(other Entry) bool {
	return e.Key == other.Key && e.Sum == other.Sum && e.Comment == other.Comment
}

func (db *DB) set(key, sum, comment string) bool {
	entered := Entry{
		Key:     key,
		Sum:     sum,
		Comment: comment,
	}

	existing, ok := db.entries[key]
	if !ok || !existing.Equal(entered) {
		db.entries[key] = entered
		return true
	}
	return false
}

func (db *DB) Set(key, sum, comment string) error {
	if !db.set(key, sum, comment) {
		return nil
	}
	return db.Save()
}

func (db *DB) Get(key string) (Entry, bool) {
	val, ok := db.entries[key]
	return val, ok
}

func Read(in io.Reader) (*DB, error) {
	db := &DB{
		entries: map[string]Entry{},
	}
	reader := csv.NewReader(in)

	for {
		record, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		if len(record) != entryLength {
			return nil, fmt.Errorf("record is not three columns wide: %q", record)
		}
		db.set(record[0], record[1], record[2])
	}

	return db, nil
}

const sumDbPerms = 0o600

func Open(p string) (*DB, error) {
	f, err := os.OpenFile(p, os.O_CREATE, sumDbPerms)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	db, err := Read(f)
	if err != nil {
		return nil, err
	}
	db.path = p
	return db, nil
}

func (db *DB) Save() error {
	f, err := os.Create(db.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return db.Write(f)
}

func (db *DB) Write(out io.Writer) error {
	writer := csv.NewWriter(out)

	for _, entry := range db.entries {
		if err := writer.Write([]string{entry.Key, entry.Sum, entry.Comment}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}
