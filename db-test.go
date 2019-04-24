package main

import (
    "fmt"
    "log"
    "database/sql"
    _ "github.com/lib/pq"
)

func main (){
  var (
      id int
      name string
  )

  db, err := sql.Open("postgres", "user=beebgar password=72beebe72 dbname=babble sslmode=disable")
  if err != nil {
    log.Fatal("bad connection string")
  }

  query, err := db.Prepare("SELECT id, username from users")
  if err != nil {
    log.Println("1")
    log.Fatal(err)
  }
  defer query.Close()

  rows, err := query.Query()
  if err != nil{
    log.Println("2")
    log.Fatal(err)
  }
  defer rows.Close()

  for rows.Next(){
    err := rows.Scan(&id, &name)
    if err != nil{
      log.Println("3")
      log.Fatal(err)
    }
    log.Println(id, name)
  }
  if err = rows.Err(); err!= nil {
    log.Println("4")
    log.Fatal(err)
  }

  fmt.Println("Made it to end")
}
