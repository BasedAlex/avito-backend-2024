package main

import (
	"banner/cmd/banner/router"
	"banner/internal/cache"
	"banner/internal/config"
	"banner/internal/db"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)


func main() {
	//Load Config
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalln("error loading config: ", err)
	}

	redis, err := cache.New(cfg.Redis)
	if err != nil {
		log.Fatalln("error connecting to redis: ", err)
	}
	log.Println("connected to redis")

	// Open DB
	dbCtx, cancel := context.WithTimeout(ctx, cfg.Cancel*time.Second)
	defer cancel()
	database, err := db.NewPostgres(dbCtx, cfg.DBConnect)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("connected to db")

	srv := &http.Server{
		Addr: fmt.Sprintf(":%s", cfg.Port),
		Handler: router.Register(cfg.UserToken, cfg.AdminToken, database, redis),
	}
	// Graceful shutdown
	go func(){
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second * 15)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	err = srv.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalln(err)
	}

}