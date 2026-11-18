package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"mindora/backend/internal/models"
)

var (
	ErrSessionNotFound          = errors.New("authentication session not found")
	ErrSessionAlreadyRotated     = errors.New("authentication session has already been rotated")
	ErrConcurrentRefreshConflict = errors.New("concurrent refresh in progress; please retry with current token")
	ErrIdempotencyExpired       = errors.New("refresh idempotency window expired")
	ErrTokenReuseDetected       = errors.New("refresh token reuse detected; all associated sessions have been revoked")
)

// RotateDistributedParams encapsulates parameters and callbacks for atomic distributed rotation.
type RotateDistributedParams struct {
	TokenHash   string
	OperationID string
	ResponseTTL time.Duration

	// OnRotate is called under the exclusive advisory lock when a fresh rotation is required.
	// It receives the current active session (with User and Role preloaded).
	OnRotate func(oldSession *models.AuthSession) (newSession *models.AuthSession, responsePayload []byte, err error)

	// EncryptFunc encrypts responsePayload with authenticated encryption (AES-256-GCM).
	EncryptFunc func(plaintext []byte) (string, error)

	// DecryptFunc decrypts cached encrypted response.
	DecryptFunc func(ciphertextB64 string) ([]byte, error)
}

type SessionRepository interface {
	CreateSession(ctx context.Context, session *models.AuthSession) error
	FindByTokenHash(ctx context.Context, tokenHash string) (*models.AuthSession, error)
	RotateSession(ctx context.Context, oldSessionID uuid.UUID, newSession *models.AuthSession) error
	RotateSessionDistributed(ctx context.Context, params RotateDistributedParams) ([]byte, error)
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error
	RevokeFamily(ctx context.Context, familyID uuid.UUID) error
	HasActiveSessionInFamily(ctx context.Context, familyID uuid.UUID) (bool, error)
}

type GormSessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *GormSessionRepository {
	return &GormSessionRepository{db: db}
}

func (r *GormSessionRepository) CreateSession(ctx context.Context, session *models.AuthSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *GormSessionRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*models.AuthSession, error) {
	var session models.AuthSession
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("User.Role").
		Where("token_hash = ?", tokenHash).
		First(&session).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	return &session, nil
}

// RotateSession marks the old session as revoked and inserts the new session inside an atomic transaction.
func (r *GormSessionRepository) RotateSession(ctx context.Context, oldSessionID uuid.UUID, newSession *models.AuthSession) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		res := tx.Model(&models.AuthSession{}).
			Where("id = ? AND revoked_at IS NULL", oldSessionID).
			Update("revoked_at", now)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrSessionAlreadyRotated
		}

		return tx.Create(newSession).Error
	})
}

// RotateSessionDistributed coordinates token rotation across multiple independent backend instances
// using PostgreSQL Transactional Advisory Locks (`pg_advisory_xact_lock`) and cryptographically bound idempotency.
// It deterministically detects in-flight lock contention using `pg_try_advisory_xact_lock` to distinguish
// concurrent collisions from sequential replay attacks without relying on arbitrary sleep windows.
func (r *GormSessionRepository) RotateSessionDistributed(ctx context.Context, params RotateDistributedParams) ([]byte, error) {
	var (
		finalPayload []byte
		isTokenReuse bool
	)

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Determine in-flight concurrency contention via pg_try_advisory_xact_lock
		// A 64-bit key space is constructed using (hashtext(tokenHash), hashtext(reverse(tokenHash))).
		var acquiredImmediately bool
		if err := tx.Raw("SELECT pg_try_advisory_xact_lock(hashtext(?), hashtext(reverse(?)))", params.TokenHash, params.TokenHash).Scan(&acquiredImmediately).Error; err != nil {
			return fmt.Errorf("failed to probe transaction advisory lock: %w", err)
		}

		hadLockContention := !acquiredImmediately

		// If another transaction is in-flight on this token, wait for it to commit or abort
		if hadLockContention {
			if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?), hashtext(reverse(?)))", params.TokenHash, params.TokenHash).Error; err != nil {
				return fmt.Errorf("failed to acquire blocking advisory lock: %w", err)
			}
		}

		// 2. Query session under exclusive transaction lock
		var session models.AuthSession
		err := tx.
			Preload("User").
			Preload("User.Role").
			Where("token_hash = ?", params.TokenHash).
			First(&session).Error

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrSessionNotFound
			}
			return err
		}

		now := time.Now().UTC()

		// 3. Evaluate if session was already rotated / revoked
		if session.RevokedAt != nil {
			// Case A: IDEMPOTENT RETRY with the exact SAME operation identifier
			// (Works both for concurrent requests sharing the same op ID and sequential network retries)
			if params.OperationID != "" && session.RotationOpID != nil && *session.RotationOpID == params.OperationID {
				if session.ResponseExpiresAt != nil && session.ResponseExpiresAt.After(now) && session.EncryptedResponse != nil {
					if params.DecryptFunc == nil {
						return errors.New("decrypt function required for idempotent response retrieval")
					}
					decrypted, err := params.DecryptFunc(*session.EncryptedResponse)
					if err != nil {
						return fmt.Errorf("failed to decrypt cached idempotent response: %w", err)
					}
					finalPayload = decrypted
					return nil
				}
				return ErrIdempotencyExpired
			}

			// Case B: DIFFERENT or EMPTY operation identifier under confirmed in-flight lock contention
			if hadLockContention {
				// A concurrent request won the rotation race while this request was waiting on the lock.
				// Reject without destroying the newly issued active family!
				return ErrConcurrentRefreshConflict
			}

			// Case C: Session already revoked, no lock contention, and operation ID does not match.
			// This is with 100% certainty a SEQUENTIAL REPLAY ATTACK by an adversary presenting an old token!
			// We revoke the family inside the transaction and return nil so GORM commits the revocation.
			_ = tx.Model(&models.AuthSession{}).
				Where("family_id = ? AND revoked_at IS NULL", session.FamilyID).
				Update("revoked_at", now).Error

			isTokenReuse = true
			return nil
		}

		// 4. Session is unrevoked: execute fresh rotation callback
		newSession, payload, err := params.OnRotate(&session)
		if err != nil {
			return err
		}

		var encrypted string
		var responseExpiresAt *time.Time
		if params.EncryptFunc != nil && len(payload) > 0 && params.ResponseTTL > 0 {
			enc, err := params.EncryptFunc(payload)
			if err != nil {
				return fmt.Errorf("failed to encrypt response for idempotency storage: %w", err)
			}
			encrypted = enc
			exp := now.Add(params.ResponseTTL)
			responseExpiresAt = &exp
		}

		// 5. Insert replacement session first so that foreign key fk_auth_sessions_replaced_by is satisfied
		if err := tx.Create(newSession).Error; err != nil {
			return fmt.Errorf("failed to insert replacement session: %w", err)
		}

		// 6. Atomically update old session with rotation metadata
		updates := map[string]interface{}{
			"revoked_at":          now,
			"replaced_by_id":      newSession.ID,
			"rotation_op_id":      params.OperationID,
			"encrypted_response":  encrypted,
			"response_expires_at": responseExpiresAt,
		}

		res := tx.Model(&models.AuthSession{}).
			Where("id = ? AND revoked_at IS NULL", session.ID).
			Updates(updates)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrSessionAlreadyRotated
		}

		finalPayload = payload
		return nil
	})

	if err != nil {
		return nil, err
	}

	if isTokenReuse {
		return nil, ErrTokenReuseDetected
	}

	return finalPayload, nil
}

// RevokeSession revokes a specific session.
func (r *GormSessionRepository) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&models.AuthSession{}).
		Where("id = ? AND revoked_at IS NULL", sessionID).
		Update("revoked_at", now).Error
}

// RevokeFamily revokes all active sessions belonging to the same token lineage when reuse is detected.
func (r *GormSessionRepository) RevokeFamily(ctx context.Context, familyID uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&models.AuthSession{}).
		Where("family_id = ? AND revoked_at IS NULL", familyID).
		Update("revoked_at", now).Error
}

// HasActiveSessionInFamily checks if there are any unrevoked sessions belonging to this family.
func (r *GormSessionRepository) HasActiveSessionInFamily(ctx context.Context, familyID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.AuthSession{}).
		Where("family_id = ? AND revoked_at IS NULL", familyID).
		Count(&count).Error
	return count > 0, err
}
