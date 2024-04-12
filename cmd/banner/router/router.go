package router

import (
	"banner/internal/db"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"banner/internal/cache"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Handler struct {
    Database *db.Postgres
	Cache cache.Cache
	userToken string
	adminToken string
}

func Register(userToken, adminToken string, database *db.Postgres, redis cache.Cache) *chi.Mux {
	handler:= &Handler{
		Database: database,
		Cache: redis,
		userToken: userToken,
		adminToken: adminToken,
	}

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders: []string{"Link"},
		AllowCredentials: true,
		MaxAge: 300,
	}))

	r.Use(middleware.Logger)
	r.Use(isAuth)

	r.Group(func(r chi.Router) {
		r.Use(guard(userToken, adminToken))
		r.Get("/user_banner", handler.getUserBanner)
	})

	r.Group(func(r chi.Router) {
		r.Use(guard(adminToken))
		r.Get("/banner", handler.getBanner)
		r.Post("/banner", handler.postBanner)
		r.Patch("/banner/{id}", handler.updateBanner)
		r.Delete("/banner/{id}", handler.deleteBanner)
	})
	
	return r
}

func isAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("token")
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func guard(tokenList ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("token")
			for _, t := range tokenList {
				if t == token {
					next.ServeHTTP(w, r)
					return
				}
			}

			w.WriteHeader(http.StatusForbidden)
		})
	}
}

func (h *Handler) getUserBanner(w http.ResponseWriter, r *http.Request) {

	tagIdStr := r.URL.Query().Get("tag_id")
	featureIdStr := r.URL.Query().Get("feature_id")
	useLastRevision := r.URL.Query().Has("use_last_revision")
	
	if tagIdStr == "" || featureIdStr == "" {
		http.Error(w, "tag_id or feature_id is not defined", http.StatusBadRequest)
		return
	}
	
	key := tagIdStr+featureIdStr

	tagId, err := strconv.Atoi(tagIdStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	featureId, err := strconv.Atoi(featureIdStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	if !useLastRevision {
		ret, err := h.Cache.Get(r.Context(), key) 
		if err != nil {
			log.Println("no cache for this key", err)
		}
		if ret != "" {
			w.Write([]byte(ret))
			return
		}
	}
	

	banner, err := h.Database.GetUserBanner(r.Context(), tagId, featureId)
	var isActive bool
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	token := r.Header.Get("token")
	if !isActive && token == h.userToken {
		http.Error(w, "not allowed", http.StatusForbidden)
	}
	err = h.Cache.Set(r.Context(), key, banner)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(banner))
}

func (h *Handler) getBanner(w http.ResponseWriter, r *http.Request) {
	tagIdStr := r.URL.Query().Get("tag_id")
	featureIdStr := r.URL.Query().Get("feature_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	var featureId, tagId, limit, offset int
	var err error

	if featureIdStr != "" {
		featureId, err = strconv.Atoi(featureIdStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if tagIdStr != "" {
		tagId, err = strconv.Atoi(tagIdStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	banners, err := h.Database.GetBanner(r.Context(), featureId, tagId, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	res, err := json.Marshal(banners)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, string(res))
}


func (h *Handler) postBanner(w http.ResponseWriter, r *http.Request) {
	var data db.Banner

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

	ctx := context.TODO()

	err = h.Database.CreateBanner(ctx, data)
	if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
	w.WriteHeader(http.StatusCreated)
}


func (h *Handler) updateBanner(w http.ResponseWriter, r *http.Request) {
	var data db.Banner

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
	ctx := context.TODO()

	err = h.Database.UpdateBanner(ctx, id, data, &data.IsActive)
	if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) deleteBanner(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

	ctx := context.TODO()
	err = h.Database.DeleteBanner(ctx, id)
	if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

	w.WriteHeader(http.StatusNoContent)
}