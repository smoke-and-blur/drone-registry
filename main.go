package main

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var cards *mongo.Collection

// Card — одна картка пошкодження, ключ = серійний номер.
type Card struct {
	SN        string            `bson:"_id"           json:"sn"`
	Model     string            `bson:"model"         json:"model"`
	Fields    map[string]string `bson:"fields"        json:"fields"`
	Checks    []string          `bson:"checks"        json:"checks"`
	UpdatedAt time.Time         `bson:"updatedAt"     json:"updatedAt"`
	CreatedAt time.Time         `bson:"createdAt"     json:"createdAt"`
}

func main() {
	loadEnvFile(".env")
	uri := os.Getenv("MONGO_URI")
	if uri == "" {
		log.Fatal("MONGO_URI not set — див. .env.example")
	}
	addr := ":" + env("PORT", "8080")

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cl, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	if err := cl.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}
	cards = cl.Database(env("MONGO_DB", "drone_registry")).Collection("cards")
	log.Println("mongo ok")

	http.HandleFunc("/api/cards", apiCards)
	http.HandleFunc("/list", page("list.html"))
	http.HandleFunc("/", page("damage-card.html"))

	log.Println("listening on http://localhost" + addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func apiCards(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		save(w, r)
	case http.MethodGet:
		list(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// POST /api/cards — створити або оновити картку за S/N.
func save(w http.ResponseWriter, r *http.Request) {
	var c Card
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&c); err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	if c.SN == "" {
		http.Error(w, "sn required", http.StatusBadRequest)
		return
	}
	if c.Checks == nil {
		c.Checks = []string{}
	}
	if c.Fields == nil {
		c.Fields = map[string]string{}
	}
	now := time.Now().UTC()

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	_, err := cards.UpdateByID(ctx, c.SN, bson.M{
		"$set": bson.M{
			"model":     c.Model,
			"fields":    c.Fields,
			"checks":    c.Checks,
			"updatedAt": now,
		},
		"$setOnInsert": bson.M{"createdAt": now},
	}, options.Update().SetUpsert(true))
	if err != nil {
		log.Println("save:", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, bson.M{"ok": true, "sn": c.SN})
}

// GET /api/cards       — усі картки, найновіші зверху.
// GET /api/cards?sn=X  — одна картка.
func list(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if sn := r.URL.Query().Get("sn"); sn != "" {
		var c Card
		switch err := cards.FindOne(ctx, bson.M{"_id": sn}).Decode(&c); err {
		case nil:
			writeJSON(w, c)
		case mongo.ErrNoDocuments:
			http.Error(w, "not found", http.StatusNotFound)
		default:
			log.Println("find:", err)
			http.Error(w, "db error", http.StatusInternalServerError)
		}
		return
	}

	cur, err := cards.Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"updatedAt": -1}).SetLimit(500))
	if err != nil {
		log.Println("list:", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	out := []Card{}
	if err := cur.All(ctx, &out); err != nil {
		log.Println("list:", err)
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, out)
}

func page(file string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, file)
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(v)
}

// loadEnvFile читає KEY=VALUE з .env, не перетираючи вже задані змінні
// оточення (у проді їх задає платформа, файла там немає).
func loadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.Trim(strings.TrimSpace(v), `"'`)
		if _, exists := os.LookupEnv(k); !exists {
			os.Setenv(k, v)
		}
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
