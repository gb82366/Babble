package routes

import (
	"log"
	"encoding/json"
	"net/http"
	sql "../sql"
	help "../helpers"
	//"io/ioutil"
	s "strings"
	c "strconv"
)

func CreateUser(writer http.ResponseWriter, req *http.Request){
	//0. define structs for this function
	type response struct{
		Error_code int
		Description string
		Data int
	}
	type reqData struct {
		Username string
		Email string
		Password string
	}

	//1. get data from body
	var data reqData
	jsonErr := json.NewDecoder(req.Body).Decode(&data)
	help.CheckErrorSafe(jsonErr)

	//2. grab route requirements
	user     := data.Username
	email    := data.Email
	password := data.Password

	//3. verify, check and prep data
	//3.1 nil check
	if user == "" || email == "" || password == ""{
		writer.WriteHeader(400)
		return
	}
	//3.2 length check
	if len(user)>64{
		//username too long
		resp := response{1,"username exceeds max length", -1}
		json, err := json.Marshal(resp)
		help.CheckErrorSafe(err)
		writer.Header().Set("content-type", "application/json")
		writer.Write(json)
		return
	}
	if len(email)>256{
		//email too long
		resp := response{2,"email exceeds max length", -1}
		json, err := json.Marshal(resp)
		help.CheckErrorSafe(err)
		writer.Header().Set("content-type", "application/json")
		writer.Write(json)
		return
	}
	if len(password)>128{
		//password too long
		resp := response{3,"password exceeds max length", -1}
		json, err := json.Marshal(resp)
		help.CheckErrorSafe(err)
		writer.Header().Set("content-type", "application/json")
		writer.Write(json)
		return
	}

	//3.3 data prep
	user     = c.QuoteToASCII(user)
	email    = c.QuoteToASCII(s.ToLower(email))
	password = c.QuoteToASCII(help.MakePasswordHash(password))

	//4.put in db
	query := `INSERT INTO users (username, email, password)
	VALUES ($1, $2, $3) RETURNING ID`
	var id int
	sqlErr := sql.Db.QueryRow(query, user, email, password).Scan(&id)

	if sqlErr != nil{
		e := sqlErr.Error()
		//duplicates
		if s.Contains(e, "duplicate"){
			if s.Contains(e, "username"){
				resp := response{4,"Username already in use",-1}
				json, _ := json.Marshal(resp)
				writer.Header().Set("content-type", "application/json")
				writer.Write(json)
				return
			} else if s.Contains(e, "email"){
				resp := response{5,"Email already in use",-1}
				json, _ := json.Marshal(resp)
				writer.Header().Set("content-type", "application/json")
				writer.Write(json)
				return
			}
		}
		log.Println(sqlErr)
	}

	//5. return succesful response object

	//5.1 make and jsonify data
	resp := response{0,"",id}
	json, err := json.Marshal(resp)

	help.CheckErrorSafe(err)

	//5.2 set up header
	writer.Header().Set("content-type", "application/json")

	//5.3 write json to response
	writer.Write(json)

	return
}

func LoginUser(writer http.ResponseWriter, req *http.Request){
	//0. Define structsa for this route
	type user struct{
		Id int
		Token string
	}
	type response struct{
		Error_code int
		Description string
		Data user
	}
	type reqData struct{
		Username string
		Password string
	}
	//1. get content from body
	var data reqData
	jsonErr := json.NewDecoder(req.Body).Decode(&data)
	if jsonErr != nil {
		log.Println(jsonErr)
	}
	//2. get necessary data
	username := data.Username
	password := data.Password

	//3. data prep/validation 
	if username == "" || password == "" {
		writer.WriteHeader(400)
		return
	}
	if len(username)>64{
		//username too long
		resp := response{1,"username exceeds max length", user{-1, ""}}
		json, err := json.Marshal(resp)
		help.CheckErrorSafe(err)
		writer.Header().Set("content-type", "application/json")
		writer.Write(json)
		return
	}
	if len(password)>128{
		//password too long
		resp := response{3,"password exceeds max length", user{-1, ""}}
		json, err := json.Marshal(resp)
		help.CheckErrorSafe(err)
		writer.Header().Set("content-type", "application/json")
		writer.Write(json)
		return
	}
	username = c.QuoteToASCII(username)

	//4. database interaction
	query := "SELECT id,password FROM users WHERE username=$1"
	var id int
	var pwd string
	err := sql.Db.QueryRow(query, username).Scan(&id, &pwd)
	if err != nil {
		if !s.HasSuffix(err.Error(), "no rows in result set"){
			log.Println(err.Error())
		}
	}
	if help.CheckPasswordHash(password, pwd) {
		//generate new token
		new_auth := help.GenerateAuthToken(24)
		//update token
		query = "UPDATE users SET auth_token=$1 WHERE id=$2"
		_, err = sql.Db.Exec(query, new_auth, id)
		if err != nil {
			log.Println(err)
		}
		//return this token
		resp := response{0, "", user{id, new_auth}}
		json, err := json.Marshal(resp)
		if err != nil {
			log.Println(err)
		}
		writer.Header().Set("content-type", "application/json")
		writer.Write(json)
		return
	}

	//return error
	var x user = user{-1, ""}
	var resp response = response{1, "invalid username/password", x} 
	json, err := json.Marshal(resp)
	writer.Header().Set("content-type", "application/json")
	writer.Write(json)
	return

}

func ChangePassword(writer http.ResponseWriter, req *http.Request){
}

func UserProfile(writer http.ResponseWriter, req *http.Request){
}

