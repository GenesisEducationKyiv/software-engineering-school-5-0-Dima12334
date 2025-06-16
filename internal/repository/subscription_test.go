package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
	customErrors "weather_forecast_sub/pkg/errors"
	"weather_forecast_sub/testutils"

	"github.com/lib/pq"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"weather_forecast_sub/internal/domain"
	"weather_forecast_sub/internal/repository"
)

func TestSubscriptionRepo(t *testing.T) {
	t.Run("Create", testSubscriptionRepoCreate)
	t.Run("Create Error", testSubscriptionRepoCreateError)
	t.Run("Create Duplication Error", testSubscriptionRepoCreateDuplicationError)
	t.Run("GetByToken", testSubscriptionRepoGetByToken)
	t.Run("GetByToken Not Found", testSubscriptionRepoGetByTokenNotFound)
	t.Run("GetByToken DB Error", testSubscriptionRepoGetByTokenDBError)
	t.Run("Confirm", testSubscriptionRepoConfirm)
	t.Run("Confirm Error", testSubscriptionRepoConfirmError)
	t.Run("SetLastSentAt", testSubscriptionRepoSetLastSentAt)
	t.Run("SetLastSentAt Error", testSubscriptionRepoSetLastSentAtError)
	t.Run("Delete", testSubscriptionRepoDelete)
	t.Run("Delete Error", testSubscriptionRepoDeleteError)
	t.Run("GetConfirmedByFrequency", testSubscriptionRepoGetConfirmedByFrequency)
	t.Run("GetConfirmedByFrequency Error", testSubscriptionRepoGetConfirmedByFrequencyError)
}

func testSubscriptionRepoCreate(t *testing.T) {
	db, mock := testutils.SetupMockDB(t)
	defer func() {
		mock.ExpectClose()
		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	repo := repository.NewSubscriptionRepo(db)

	now := time.Now()
	sub := domain.Subscription{
		CreatedAt:  time.Now(),
		Email:      "test@example.com",
		City:       "Kyiv",
		Token:      "token123",
		Frequency:  "daily",
		Confirmed:  false,
		LastSentAt: &now,
	}

	mock.ExpectExec("INSERT INTO subscriptions").
		WithArgs(sub.CreatedAt, sub.Email, sub.City, sub.Token, sub.Frequency, sub.Confirmed, sub.LastSentAt).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Create(context.Background(), sub)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func testSubscriptionRepoCreateError(t *testing.T) {
	db, mock := testutils.SetupMockDB(t)
	defer func() {
		mock.ExpectClose()
		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	repo := repository.NewSubscriptionRepo(db)

	sub := domain.Subscription{
		CreatedAt:  time.Now(),
		Email:      "error@example.com",
		City:       "Kyiv",
		Token:      "errorToken",
		Frequency:  "daily",
		Confirmed:  false,
		LastSentAt: nil,
	}

	mock.ExpectExec("INSERT INTO subscriptions").
		WithArgs(sub.CreatedAt, sub.Email, sub.City, sub.Token, sub.Frequency, sub.Confirmed, sub.LastSentAt).
		WillReturnError(errors.New("some db error"))

	err := repo.Create(context.Background(), sub)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "some db error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func testSubscriptionRepoCreateDuplicationError(t *testing.T) {
	db, mock := testutils.SetupMockDB(t)
	defer func() {
		mock.ExpectClose()
		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	repo := repository.NewSubscriptionRepo(db)

	sub := domain.Subscription{
		CreatedAt:  time.Now(),
		Email:      "error@example.com",
		City:       "Kyiv",
		Token:      "errorToken",
		Frequency:  "daily",
		Confirmed:  false,
		LastSentAt: nil,
	}

	duplicateError := pq.Error{Code: customErrors.PgUniqueViolationCode}
	mock.ExpectExec("INSERT INTO subscriptions").
		WithArgs(sub.CreatedAt, sub.Email, sub.City, sub.Token, sub.Frequency, sub.Confirmed, sub.LastSentAt).
		WillReturnError(&duplicateError)

	err := repo.Create(context.Background(), sub)
	assert.Error(t, err)
	assert.ErrorIs(t, err, customErrors.ErrSubscriptionAlreadyExists)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func testSubscriptionRepoGetByToken(t *testing.T) {
	db, mock := testutils.SetupMockDB(t)
	defer func() {
		mock.ExpectClose()
		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	repo := repository.NewSubscriptionRepo(db)

	token := "test-token"
	now := time.Now()
	expected := domain.Subscription{
		ID:         "68501cb6-0bf0-800e-81ba-bae3763ecdd2",
		CreatedAt:  time.Now(),
		Email:      "user@example.com",
		City:       "Kyiv",
		Token:      token,
		Frequency:  "daily",
		Confirmed:  true,
		LastSentAt: &now,
	}

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "email", "city", "token", "frequency", "confirmed", "last_sent_at",
	}).AddRow(
		expected.ID, expected.CreatedAt, expected.Email, expected.City, expected.Token,
		expected.Frequency, expected.Confirmed, expected.LastSentAt,
	)

	mock.ExpectQuery("SELECT .* FROM subscriptions WHERE token =").
		WithArgs(token).
		WillReturnRows(rows)

	got, err := repo.GetByToken(context.Background(), token)
	assert.NoError(t, err)
	assert.Equal(t, expected.Email, got.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func testSubscriptionRepoGetByTokenNotFound(t *testing.T) {
	db, mock := testutils.SetupMockDB(t)
	defer func() {
		mock.ExpectClose()
		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	repo := repository.NewSubscriptionRepo(db)

	token := "missing-token"

	mock.ExpectQuery("SELECT .* FROM subscriptions WHERE token =").
		WithArgs(token).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.GetByToken(context.Background(), token)
	assert.ErrorIs(t, err, customErrors.ErrSubscriptionNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func testSubscriptionRepoGetByTokenDBError(t *testing.T) {
	db, mock := testutils.SetupMockDB(t)
	defer func() {
		mock.ExpectClose()
		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	repo := repository.NewSubscriptionRepo(db)

	token := "error-token"

	mock.ExpectQuery("SELECT .* FROM subscriptions WHERE token =").
		WithArgs(token).
		WillReturnError(errors.New("db failure"))

	_, err := repo.GetByToken(context.Background(), token)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func testSubscriptionRepoConfirm(t *testing.T) {
	db, mock := testutils.SetupMockDB(t)
	defer func() {
		mock.ExpectClose()
		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	repo := repository.NewSubscriptionRepo(db)

	mock.ExpectExec("UPDATE subscriptions SET confirmed = true").
		WithArgs("token123").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Confirm(context.Background(), "token123")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func testSubscriptionRepoConfirmError(t *testing.T) {
	db, mock := testutils.SetupMockDB(t)
	defer func() {
		mock.ExpectClose()
		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	repo := repository.NewSubscriptionRepo(db)

	mock.ExpectExec("UPDATE subscriptions SET confirmed = true").
		WithArgs("token123").
		WillReturnError(errors.New("update error"))

	err := repo.Confirm(context.Background(), "token123")
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func testSubscriptionRepoSetLastSentAt(t *testing.T) {
	db, mock := testutils.SetupMockDB(t)
	defer func() {
		mock.ExpectClose()
		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	repo := repository.NewSubscriptionRepo(db)

	now := time.Now()
	tokens := []string{"token1", "token2"}

	mock.ExpectExec("UPDATE subscriptions SET last_sent_at").
		WithArgs(now, pq.Array(tokens)).
		WillReturnResult(sqlmock.NewResult(1, 2))

	err := repo.SetLastSentAt(now, tokens)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func testSubscriptionRepoSetLastSentAtError(t *testing.T) {
	db, mock := testutils.SetupMockDB(t)
	defer func() {
		mock.ExpectClose()
		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	repo := repository.NewSubscriptionRepo(db)

	now := time.Now()
	tokens := []string{"token1", "token2"}

	mock.ExpectExec("UPDATE subscriptions SET last_sent_at").
		WithArgs(now, pq.Array(tokens)).
		WillReturnError(errors.New("update error"))

	err := repo.SetLastSentAt(now, tokens)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func testSubscriptionRepoDelete(t *testing.T) {
	db, mock := testutils.SetupMockDB(t)
	defer func() {
		mock.ExpectClose()
		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	repo := repository.NewSubscriptionRepo(db)

	mock.ExpectExec("DELETE FROM subscriptions").
		WithArgs("token123").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.Delete(context.Background(), "token123")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func testSubscriptionRepoDeleteError(t *testing.T) {
	db, mock := testutils.SetupMockDB(t)
	defer func() {
		mock.ExpectClose()
		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	repo := repository.NewSubscriptionRepo(db)

	mock.ExpectExec("DELETE FROM subscriptions").
		WithArgs("token123").
		WillReturnError(errors.New("delete error"))

	err := repo.Delete(context.Background(), "token123")
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func testSubscriptionRepoGetConfirmedByFrequency(t *testing.T) {
	db, mock := testutils.SetupMockDB(t)
	defer func() {
		mock.ExpectClose()
		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	repo := repository.NewSubscriptionRepo(db)

	now := time.Now()
	expected := domain.Subscription{
		ID:         "68501cb6-0bf0-800e-81ba-bae3763ecdd2",
		CreatedAt:  time.Now(),
		Email:      "test@example.com",
		City:       "Lviv",
		Token:      "token321",
		Frequency:  "weekly",
		Confirmed:  true,
		LastSentAt: &now,
	}

	rows := sqlmock.NewRows([]string{
		"id", "created_at", "email", "city", "token", "frequency", "confirmed", "last_sent_at",
	}).AddRow(
		expected.ID, expected.CreatedAt, expected.Email, expected.City, expected.Token,
		expected.Frequency, expected.Confirmed, expected.LastSentAt,
	)

	mock.ExpectQuery("SELECT .* FROM subscriptions WHERE confirmed = true").
		WithArgs("weekly").
		WillReturnRows(rows)

	subs, err := repo.GetConfirmedByFrequency("weekly")
	assert.NoError(t, err)
	assert.Len(t, subs, 1)
	assert.Equal(t, expected.Email, subs[0].Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func testSubscriptionRepoGetConfirmedByFrequencyError(t *testing.T) {
	db, mock := testutils.SetupMockDB(t)
	defer func() {
		mock.ExpectClose()
		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()

	repo := repository.NewSubscriptionRepo(db)

	mock.ExpectQuery("SELECT .* FROM subscriptions WHERE confirmed = true").
		WithArgs("weekly").
		WillReturnError(errors.New("query error"))

	_, err := repo.GetConfirmedByFrequency("weekly")
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
