package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

func DefaultConfig() Config {
	return Config{
		MaxWriters: 1,
	}
}

func DBMaxWriters(N uint) ConfigFunc {
	if N <= 0 {
		return func(c *Config) {
		}
	}
	return func(c *Config) {
		c.MaxWriters = N
	}
}

func create(name string, configs ...ConfigFunc) *DBObj {
	os.Remove(name)
	db, err := sql.Open("sqlite3", name)
	if err != nil {
		log.Fatal(err)
	}

	query := `
  create table test (key integer primary key not null, value text not null)
  `

	config := DefaultConfig()

	for _, conf := range configs {
		conf(&config)
	}

	_, err = db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}

	return &DBObj{
		name:   name,
		db:     db,
		status: true,
		Config: config,
	}
}

func (d *DBObj) GlobDeleteDB() {
	files, err := filepath.Glob(fmt.Sprintf("%s*", d.name))
	if err != nil {
		log.Fatal("Failed to create filepath.Glob, ", err)
	}

	for _, f := range files {
		if err := os.Remove(f); err != nil {
			log.Fatal("Failed to delete, ", f)
		}
	}
}

func (d *DBObj) CloseDB() {
	d.db.Close()
	d.status = false
	d.GlobDeleteDB()
}

func (d *DBObj) GetJournalMode() string {
	rows, err := d.db.Query("PRAGMA journal_mode")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var mode string
		err = rows.Scan(&mode)
		if err != nil {
			log.Fatal(err)
		}
		if strings.Compare(strings.ToLower("DELETE"), mode) == 0 {
			rows.Close()
			return ModeDelete
		} else if strings.Compare(strings.ToLower("WAL"), mode) == 0 {
			rows.Close()
			return ModeWAL
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return ModeDelete
}

func (d *DBObj) SetJournalMode(mode string) {
	_ = d.GetJournalMode()

	var modeStr string
	switch mode {
	case ModeDelete:
		{
			modeStr = "DELETE"
		}
	case ModeWAL:
		{
			modeStr = "WAL"
		}
	default:
		return
	}

	query := fmt.Sprintf("PRAGMA journal_mode=%s", modeStr)
	rows, err := d.db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var result string
		err = rows.Scan(&result)
		if err != nil {
			log.Fatal(err)
		}
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}
}

func InsertObjToDB(db *sql.DB, rec Record) error {
	if rec.Key < 0 || rec.Value == "" {
		return errors.New("bad record")
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatal("db.Begin failed", err)
	}

	query := "insert into test values (?, ?)"
	stmt, err := tx.Prepare(query)
	if err != nil {
		log.Fatal("tx.Prepare failed", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(rec.Key, rec.Value)
	if err != nil {
		log.Fatal("stmt.Exec failed", err)
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal("tx.Commit failed", err)
	}

	return nil
}

func (d *DBObj) ClearTable() {
	tx, err := d.db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("DELETE FROM test")
	defer stmt.Close()
	if err != nil {
		log.Fatal(err)
	}
	_, err = stmt.Exec()
	if err != nil {
		log.Fatal(err)
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
}

func Sample() {
	count := 100
	base := 1
	for i := 0; i < count; i++ {
		base = base * i
	}
}

func (d *DBObj) InsertNRandom(N int) {
	var wg sync.WaitGroup
	d.ClearTable()
	semaphore := make(chan struct{}, d.Config.MaxWriters)
	for i := 0; i < N; i++ {
		semaphore <- struct{}{}
		wg.Add(1)
		go func() {
			defer func() {
				<-semaphore
				wg.Done()
			}()
			rec := Record{
				Key:   i,
				Value: fmt.Sprintf("value_%d", i),
			}
			InsertObjToDB(d.db, rec)
		}()
	}
	wg.Wait()
}

func main() {
	dbObject := create("test.db", DBMaxWriters(1))
	defer dbObject.CloseDB()

	dbObject.SetJournalMode(ModeWAL)
	dbObject.InsertNRandom(MAX_INSERTS)

	dbObject.GlobDeleteDB()
}
