// Package store provides a tiny JSON-file-backed document store used
// in place of SQLite. The API mirrors enough of database/sql to keep
// the rest of the codebase familiar, but the implementation is pure
// Go with sync.Mutex for concurrency.
//
// Design:
//   - Each "table" is a JSON file on disk under <dataDir>/store/.
//   - The whole table is loaded into memory at startup and reloaded
//     on each transaction.
//   - Writes are atomic (write to .tmp then rename).
//   - The legal_docs and bieu_mau tables are seeded from embedded
//     JSON snapshots on first run.
//
// This trades query power (no SQL) for: zero external deps, fully
// portable, no CGO, no libsqlite3. For an offline prosecutor toolkit
// with thousands (not millions) of records, this is plenty fast.
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// Store is the on-disk JSON store.
type Store struct {
	root string
	mu   sync.RWMutex
}

// New opens (or creates) a store rooted at root/store.
func New(root string) (*Store, error) {
	s := &Store{root: filepath.Join(root, "store")}
	if err := os.MkdirAll(s.root, 0755); err != nil {
		return nil, err
	}
	return s, nil
}

// Table returns a handle to one named table.
func (s *Store) Table(name string) *Table {
	return &Table{store: s, name: name}
}

// Table is a single JSON file. Each row is a map[string]any.
type Table struct {
	store *Store
	name  string
}

// Row is one record.
type Row = map[string]interface{}

// All loads every row in the table.
func (t *Table) All() []Row {
	t.store.mu.RLock()
	defer t.store.mu.RUnlock()
	path := t.store.tablePath(t.name)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var rows []Row
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil
	}
	return rows
}

// Filter returns rows matching predicate.
func (t *Table) Filter(pred func(Row) bool) []Row {
	rows := t.All()
	out := make([]Row, 0, len(rows))
	for _, r := range rows {
		if pred(r) {
			out = append(out, r)
		}
	}
	return out
}

// First returns the first matching row or nil.
func (t *Table) First(pred func(Row) bool) Row {
	for _, r := range t.All() {
		if pred(r) {
			return r
		}
	}
	return nil
}

// Insert appends a row.
func (t *Table) Insert(row Row) error {
	t.store.mu.Lock()
	defer t.store.mu.Unlock()
	rows, _ := t.loadLocked()
	rows = append(rows, row)
	return t.saveLocked(rows)
}

// Update modifies the first row matching pred.
func (t *Table) Update(pred func(Row) bool, patch Row) error {
	t.store.mu.Lock()
	defer t.store.mu.Unlock()
	rows, _ := t.loadLocked()
	updated := 0
	for i, r := range rows {
		if pred(r) {
			for k, v := range patch {
				r[k] = v
			}
			rows[i] = r
			updated++
			break
		}
	}
	if updated == 0 {
		return fmt.Errorf("no matching row")
	}
	return t.saveLocked(rows)
}

// UpsertAll replaces the entire table atomically.
func (t *Table) UpsertAll(rows []Row) error {
	t.store.mu.Lock()
	defer t.store.mu.Unlock()
	return t.saveLocked(rows)
}

// Delete removes rows matching pred.
func (t *Table) Delete(pred func(Row) bool) (int, error) {
	t.store.mu.Lock()
	defer t.store.mu.Unlock()
	rows, _ := t.loadLocked()
	kept := make([]Row, 0, len(rows))
	removed := 0
	for _, r := range rows {
		if pred(r) {
			removed++
			continue
		}
		kept = append(kept, r)
	}
	if err := t.saveLocked(kept); err != nil {
		return 0, err
	}
	return removed, nil
}

// Count returns the number of rows.
func (t *Table) Count() int {
	return len(t.All())
}

// SortedCopy returns rows sorted by less.
func (t *Table) SortedCopy(less func(a, b Row) bool) []Row {
	rows := t.All()
	sort.SliceStable(rows, func(i, j int) bool { return less(rows[i], rows[j]) })
	return rows
}

// ────────────── internals ──────────────

func (t *Table) loadLocked() ([]Row, error) {
	data, err := os.ReadFile(t.store.tablePath(t.name))
	if err != nil {
		if os.IsNotExist(err) {
			return []Row{}, nil
		}
		return nil, err
	}
	var rows []Row
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []Row{}
	}
	return rows, nil
}

func (t *Table) saveLocked(rows []Row) error {
	if rows == nil {
		rows = []Row{}
	}
	data, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return err
	}
	path := t.store.tablePath(t.name)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (s *Store) tablePath(name string) string {
	return filepath.Join(s.root, name+".json")
}

// ────────────── helpers ──────────────

// ILike matches a SQL LIKE pattern case-insensitively.
func ILike(value, pattern string) bool {
	if pattern == "" {
		return true
	}
	pattern = strings.Trim(pattern, "%")
	return strings.Contains(strings.ToLower(value), strings.ToLower(pattern))
}

// Equal is a literal equality predicate on a field.
func Equal(field string, want interface{}) func(Row) bool {
	return func(r Row) bool {
		v, ok := r[field]
		if !ok {
			return false
		}
		return fmt.Sprintf("%v", v) == fmt.Sprintf("%v", want)
	}
}

// FieldStr returns r[field] as string.
func FieldStr(r Row, field string) string {
	if v, ok := r[field]; ok {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

// FieldInt returns r[field] as int.
func FieldInt(r Row, field string) int {
	if v, ok := r[field]; ok {
		switch n := v.(type) {
		case int:
			return n
		case int64:
			return int(n)
		case float64:
			return int(n)
		case json.Number:
			i, _ := n.Int64()
			return int(i)
		}
	}
	return 0
}

// FieldBool returns r[field] as bool.
func FieldBool(r Row, field string) bool {
	if v, ok := r[field]; ok {
		switch b := v.(type) {
		case bool:
			return b
		case int:
			return b != 0
		case float64:
			return b != 0
		}
	}
	return false
}

// NextID returns the next integer id for a table.
func (t *Table) NextID() int {
	max := 0
	for _, r := range t.All() {
		if i := FieldInt(r, "id"); i > max {
			max = i
		}
	}
	return max + 1
}
