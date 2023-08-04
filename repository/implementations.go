package repository

import (
	"context"
	"errors"

	"github.com/SawitProRecruitment/UserService/generated"
	"golang.org/x/crypto/bcrypt"
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
		return "", err
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO public.login (user_id, success_login) VALUES ($1, $2)", userID, 0)
	if err != nil {
		return "", err
	}

	_, err = tx.ExecContext(ctx, "INSERT INTO public.password (user_id, password, salt) VALUES ($1, $2, $3)",
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

func (r *Repository) Login(ctx context.Context, loginRequest generated.LoginRequest) (string, error) {
	tx, err := r.Db.Begin()
	if err != nil {
		return "", err
	}

	var userID string
	if err := tx.QueryRowContext(ctx, "SELECT id FROM public.user WHERE phone_number = $1",
		loginRequest.PhoneNumber).Scan(&userID); err != nil {
		return "", err
	}

	if userID == "" {
		return "", errors.New("User not found.")
	}

	var hashedPassword, salt string
	if err := tx.QueryRowContext(ctx, "SELECT password, salt FROM public.password WHERE user_id = $1",
		userID).Scan(&hashedPassword, &salt); err != nil {
		return "", err
	}

	providedPasswordWithSalt := []byte(loginRequest.Password + salt)

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), providedPasswordWithSalt)
	if err == nil {
		_, err := tx.ExecContext(ctx,
			"UPDATE public.login SET success_login = success_login + 1 WHERE user_id = $1", userID)
		if err != nil {
			return "", err
		}
	} else {
		return "", errors.New("Wrong password")
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return "", err
	}

	return userID, nil
}

func (r *Repository) GetUserProfile(ctx context.Context, userID string) (generated.UserProfile, error) {
	var userProfile generated.UserProfile
	sqlStmt := "SELECT full_name, phone_number FROM public.user WHERE id = $1"
	if err := r.Db.QueryRowContext(ctx, sqlStmt, userID).Scan(&userProfile.FullName, &userProfile.PhoneNumber); err != nil {
		return generated.UserProfile{}, err
	}
	return userProfile, nil
}
