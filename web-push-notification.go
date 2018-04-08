package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	webnotification "github.com/hbouvier/web-notifications/notification"
	"github.com/hbouvier/web-notifications/storage"
)

const defaultPort string = "0.0.0.0:8000"

func main() {
	portPtr := flag.String("port", "", "http port to listen to.")
	flag.Parse()

	if len(flag.Args()) > 0 {
		log.Panicf("USAGE: web-push-notification -port %s", defaultPort)
	}
	httpServer(getPortOr(*portPtr, "PORT", defaultPort))
}

func getPortOr(port string, envName string, defaultPort string) string {
	if port == "" {
		port = os.Getenv(envName)
		if port == "" {
			port = defaultPort
		}
	}
	if strings.Index(port, ":") == -1 {
		port = "0.0.0.0:" + port
	}
	return port
}

func httpServer(port string) {
	vapid := storage.GetOrCreateVAPID("./vapid.json")
	db := storage.Open("./registrations.json")

	routes(vapid, db)
	log.Printf("Listening on %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Panicf("Unable to listen to %s => %+v", port, err)
	}
}

func routes(vapid *storage.VAPID, db *storage.DB) {
	fs := http.FileServer(http.Dir("web"))
	http.HandleFunc("/api/v1/push", notificationHandler(vapid, db))
	http.HandleFunc("/api/v1/register", registrationHandler(db))
	http.HandleFunc("/scripts/", templateHandler(vapid))
	http.HandleFunc("/service-worker.js", templateHandler(vapid))
	http.Handle("/", fs)
}

func templateHandler(vapid *storage.VAPID) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Host, r.Method, r.RequestURI)
		filename := fmt.Sprintf("web/%s", r.RequestURI)
		log.Printf("Read filename %s ", filename)
		buffer, err := ioutil.ReadFile(filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		rendered := strings.Replace(string(buffer), "%{PUBLIC_KEY}%", vapid.Public, -1)
		w.Header().Set("Content-Type", "application/javascript")
		w.Write([]byte(rendered))
	}
}

func notificationHandler(vapid *storage.VAPID, db *storage.DB) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		buffer, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Println(err.Error())
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		var notification webnotification.Notification
		if err = json.Unmarshal(buffer, &notification); err != nil {
			log.Println(err.Error())
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		errs := make([]error, 0)
		userRegistrations := db.FindRegistration(notification.Subscriber)
		userRegistrations.Each(func(registration storage.Registration) {
			if err = notification.Push(registration.Subscription, vapid); err != nil {
				errs = append(errs, err)
			}
		})

		nbRegistration := userRegistrations.Length()
		nbError := len(errs)

		if nbError == nbRegistration && nbError > 0 {
			log.Println(errs[0].Error())
			http.Error(res, errs[0].Error(), http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.Write([]byte(fmt.Sprintf("{\"ok\":true,\"success\":%d,\"failed\":%d}", nbRegistration, nbError)))
	}
}

func registrationHandler(db *storage.DB) func(http.ResponseWriter, *http.Request) {
	return func(res http.ResponseWriter, req *http.Request) {
		buffer, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Println(err.Error())
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		var registration storage.Registration
		if err = json.Unmarshal(buffer, &registration); err != nil {
			log.Println(err.Error())
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		registrationsFound := db.Filter(func(aRegistration storage.Registration) bool {
			return aRegistration.Subscriber == registration.Subscriber &&
				aRegistration.Subscription == registration.Subscription
		})
		alreadyExist := registrationsFound.Length() > 0

		if req.Method == "POST" {
			if !alreadyExist {
				db.Register(registration)
				if err := db.WriteRegistrations(); err != nil {
					log.Println(err.Error())
					http.Error(res, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		} else if req.Method == "DELETE" {
			if alreadyExist {
				db.Unregister(registration)
				if err := db.WriteRegistrations(); err != nil {
					log.Println(err.Error())
					http.Error(res, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		} else {
			err = fmt.Errorf("Unknown verb %s", req.Method)
			log.Println(err.Error())
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.Write([]byte("{\"ok\":true}"))
	}
}
