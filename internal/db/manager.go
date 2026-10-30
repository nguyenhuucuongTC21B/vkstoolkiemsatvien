// Package db provides a thin wrapper around the JSON-backed store
// (internal/store) plus an embedded snapshot of seed legal_docs and
// bieu_mau. The wrapper exists so module code can hold a single
// *Manager and reach both application tables (cases, evidence, audit_log,
// calc_history, spell_rules, settings) and the read-only legal tables
// (legal_docs, bieu_mau) through a uniform API.
package db

import (
	"kiem_sat_toolkit/internal/store"
)

// Manager wraps two Store instances:
//   - App   : user data (case files, evidence, audit log, OCR cache)
//   - Legal : read-mostly legal documents (laws, decrees, circulars)
type Manager struct {
	App   *store.Store
	Legal *store.Store
}

// NewManager opens the user store under dataDir and seeds the legal
// store from embedded snapshots if not already present.
func NewManager(dataDir string) (*Manager, error) {
	appStore, err := store.New(dataDir)
	if err != nil {
		return nil, err
	}
	legalStore, err := store.New(dataDir + "/legal")
	if err != nil {
		return nil, err
	}
	mgr := &Manager{App: appStore, Legal: legalStore}
	if err := mgr.seedLegal(); err != nil {
		// non-fatal — UI will just see an empty legal DB
		_ = err
	}
	if err := mgr.seedSettings(); err != nil {
		_ = err
	}
	if err := mgr.seedSpellRules(); err != nil {
		_ = err
	}
	return mgr, nil
}

// Close is a no-op for the JSON store (no connection pool).
func (m *Manager) Close() error { return nil }

// seedLegal populates legal_docs and bieu_mau tables from embedded JSON
// the first time the app starts.
func (m *Manager) seedLegal() error {
	tbl := m.Legal.Table("legal_docs")
	if tbl.Count() > 0 {
		return nil
	}
	rows := []store.Row{}
	for _, d := range legalDocsSeed {
		rows = append(rows, store.Row(d))
	}
	if err := m.Legal.Table("legal_docs").UpsertAll(rows); err != nil {
		return err
	}
	bmRows := []store.Row{}
	for _, b := range bieuMauSeed {
		bmRows = append(bmRows, store.Row(b))
	}
	return m.Legal.Table("bieu_mau").UpsertAll(bmRows)
}

func (m *Manager) seedSettings() error {
	tbl := m.App.Table("settings")
	if tbl.Count() > 0 {
		return nil
	}
	defaults := []store.Row{
		{"key": "theme", "value": "light"},
		{"key": "language", "value": "vi"},
		{"key": "ocr_lang_default", "value": "vie"},
		{"key": "spellcheck_strict", "value": "1"},
		{"key": "export_signature", "value": "0"},
		{"key": "audit_max_age_days", "value": "365"},
	}
	return tbl.UpsertAll(defaults)
}

func (m *Manager) seedSpellRules() error {
	tbl := m.App.Table("spell_rules")
	if tbl.Count() > 0 {
		return nil
	}
	rules := []store.Row{}
	for i, r := range spellRulesSeed {
		rules = append(rules, store.Row{
			"id":          i + 1,
			"pattern":     r.Pattern,
			"replacement": r.Replacement,
			"rule_type":   r.RuleType,
			"enabled":     1,
			"notes":       r.Notes,
		})
	}
	return tbl.UpsertAll(rules)
}
