package routes

import (
	"log"
	"encoding/json"
	"net/http"
	sql "../sql" 
	help "../helpers"
	"github.com/lib/pq"
)

func ThreadVote(writer http.ResponseWriter, req *http.Request){
	//0. Define Structs
	type reqData struct{
		User_ID int
		Thread_ID int
		Auth_Token string
		Vote int
	}
	type response struct{
		Error_code int
		Description string
		Data int
	}
	//1. get data from body
	var data reqData
	jsonErr := json.NewDecoder(req.Body).Decode(&data)
	help.CheckErrorSafe(jsonErr)
	//2. grab route requirements 
	user   := data.User_ID
	thread := data.Thread_ID
	auth   := data.Auth_Token
	vote   := data.Vote

	if user == 0 || thread == 0 || auth == "" || vote > 1 || vote < -1 {
		writer.WriteHeader(400)
		return
	}

	//3. check user auth token
	if !sql.VerifyAuthToken(auth, user){ 
		resp := response{6, "user is not authorized to make this request", -1}
		json, err := json.Marshal(resp)
		help.CheckErrorSafe(err)
		writer.Header().Set("content-type", "application/json")
		writer.Write(json)
		return
	}
	//4. database interaction
	var vote_id int
	scoreQuery := "UPDATE threads SET score = (SELECT sum(vote) from thread_votes where thread_id=$1 GROUP BY thread_id) WHERE id=$1 RETURNING score"
	voteQuery  := "INSERT INTO thread_votes (thread_id, voter_id, vote) VALUES ($1, $2, $3) RETURNING id"

	//4.1 try to put the vote in
	sqlErr := sql.Db.QueryRow(voteQuery, thread, user, vote).Scan(&vote_id)
	pgErr := sqlErr.(*pq.Error)
	if pgErr != nil {
		if pgErr.Code == "23505" {//unique violation
			query := "UPDATE thread_votes SET vote=$1 WHERE thread_id=$2 AND voter_id=$3 RETURNING id"
			sqlErr = sql.Db.QueryRow(query, vote, thread, user).Scan(&vote_id)

		}else if pgErr.Code == "23503"{//foreign violation
			resp := response{14, "thread not found", -1}
			json, err := json.Marshal(resp)
			help.CheckErrorSafe(err)
			writer.Header().Set("content-type", "application/json")
			writer.Write(json)
			return

		}else{
			log.Println(sqlErr)
		}
	}

	//4.2 update thread's score
	var newScore int
	sqlErr = sql.Db.QueryRow(scoreQuery, thread).Scan(&newScore)
	if sqlErr != nil {
		log.Println(sqlErr)
	}
	if newScore < -5 {
		delQuery := "DELETE from threads where id=$1"
		sqlErr := sql.Db.QueryRow(delQuery, thread).Scan()
		help.CheckErrorSafe(sqlErr)
	}
	//5. handle return
	resp := response{0, "", vote_id}
	json, err := json.Marshal(resp)
	help.CheckErrorSafe(err)
	writer.Header().Set("content-type", "application/json")
	writer.Write(json)
	return
}

func CommentVote(writer http.ResponseWriter, req *http.Request){
	//0. Define structs
	type reqData struct{
		User_ID int
		Comment_ID int
		Auth_Token string
		Vote int
	}
	type response struct{
		Error_code int
		Description string
		Data int
	}
	//1. get data from body
	var data reqData
	jsonErr := json.NewDecoder(req.Body).Decode(&data)
	help.CheckErrorSafe(jsonErr)
	//2. grab route requirements 
	user    := data.User_ID
	comment := data.Comment_ID
	auth    := data.Auth_Token
	vote    := data.Vote

	if user == 0 || comment == 0 || auth == "" || vote > 1 || vote < -1 {
		writer.WriteHeader(400)
		return
	}
	//3. check user auth token
	if !sql.VerifyAuthToken(auth, user){ 
		resp := response{6, "user is not authorized to make this request", -1}
		json, err := json.Marshal(resp)
		help.CheckErrorSafe(err)
		writer.Header().Set("content-type", "application/json")
		writer.Write(json)
		return
	}
	//4. database interaction
	var vote_id int
	scoreQuery :=  "UPDATE comments SET score = (SELECT sum(vote) from comment_votes where comment_id=$1 GROUP BY comment_id) WHERE id=$1 RETURNING score"
	voteQuery  := "INSERT INTO comment_votes (comment_id, voter_id, vote) VALUES ($1, $2, $3) RETURNING id"

	//4.1 try to put vote in
	sqlErr := sql.Db.QueryRow(voteQuery, comment, user, vote).Scan(&vote_id)
	pgErr := sqlErr.(*pq.Error)
	if pgErr != nil {

		if pgErr.Code == "23505" {//unique violation 
			query := "UPDATE comment_votes SET vote=$1 WHERE comment_id=$2 AND voter_id=$3 RETURNING id"
			sqlErr = sql.Db.QueryRow(query, vote, comment, user).Scan(&vote_id)

		}else if pgErr.Code == "23503"{//foreign key violation
			resp := response{15, "comment not found", -1}
			json, err := json.Marshal(resp)
			help.CheckErrorSafe(err)
			writer.Header().Set("content-type", "application/json")
			writer.Write(json)
			return

		}else{
			log.Println(sqlErr)
		}
	}

	//4.2 update score and delete if necessary
	var newScore int
	sqlErr = sql.Db.QueryRow(scoreQuery, comment).Scan(&newScore)
	if sqlErr != nil {
		log.Println(sqlErr)
	}
	if newScore < -5 {
		delQuery := "DELETE from comments where id=$1"
		sqlErr := sql.Db.QueryRow(delQuery, comment).Scan()
		help.CheckErrorSafe(sqlErr)
	}

	//5. handle return
	resp := response{0, "", vote_id}
	json, err := json.Marshal(resp)
	help.CheckErrorSafe(err)
	writer.Header().Set("content-type", "application/json")
	writer.Write(json)
	return

}

