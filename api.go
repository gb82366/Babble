package main

import (
    "log"
    "net/http"
    "github.com/gorilla/mux"
    routes "./routes"
)

func main (){
  router := mux.NewRouter()
  //templates for copy/paste
  //router.HandleFunc("", routes.).Methods("POST")
  //router.HandleFunc("", routes.).Methods("GET")
  router.HandleFunc("/users/create", routes.CreateUser).Methods("POST")
  router.HandleFunc("/users/login", routes.LoginUser).Methods("POST")
  router.HandleFunc("/users/changePassword", routes.ChangePassword).Methods("POST")
  router.HandleFunc("/users/profile", routes.UserProfile).Methods("GET")
  router.HandleFunc("/threads/post", routes.PostThread).Methods("POST")
  router.HandleFunc("/threads/view/nearby", routes.ViewAreaThreads).Methods("GET")
  router.HandleFunc("/threads/view/single", routes.ViewSingleThread).Methods("GET")
  router.HandleFunc("/comments/post", routes.PostComment).Methods("POST")
  router.HandleFunc("/comments/thread", routes.CommentsForThread).Methods("GET")
  router.HandleFunc("/vote/thread", routes.ThreadVote).Methods("POST")
  router.HandleFunc("/vote/comment", routes.CommentVote).Methods("POST")

  log.Println("Serving on port 46442")
  log.Println(http.ListenAndServe(":46442", router))
}
