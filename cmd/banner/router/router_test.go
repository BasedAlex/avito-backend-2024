package router

import (
	"banner/internal/cache"
	"banner/internal/config"
	"banner/internal/db"
	"context"
	"database/sql"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/pressly/goose"

	_ "github.com/lib/pq"
)

var mockCache cache.Cache
var mockDB *db.Postgres

func TestMain(m *testing.M) {
	clear := StartTestServer()
	code := m.Run()
	clear()
	os.Exit(code)
}

func StartTestServer() func() {
	cfg, err := config.Load("../../../config.test.yaml")

	if err != nil {
		log.Fatalln("error loading config: ", err)
	}

	ctx := context.Background()

	mockDB, err = db.NewPostgres(ctx, cfg.DBConnect)
	if err != nil {
		log.Fatalln("error loading db: ", err)
	}
	migrationDb, err := sql.Open("postgres", cfg.DBConnect)
	if err != nil {
		log.Fatalln("error loading db: ", err)
	}

	goose.Up(migrationDb, "../../../internal/migrations/")

	_, err = migrationDb.ExecContext(ctx, `INSERT INTO features (id) VALUES (3);`)
	if err != nil {
		log.Fatalln("couldn't seed db: ", err)
	}
	_, err = migrationDb.ExecContext(ctx, `INSERT INTO tags (id) VALUES (3);`)
	if err != nil {
		log.Fatalln("couldn't seed db: ", err)
	}
	row := migrationDb.QueryRowContext(ctx, `INSERT INTO banners (feature_id, content, is_active, created_at, updated_at)
    VALUES (3, '{"url": "123", "text": "123", "title": "321"}', true, now(), now()) 
	RETURNING id`)
	if err != nil {
		log.Fatalln("couldn't seed db: ", err)
	}
	var id int
	err = row.Scan(&id)
	if err != nil {
		log.Fatalln("couldn't seed db: ", err)
	}

	_, err = migrationDb.ExecContext(ctx, `INSERT INTO banners_tags (tag_id, banner_id)
	VALUES (3, $1);`, id)
	if err != nil {
		log.Fatalln("couldn't seed db: ", err)
	}

	mockCache, err = cache.New(cfg.Redis)
	if err != nil {
		log.Fatalln("error loading cache: ", err)
	}
	log.Println("connected to redis")
	return func() {
		goose.Down(migrationDb, "../../../internal/migrations/")
	}
}

func TestGetUserBannerHandler(t *testing.T) {
	handler := &Handler{
		Database:   mockDB,
		Cache:      mockCache,
		userToken:  "userToken",
		adminToken: "adminToken",
	}

	req, err := http.NewRequest("GET", "/user_banner?tag_id=3&feature_id=3", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	handler.getUserBanner(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	expected := `{"url": "123", "text": "123", "title": "321"}`

	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}