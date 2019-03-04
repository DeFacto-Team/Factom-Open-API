package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"time"

	//	"github.com/DeFacto-Team/Factom-Open-API/api"
	"github.com/DeFacto-Team/Factom-Open-API/config"
	"github.com/DeFacto-Team/Factom-Open-API/model"
	"github.com/DeFacto-Team/Factom-Open-API/service"
	"github.com/DeFacto-Team/Factom-Open-API/store"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func main() {

	var err error

	name := flag.String("name", "", "User name")
	configFile := flag.String("c", "config/config.yaml", "Path to config file")
	flag.Parse()

	var conf *config.Config
	if conf, err = config.NewConfig(*configFile); err != nil {
		log.Fatal(err)
	}
	// Создаем сторедж
	store, err := store.NewStore(conf)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	user := &model.User{}
	user.AccessToken = GenerateApiKey(32)
	user.Name = *name

	us := service.NewUserService(store)

	created, err := us.CreateUser(user)
	if err != nil {
		log.Fatal(err)
	}

	user_json, _ := json.Marshal(created)

	fmt.Printf(string(user_json) + "\n")

}

func GenerateApiKey(n int) string {
	b := make([]rune, n)
	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
