package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/paranoiachains/metrics/internal/collector"
	"github.com/paranoiachains/metrics/internal/flags"
)

var retryDelays = []time.Duration{1, 3, 5}

var Storage = NewMemStorage()

// Depends on the flags used while running the program
var CurrentStorage Database

// flexibility
type Database interface {
	Update(ctx context.Context, mtype string, id string, value any) error
	UpdateBatch(ctx context.Context, metrics collector.Metrics) error
	Return(ctx context.Context, mtype string, id string) (*collector.Metric, error)
}

type FileHandler interface {
	Write(filename string) error
	Read(filename string) (*collector.Metric, error)
	Restore(filename string) error
	ClearFile(filename string) error
}

// function for determining whether to use memory storage or postgres
func DetermineStorage() (Database, error) {
	var s Database
	if flags.DBEndpoint != "" {
		databaseDSN := flags.DBEndpoint
		db, err := ConnectAndPing("pgx", databaseDSN)
		if err != nil {
			return nil, err
		}
		s = db
		fmt.Println("Using POSTGRESQL")
	} else {
		s = Storage
		fmt.Println("Using MemStorage")
	}
	return s, nil
}

// ----- MEMORY STORAGE -----

// store values (temporary choice)
type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

// creates new memory storage
func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
}

// clears memory storage
func (s *MemStorage) Clear() {
	s.Gauge = make(map[string]float64)
	s.Counter = make(map[string]int64)
}

// updates memory storage
func (s *MemStorage) Update(ctx context.Context, mtype string, id string, value any) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	switch mtype {
	case "gauge":
		v, ok := value.(float64)
		if !ok {
			return fmt.Errorf("type assertion error while updating memory storage")
		}
		s.Gauge[id] = v

	case "counter":
		v, ok := value.(int64)
		if !ok {
			return fmt.Errorf("type assertion error while updating memory storage")
		}
		s.Counter[id] += v
	}
	return nil
}

func (s *MemStorage) UpdateBatch(ctx context.Context, metrics collector.Metrics) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	for _, metric := range metrics {
		switch metric.MType {
		case "gauge":
			s.Update(ctx, metric.MType, metric.ID, *metric.Value)
		case "counter":
			s.Update(ctx, metric.MType, metric.ID, *metric.Delta)
		}
	}
	return nil
}

// retrieves value from memory storage
func (s MemStorage) Return(ctx context.Context, mtype string, id string) (*collector.Metric, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	switch mtype {
	case "gauge":
		v, ok := s.Gauge[id]
		if !ok {
			return nil, fmt.Errorf("no such gauge metric")
		}
		return &collector.Metric{ID: id, MType: mtype, Value: &v}, nil

	case "counter":
		v, ok := s.Counter[id]
		if !ok {
			return nil, fmt.Errorf("no such counter metric")
		}
		return &collector.Metric{ID: id, MType: mtype, Delta: &v}, nil
	}

	return nil, fmt.Errorf("unknown metric type")
}

// writes to memory storage
func (s MemStorage) Write(filename string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	var metrics collector.Metrics
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	for name := range s.Gauge {
		metric, err := s.Return(context.Background(), "gauge", name)
		if err != nil {
			return err
		}
		metrics = append(metrics, *metric)
	}

	for name := range s.Counter {
		metric, err := s.Return(context.Background(), "counter", name)
		if err != nil {
			return err
		}
		metrics = append(metrics, *metric)
	}
	if err := encoder.Encode(metrics); err != nil {
		return err
	}
	return nil
}

// FileHandler interface implementation of MemStorage type
func (s MemStorage) Read(filename string) (*collector.Metric, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var metric collector.Metric
	if err := json.Unmarshal(data, &metric); err != nil {
		return nil, err
	}

	return &metric, nil
}

func (s *MemStorage) Restore(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	for {
		var metrics collector.Metrics
		if err := decoder.Decode(&metrics); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		for _, metric := range metrics {
			switch metric.MType {
			case "gauge":
				s.Gauge[metric.ID] = *metric.Value
			case "counter":
				s.Counter[metric.ID] = *metric.Delta
			}
		}
	}
	return nil
}

func (s *MemStorage) ClearFile(filename string) error {
	if err := os.Truncate(filename, 0); err != nil {
		return err
	}
	return nil
}

func WriteWithInterval(file FileHandler, filename string, storeInterval int) {
	// lol
	if storeInterval == 0 {
		storeInterval = 1
	}
	// doesnt work without this line idk why
	time.Sleep(time.Second)
	for {
		file.ClearFile(filename)
		if err := file.Write(filename); err != nil {
			log.Fatal(err)
		}
		time.Sleep(time.Duration(storeInterval) * time.Second)
	}
}

// ----- POSTGRES DATABASE -----

// redeclaration for Database interface implementation
type DBStorage struct {
	*sql.DB
}

func (db DBStorage) CreateIfNotExists(ctx context.Context) error {
	createQuery := `
	CREATE TABLE IF NOT EXISTS metrics (
    id VARCHAR(255) PRIMARY KEY,
    mtype VARCHAR(50) NOT NULL,
    value DOUBLE PRECISION,
    delta BIGINT
);`
	err := withRetry(func() error {
		_, err := db.ExecContext(ctx, createQuery)
		return err
	})
	if err != nil {
		return err
	}
	return nil
}

func (db DBStorage) Update(ctx context.Context, mtype string, id string, value any) error {
	insertQuery := `
	INSERT INTO metrics (id, mtype, value, delta)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (id) DO UPDATE
		SET value = EXCLUDED.value, delta = EXCLUDED.delta;`

	counterDeltaQuery := `
	SELECT delta FROM metrics
	WHERE id=$1;
	`

	switch mtype {
	case "gauge":
		v, ok := value.(float64)
		if !ok {
			return fmt.Errorf("type assertion error while updating database")
		}
		return withRetry(func() error {
			_, err := db.ExecContext(ctx, insertQuery, id, mtype, v, nil)
			return err
		})

	case "counter":
		v, ok := value.(int64)
		if !ok {
			return fmt.Errorf("type assertion error while updating database")
		}
		return withRetry(func() error {
			var currentDelta sql.NullInt64
			row := db.QueryRowContext(ctx, counterDeltaQuery, id)
			err := row.Scan(&currentDelta)
			if err != nil && err != sql.ErrNoRows {
				return err
			}
			newDelta := v
			if currentDelta.Valid {
				newDelta += currentDelta.Int64
			}
			_, err = db.ExecContext(ctx, insertQuery, id, mtype, nil, newDelta)
			return err
		})
	default:
		return fmt.Errorf("unknown metric type: %s", mtype)
	}
}

func (db DBStorage) UpdateBatch(ctx context.Context, metrics collector.Metrics) error {
	insertQuery := `
	INSERT INTO metrics (id, mtype, value, delta)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (id) DO UPDATE
		SET value = EXCLUDED.value, delta = EXCLUDED.delta;`
	counterDeltaQuery := `
	SELECT delta FROM metrics
	WHERE id=$1;
	`

	return withRetry(func() error {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		stmt, err := tx.PrepareContext(ctx, insertQuery)
		if err != nil {
			tx.Rollback()
			return err
		}
		defer stmt.Close()

		for _, metric := range metrics {
			switch metric.MType {
			case "gauge":
				if _, err := stmt.ExecContext(ctx, metric.ID, metric.MType, *metric.Value, nil); err != nil {
					tx.Rollback()
					return err
				}
			case "counter":
				var currentDelta sql.NullInt64
				row := tx.QueryRowContext(ctx, counterDeltaQuery, metric.ID)
				err := row.Scan(&currentDelta)
				if err != nil && err != sql.ErrNoRows {
					tx.Rollback()
					return err
				}
				newDelta := *metric.Delta
				if currentDelta.Valid {
					newDelta += currentDelta.Int64
				}
				if _, err := stmt.ExecContext(ctx, metric.ID, metric.MType, nil, newDelta); err != nil {
					tx.Rollback()
					return err
				}
			default:
				tx.Rollback()
				return fmt.Errorf("unknown metric type: %s", metric.MType)
			}
		}

		return tx.Commit()
	})
}

func (db DBStorage) Return(ctx context.Context, mtype string, id string) (*collector.Metric, error) {
	selectQuery := `
	SELECT id, mtype, value, delta 
	FROM metrics 
	WHERE id=$1 AND mtype=$2;`

	var metric collector.Metric
	err := withRetry(func() error {
		row := db.QueryRowContext(ctx, selectQuery, id, mtype)
		return row.Scan(&metric.ID, &metric.MType, &metric.Value, &metric.Delta)
	})
	if err != nil {
		return nil, err
	}
	fmt.Println("Successfully returned value from table metrics")
	return &metric, nil
}

func ConnectAndPing(driverName string, dataSourceName string) (*DBStorage, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	fmt.Println("Successfully connected to db")

	newDB := &DBStorage{db}
	if err := newDB.CreateIfNotExists(ctx); err != nil {
		return nil, err
	}
	return newDB, nil
}

func withRetry(fn func() error) error {
	var lastErr error
	for _, delay := range retryDelays {
		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			fmt.Printf("DB connection failed, retrying in %v...\n", delay*time.Second)
			time.Sleep(delay * time.Second)
			continue
		}
		return err // не retriable
	}
	return lastErr
}
