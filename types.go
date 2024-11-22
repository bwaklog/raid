package main

import (
	"database/sql"
)

type Config struct {
	MaxWriters uint
}

type ConfigFunc func(c *Config)

type DBObj struct {
	db     *sql.DB
	name   string
	status bool
	Config
}

type Record struct {
	Value string `json:"value"`
	Key   int    `json:"key"`
}

const MAX_ROUTINES = 1
const MAX_INSERTS = 100000

const (
	ModeDelete = "DELETE"
	ModeWAL    = "WAL"
)
