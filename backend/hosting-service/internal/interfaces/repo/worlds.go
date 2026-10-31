package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain"
	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	unique_violation = "23505"
)

type WorldsRepo struct {
	pool *pgxpool.Pool
}

func NewWorldsRepo(pool *pgxpool.Pool) *WorldsRepo {
	return &WorldsRepo{pool: pool}
}

func (r *WorldsRepo) Upsert(ctx context.Context, server *models.MinecraftServer) (*models.MinecraftServer, error) {
	sql := `
		INSERT INTO servers (user_id, name, address, port, edition) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
		name = EXCLUDED.name,
		address = EXCLUDED.address,
		port = EXCLUDED.port,
		edition = EXCLUDED.edition,
		updated_at = NOW()
		RETURNING id, user_id, name, address, port, edition, created_at, updated_at`

	var res models.MinecraftServer

	if err := r.pool.QueryRow(ctx, sql, server.UserID, server.Name, server.Address, server.Port, server.Edition).Scan(
		&res.ID,
		&res.UserID,
		&res.Name,
		&res.Address,
		&res.Port,
		&res.Edition,
		&res.CreatedAt,
		&res.UpdatedAt,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == unique_violation && pgErr.ConstraintName == "servers_endpoint_unique" {
			return nil, fmt.Errorf("server endpoint already exists: %w", domain.ErrAlreadyExists)
		}

		return nil, fmt.Errorf("r.pool.QueryRow: %w", err)
	}

	return &res, nil
}

func (r *WorldsRepo) GetAll(ctx context.Context) ([]models.MinecraftServer, error) {
	sql := `
		SELECT id, user_id, name, address, port, edition, created_at, updated_at
		FROM servers ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("r.pool.Query: %w", err)
	}
	defer rows.Close()

	servers := make([]models.MinecraftServer, 0)

	for rows.Next() {
		var server models.MinecraftServer

		if err := rows.Scan(
			&server.ID,
			&server.UserID,
			&server.Name,
			&server.Address,
			&server.Port,
			&server.Edition,
			&server.CreatedAt,
			&server.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("rows.Scan: %w", err)
		}

		servers = append(servers, server)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows.Err: %w", err)
	}

	return servers, nil
}

func (r *WorldsRepo) GetByID(ctx context.Context, id string) (*models.MinecraftServer, error) {
	sql := ` SELECT id, user_id, name, address, port, edition, created_at, updated_at FROM servers WHERE id = $1 `

	var server models.MinecraftServer

	if err := r.pool.QueryRow(ctx, sql, id).Scan(
		&server.ID,
		&server.UserID,
		&server.Name,
		&server.Address,
		&server.Port,
		&server.Edition,
		&server.CreatedAt,
		&server.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}

		return nil, fmt.Errorf("r.pool.QueryRow: %w", err)
	}

	return &server, nil
}

func (r *WorldsRepo) Delete(ctx context.Context, id, userID string) error {
	sql := ` DELETE FROM servers WHERE id = $1 AND user_id = $2`

	tag, err := r.pool.Exec(ctx, sql, id, userID)
	if err != nil {
		return fmt.Errorf("r.pool.Exec: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
