package routes

import (
	"log"
	"encoding/json"
	"net/http"
	sql "../sql" 
	help "../helpers"
	c "strconv"
	s "strings"
	"time"
)

func PostComment(writer http.ResponseWriter, req *http.Request){
	//0. define structs
	type reqData struct{
		User_ID int
		Auth_Token string
		Thread_ID int
		Text string
		Username bool
	}
	type response struct{
		Error_code int
		Description string
		Data int
	}

	//1. get content from request body
	var data reqData
	jsonErr := json.NewDecoder(req.Body).Decode(&data)
	help.CheckErrorSafe(jsonErr)

	//2. pull data from object
	user   := data.User_ID
	auth   := data.Auth_Token
	thread := data.Thread_ID
	text   := data.Text
	un     := data.Username

	//3. data prep/validation
	if user == 0 || auth=="" || thread == 0 || text == "" {
		writer.WriteHeader(400)
		return
	}

	if len(text)>500{
		resp := response{12, "Comment is too long to post", -1}
		json, err := json.Marshal(resp)
		help.CheckErrorSafe(err)
		writer.Header().Set("content-type", "application/json")
		writer.Write(json)
		return
	}

	text = c.QuoteToASCII(text)
	text = text[1:len(text)-1]

	//4. you already know who it is (ft. database)

	if !sql.VerifyAuthToken(auth, user){
		resp := response{12, "Comment is too long to post", -1}
		json, err := json.Marshal(resp)
		help.CheckErrorSafe(err)
		writer.Header().Set("content-type", "application/json")
		writer.Write(json)
		return
	}

	//4.1 username bits
	query := "SELECT username FROM thread_names WHERE user_id=$1 AND thread_id=$2"
	var username string
	sqlErr := sql.Db.QueryRow(query, user, thread).Scan(&username)
	if sqlErr != nil {
		switch{
		//if they don't already have a username entered
		case s.HasSuffix(sqlErr.Error(), "no rows in result set"):
			var name, icon, color string
				//make sure generated name is unique
				for{
					newName, col, i := help.GenerateUsername()
					var temp int
					query = "SELECT id FROM thread_names where icon=$1 and color=$2 and thread_id=$3"
					sqlErr = sql.Db.QueryRow(query, icon, color, thread).Scan(&temp)
					if s.HasSuffix(sqlErr.Error(), "no rows in result set"){
						if !un{
							name = newName
						} else{
							//get username
							query = "SELECT username from users where user_id = $1"
							sqlErr = sql.Db.QueryRow(query, user).Scan(&name)
							help.CheckErrorSafe(sqlErr)
						}
						color, icon = col, i
						break
					}
				}
			var i int
			query = "INSERT INTO thread_names (thread_id, user_id, icon, color, username) VALUES ($1, $2, $3, $4, $5) RETURNING id"
			sqlErr = sql.Db.QueryRow(query, thread, user, icon, color, name).Scan(&i)
			help.CheckErrorSafe(sqlErr)

		default:
			log.Println(sqlErr)
		}
	}
	//4.2 now lets actually put the comment in the database
	query = "INSERT into comments (poster_id, parent_thread, item_text, posted_on) VALUES ($1, $2, $3, now())"
	var id int
	sqlErr = sql.Db.QueryRow(query, user, thread, text).Scan(&id)

	//5. Now that we're done there we can handles the response to the user
	resp := response{0,"",id}
	json, err := json.Marshal (resp) 
	help.CheckErrorSafe(err)
	writer.Header().Set("content-type","application/json")
	writer.Write(json)
	return
}

func CommentsForThread(writer http.ResponseWriter, req *http.Request){
	//0. define structs
	type comment struct {
		ID int
		Text string
		Poster string
		Posted string
		Score int
		Icon string
		Color string
	}
	type response struct {
		Error_code int
		Description string
		Data [help.PAGE_SIZE]comment
	}
	//1. get data from request
	thread_idURL, thread_idOK := req.URL.Query()["Thread_ID"]
	pageURL, pageOK := req.URL.Query()["Page"]

	//2. Data validation/prep
	if !thread_idOK || len(thread_idURL[0])<1 || !help.IsInt(thread_idURL[0]){
		var empty [help.PAGE_SIZE]comment
		for i:=0; i<help.PAGE_SIZE; i++{
			empty[i] = comment{-1, "", "", "", -1337, "", ""}
		}
		resp := response{11, "No thread by that id found", empty}
		json, err := json.Marshal(resp)
		help.CheckErrorSafe(err)
		writer.Header().Set("content-type", "application/json")
		writer.Write(json)
		return
	}

	var page int

	if pageOK && len(pageURL[0])>=1 && help.IsInt(pageURL[0]) {
		page, _ = c.Atoi(pageURL[0])
	}else{
		page = 0
	}

	thread_id, _ := c.Atoi(thread_idURL[0])

	//3. database interaction
	query := "SELECT id, item_text, username, posted_on, score, icon, color FROM comments LEFT JOIN thread_names on comments.id = thread_names.user_id WHERE comments.parent_thread = $1 OFFSET $2 LIMIT $3"

	rows, err := sql.Db.Query(query, thread_id, help.PAGE_SIZE*page, help.PAGE_SIZE)

	help.CheckErrorSafe(err)

	var comments [help.PAGE_SIZE]comment
	var cnt int

	for rows.Next(){
		var post time.Time
		var curr comment
		err = rows.Scan(&curr.ID, &curr.Text, &curr.Poster, &post, &curr.Score, &curr.Icon, &curr.Color)
		help.CheckErrorSafe(err)
		comments[cnt] = curr
		cnt++
	}
	for cnt<help.PAGE_SIZE{
		comments[cnt] = comment{-1, "", "", "", -1337, "", ""}
		cnt++
	}

	resp := response{0, "", comments}
	json, err := json.Marshal(resp)
	writer.Header().Set("content-type", "application/json")
	writer.Write(json)
	return
}

