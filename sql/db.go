package sql

import (
    "database/sql"
    _ "github.com/lib/pq"
    //"log"
    "io/ioutil"
    s "strings"
    help "../helpers"
)

var Db *sql.DB

var user,dbname,sslmode,pass string

func init(){
  settings, err :=ioutil.ReadFile("config/database.cfg")
  help.CheckErrorFatal(err)
  setList := s.Split(string(settings), "\n")
  //log.Println(setList)
  user    = s.Split(setList[0], "=")[1]
  pass    = s.Split(setList[1], "=")[1]
  dbname  = s.Split(setList[2], "=")[1]
  sslmode = s.Split(setList[3], "=")[1]
  connString := "user="+user+" password="+pass+" dbname="+dbname+" sslmode="+sslmode
  //log.Println(connString)
  Db, err = sql.Open("postgres",connString)
  help.CheckErrorFatal(err)
}

func VerifyAuthToken(authToken string, userId int) bool {
  query :="SELECT auth_token FROM users WHERE id=$1"
  var userToken string
  err := Db.QueryRow(query, userId).Scan(&userToken)
  if err !=nil{
    return false
  }
  if userToken == authToken{
    return true
  }
  return false
}
