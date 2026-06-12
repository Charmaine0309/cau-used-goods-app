package favorite

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Add(ctx context.Context, userID, productID uint64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin add favorite tx: %w", err)
	}
	defer tx.Rollback()

	var isDeleted bool
	err = tx.QueryRowContext(ctx, `
		SELECT is_deleted
		FROM favorites
		WHERE user_id = ? AND product_id = ?
		FOR UPDATE
	`, userID, productID).Scan(&isDeleted)

	if err == sql.ErrNoRows {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO favorites (user_id, product_id, is_deleted)
			VALUES (?, ?, 0)
		`, userID, productID); err != nil {
			return fmt.Errorf("insert favorite: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE products
			SET favorite_count = favorite_count + 1,
			    update_time = CURRENT_TIMESTAMP
			WHERE id = ? AND is_deleted = 0
		`, productID); err != nil {
			return fmt.Errorf("increment favorite count: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("query favorite: %w", err)
	} else if isDeleted {
		if _, err := tx.ExecContext(ctx, `
			UPDATE favorites
			SET is_deleted = 0
			WHERE user_id = ? AND product_id = ?
		`, userID, productID); err != nil {
			return fmt.Errorf("restore favorite: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE products
			SET favorite_count = favorite_count + 1,
			    update_time = CURRENT_TIMESTAMP
			WHERE id = ? AND is_deleted = 0
		`, productID); err != nil {
			return fmt.Errorf("increment favorite count: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit add favorite tx: %w", err)
	}
	return nil
}

func (r *Repository) Remove(ctx context.Context, userID, productID uint64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin remove favorite tx: %w", err)
	}
	defer tx.Rollback()

	var isDeleted bool
	err = tx.QueryRowContext(ctx, `
		SELECT is_deleted
		FROM favorites
		WHERE user_id = ? AND product_id = ?
		FOR UPDATE
	`, userID, productID).Scan(&isDeleted)

	if err == sql.ErrNoRows {
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit remove favorite tx: %w", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("query favorite: %w", err)
	}

	if !isDeleted {
		if _, err := tx.ExecContext(ctx, `
			UPDATE favorites
			SET is_deleted = 1
			WHERE user_id = ? AND product_id = ?
		`, userID, productID); err != nil {
			return fmt.Errorf("remove favorite: %w", err)
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE products
			SET favorite_count = CASE
			        WHEN favorite_count > 0 THEN favorite_count - 1
			        ELSE 0
			    END,
			    update_time = CURRENT_TIMESTAMP
			WHERE id = ? AND is_deleted = 0
		`, productID); err != nil {
			return fmt.Errorf("decrement favorite count: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit remove favorite tx: %w", err)
	}
	return nil
}

func (r *Repository) IsFavorited(ctx context.Context, userID, productID uint64) (bool, error) {
	query := `SELECT COUNT(*) FROM favorites WHERE user_id = ? AND product_id = ? AND is_deleted = 0`
	var count int
	if err := r.db.QueryRowContext(ctx, query, userID, productID).Scan(&count); err != nil {
		return false, fmt.Errorf("check favorite: %w", err)
	}
	return count > 0, nil
}

func (r *Repository) ListByUser(ctx context.Context, userID uint64, page, pageSize int) ([]FavoriteDetail, int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM favorites WHERE user_id = ? AND is_deleted = 0`
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count favorites: %w", err)
	}

	query := `
		SELECT f.id, f.user_id, f.product_id, f.create_time,
			p.title, p.price, p.status, pi.image_url, u.nickname
		FROM favorites f
		JOIN products p ON p.id = f.product_id
		LEFT JOIN product_images pi ON pi.product_id = p.id AND pi.sort_order = 0
		LEFT JOIN users u ON u.id = p.seller_id
		WHERE f.user_id = ? AND f.is_deleted = 0
		ORDER BY f.create_time DESC
		LIMIT ? OFFSET ?
	`
	rows, err := r.db.QueryContext(ctx, query, userID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list favorites: %w", err)
	}
	defer rows.Close()

	var items []FavoriteDetail
	for rows.Next() {
		var fd FavoriteDetail
		var img, nick sql.NullString
		err := rows.Scan(&fd.ID, &fd.UserID, &fd.ProductID, &fd.CreateTime,
			&fd.ProductTitle, &fd.ProductPrice, &fd.ProductStatus, &img, &nick)
		if err != nil {
			return nil, 0, fmt.Errorf("scan favorite: %w", err)
		}
		if img.Valid {
			fd.ProductImage = &img.String
		}
		if nick.Valid {
			fd.SellerNickname = &nick.String
		}
		items = append(items, fd)
	}
	return items, total, nil
}
