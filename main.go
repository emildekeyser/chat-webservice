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
	commentChannel = make(chan comment)
	messages = make([]*message, 0, 100)
)

type user struct {
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Status   string   `json:"status"`
	Friends  []string `json:"friends"`
}

type comment struct {
	Name string `json:"name"`
	Rating string `json:"rating"`
	Text string `json:"text"`
}

type friend struct {
	Name string `json:name`
	Email string `json:email`
	Status string `json:status`
}

type message struct {
	From string `json:from`
	To string `json:to`
	Msg string `json:msg`
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

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")
		name := r.FormValue("name")
		password := r.FormValue("password")
		_, exists := users[email]
		valid := !exists && email != "" && password != "" && name != ""
		if valid {
			users[email] = &user{
				Name: name,
				Email: email,
				Password: password,
				Status:   "offline",
				Friends:  make([]string, 0)}
			file, _ := ioutil.ReadFile("web/login.html")
			fmt.Fprintf(w, "%s", file)
		} else {
			file, _ := ioutil.ReadFile("web/register.html")
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
		friendEmail := r.FormValue("newFriend")
		session, _ := store.Get(r, "auth")
		email := session.Values["authenticated"].(string)

		if users[email] == nil || users[friendEmail] == nil {
			return
		}
		if !contains(users[email].Friends, friendEmail) {
			users[email].Friends = append(users[email].Friends, friendEmail)
		}
		if !contains(users[friendEmail].Friends, email) {
			users[friendEmail].Friends = append(users[friendEmail].Friends, email)
		}
		// fmt.Printf("List: %+v", users[email].Friends)
	}))

	http.HandleFunc("/friends", authenticate(func(w http.ResponseWriter, r *http.Request) {
		// email := r.URL.Path[len("/friends/"):]
		email := ""
		// fmt.Printf("email: %s\n", email)
		if email == "" {
			session, err := store.Get(r, "auth")
			if err == nil {
				email = session.Values["authenticated"].(string)
			} else {
				fmt.Fprintf(w, "%s", "")
				return
			}
		}
		friendsJson, _ := json.Marshal(MakeFriendlist(email))
		fmt.Fprintf(w, "%s", string(friendsJson))
	}))

	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		usersSlice := make([]*user, 0, len(users))
		for _, value := range users {
			usersSlice = append(usersSlice, value)
		}
		usersJson, _ := json.Marshal(usersSlice)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprintf(w, "%s", string(usersJson))
	})

	http.HandleFunc("/comment", func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil)
		defer conn.Close()
		commentConnections[conn] = true
		for {
			var msg comment
			err := conn.ReadJSON(&msg)
			if err != nil {
				delete(commentConnections, conn)
				return
			}
			fmt.Printf("%s sent: %+v with type %d\n", conn.RemoteAddr(), msg)
			commentChannel <- msg
		}
	})
	go handlecomments()

	http.HandleFunc("/sendmsg", func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "auth")
		email, _ := session.Values["authenticated"].(string)
		to := r.FormValue("msgreceiver")
		msg := r.FormValue("msg")
		valid := users[email] != nil && users[to] != nil
		newMsg := &message{
			From: email,
			To: to,
			Msg: msg }
		if valid {
			messages = append(messages, newMsg)
		}
		// need to return something
		fmt.Fprintf(w, "{}")
	})

	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "auth")
		from, _ := session.Values["authenticated"].(string)
		to := r.FormValue("msgreceiver")
		valid := users[from] != nil && users[to] != nil
		correspondance := make([]*message, 0, 100)
		if valid {
			for _, m := range messages{
				if (m.From == from && m.To == to) ||
					(m.From == to && m.To == from){
					correspondance = append(correspondance, m)
				}
			}
		}
		cj, _ := json.Marshal(correspondance)
		fmt.Fprintf(w, string(cj))

	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		file, _ := ioutil.ReadFile("web/login.html")
		fmt.Fprintf(w, "%s", file)
	})

	http.Handle("/web/js/", http.StripPrefix("/web/js", http.FileServer(http.Dir("./web/js"))))
	http.ListenAndServe(":8080", nil)
}

func handlecomments() {
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
		// session, _ := store.Get(r, "auth")
		// email, ok := session.Values["authenticated"].(string)
		// if !ok || email == "" {
		// 	http.Error(w, "Forbidden", http.StatusForbidden)
		// 	return
		// }
		f(w, r)
	}
}

func MakeFriendlist(email string) []*friend {
	var friendList []*friend
	if users[email] == nil {
		return friendList
	}
	for _, friendEmail := range users[email].Friends {
		f := &friend{
			Name: users[friendEmail].Name,
			Email: users[friendEmail].Email,
			Status: users[friendEmail].Status,
		}
		friendList = append(friendList, f)
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

func contains(slice []string, element string) bool {
	for _, s := range slice {
		if s == element {
			return true
		}
	}
	return false
}
