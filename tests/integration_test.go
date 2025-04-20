package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"pvz/internal/config"
	"pvz/internal/models"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

const baseURL = "http://localhost:8081"

func cleanupDB(t *testing.T) {
	t.Helper()

	cfg := config.Config{
		DB: config.DataBaseCfg{
			Host:     "localhost",
			Port:     "5555",
			User:     "test_user",
			Password: "123",
			DBName:   "test_database",
		},
		App: config.AppCfg{
			Address: "localhost",
			Port:    "8081",
		},
	}

	dsn := cfg.DB.GetDsn()
	db, err := sql.Open("postgres", dsn)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec(`TRUNCATE TABLE products, receptions, pvz, users RESTART IDENTITY CASCADE;`)
	require.NoError(t, err)
}

func login(t *testing.T, role string) string {
	t.Helper()
	body, err := json.Marshal(map[string]string{"role": role})
	require.NoError(t, err)

	resp, err := http.Post(baseURL+"/dummyLogin", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var token string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&token))
	require.NotEmpty(t, token)
	return token
}

func TestFullLifeСycle(t *testing.T) {
	cleanupDB(t)

	client := &http.Client{}

	modToken := login(t, "moderator")
	pvzReq := map[string]string{"city": "Москва"}
	pvzBody, _ := json.Marshal(pvzReq)
	req, _ := http.NewRequest("POST", baseURL+"/pvz", bytes.NewReader(pvzBody))
	req.Header.Set("Authorization", "Bearer "+modToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var pvz models.PVZ
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pvz))
	fmt.Println(pvz)
	require.NotEmpty(t, pvz.ID)

	empToken := login(t, "employee")
	recvReq := map[string]string{"pvzId": pvz.ID}
	recvBody, _ := json.Marshal(recvReq)
	req, _ = http.NewRequest("POST", baseURL+"/receptions", bytes.NewReader(recvBody))
	req.Header.Set("Authorization", "Bearer "+empToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	for i := 0; i < 50; i++ {
		prodReq := map[string]string{"pvzId": pvz.ID, "type": "электроника"}
		prodBody, _ := json.Marshal(prodReq)
		req, _ = http.NewRequest("POST", baseURL+"/products", bytes.NewReader(prodBody))
		req.Header.Set("Authorization", "Bearer "+empToken)
		req.Header.Set("Content-Type", "application/json")
		resp, err = client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	time.Sleep(100 * time.Millisecond)
	closeURL := fmt.Sprintf("%s/pvz/%s/close_last_reception", baseURL, pvz.ID)
	req, _ = http.NewRequest("POST", closeURL, nil)
	req.Header.Set("Authorization", "Bearer "+empToken)
	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
