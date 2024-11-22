package main

import (
	"os"
	"testing"
)

const MAX_TEST_INSERTS = 100000
const MAX_TEST_WRITERS = 3

func BenchmarkWALMode(b *testing.B) {
	os.Remove("test-wal.db")

	dbObject := create("test-wal.db", DBMaxWriters(MAX_TEST_WRITERS), WithPragma([]PragmaConfig{
		{"journal_mode", "WAL"},
		{"busy_timeout", "5000"},
		{"synchronous", "NORMAL"},
	}))
	defer dbObject.CloseDB()

	for i := 0; i < b.N; i++ {
		dbObject.InsertNRandom(MAX_TEST_INSERTS)
	}
}

func BenchmarkDeleteMode(b *testing.B) {
	os.Remove("test-delete.db")

	// BUG: Yet to add re-trys on db locking
	dbObject := create("test-delete.db", DBMaxWriters(MAX_TEST_WRITERS), WithPragma([]PragmaConfig{
		{"journal_mode", "WAL"},
		{"busy_timeout", "5000"},
		{"synchronous", "NORMAL"},
	}))
	defer dbObject.CloseDB()

	dbObject.SetJournalMode(ModeDelete)

	for i := 0; i < b.N; i++ {
		dbObject.InsertNRandom(MAX_TEST_INSERTS)
	}
}
