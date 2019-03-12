package main

import (
	"fmt"
	//"html"
	//"log"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
)

var (
	key      = []byte("super-secret-key")
	store    = sessions.NewCookieStore(key)
	users    = MakeUsers()
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	commentConnections = make(map[*websocket.Conn]bool)
	commentChannel = make(chan Comment)
)

type user struct {
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Status   string   `json:"status"`
	Friends  []string `json:"friends"`
}

type Comment struct {
	Name string `json:"name"`
	Rating string `json:"rating"`
	Text string `json:"text"`
}

func main() {
	fmt.Printf("server up\n")

	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")
		password := r.FormValue("password")
		if _, exists := users[email]; exists && users[email].Password == password {
			session, _ := store.Get(r, "auth")
			session.Values["authenticated"] = email
			session.Save(r, w)
			users[email].Status = "online"
			file, _ := ioutil.ReadFile("web/chat.html")
			fmt.Fprintf(w, "%s", file)
		} else {
			file, _ := ioutil.ReadFile("web/login.html")
			fmt.Fprintf(w, "%s", file)
		}
	})

	http.HandleFunc("/logout", authenticate(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "auth")
		session.Values["authenticated"] = nil
		session.Save(r, w)
		http.ServeFile(w, r, "web/login.html")
	}))

	http.HandleFunc("/status", authenticate(func(w http.ResponseWriter, r *http.Request) {
		status := r.FormValue("newStatus")
		session, _ := store.Get(r, "auth")
		email := session.Values["authenticated"].(string)
		users[email].Status = status
	}))

	http.HandleFunc("/addfriend", authenticate(func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("newFriend")
		session, _ := store.Get(r, "auth")
		email := session.Values["authenticated"].(string)
		friendEmail := name + "@ucll.be"
		users[email].Friends = append(users[email].Friends, friendEmail)
		users[friendEmail].Friends = append(users[friendEmail].Friends, email)
		fmt.Printf("List: %+v", users[email].Friends)
	}))

	http.HandleFunc("/friends", authenticate(func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "auth")
		email := session.Values["authenticated"].(string)
		friendsJson, _ := json.Marshal(MakeFriendlist(email))
		//fmt.Printf("%s\n", string(friendsJson))
		fmt.Fprintf(w, "%s", string(friendsJson))
	}))

	http.HandleFunc("/comment", func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()
		commentConnections[conn] = true
		for {
			var msg Comment
			err := conn.ReadJSON(&msg)
			if err != nil {
				delete(commentConnections, conn)
				return
			}
			fmt.Printf("%s sent: %+v with type %d\n", conn.RemoteAddr(), msg)
			commentChannel <- msg
		}
	})
	go handleComments()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		file, _ := ioutil.ReadFile("web/login.html")
		fmt.Fprintf(w, "%s", file)
	})

	http.Handle("/web/js/", http.StripPrefix("/web/js", http.FileServer(http.Dir("./web/js"))))
	http.ListenAndServe(":8080", nil)
}

func handleComments() {
	for {
		msg := <-commentChannel
		for conn := range commentConnections {
			if err := conn.WriteJSON(msg); err != nil {
				conn.Close()
				delete(commentConnections, conn)
			}
		}
	}
}

func authenticate(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "auth")
		email, ok := session.Values["authenticated"].(string)
		if !ok || email == "" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		f(w, r)
	}
}

func MakeFriendlist(email string) map[string]string {
	friendList := make(map[string]string)
	for _, friend := range users[email].Friends {
		name := users[friend].Name
		status := users[friend].Status
		friendList[name] = status
	}
	return friendList
}

func MakeUsers() map[string]*user {
	users := make(map[string]*user)
	users["jan@ucll.be"] = &user{
		Name:     "jan",
		Email:    "jan@ucll.be",
		Password: "t",
		Status:   "offline",
		Friends:  []string{"an@ucll.be", "artyom@ucll.be"}}
	users["an@ucll.be"] = &user{
		Name:     "an",
		Email:    "an@ucll.be",
		Password: "x",
		Status:   "offline",
		Friends:  []string{"artyom@ucll.be"}}
	users["artyom@ucll.be"] = &user{
		Name:     "artyom",
		Email:    "artyom@ucll.be",
		Password: "o",
		Status:   "offline",
		Friends:  []string{"an@ucll.be"}}
	users["tony@ucll.be"] = &user{
		Name:     "tony",
		Email:    "tony@ucll.be",
		Password: "o",
		Status:   "offline",
		Friends:  []string{"an@ucll.be"}}
	return users
}
