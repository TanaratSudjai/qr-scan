package repository

import (
	"database/sql"
	"qr-scan/internal/models"
)

type MemberRepository struct {
	db *sql.DB
}

func NewMemberRepository(db *sql.DB) *MemberRepository {
	return &MemberRepository{db: db}
}

func (r *MemberRepository) GetByPhoneNumber(phone string) (*models.Member, error) {
	query := `SELECT id, timestamp, fullName, address, province, postalCode, phoneNumber, email, organization, position, responsibility, expectation, COALESCE(count_checkin, 0) FROM member_ai WHERE phoneNumber = ? LIMIT 1`
	row := r.db.QueryRow(query, phone)

	var m models.Member
	var resp sql.NullString
	var exp sql.NullString
	var ts sql.NullString

	err := row.Scan(&m.ID, &ts, &m.FullName, &m.Address, &m.Province, &m.PostalCode, &m.PhoneNumber, &m.Email, &m.Organization, &m.Position, &resp, &exp, &m.CountCheckin)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if resp.Valid {
		m.Responsibility = resp.String
	}
	if exp.Valid {
		m.Expectation = exp.String
	}

	return &m, nil
}

func (r *MemberRepository) IncrementCheckin(id int) error {
	_, err := r.db.Exec(`UPDATE member_ai SET count_checkin = COALESCE(count_checkin, 0) + 1 WHERE id = ?`, id)
	return err
}

func (r *MemberRepository) GetAll() ([]models.Member, error) {
	query := `SELECT id, fullName, phoneNumber, COALESCE(count_checkin, 0) FROM member_ai ORDER BY count_checkin DESC, id ASC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []models.Member
	for rows.Next() {
		var m models.Member
		err := rows.Scan(&m.ID, &m.FullName, &m.PhoneNumber, &m.CountCheckin)
		if err != nil {
			continue
		}
		members = append(members, m)
	}
	return members, nil
}
