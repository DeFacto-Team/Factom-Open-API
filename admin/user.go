package main

import (
	"flag"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/DeFacto-Team/Factom-Open-API/config"
	"github.com/DeFacto-Team/Factom-Open-API/model"
	"github.com/DeFacto-Team/Factom-Open-API/store"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

const (
	ConfigFile = ".config/config.yaml"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func main() {

	var err error
	var action, name, param string

	var conf *config.Config
	if conf, err = config.NewConfig(ConfigFile); err != nil {
		log.Fatal(err)
	}

	args := flag.Args()

	if len(args) == 0 {
		log.Fatal("No params provided")
	}

	if len(args) >= 1 {
		action = args[0]
	}

	if len(args) >= 2 {
		name = args[1]
	}

	if len(args) >= 3 {
		param = args[2]
	}

	log.Info("action=", action, ", name=", name, ", param=", param)

	store, err := store.NewStore(conf)
	if err != nil {
		log.Fatal(err)
	}
	defer store.Close()

	user := &model.User{}

	if name == "" && action != "ls" && action != "help" {
		log.Fatal("Name can not be null for action ", action)
	}

	if name != "" && action != "create" && action != "ls" && action != "help" {
		user.Name = name
		user = store.GetUser(user)
		if user == nil {
			log.Fatal("User ", name, " not found")
		}
	}

	switch action {
	case "help":

		fmt.Printf("User management tool for Factom Open API:\n")
		fmt.Printf("user help — Show help\n")
		fmt.Printf("user create john — Create user 'john' and generate API access key\n")
		fmt.Printf("user disable john — Disable access to API for user 'john'\n")
		fmt.Printf("user enable john — Enable access to API for user 'john'\n")
		fmt.Printf("user delete john — Delete user 'john'\n")
		fmt.Printf("user rotate-key john — Rotate API access key for user 'john'\n")
		fmt.Printf("user set-limit john 1000 — Set writes limit for user 'john' to 1000\n")
		fmt.Printf("user ls — Show all API users, their API keys, statuses & limits\n")

	case "create":

		user.Name = name
		user.AccessToken = generateAPIKey(32)

		err = store.CreateUser(user)
		if err != nil {
			log.Fatal(err)
		}

		log.Info("User ", user.Name, " created, API access key: ", user.AccessToken)

	case "delete":

		err = store.DeleteUser(user)
		if err != nil {
			log.Fatal(err)
		}

		log.Info("User ", user.Name, " deleted")

	case "enable":

		user.Status = 1

		err = store.UpdateUser(user)
		if err != nil {
			log.Fatal(err)
		}

		log.Info("User ", user.Name, " enabled")

	case "disable":

		user.Status = -1

		err = store.UpdateUser(user)
		if err != nil {
			log.Fatal(err)
		}

		log.Info("User ", user.Name, " disabled")

	case "rotate-key":

		user.AccessToken = generateAPIKey(32)

		err = store.UpdateUser(user)
		if err != nil {
			log.Fatal(err)
		}

		log.Info("Access key changed for user ", user.Name, ", new API access key: ", user.AccessToken)

	case "set-limit":

		if param == "" {
			log.Fatal("You have to provide a numeric param for action ", action)
		}

		var limit int
		if limit, err = strconv.Atoi(param); err != nil {
			log.Fatal("You have to provide a numeric param for action ", action)
		}

		user.UsageLimit = limit

		if limit == 0 {
			err = store.DisableUserUsageLimit(user)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			err = store.UpdateUser(user)
			if err != nil {
				log.Fatal(err)
			}
		}

		log.Info("Usage limit for user ", user.Name, " set to ", user.UsageLimit, " write(s)")

	case "ls":

		users := store.GetUsers(&model.User{})

		if users == nil {
			log.Info("No users found")
		} else {
			for _, u := range users {
				var status string
				if u.Status == 1 {
					status = "enabled"
				} else {
					status = "disabled"
				}
				log.Info("id=", u.ID, ", name=", u.Name, ", accessToken=", u.AccessToken, ", status=", status, ", usage=", u.Usage, ", usageLimit=", u.UsageLimit)
			}
		}

	default:

		log.Fatal("Incorrect action: ", action)

	}

}

func generateAPIKey(n int) string {
	b := make([]rune, n)
	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
