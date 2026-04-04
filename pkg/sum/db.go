package sum

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
)

// DB is a key-value store for checksum entries.
type DB struct {
	path    string
	entries map[string]Entry
}

// entryLength is used when reading records to validate the number of fields.
const entryLength = 5

// Entry represents a single checksum record.
type Entry struct {
	URLTemplate string
	OS          string
	Arch        string
	Sum         string
	Comment     string
}

// Equal returns true if the entry and other are identical.
func (e Entry) Equal(other Entry) bool {
	return e.URLTemplate == other.URLTemplate &&
		e.OS == other.OS &&
		e.Arch == other.Arch &&
		e.Sum == other.Sum &&
		e.Comment == other.Comment
}

func (e Entry) key() string {
	return fmt.Sprintf("%s#%s#%s", e.URLTemplate, e.OS, e.Arch)
}

func (e Entry) records() []string {
	return []string{
		e.URLTemplate,
		e.OS,
		e.Arch,
		e.Sum,
		e.Comment,
	}
}

// Set stores an entry.
func (db *DB) Set(urlTemplate, os, arch, sum, comment string) {
	e := Entry{
		URLTemplate: urlTemplate,
		OS:          os,
		Arch:        arch,
		Sum:         sum,
		Comment:     comment,
	}

	db.entries[e.key()] = e
}

// Get retrieves an entry by key.
func (db *DB) Get(urlTemplate, os, arch string) (Entry, bool) {
	key := Entry{
		URLTemplate: urlTemplate,
		OS:          os,
		Arch:        arch,
	}.key()

	val, ok := db.entries[key]
	return val, ok
}

// GetAll retrieves all entries for the given URL template.
func (db *DB) GetAll(urlTemplate string) []Entry {
	es := []Entry{}
	for _, entry := range db.entries {
		if entry.URLTemplate != urlTemplate {
			continue
		}
		es = append(es, entry)
	}
	return es
}

// Read parses a DB from in.
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
		db.Set(record[0], record[1], record[2], record[3], record[4])
	}

	return db, nil
}

const sumDbPerms = 0o600

// Open opens a DB at path p.
// If the path does not exist an empty DB will be returned.
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

// Save writes the DB to its file path.
func (db *DB) Save() error {
	f, err := os.Create(db.path)
	if err != nil {
		return err
	}
	defer f.Close()
	return db.Write(f)
}

// Write writes the DB to the given writer.
func (db *DB) Write(out io.Writer) error {
	writer := csv.NewWriter(out)

	keys := slices.Collect(maps.Keys(db.entries))
	slices.Sort(keys)

	for _, key := range keys {
		entry := db.entries[key]
		if err := writer.Write(entry.records()); err != nil {
			return err
		}
	}

	writer.Flush()
	return writer.Error()
}
