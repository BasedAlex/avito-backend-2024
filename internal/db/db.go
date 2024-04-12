package db

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type Postgres struct {
	db *pgx.Conn
}

func NewPostgres(ctx context.Context, dbConnect string) (*Postgres, error) {
	db, err := pgx.Connect(ctx, dbConnect)
	if err != nil {
		return nil, err
	}
	return &Postgres{db:db}, nil
}

type Banner struct {
	Id int `json:"id"`
	FeatureId int `json:"feature_id"`
	Content map[string]interface{} `json:"content"`
	IsActive bool `json:"is_active"`
	TagIds []int `json:"tag_ids"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (db *Postgres) GetBanner(ctx context.Context, featureId, tagId, limit, offset int) ([]Banner, error ) {

	stmt := `SELECT b.id, b.feature_id, b.content, b.is_active, b.created_at, b.updated_at, 
			(SELECT ARRAY_AGG(bt.tag_id) FROM banners_tags bt WHERE bt.banner_id = b.id) AS tag_ids
			FROM banners b
			JOIN banners_tags bt ON b.id = bt.banner_id`

	var args []any
	if featureId != 0 || tagId != 0 {
		stmt += ` WHERE `
	}
	if featureId != 0 {
		stmt += `feature_id = $1`
		args = append(args, featureId)
	}
	if tagId != 0 {
		if featureId != 0 {
			stmt += ` AND tag_id = $2`
		} else {
			stmt += `tag_id = $1`
		}
		args = append(args, tagId)
	}

	stmt+= ` GROUP BY b.id, b.feature_id, b.content, b.is_active, b.created_at, b.updated_at ORDER BY b.id `

	switch {
	case limit != 0 && offset == 0: stmt+=fmt.Sprintf("LIMIT %d", limit) 
	case limit == 0 && offset != 0: stmt+=fmt.Sprintf("OFFSET %d", offset)
	case limit != 0 && offset != 0: stmt+=fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	}

	rows, err := db.db.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var banners []Banner
	bannerIndex := map[int]int{}

	for rows.Next() {
		var banner Banner
		var tagIds []int
		
		err = rows.Scan(&banner.Id, &banner.FeatureId, &banner.Content, &banner.IsActive, &banner.CreatedAt, &banner.UpdatedAt, &tagIds)
		if err != nil {
			return nil, err
		}

		banner.TagIds = tagIds

		b, ok := bannerIndex[banner.Id]
		if ok {
			banners[b] = banner
		} else {
			banners = append(banners, banner)
			bannerIndex[banner.Id] = len(banners) - 1
		}
	}
	return banners, nil 
}

func (db *Postgres) UpdateBanner(ctx context.Context, id int, banner Banner, isActive *bool) error {
	content, err := json.Marshal(banner.Content)
	if err != nil {
		return err
	}

	var args []any

	var queryParts []string

	if banner.FeatureId != 0 {
		queryParts = append(queryParts, "feature_id = $"+strconv.Itoa(len(args)+1))
		args = append(args, banner.FeatureId)
	}
	if isActive != nil {
		queryParts = append(queryParts, "is_active = $"+strconv.Itoa(len(args)+1))
		args = append(args, banner.IsActive)
	}
	if len(banner.Content) > 0 {
		queryParts = append(queryParts, "content = $"+strconv.Itoa(len(args)+1))
		args = append(args, content)
	}
	queryParts = append(queryParts, "updated_at = $"+strconv.Itoa(len(args)+1))
	args = append(args, time.Now())

	stmt := `
	UPDATE banners
	SET ` + strings.Join(queryParts, ", ") + `
	WHERE id = $` + strconv.Itoa(len(args)+1) + ";"

	args = append(args, id)

	_, err = db.db.Exec(ctx, stmt, args...)
	if err != nil {
		return err
	}

	stmtDeleteTags := `
    DELETE FROM banners_tags
    WHERE banner_id = $1;`

	_, err = db.db.Exec(ctx, stmtDeleteTags, id)
	if err != nil {
		return err
	}

	if len(banner.TagIds) > 0 {
		stmtInsertTags := `
		INSERT INTO banners_tags (tag_id, banner_id)
		VALUES `

		var valueStrings []string
		var valueArgs []interface{}

		for _, tagID := range banner.TagIds {
			valueStrings = append(valueStrings, "($"+strconv.Itoa(len(valueArgs)+1)+", $"+strconv.Itoa(len(valueArgs)+2)+")")
			valueArgs = append(valueArgs, tagID, id)
		}

		stmtInsertTags += strings.Join(valueStrings, ", ")

		_, err = db.db.Exec(ctx, stmtInsertTags, valueArgs...)
		if err != nil {
			return err
		}
	}

	return nil
}



func (db *Postgres) GetUserBanner(ctx context.Context, tagId int, featureId int) (string, error) {
	
	stmt := `
	SELECT content FROM banners 
	INNER JOIN banners_tags 
	ON banners_tags.banner_id=banners.id 
	WHERE feature_id=$1 
	AND banners_tags.tag_id=$2 
	LIMIT 1;`

	var content string
	fmt.Println(featureId, tagId)

	err := db.db.QueryRow(ctx, stmt, featureId, tagId).Scan(&content)
	if err != nil {
		fmt.Println("here", err)
		return "", err
	}
	

	return content, nil
}

func (db *Postgres) CreateBanner(ctx context.Context, banner Banner) error {
	content, err := json.Marshal(banner.Content)
	if err != nil {
		return err
	}
    stmt := `
    INSERT INTO banners (feature_id, content, is_active, created_at, updated_at)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING id;`

    var bannerID int
    err = db.db.QueryRow(ctx, stmt, banner.FeatureId, content, banner.IsActive, time.Now(), time.Now()).Scan(&bannerID)
    if err != nil {
        return err
    }

	stmtInsertTags := `
	INSERT INTO banners_tags (tag_id, banner_id)
	VALUES ($1, $2);`

	for _, tagID := range banner.TagIds {
		_, err := db.db.Exec(ctx, stmtInsertTags, tagID, bannerID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *Postgres) DeleteBanner(ctx context.Context, id int) error {

   stmtDeleteTags := `
   DELETE FROM banners_tags
   WHERE banner_id = $1;`

   _, err := db.db.Exec(ctx, stmtDeleteTags, id)
   if err != nil {
	   return err
   }

   stmtDeleteBanner := "DELETE FROM banners WHERE id=$1;"
   _, err = db.db.Exec(ctx, stmtDeleteBanner, id)
   if err != nil {
		return err
   }

	return nil
}