package db

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"
)

// SportsRepo provides repository access to sports.
type SportsRepo interface {
	// Init will initialise our sports repository.
	Init() error

	// List will return a list of races.
	List(filter *sports.ListEventsRequestFilter) ([]*sports.Events, error)
}

type sportsRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewSportsRepo creates a new sports repository.
func NewSportsRepo(db *sql.DB) SportsRepo {
	return &sportsRepo{db: db}
}

// Init prepares the sports repository dummy data.
func (r *sportsRepo) Init() error {
	var err error

	r.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy races.
		err = r.seed()
	})

	return err
}

func (r *sportsRepo) List(filter *sports.ListEventsRequestFilter) ([]*sports.Events, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getEventsQueries()[sportsList]

	query, args = r.applyFilter(query, filter)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	return r.scanEvents(rows)
}

func (r *sportsRepo) applyFilter(query string, filter *sports.ListEventsRequestFilter) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	if filter == nil {
		return query, args
	}

	if len(filter.Ids) > 0 {
		clauses = append(clauses, "id IN ("+strings.Repeat("?,", len(filter.Ids)-1)+"?)")

		for _, ID := range filter.Ids {
			args = append(args, ID)
		}
	}

	if len(clauses) != 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	return query, args
}

func (m *sportsRepo) scanEvents(
	rows *sql.Rows,
) ([]*sports.Events, error) {
	var sports []*sports.Events

	for rows.Next() {
		var events sports.Events
		var advertisedStart time.Time

		if err := rows.Scan(&events.Id, &events.Name, &advertisedStart); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}

		ts, err := ptypes.TimestampProto(advertisedStart)
		if err != nil {
			return nil, err
		}

		events.AdvertisedStartTime = ts

		sports = append(events, &events)
	}

	return sports, nil
}
