package acl

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	pkgacl "go-fiber/pkg/acl"
	"go-fiber/pkg/db"
)

const (
	dbTableGroups        = "acl_groups"
	dbTableMembers       = "acl_group_members"
	dbTableInvitations   = "acl_invitations"
	dbTableDeleteRequest = "acl_delete_requests"
	dbTableActivityLogs  = "acl_activity_logs"
)

// QueryObserver records DB query metrics.
type QueryObserver interface {
	ObserveQuery(operation, table string, duration time.Duration, err error)
}

// Repository handles all ACL database operations.
type Repository struct {
	db       db.DB
	observer QueryObserver
}

// NewRepository returns a new Repository.
func NewRepository(database db.DB, observer QueryObserver) *Repository {
	return &Repository{db: database, observer: observer}
}

// dbScanner abstracts *sql.Row and *sql.Rows for shared scan helpers.
type dbScanner interface {
	Scan(dest ...any) error
}

type observedScanner struct {
	inner    dbScanner
	observe  func(error)
	observed bool
}

func (s *observedScanner) Scan(dest ...any) error {
	err := s.inner.Scan(dest...)
	if s.observe != nil && !s.observed {
		s.observed = true
		s.observe(err)
	}
	return err
}

// --- scan helpers ---

func scanGroup(s dbScanner) (*Group, error) {
	var g Group
	var deletedAt sql.NullTime
	if err := s.Scan(
		&g.ID, &g.ResourceType, &g.ResourceID, &g.Name, &g.AccessLevel,
		&g.CreatedBy, &g.CreatedAt, &g.UpdatedAt, &deletedAt,
	); err != nil {
		return nil, err
	}
	if deletedAt.Valid {
		g.DeletedAt = &deletedAt.Time
	}
	return &g, nil
}

func scanMember(s dbScanner) (*GroupMember, error) {
	var m GroupMember
	var sourceID sql.NullString
	var removedAt sql.NullTime
	if err := s.Scan(
		&m.ID, &m.GroupID, &m.UserID, &m.SourceType,
		&sourceID, &m.CreatedAt, &removedAt,
	); err != nil {
		return nil, err
	}
	if sourceID.Valid {
		m.SourceID = &sourceID.String
	}
	if removedAt.Valid {
		m.RemovedAt = &removedAt.Time
	}
	return &m, nil
}

func scanInvitation(s dbScanner) (*Invitation, error) {
	var inv Invitation
	var consumedAt sql.NullTime
	var acceptedBy sql.NullString
	if err := s.Scan(
		&inv.ID, &inv.GroupID, &inv.CreatedBy, &inv.CodeHash,
		&inv.CodeExpiresAt, &inv.Status,
		&consumedAt, &acceptedBy,
		&inv.CreatedAt, &inv.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if consumedAt.Valid {
		inv.ConsumedAt = &consumedAt.Time
	}
	if acceptedBy.Valid {
		inv.AcceptedBy = &acceptedBy.String
	}
	return &inv, nil
}

func scanDeleteRequest(s dbScanner) (*DeleteRequest, error) {
	var dr DeleteRequest
	var reason sql.NullString
	var approvedBy sql.NullString
	var approvedAt sql.NullTime
	var rejectedBy sql.NullString
	var rejectedAt sql.NullTime
	var executedAt sql.NullTime
	var expiresAt sql.NullTime
	if err := s.Scan(
		&dr.ID, &dr.ResourceType, &dr.ResourceID, &dr.RequestedBy,
		&dr.Status, &reason,
		&dr.CreatedAt, &dr.UpdatedAt,
		&approvedBy, &approvedAt,
		&rejectedBy, &rejectedAt,
		&executedAt, &expiresAt,
	); err != nil {
		return nil, err
	}
	if reason.Valid {
		dr.Reason = &reason.String
	}
	if approvedBy.Valid {
		dr.ApprovedBy = &approvedBy.String
	}
	if approvedAt.Valid {
		dr.ApprovedAt = &approvedAt.Time
	}
	if rejectedBy.Valid {
		dr.RejectedBy = &rejectedBy.String
	}
	if rejectedAt.Valid {
		dr.RejectedAt = &rejectedAt.Time
	}
	if executedAt.Valid {
		dr.ExecutedAt = &executedAt.Time
	}
	if expiresAt.Valid {
		dr.ExpiresAt = &expiresAt.Time
	}
	return &dr, nil
}

func scanActivityLog(s dbScanner) (*ActivityLog, error) {
	var log ActivityLog
	var actorID sql.NullString
	var correlationID sql.NullString
	if err := s.Scan(
		&log.ID, &actorID, &log.EntityType, &log.EntityID,
		&log.Action, &log.ResourceType, &log.ResourceID,
		&log.Metadata, &correlationID, &log.CreatedAt,
	); err != nil {
		return nil, err
	}
	if actorID.Valid {
		log.ActorUserID = &actorID.String
	}
	if correlationID.Valid {
		log.CorrelationID = &correlationID.String
	}
	return &log, nil
}

// --- Group ---

func (r *Repository) CreateGroup(ctx context.Context, g *Group) (*Group, error) {
	row := r.queryRow(ctx, "INSERT", dbTableGroups, queryInsertGroup,
		g.ID, g.ResourceType, g.ResourceID, g.Name, g.AccessLevel, g.CreatedBy)
	result, err := scanGroup(row)
	if err != nil {
		if isDuplicateKey(err) {
			return nil, pkgacl.ErrGroupNameDuplicate
		}
		return nil, wrapInternal(err)
	}
	return result, nil
}

func (r *Repository) GetGroupByID(ctx context.Context, id string) (*Group, error) {
	row := r.queryRow(ctx, "SELECT", dbTableGroups, queryGetGroupByID, id)
	g, err := scanGroup(row)
	if isNotFoundLookup(err) {
		return nil, pkgacl.ErrNotFound
	}
	if err != nil {
		return nil, wrapInternal(err)
	}
	return g, nil
}

func (r *Repository) ListGroupsByResource(ctx context.Context, resourceType, resourceID string) ([]*Group, error) {
	rows, err := r.query(ctx, "SELECT", dbTableGroups, queryListGroupsByResource, resourceType, resourceID)
	if err != nil {
		return nil, wrapInternal(err)
	}
	defer rows.Close()

	var groups []*Group
	for rows.Next() {
		g, err := scanGroup(rows)
		if err != nil {
			return nil, wrapInternal(err)
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

func (r *Repository) UpdateGroup(ctx context.Context, id, name, accessLevel string) (*Group, error) {
	row := r.queryRow(ctx, "UPDATE", dbTableGroups, queryUpdateGroup, id, name, accessLevel)
	g, err := scanGroup(row)
	if isNotFoundLookup(err) {
		return nil, pkgacl.ErrNotFound
	}
	if err != nil {
		if isDuplicateKey(err) {
			return nil, pkgacl.ErrGroupNameDuplicate
		}
		return nil, wrapInternal(err)
	}
	return g, nil
}

func (r *Repository) SoftDeleteGroup(ctx context.Context, groupID string) error {
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		if _, err := r.txExec(ctx, tx, "UPDATE", dbTableInvitations, queryCancelPendingInvitesByGroup, groupID); err != nil {
			if isInvalidUUID(err) {
				return pkgacl.ErrNotFound
			}
			return wrapInternal(err)
		}
		if _, err := r.txExec(ctx, tx, "UPDATE", dbTableMembers, querySoftRemoveMembersByGroup, groupID); err != nil {
			if isInvalidUUID(err) {
				return pkgacl.ErrNotFound
			}
			return wrapInternal(err)
		}
		res, err := r.txExec(ctx, tx, "UPDATE", dbTableGroups, querySoftDeleteGroup, groupID)
		if err != nil {
			if isInvalidUUID(err) {
				return pkgacl.ErrNotFound
			}
			return wrapInternal(err)
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return pkgacl.ErrNotFound
		}
		return nil
	})
}

// --- GroupMember ---

func (r *Repository) InsertMember(ctx context.Context, m *GroupMember) (*GroupMember, error) {
	row := r.queryRow(ctx, "INSERT", dbTableMembers, queryInsertMember,
		m.ID, m.GroupID, m.UserID, m.SourceType, nullStr(m.SourceID))
	result, err := scanMember(row)
	if err != nil {
		if isDuplicateKey(err) {
			return nil, pkgacl.ErrMemberAlreadyExists
		}
		return nil, wrapInternal(err)
	}
	return result, nil
}

func (r *Repository) ListMembersByGroup(ctx context.Context, groupID string) ([]*GroupMember, error) {
	rows, err := r.query(ctx, "SELECT", dbTableMembers, queryListMembersByGroup, groupID)
	if err != nil {
		return nil, wrapInternal(err)
	}
	defer rows.Close()

	var members []*GroupMember
	for rows.Next() {
		m, err := scanMember(rows)
		if err != nil {
			return nil, wrapInternal(err)
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (r *Repository) GetMemberByID(ctx context.Context, memberID string) (*GroupMember, error) {
	row := r.queryRow(ctx, "SELECT", dbTableMembers, queryGetMemberByID, memberID)
	m, err := scanMember(row)
	if isNotFoundLookup(err) {
		return nil, pkgacl.ErrNotFound
	}
	if err != nil {
		return nil, wrapInternal(err)
	}
	return m, nil
}

func (r *Repository) SoftRemoveMember(ctx context.Context, memberID, groupID string) error {
	res, err := r.exec(ctx, "UPDATE", dbTableMembers, querySoftRemoveMember, memberID, groupID)
	if err != nil {
		if isInvalidUUID(err) {
			return pkgacl.ErrNotFound
		}
		return wrapInternal(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return pkgacl.ErrNotFound
	}
	return nil
}

// --- Invitation ---

func (r *Repository) InsertInvitation(ctx context.Context, inv *Invitation) (*Invitation, error) {
	row := r.queryRow(ctx, "INSERT", dbTableInvitations, queryInsertInvitation,
		inv.ID, inv.GroupID, inv.CreatedBy, inv.CodeHash, inv.CodeExpiresAt)
	result, err := scanInvitation(row)
	if err != nil {
		return nil, wrapInternal(err)
	}
	return result, nil
}

func (r *Repository) GetInvitationByCodeHash(ctx context.Context, hash string) (*Invitation, error) {
	row := r.queryRow(ctx, "SELECT", dbTableInvitations, queryGetInvitationByCodeHash, hash)
	inv, err := scanInvitation(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, pkgacl.ErrNotFound
	}
	if err != nil {
		return nil, wrapInternal(err)
	}
	return inv, nil
}

func (r *Repository) GetInvitationByID(ctx context.Context, id string) (*Invitation, error) {
	row := r.queryRow(ctx, "SELECT", dbTableInvitations, queryGetInvitationByID, id)
	inv, err := scanInvitation(row)
	if isNotFoundLookup(err) {
		return nil, pkgacl.ErrNotFound
	}
	if err != nil {
		return nil, wrapInternal(err)
	}
	return inv, nil
}

func (r *Repository) ListInvitationsByResource(ctx context.Context, resourceType, resourceID string) ([]*Invitation, error) {
	rows, err := r.query(ctx, "SELECT", dbTableInvitations, queryListInvitationsByResource, resourceType, resourceID)
	if err != nil {
		return nil, wrapInternal(err)
	}
	defer rows.Close()

	var invitations []*Invitation
	for rows.Next() {
		inv, err := scanInvitation(rows)
		if err != nil {
			return nil, wrapInternal(err)
		}
		invitations = append(invitations, inv)
	}
	return invitations, rows.Err()
}

// AcceptInvitationTx runs the accept-invitation flow inside a single transaction.
// It locks the invitation row, validates state, inserts the member, and marks the invitation accepted.
func (r *Repository) AcceptInvitationTx(ctx context.Context, codeHash, actorID, memberID string) error {
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		inv, err := scanInvitation(r.txQueryRow(ctx, tx, "SELECT", dbTableInvitations, queryGetInvitationByCodeHashForUpdate, codeHash))
		if errors.Is(err, sql.ErrNoRows) {
			return pkgacl.ErrNotFound
		}
		if err != nil {
			return wrapInternal(err)
		}

		if inv.Status != "pending" {
			if inv.Status == "accepted" {
				return pkgacl.ErrInviteAlreadyConsumed
			}
			return pkgacl.ErrInviteInvalidState
		}

		if time.Now().After(inv.CodeExpiresAt) {
			return pkgacl.ErrInviteExpired
		}

		grp, err := scanGroup(r.txQueryRow(ctx, tx, "SELECT", dbTableGroups, queryGetGroupByID, inv.GroupID))
		if errors.Is(err, sql.ErrNoRows) {
			return pkgacl.ErrInviteInvalidState
		}
		if err != nil {
			return wrapInternal(err)
		}

		if _, err := r.txExec(ctx, tx, "INSERT", dbTableMembers, queryInsertMemberTx,
			memberID, inv.GroupID, actorID, inv.ID,
		); err != nil {
			if isDuplicateKey(err) {
				return pkgacl.ErrMemberAlreadyExists
			}
			return wrapInternal(err)
		}

		if _, err := r.txExec(ctx, tx, "UPDATE", dbTableInvitations, queryAcceptInvitation, inv.ID, actorID); err != nil {
			return wrapInternal(err)
		}

		logID, _ := r.newID()
		_, _ = r.txExec(ctx, tx, "INSERT", dbTableActivityLogs, queryInsertActivityLog,
			logID, actorID, "invitation", inv.ID, "accept",
			grp.ResourceType, grp.ResourceID, []byte(`{}`), nil,
		)

		return nil
	})
}

// RejectInvitationTx rejects a pending invitation by code hash.
func (r *Repository) RejectInvitationTx(ctx context.Context, codeHash string) error {
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		inv, err := scanInvitation(r.txQueryRow(ctx, tx, "SELECT", dbTableInvitations, queryGetInvitationByCodeHashForUpdate, codeHash))
		if errors.Is(err, sql.ErrNoRows) {
			return pkgacl.ErrNotFound
		}
		if err != nil {
			return wrapInternal(err)
		}
		if inv.Status != "pending" {
			return pkgacl.ErrInviteInvalidState
		}
		if _, err := r.txExec(ctx, tx, "UPDATE", dbTableInvitations, queryRejectInvitation, inv.ID); err != nil {
			return wrapInternal(err)
		}
		return nil
	})
}

// CancelInvitationTx cancels a pending invitation by ID.
func (r *Repository) CancelInvitationTx(ctx context.Context, invitationID string) error {
	return r.db.WithTx(ctx, func(tx *sql.Tx) error {
		inv, err := scanInvitation(r.txQueryRow(ctx, tx, "SELECT", dbTableInvitations, queryGetInvitationByIDForUpdate, invitationID))
		if isNotFoundLookup(err) {
			return pkgacl.ErrNotFound
		}
		if err != nil {
			return wrapInternal(err)
		}
		if inv.Status != "pending" {
			return pkgacl.ErrInviteInvalidState
		}
		if _, err := r.txExec(ctx, tx, "UPDATE", dbTableInvitations, queryCancelInvitation, inv.ID); err != nil {
			return wrapInternal(err)
		}
		return nil
	})
}

// --- Authorization ---

func (r *Repository) GetEffectiveAccessLevel(ctx context.Context, userID, resourceType, resourceID string) (string, error) {
	var accessLevel string
	err := r.queryRow(ctx, "SELECT", dbTableMembers, queryGetEffectiveAccessLevel,
		resourceType, resourceID, userID,
	).Scan(&accessLevel)
	if errors.Is(err, sql.ErrNoRows) {
		return "", pkgacl.ErrNotFound
	}
	if err != nil {
		return "", wrapInternal(err)
	}
	return accessLevel, nil
}

func (r *Repository) ListAccessibleResourceAccessLevels(ctx context.Context, userID, resourceType string) ([]ResourceAccessLevel, error) {
	rows, err := r.query(ctx, "SELECT", dbTableMembers, queryListAccessibleResourceAccessLevels, resourceType, userID)
	if err != nil {
		return nil, wrapInternal(err)
	}
	defer rows.Close()

	var results []ResourceAccessLevel
	for rows.Next() {
		var rr ResourceAccessLevel
		if err := rows.Scan(&rr.ResourceID, &rr.EffectiveAccessLevel); err != nil {
			return nil, wrapInternal(err)
		}
		results = append(results, rr)
	}
	return results, rows.Err()
}

// --- Delete requests ---

func (r *Repository) InsertDeleteRequest(ctx context.Context, dr *DeleteRequest) (*DeleteRequest, error) {
	row := r.queryRow(ctx, "INSERT", dbTableDeleteRequest, queryInsertDeleteRequest,
		dr.ID, dr.ResourceType, dr.ResourceID, dr.RequestedBy, dr.Status, nullStr(dr.Reason), nullTime(dr.ExpiresAt),
	)
	created, err := scanDeleteRequest(row)
	if err != nil {
		if isDuplicateKey(err) {
			return nil, pkgacl.ErrDeleteRequestAlreadyOpen
		}
		return nil, wrapInternal(err)
	}
	return created, nil
}

func (r *Repository) GetDeleteRequestByID(ctx context.Context, id string) (*DeleteRequest, error) {
	row := r.queryRow(ctx, "SELECT", dbTableDeleteRequest, queryGetDeleteRequestByID, id)
	dr, err := scanDeleteRequest(row)
	if isNotFoundLookup(err) {
		return nil, pkgacl.ErrNotFound
	}
	if err != nil {
		return nil, wrapInternal(err)
	}
	return dr, nil
}

func (r *Repository) ListDeleteRequestsByResource(ctx context.Context, resourceType, resourceID string, requestedBy, status *string, limit int) ([]*DeleteRequest, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.query(ctx, "SELECT", dbTableDeleteRequest, queryListDeleteRequestsByResource,
		resourceType, resourceID, nullStr(requestedBy), nullStr(status), limit,
	)
	if err != nil {
		return nil, wrapInternal(err)
	}
	defer rows.Close()

	var result []*DeleteRequest
	for rows.Next() {
		dr, err := scanDeleteRequest(rows)
		if err != nil {
			return nil, wrapInternal(err)
		}
		result = append(result, dr)
	}
	return result, rows.Err()
}

func (r *Repository) ApproveDeleteRequestTx(ctx context.Context, requestID, approvedBy string) (*DeleteRequest, error) {
	var out *DeleteRequest
	err := r.db.WithTx(ctx, func(tx *sql.Tx) error {
		dr, err := scanDeleteRequest(r.txQueryRow(ctx, tx, "SELECT", dbTableDeleteRequest, queryGetDeleteRequestByIDForUpdate, requestID))
		if isNotFoundLookup(err) {
			return pkgacl.ErrNotFound
		}
		if err != nil {
			return wrapInternal(err)
		}
		if dr.Status != "pending" {
			return pkgacl.ErrDeleteRequestInvalidState
		}

		res, err := r.txExec(ctx, tx, "UPDATE", dbTableDeleteRequest, queryApproveDeleteRequest, requestID, approvedBy)
		if err != nil {
			return wrapInternal(err)
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return pkgacl.ErrDeleteRequestInvalidState
		}

		updated, err := scanDeleteRequest(r.txQueryRow(ctx, tx, "SELECT", dbTableDeleteRequest, queryGetDeleteRequestByID, requestID))
		if err != nil {
			return wrapInternal(err)
		}
		out = updated
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) RejectDeleteRequestTx(ctx context.Context, requestID, rejectedBy string) (*DeleteRequest, error) {
	var out *DeleteRequest
	err := r.db.WithTx(ctx, func(tx *sql.Tx) error {
		dr, err := scanDeleteRequest(r.txQueryRow(ctx, tx, "SELECT", dbTableDeleteRequest, queryGetDeleteRequestByIDForUpdate, requestID))
		if isNotFoundLookup(err) {
			return pkgacl.ErrNotFound
		}
		if err != nil {
			return wrapInternal(err)
		}
		if dr.Status != "pending" {
			return pkgacl.ErrDeleteRequestInvalidState
		}

		res, err := r.txExec(ctx, tx, "UPDATE", dbTableDeleteRequest, queryRejectDeleteRequest, requestID, rejectedBy)
		if err != nil {
			return wrapInternal(err)
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return pkgacl.ErrDeleteRequestInvalidState
		}

		updated, err := scanDeleteRequest(r.txQueryRow(ctx, tx, "SELECT", dbTableDeleteRequest, queryGetDeleteRequestByID, requestID))
		if err != nil {
			return wrapInternal(err)
		}
		out = updated
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) MarkDeleteRequestExecutedTx(ctx context.Context, requestID string) (*DeleteRequest, error) {
	var out *DeleteRequest
	err := r.db.WithTx(ctx, func(tx *sql.Tx) error {
		if _, err := scanDeleteRequest(r.txQueryRow(ctx, tx, "SELECT", dbTableDeleteRequest, queryGetDeleteRequestByIDForUpdate, requestID)); err != nil {
			if isNotFoundLookup(err) {
				return pkgacl.ErrNotFound
			}
			return wrapInternal(err)
		}
		res, err := r.txExec(ctx, tx, "UPDATE", dbTableDeleteRequest, queryMarkDeleteRequestExecuted, requestID)
		if err != nil {
			return wrapInternal(err)
		}
		n, _ := res.RowsAffected()
		if n == 0 {
			return pkgacl.ErrDeleteRequestInvalidState
		}
		updated, err := scanDeleteRequest(r.txQueryRow(ctx, tx, "SELECT", dbTableDeleteRequest, queryGetDeleteRequestByID, requestID))
		if err != nil {
			return wrapInternal(err)
		}
		out = updated
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// --- Activity log ---

func (r *Repository) InsertActivityLog(ctx context.Context, log *ActivityLog) error {
	_, err := r.exec(ctx, "INSERT", dbTableActivityLogs, queryInsertActivityLog,
		log.ID, nullStr(log.ActorUserID), log.EntityType, log.EntityID,
		log.Action, log.ResourceType, log.ResourceID, log.Metadata, nullStr(log.CorrelationID),
	)
	if err != nil {
		return wrapInternal(err)
	}
	return nil
}

func (r *Repository) ListActivityLogsByResource(ctx context.Context, resourceType, resourceID string, actorUserID, action *string, from, to *time.Time, limit int) ([]*ActivityLog, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := r.query(ctx, "SELECT", dbTableActivityLogs, queryListActivityLogsByResource,
		resourceType, resourceID, nullStr(actorUserID), nullStr(action), nullTime(from), nullTime(to), limit,
	)
	if err != nil {
		return nil, wrapInternal(err)
	}
	defer rows.Close()

	var out []*ActivityLog
	for rows.Next() {
		entry, err := scanActivityLog(rows)
		if err != nil {
			return nil, wrapInternal(err)
		}
		out = append(out, entry)
	}
	return out, rows.Err()
}

// --- helpers ---

func (r *Repository) query(ctx context.Context, operation, table, query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	rows, err := r.db.Raw().QueryContext(ctx, query, args...)
	r.observeQuery(operation, table, start, err)
	return rows, err
}

func (r *Repository) queryRow(ctx context.Context, operation, table, query string, args ...any) dbScanner {
	start := time.Now()
	row := r.db.Raw().QueryRowContext(ctx, query, args...)
	return &observedScanner{
		inner: row,
		observe: func(err error) {
			r.observeQuery(operation, table, start, err)
		},
	}
}

func (r *Repository) exec(ctx context.Context, operation, table, query string, args ...any) (sql.Result, error) {
	start := time.Now()
	result, err := r.db.Raw().ExecContext(ctx, query, args...)
	r.observeQuery(operation, table, start, err)
	return result, err
}

func (r *Repository) txQueryRow(ctx context.Context, tx *sql.Tx, operation, table, query string, args ...any) dbScanner {
	start := time.Now()
	row := tx.QueryRowContext(ctx, query, args...)
	return &observedScanner{
		inner: row,
		observe: func(err error) {
			r.observeQuery(operation, table, start, err)
		},
	}
}

func (r *Repository) txExec(ctx context.Context, tx *sql.Tx, operation, table, query string, args ...any) (sql.Result, error) {
	start := time.Now()
	result, err := tx.ExecContext(ctx, query, args...)
	r.observeQuery(operation, table, start, err)
	return result, err
}

func (r *Repository) observeQuery(operation, table string, start time.Time, err error) {
	if r == nil || r.observer == nil {
		return
	}
	r.observer.ObserveQuery(operation, table, time.Since(start), err)
}

func (r *Repository) newID() (string, error) {
	return newUUID()
}

func isDuplicateKey(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" // unique_violation
	}
	return false
}

func isInvalidUUID(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	if pgErr.Code != "22P02" {
		return false
	}
	return strings.Contains(strings.ToLower(pgErr.Message), "uuid")
}

func isNotFoundLookup(err error) bool {
	return errors.Is(err, sql.ErrNoRows) || isInvalidUUID(err)
}

func wrapInternal(err error) error {
	return &pkgacl.ACLError{
		Code:    "internal-error",
		Message: "Internal server error",
		Cause:   err,
	}
}

func nullStr(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

func nullTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}
