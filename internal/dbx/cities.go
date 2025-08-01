package dbx

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

const citiesTable = "cities"

type CityModels struct {
	ID        uuid.UUID `db:"id"`
	CountryID uuid.UUID `db:"country_id"`
	Name      string    `db:"name"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type CitiesQ struct {
	db       *sql.DB
	selector sq.SelectBuilder
	inserter sq.InsertBuilder
	updater  sq.UpdateBuilder
	deleter  sq.DeleteBuilder
	counter  sq.SelectBuilder
}

func NewCities(db *sql.DB) CitiesQ {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	return CitiesQ{
		db:       db,
		selector: builder.Select("*").From(citiesTable),
		inserter: builder.Insert(citiesTable),
		updater:  builder.Update(citiesTable),
		deleter:  builder.Delete(citiesTable),
		counter:  builder.Select("COUNT(*) AS count").From(citiesTable),
	}
}

func (q CitiesQ) New() CitiesQ {
	return NewCities(q.db)
}

func (q CitiesQ) Insert(ctx context.Context, input CityModels) error {
	values := map[string]interface{}{
		"id":         input.ID,
		"country_id": input.CountryID,
		"name":       input.Name,
		"status":     input.Status,
		"created_at": input.CreatedAt,
		"updated_at": input.UpdatedAt,
	}

	query, args, err := q.inserter.SetMap(values).ToSql()
	if err != nil {
		return fmt.Errorf("building inserter query for table: %s: %w", citiesTable, err)
	}

	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = q.db.ExecContext(ctx, query, args...)
	}
	return err
}

func (q CitiesQ) Get(ctx context.Context) (CityModels, error) {
	query, args, err := q.selector.Limit(1).ToSql()
	if err != nil {
		return CityModels{}, fmt.Errorf("building selector query for table: %s: %w", citiesTable, err)
	}

	var city CityModels
	var row *sql.Row
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		row = tx.QueryRowContext(ctx, query, args...)
	} else {
		row = q.db.QueryRowContext(ctx, query, args...)
	}

	err = row.Scan(
		&city.ID,
		&city.CountryID,
		&city.Name,
		&city.Status,
		&city.CreatedAt,
		&city.UpdatedAt,
	)

	return city, err
}

func (q CitiesQ) Select(ctx context.Context) ([]CityModels, error) {
	query, args, err := q.selector.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building selector query for table: %s: %w", citiesTable, err)
	}

	var cities []CityModels
	var rows *sql.Rows
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		rows, err = tx.QueryContext(ctx, query, args...)
	} else {
		rows, err = q.db.QueryContext(ctx, query, args...)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var city CityModels
		if err := rows.Scan(
			&city.ID,
			&city.CountryID,
			&city.Name,
			&city.Status,
			&city.CreatedAt,
			&city.UpdatedAt,
		); err != nil {
			return nil, err
		}
		cities = append(cities, city)
	}

	return cities, nil
}

type CityUpdate struct {
	CountryID *uuid.UUID
	Name      *string
	Status    *string
	UpdatedAt time.Time
}

func (q CitiesQ) Update(ctx context.Context, input CityUpdate) error {
	updates := map[string]interface{}{
		"updated_at": input.UpdatedAt,
	}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Status != nil {
		updates["status"] = *input.Status
	}
	if input.CountryID != nil {
		updates["country_id"] = *input.CountryID
	}

	query, args, err := q.updater.SetMap(updates).ToSql()
	if err != nil {
		return fmt.Errorf("building updater query for table: %s: %w", citiesTable, err)
	}

	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = q.db.ExecContext(ctx, query, args...)
	}

	return err
}

func (q CitiesQ) Delete(ctx context.Context) error {
	query, args, err := q.deleter.ToSql()
	if err != nil {
		return fmt.Errorf("building deleter query for table: %s: %w", citiesTable, err)
	}

	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		_, err = tx.ExecContext(ctx, query, args...)
	} else {
		_, err = q.db.ExecContext(ctx, query, args...)
	}

	return err
}

func (q CitiesQ) FilterID(ID uuid.UUID) CitiesQ {
	q.selector = q.selector.Where(sq.Eq{"id": ID})
	q.counter = q.counter.Where(sq.Eq{"id": ID})
	q.deleter = q.deleter.Where(sq.Eq{"id": ID})
	q.updater = q.updater.Where(sq.Eq{"id": ID})
	return q
}

func (q CitiesQ) FilterCountryID(countryID uuid.UUID) CitiesQ {
	q.selector = q.selector.Where(sq.Eq{"country_id": countryID})
	q.counter = q.counter.Where(sq.Eq{"country_id": countryID})
	q.deleter = q.deleter.Where(sq.Eq{"country_id": countryID})
	q.updater = q.updater.Where(sq.Eq{"country_id": countryID})
	return q
}

func (q CitiesQ) FilterStatus(status string) CitiesQ {
	q.selector = q.selector.Where(sq.Eq{"status": status})
	q.counter = q.counter.Where(sq.Eq{"status": status})
	q.deleter = q.deleter.Where(sq.Eq{"status": status})
	q.updater = q.updater.Where(sq.Eq{"status": status})
	return q
}

func (q CitiesQ) Count(ctx context.Context) (uint64, error) {
	query, args, err := q.counter.ToSql()
	if err != nil {
		return 0, fmt.Errorf("building count query for table: %s: %w", citiesTable, err)
	}

	var count uint64
	if tx, ok := ctx.Value(TxKey).(*sql.Tx); ok {
		err = tx.QueryRowContext(ctx, query, args...).Scan(&count)
	} else {
		err = q.db.QueryRowContext(ctx, query, args...).Scan(&count)
	}

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (q CitiesQ) Page(limit, offset uint64) CitiesQ {
	q.counter = q.counter.Limit(limit).Offset(offset)
	q.selector = q.selector.Limit(limit).Offset(offset)
	return q
}
