package repository

import (
	"database/sql"
	"pvz/internal/models"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func setup(t *testing.T) (*Repository, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}
	return NewRepository(db), mock
}

func TestRegisterUserSuccess(t *testing.T) {
	repo, mock := setup(t)
	defer repo.DB.Close()
	email, hash, role := "u@e.com", "h", models.Role("employee")
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1);`)).
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO users (id, email, password, role, token)
	VALUES ($1, $2, $3, $4, $5);`)).
		WithArgs(sqlmock.AnyArg(), email, hash, role, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	u, err := repo.RegisterUser(email, hash, role)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.Email != email || u.Role != role {
		t.Errorf("got %+v", u)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestRegisterUserAlreadyExists(t *testing.T) {
	repo, mock := setup(t)
	defer repo.DB.Close()
	email := "u@e.com"
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1);`)).
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	_, err := repo.RegisterUser(email, "", models.Role("employee"))
	if err == nil {
		t.Fatal("expected error")
	}
}
func TestLoginUserSuccess(t *testing.T) {
	repo, mock := setup(t)
	defer repo.DB.Close()

	email := "user@example.com"
	plain := "supersecret"
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt failed: %v", err)
	}
	hash := string(hashBytes)
	token := "jwt-token"

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT password, token FROM users WHERE email = $1;`)).
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{"password", "token"}).
			AddRow(hash, token),
		)

	got, err := repo.LoginUser(email, plain)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != models.Token(token) {
		t.Errorf("token = %q, want %q", got, token)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestLoginUserInvalidPassword(t *testing.T) {
	repo, mock := setup(t)
	defer repo.DB.Close()

	email := "user@example.com"
	realPass := "correct"
	wrongPass := "wrong!"
	hashBytes, _ := bcrypt.GenerateFromPassword([]byte(realPass), bcrypt.DefaultCost)
	hash := string(hashBytes)

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT password, token FROM users WHERE email = $1;`)).
		WithArgs(email).
		WillReturnRows(sqlmock.NewRows([]string{"password", "token"}).
			AddRow(hash, "irrelevant-token"),
		)

	_, err := repo.LoginUser(email, wrongPass)
	if err != ErrInvalidPassword {
		t.Fatalf("error = %v, want ErrInvalidPassword", err)
	}
}

func TestLoginUserNotFound(t *testing.T) {
	repo, mock := setup(t)
	defer repo.DB.Close()
	email := "x"
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT password, token FROM users WHERE email = $1;`)).
		WithArgs(email).
		WillReturnError(sql.ErrNoRows)
	_, err := repo.LoginUser(email, "")
	if err != ErrUserNotFound {
		t.Fatalf("want ErrUserNotFound, got %v", err)
	}
}

func TestCreatePVZSuccess(t *testing.T) {
	repo, mock := setup(t)
	defer repo.DB.Close()
	city := models.City("Москва")
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO pvz (id, create_date, city) VALUES ($1, $2, $3);`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), city).
		WillReturnResult(sqlmock.NewResult(1, 1))
	pvz, err := repo.CreatePVZ(city)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := uuid.Parse(pvz.ID); err != nil {
		t.Error("invalid uuid")
	}
}

func TestCreateReceptionFlows(t *testing.T) {
	repo, mock := setup(t)
	defer repo.DB.Close()
	pvzID := "pvz"
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS (SELECT 1 FROM pvz WHERE id = $1);`)).
		WithArgs(pvzID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	_, err := repo.CreateReception(pvzID)
	if err != ErrPvzNotFound {
		t.Fatal(err)
	}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS (SELECT 1 FROM pvz WHERE id = $1);`)).
		WithArgs(pvzID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS (SELECT 1 FROM receptions WHERE pvz_id = $1 AND status = $2);`)).
		WithArgs(pvzID, models.StatusInProgress).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	_, err = repo.CreateReception(pvzID)
	if err != ErrReceptionInProgress {
		t.Fatal(err)
	}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS (SELECT 1 FROM pvz WHERE id = $1);`)).
		WithArgs(pvzID).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT EXISTS (SELECT 1 FROM receptions WHERE pvz_id = $1 AND status = $2);`)).
		WithArgs(pvzID, models.StatusInProgress).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO receptions (id, create_date, pvz_id, status)
	VALUES ($1, $2, $3, $4);`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), pvzID, models.StatusInProgress).
		WillReturnResult(sqlmock.NewResult(1, 1))
	_, err = repo.CreateReception(pvzID)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateProductSuccessAndNoReception(t *testing.T) {
	repo, mock := setup(t)
	defer repo.DB.Close()
	pvzID := "p"
	productType := models.ProductType("электроника")
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM receptions WHERE pvz_id = $1 AND status = $2
	ORDER BY create_date DESC FOR UPDATE
	LIMIT 1;`)).
		WithArgs(pvzID, models.StatusInProgress).
		WillReturnError(sql.ErrNoRows)
	_, err := repo.CreateProduct(pvzID, productType)
	if err != ErrNoActiveReception {
		t.Fatal(err)
	}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM receptions WHERE pvz_id = $1 AND status = $2
	ORDER BY create_date DESC FOR UPDATE
	LIMIT 1;`)).
		WithArgs(pvzID, models.StatusInProgress).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("r1"))
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO products (id, create_date, type, reception_id)
	VALUES ($1, $2, $3, $4);`)).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), productType, "r1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	p, err := repo.CreateProduct(pvzID, productType)
	if err != nil {
		t.Fatal(err)
	}
	if p.ReceptionID != "r1" {
		t.Fatalf("got %v", p)
	}
}

func TestCloseLastReceptionSuccessAndEmpty(t *testing.T) {
	repo, mock := setup(t)
	defer repo.DB.Close()
	pvzID := "x"
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, create_date FROM receptions WHERE pvz_id = $1 AND status = $2
	ORDER BY create_date DESC FOR UPDATE LIMIT 1;`)).
		WithArgs(pvzID, models.StatusInProgress).
		WillReturnError(sql.ErrNoRows)
	_, err := repo.CloseLastReception(pvzID)
	if err != ErrNoActiveReception {
		t.Fatal(err)
	}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, create_date FROM receptions WHERE pvz_id = $1 AND status = $2
	ORDER BY create_date DESC FOR UPDATE LIMIT 1;`)).
		WithArgs(pvzID, models.StatusInProgress).
		WillReturnRows(sqlmock.NewRows([]string{"id", "create_date"}).AddRow("r1", time.Now()))
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE receptions SET status = $1 WHERE id = $2;`)).
		WithArgs(models.StatusClose, "r1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	rec, err := repo.CloseLastReception(pvzID)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Status != models.StatusClose {
		t.Fatal("status")
	}
}

func TestDeleteLastProductFlows(t *testing.T) {
	repo, mock := setup(t)
	defer repo.DB.Close()
	pvzID := "p"
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM receptions WHERE pvz_id = $1 AND status = $2
	ORDER BY create_date DESC FOR UPDATE LIMIT 1;`)).
		WithArgs(pvzID, models.StatusInProgress).
		WillReturnError(sql.ErrNoRows)
	if err := repo.DeleteLastProduct(pvzID); err != ErrNoActiveReception {
		t.Fatal(err)
	}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM receptions WHERE pvz_id = $1 AND status = $2
	ORDER BY create_date DESC FOR UPDATE LIMIT 1;`)).
		WithArgs(pvzID, models.StatusInProgress).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("r"))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM products WHERE reception_id = $1
	ORDER BY create_date desc FOR UPDATE LIMIT 1;`)).
		WithArgs("r").
		WillReturnError(sql.ErrNoRows)
	if err := repo.DeleteLastProduct(pvzID); err != ErrNoProductsInReception {
		t.Fatal(err)
	}
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM receptions WHERE pvz_id = $1 AND status = $2
	ORDER BY create_date DESC FOR UPDATE LIMIT 1;`)).
		WithArgs(pvzID, models.StatusInProgress).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("r"))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM products WHERE reception_id = $1
	ORDER BY create_date desc FOR UPDATE LIMIT 1;`)).
		WithArgs("r").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("p1"))
	mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM products WHERE id = $1;`)).
		WithArgs("p1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	if err := repo.DeleteLastProduct(pvzID); err != nil {
		t.Fatal(err)
	}
}

func TestGetPVZInfoEmptyAndOne(t *testing.T) {
	repo, mock := setup(t)
	defer repo.DB.Close()
	start, end := time.Now(), time.Now()
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT DISTINCT  pvz.id, pvz.create_date, pvz.city
	FROM pvz JOIN receptions r ON r.pvz_id = pvz.id
	WHERE r.create_date BETWEEN $1 AND $2
	ORDER BY pvz.create_date
	LIMIT $3 OFFSET $4;`)).
		WithArgs(start, end, 10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "create_date", "city"}))
	list, err := repo.GetPVZInfo(start, end, 1, 10)
	if err != nil || len(list) != 0 {
		t.Fatalf("got %v, %v", list, err)
	}
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT DISTINCT  pvz.id, pvz.create_date, pvz.city
	FROM pvz JOIN receptions r ON r.pvz_id = pvz.id
	WHERE r.create_date BETWEEN $1 AND $2
	ORDER BY pvz.create_date
	LIMIT $3 OFFSET $4;`)).
		WithArgs(start, end, 10, 0).
		WillReturnRows(sqlmock.NewRows([]string{"id", "create_date", "city"}).AddRow("pvz1", start, "Казань"))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, create_date, status, pvz_id FROM receptions
	WHERE pvz_id = ANY($1) AND create_date >= $2 AND create_date <= $3
	ORDER BY create_date;`)).
		WithArgs(sqlmock.AnyArg(), start, end).
		WillReturnRows(sqlmock.NewRows([]string{"id", "create_date", "status", "pvz_id"}).
			AddRow("r1", start, models.StatusInProgress, "pvz1"))
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, create_date, type, reception_id
	FROM products WHERE reception_id = ANY($1)
	ORDER BY create_date;`)).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "create_date", "type", "reception_id"}).
			AddRow("p1", start, models.ProductType("электроника"), "r1"))
	out, err := repo.GetPVZInfo(start, end, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 || len(out[0].Receptions) != 1 || len(out[0].Receptions[0].Products) != 1 {
		t.Fatalf("unexpected %+v", out)
	}
}
