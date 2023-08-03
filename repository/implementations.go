package repository

import (
	"context"
	"fmt"

	"github.com/SawitProRecruitment/UserService/generated"
)

func (r *Repository) IsPhoneNumberExists(ctx context.Context, phoneNumber string) (bool, error) {
	var nData int
	sqlStmt := "SELECT count(phone_number) FROM public.user WHERE phone_number = $1"

	if err := r.Db.QueryRowContext(ctx, sqlStmt, phoneNumber).Scan(&nData); err != nil {
		return false, err
	}

	if nData > 0 {
		return true, nil
	}
	return false, nil
}

func (r *Repository) Register(ctx context.Context, regRequest generated.RegistrationRequest,
	salt string) (string, error) {
	var userID string

	tx, err := r.Db.Begin()
	if err != nil {
		return "", err
	}

	if err := tx.QueryRowContext(ctx, `INSERT INTO public.user (full_name, phone_number) VALUES ($1, $2) RETURNING id`,
		regRequest.FullName, regRequest.PhoneNumber).Scan(&userID); err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	_, err = tx.Exec("INSERT INTO public.password (user_id, password, salt) VALUES ($1, $2, $3)",
		userID, regRequest.Password, salt)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return "", err
	}

	return userID, nil
}
