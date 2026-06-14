package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func (r *Repository) FindPublicProfile(ctx context.Context, userID uint64) (*PublicProfile, error) {
	var item PublicProfile
	var authStatus, accountStatus string
	if err := r.db.QueryRowContext(ctx, `
		SELECT id, nickname, avatar_url, auth_status, account_status
		FROM users
		WHERE id = ?
		  AND is_deleted = 0
		  AND account_status IN ('NORMAL', 'DISABLED')
		LIMIT 1
	`, userID).Scan(&item.ID, &item.Nickname, &item.AvatarURL, &authStatus, &accountStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find public profile: %w", err)
	}
	item.AuthStatus = authStatus
	item.TradeAvailable = accountStatus == accountStatusNormal
	return &item, nil
}

func (r *Repository) FindRestriction(ctx context.Context, userID uint64) (*Restriction, error) {
	var item Restriction
	if err := r.db.QueryRowContext(ctx, `SELECT account_status FROM users WHERE id = ? LIMIT 1`, userID).Scan(&item.AccountStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find account restriction: %w", err)
	}
	if item.AccountStatus != accountStatusDisabled && item.AccountStatus != accountStatusBanned {
		return &item, nil
	}

	var relatedID sql.NullInt64
	var operationType, reason, relatedType, createTime sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT operation_type, description, related_type, related_id,
		       DATE_FORMAT(create_time, '%Y-%m-%d %H:%i:%s')
		FROM admin_logs
		WHERE target_type = 'USER'
		  AND target_id = ?
		  AND operation_type IN ('USER_DISABLE', 'USER_BAN')
		ORDER BY create_time DESC, id DESC
		LIMIT 1
	`, userID).Scan(&operationType, &reason, &relatedType, &relatedID, &createTime)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &item, nil
		}
		return nil, fmt.Errorf("find restriction reason: %w", err)
	}
	item.OperationType = nullStringPtr(operationType)
	item.Reason = nullStringPtr(reason)
	item.RelatedType = nullStringPtr(relatedType)
	item.CreateTime = nullStringPtr(createTime)
	if relatedID.Valid {
		value := uint64(relatedID.Int64)
		item.RelatedID = &value
	}
	return &item, nil
}
