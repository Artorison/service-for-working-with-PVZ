package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"pvz/internal/models"
	"pvz/pkg/utils"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

var (
	ErrUserNotFound          = errors.New("user not found")
	ErrNoActiveReception     = errors.New("no active reception")
	ErrReceptionInProgress   = errors.New("previous reception not closed")
	ErrPvzNotFound           = errors.New("pvz not found")
	ErrNoProductsInReception = errors.New("no products in reception")
	ErrInvalidPassword       = errors.New("invalid password")
	ErrBeginTransaction      = errors.New("failed to begin transaction")
	ErrCommitTransaction     = errors.New("failed to commit transaction")
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) RegisterUser(email, passwordHash string, role models.Role) (models.User, error) {
	var exists bool
	const checkQuery = `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1);`
	err := r.DB.QueryRow(checkQuery, email).Scan(&exists)
	if err != nil {
		return models.User{}, models.Wrap("failed to check user existence", err)
	}
	if exists {
		return models.User{}, fmt.Errorf("user with email %s already exists", email)
	}

	id := uuid.NewString()

	token, err := utils.GetJWToken(role, id)
	if err != nil {
		return models.User{}, models.Wrap("failed to generate token", err)
	}

	const insertQuery = `INSERT INTO users (id, email, password, role, token)
	VALUES ($1, $2, $3, $4, $5);`
	_, err = r.DB.Exec(insertQuery, id, email, passwordHash, role, token)
	if err != nil {
		return models.User{}, models.Wrap("failed to insert user", err)
	}

	return models.User{
		ID:    id,
		Email: email,
		Role:  role,
	}, nil
}

func (r *Repository) LoginUser(email, password string) (models.Token, error) {

	const query = `SELECT password, token FROM users WHERE email = $1;`

	var token models.Token
	var hash string
	err := r.DB.QueryRow(query, email).Scan(&hash, &token)
	if err == sql.ErrNoRows {
		return "", ErrUserNotFound
	}
	if err != nil {
		return "", err
	}
	if !utils.CheckHashPassword(hash, password) {
		return "", ErrInvalidPassword
	}
	return token, nil
}
func (r *Repository) CreatePVZ(city models.City) (models.PVZ, error) {
	var pvz = models.PVZ{
		ID:               uuid.NewString(),
		RegistrationDate: time.Now().UTC().Round(time.Millisecond),
		City:             city,
	}

	const query = `INSERT INTO pvz (id, create_date, city) VALUES ($1, $2, $3);`

	_, err := r.DB.Exec(query, pvz.ID, pvz.RegistrationDate, pvz.City)
	if err != nil {
		return models.PVZ{}, models.Wrap("can't create pvz", err)
	}
	return pvz, nil
}

func (r *Repository) CreateReception(pvzID string) (models.Reception, error) {
	var exists bool

	const checkPvzQuery = `SELECT EXISTS (SELECT 1 FROM pvz WHERE id = $1);`
	err := r.DB.QueryRow(checkPvzQuery, pvzID).Scan(&exists)
	if err != nil {
		return models.Reception{}, models.Wrap("failed to check pvz existence", err)
	}
	if !exists {
		return models.Reception{}, ErrPvzNotFound
	}

	const checkRecQuery = `SELECT EXISTS (SELECT 1 FROM receptions WHERE pvz_id = $1 AND status = $2);`
	err = r.DB.QueryRow(checkRecQuery, pvzID, models.StatusInProgress).Scan(&exists)
	if err != nil {
		return models.Reception{}, models.Wrap("failed to check reception existence", err)
	}
	if exists {
		return models.Reception{}, ErrReceptionInProgress
	}

	rec := models.Reception{
		ID:       uuid.NewString(),
		DateTime: time.Now().UTC().Round(time.Millisecond),
		PvzID:    pvzID,
		Status:   models.StatusInProgress,
	}

	const insertQuery = `INSERT INTO receptions (id, create_date, pvz_id, status)
	VALUES ($1, $2, $3, $4);`
	if _, err := r.DB.Exec(insertQuery, rec.ID, rec.DateTime, rec.PvzID, models.StatusInProgress); err != nil {
		return models.Reception{}, models.Wrap("failed to insert reception", err)
	}

	return rec, nil
}

func (r *Repository) CreateProduct(pvzID string, productType models.ProductType) (models.Product, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return models.Product{}, ErrBeginTransaction
	}
	defer tx.Rollback()

	const getReceptionIDQuery = `SELECT id FROM receptions WHERE pvz_id = $1 AND status = $2
	ORDER BY create_date DESC FOR UPDATE
	LIMIT 1;`

	var receptionID string
	err = tx.QueryRow(getReceptionIDQuery, pvzID, models.StatusInProgress).Scan(&receptionID)
	if err == sql.ErrNoRows {
		return models.Product{}, ErrNoActiveReception
	}
	if err != nil {
		return models.Product{}, models.Wrap("failed to get reception id", err)
	}

	const insertProductQuery = `INSERT INTO products (id, create_date, type, reception_id)
	VALUES ($1, $2, $3, $4);`
	productID := uuid.NewString()

	regTime := time.Now().UTC().Round(time.Millisecond)
	_, err = tx.Exec(insertProductQuery, productID, regTime, productType, receptionID)
	if err != nil {
		return models.Product{}, models.Wrap("failed to insert product", err)
	}

	if err := tx.Commit(); err != nil {
		return models.Product{}, ErrCommitTransaction
	}

	return models.Product{
		ID:          productID,
		DateTime:    regTime,
		Type:        productType,
		ReceptionID: receptionID,
	}, nil
}

func (r *Repository) CloseLastReception(pvzID string) (models.Reception, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return models.Reception{}, ErrBeginTransaction
	}
	defer tx.Rollback()

	const getRecInfoQuery = `SELECT id, create_date FROM receptions WHERE pvz_id = $1 AND status = $2
	ORDER BY create_date DESC FOR UPDATE LIMIT 1;`

	var rec models.Reception
	if err := tx.QueryRow(getRecInfoQuery, pvzID, models.StatusInProgress).Scan(&rec.ID, &rec.DateTime); err != nil {
		if err == sql.ErrNoRows {
			return models.Reception{}, ErrNoActiveReception
		}
		return models.Reception{}, models.Wrap("select reception", err)
	}
	rec.PvzID = pvzID
	rec.Status = models.StatusClose
	const updateQuery = `UPDATE receptions SET status = $1 WHERE id = $2;`
	_, err = tx.Exec(updateQuery, rec.Status, rec.ID)
	if err != nil {
		return models.Reception{}, models.Wrap("update reception", err)
	}

	if err := tx.Commit(); err != nil {
		return models.Reception{}, ErrCommitTransaction
	}

	return rec, nil
}

func (r *Repository) DeleteLastProduct(pvzID string) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return ErrBeginTransaction
	}
	defer tx.Rollback()

	const getReceptionIDQuery = `SELECT id FROM receptions WHERE pvz_id = $1 AND status = $2
	ORDER BY create_date DESC FOR UPDATE LIMIT 1;`

	var receptionID string
	if err := tx.QueryRow(getReceptionIDQuery, pvzID, models.StatusInProgress).Scan(&receptionID); err != nil {
		if err == sql.ErrNoRows {
			return ErrNoActiveReception
		}
		return models.Wrap("select reception", err)
	}

	const getProductIDQuery = `SELECT id FROM products WHERE reception_id = $1
	ORDER BY create_date desc FOR UPDATE LIMIT 1;`

	var productID string
	if err := tx.QueryRow(getProductIDQuery, receptionID).Scan(&productID); err != nil {
		if err == sql.ErrNoRows {
			return ErrNoProductsInReception
		}
		return models.Wrap("select product", err)
	}

	const deleteQuery = `DELETE FROM products WHERE id = $1;`
	if _, err := tx.Exec(deleteQuery, productID); err != nil {
		return models.Wrap("delete product", err)
	}

	if err := tx.Commit(); err != nil {
		return ErrCommitTransaction
	}
	return nil
}

func (r *Repository) GetPVZInfo(start, end time.Time, page, limit int) ([]models.PVZInfo, error) {

	const selectPVZList = `SELECT DISTINCT  pvz.id, pvz.create_date, pvz.city
	FROM pvz JOIN receptions r ON r.pvz_id = pvz.id
	WHERE r.create_date BETWEEN $1 AND $2
	ORDER BY pvz.create_date
	LIMIT $3 OFFSET $4;`

	rows, err := r.DB.Query(selectPVZList, start, end, limit, (page-1)*limit)
	if err != nil {
		return nil, models.Wrap("select pvz", err)
	}
	defer rows.Close()

	pvzInfoList := make([]models.PVZInfo, 0, limit)
	pvzIDList := make([]string, 0, limit)

	for rows.Next() {
		var pvzInfo models.PVZInfo
		err := rows.Scan(&pvzInfo.Pvz.ID, &pvzInfo.Pvz.RegistrationDate, &pvzInfo.Pvz.City)
		if err != nil {
			return nil, models.Wrap("pvz rows scan", err)
		}
		pvzInfoList = append(pvzInfoList, pvzInfo)
		pvzIDList = append(pvzIDList, pvzInfo.Pvz.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, models.Wrap("rows err pvz", err)
	}
	if len(pvzIDList) == 0 {
		return pvzInfoList, nil
	}

	const selectRecList = `SELECT id, create_date, status, pvz_id FROM receptions
	WHERE pvz_id = ANY($1) AND create_date >= $2 AND create_date <= $3
	ORDER BY create_date;`
	recRows, err := r.DB.Query(selectRecList, pq.Array(pvzIDList), start, end)
	if err != nil {
		return nil, models.Wrap("select reception", err)
	}
	defer recRows.Close()

	recList := make([]models.ReceptionWithProducts, 0)
	recIDList := make([]string, 0)
	for recRows.Next() {
		var recWithProducts models.ReceptionWithProducts
		if err := recRows.Scan(
			&recWithProducts.Reception.ID,
			&recWithProducts.Reception.DateTime,
			&recWithProducts.Reception.Status,
			&recWithProducts.Reception.PvzID,
		); err != nil {
			return nil, models.Wrap("reception rows scan", err)
		}
		recList = append(recList, recWithProducts)
		recIDList = append(recIDList, recWithProducts.Reception.ID)
	}
	if err := recRows.Err(); err != nil {
		return nil, models.Wrap("rows err reception", err)
	}

	const selectProductList = `SELECT id, create_date, type, reception_id
	FROM products WHERE reception_id = ANY($1)
	ORDER BY create_date;`
	prodRows, err := r.DB.Query(selectProductList, pq.Array(recIDList))
	if err != nil {
		return nil, models.Wrap("select product", err)
	}
	defer prodRows.Close()

	productsMap := make(map[string][]models.Product, len(recIDList))
	for prodRows.Next() {
		var product models.Product
		if err := prodRows.Scan(
			&product.ID,
			&product.DateTime,
			&product.Type,
			&product.ReceptionID,
		); err != nil {
			return nil, models.Wrap("product rows scan", err)
		}
		productsMap[product.ReceptionID] = append(productsMap[product.ReceptionID], product)
	}
	if err := prodRows.Err(); err != nil {
		return nil, models.Wrap("rows err product", err)
	}

	pvzMap := make(map[string]*models.PVZInfo, len(pvzInfoList))
	for i := range pvzInfoList {
		pvzInfoList[i].Receptions = []models.ReceptionWithProducts{}
		pvzMap[pvzInfoList[i].Pvz.ID] = &pvzInfoList[i]
	}
	for i := range recList {
		rwp := recList[i]
		products := productsMap[rwp.Reception.ID]
		if products == nil {
			products = []models.Product{}
		}
		rwp.Products = products
		pvzMap[rwp.Reception.PvzID].Receptions = append(pvzMap[rwp.Reception.PvzID].Receptions, rwp)
	}
	return pvzInfoList, nil
}
