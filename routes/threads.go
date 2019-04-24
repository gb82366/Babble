package routes

import(
	"encoding/json"
	"net/http"
	"log"
	c "strconv"
	s "strings"
	sql "../sql"
	help "../helpers"
	"time"
)

func PostThread(writer http.ResponseWriter, req *http.Request){
	//0. define structs for this route
	type reqData struct {
		User_ID int
		Auth_Token string
		Lat float64
		Long float64
		Text string
	}
	type response struct {
		Error_code int
		Description string
		Data int
	}

	//1. get content from request body
	var data reqData
	jsonErr := json.NewDecoder(req.Body).Decode(&data)
	if jsonErr != nil{
		log.Println(jsonErr)
	}

	//2. pull data from object
	id          := data.User_ID
	auth        := data.Auth_Token
	lat         := c.FormatFloat(data.Lat, 'f', -1, 64)
	long        := c.FormatFloat(data.Long, 'f', -1, 64)
	pointString := "ST_GeomFromText('POINT("+long+" "+lat+")')"
	text        := data.Text

	//3. data prep/validation

	if id == 0 || auth == "" || text == "" {
		writer.WriteHeader(400)
		return
	}

	if len(text)>1000{
		resp := response{7, "Thread is too long to post", -1}
		json, err := json.Marshal(resp)
		help.CheckErrorSafe(err)
		writer.Header().Set("content-type", "application/json")
		writer.Write(json)
		return
	}

	text = c.QuoteToASCII(text)
	text = text[1:len(text)-1]

	//4. database time
	if !sql.VerifyAuthToken(auth, id){
		resp := response{6, "user is not authorized to make this request", -1}
		json, err := json.Marshal(resp)
		help.CheckErrorSafe(err)
		writer.Header().Set("content-type", "application/json")
		writer.Write(json)
		return
	}

	var thread_id int
	t  := time.Now()
	query := "INSERT INTO threads (poster_id, item_text, location, posted_on) VALUES ($1, $2, "+pointString+", $3) RETURNING ID"
	sqlErr := sql.Db.QueryRow(query, id, text, t).Scan(&thread_id)
	help.CheckErrorSafe(sqlErr)

	//5. handle reponse
	resp := response{0, "", thread_id}
	json, err := json.Marshal(resp)
	help.CheckErrorSafe(err)
	writer.Header().Set("content-type", "application/json")
	writer.Write(json)
	return
}

func ViewAreaThreads(writer http.ResponseWriter, req *http.Request){
	//0. set up structs/helpers
	type thread struct{
		ID int
		Text string
		Score int
		Posted string
	}
	type response struct{
		Error_code int
		Description string
		Data [help.PAGE_SIZE]thread
	}

	//1. get data from request

	latURL, laOK  := req.URL.Query()["Lat"]
	longURL, loOK := req.URL.Query()["Long"]
	distURL, dOK  := req.URL.Query()["Distance"]
	pageURL, pOK  := req.URL.Query()["Page"]


	//2. data prep/validation

	//2.1 latitude
	if !laOK || len(latURL[0])<1 || !help.IsFloat(latURL[0]){
		var empty [help.PAGE_SIZE]thread
		for i:=0; i<help.PAGE_SIZE; i++{
			empty[i] = thread{-1, "", -1337, ""}
		}
		resp := response{8, "latitude missing/malformed", empty}
		json, err := json.Marshal(resp)
		help.CheckErrorSafe(err)
		writer.Header().Set("content-type", "application/json")
		writer.Write(json)
		return
	}
	lat := latURL[0]

	//longitude
	if !loOK || len(longURL[0])<1 || !help.IsFloat(longURL[0]){
		var empty [help.PAGE_SIZE]thread
		for i:=0; i<help.PAGE_SIZE; i++{
			empty[i] = thread{-1, "", -1337, ""}
		}
		resp := response{9, "longitude missing/malformed", empty}
		json, err := json.Marshal(resp)
		help.CheckErrorSafe(err)
		writer.Header().Set("content-type", "application/json")
		writer.Write(json)
		return
	}
	long := longURL[0]

	//search radius
	var dist float64 = 1.0
	if dOK && len(distURL[0])>0 {
		d, derr := c.ParseFloat(distURL[0], 64)
		if derr != nil {
			log.Println("defaulting distance")
			dist = 1.0
		}
		dist = d
	}

	//page
	var page int64 = 0
	if pOK && len(pageURL[0])>0 {
		p, perr := c.ParseInt(pageURL[0], 10, 32)
		log.Println("parsed pg #", p)
		if perr != nil {
			log.Println("defaulting page")
			page = 0
		}
		page = p
	}

	pointString := "ST_GeomFromText('POINT("+long+" "+lat+")')"

	if page<0{
		page = 0
	}
	if dist<0{
		dist = 1.0
	}

	//3. database interaction
	query := "SELECT id, item_text, score, posted_on FROM threads WHERE CAST(ST_DISTANCE(location, "+pointString+") AS numeric)/1609.344 < $1 ORDER BY posted_on DESC OFFSET $2 LIMIT $3"

	rows, err := sql.Db.Query(query, dist, page*help.PAGE_SIZE, help.PAGE_SIZE)

	if err != nil {
		log.Println(err)
	}

	defer rows.Close()

	var threads [help.PAGE_SIZE]thread
	var cnt int

	for rows.Next(){ // <--- I still hate this
		var post time.Time
		var curr thread
		err = rows.Scan(&curr.ID, &curr.Text, &curr.Score, &post)
		if err != nil {
			log.Println(err)
		}
		curr.Posted = help.TimeDelta(post)

		threads[cnt]=curr
		cnt +=1
	}
	for cnt<help.PAGE_SIZE{
		threads[cnt]=thread{-1, "", -1337, ""}
		cnt +=1
	}

	//4. handle response 
	resp := response{0,"",threads}
	json, err := json.Marshal(resp)
	help.CheckErrorSafe(err)
	writer.Header().Set("content-type", "application/json")
	writer.Write(json)
	return
}

func ViewSingleThread(writer http.ResponseWriter, req *http.Request){
	//0. define structs
	type thread struct{
		ID int
		Text string
		Score int
		Posted string
	}
	type response struct{
		Error_code int
		Description string
		Data thread
	}
	//1. get data from request
	idURL, idOK  := req.URL.Query()["ID"]
	//2. data prep/validation
	if !idOK || len(idURL[0])<1 || !help.IsInt(idURL[0]){
		empty := thread{-1, "", -1337, ""}
		resp := response{10, "id missing/malformed", empty}
		json, err := json.Marshal(resp)
		help.CheckErrorSafe(err)
		writer.Header().Set("content-type", "application/json")
		writer.Write(json)
		return
	}
	id := idURL[0]

	//3. database interaction
	query := "SELECT id, item_text, score, posted_on FROM threads WHERE threads.id=$1"

	var t time.Time
	var thr thread

	sqlErr := sql.Db.QueryRow(query, id).Scan(&thr.ID, &thr.Text, &thr.Score, &t)

	if sqlErr != nil {
		switch{
		case s.HasSuffix(sqlErr.Error(), "no rows in result set"):
			empty := thread{-1, "", -1337, ""}
			resp := response{11, "No thread by that ID found", empty}
			json, err := json.Marshal(resp)
			help.CheckErrorSafe(err)
			writer.Header().Set("content-type", "application/json")
			writer.Write(json)
			return

		default:
			log.Println(sqlErr)
		}
	}
	//4. handle response
	thr.Posted = help.TimeDelta(t)
	resp := response{0,"",thr}
	json, err := json.Marshal(resp)
	help.CheckErrorSafe(err)
	writer.Header().Set("content-type", "application/json")
	writer.Write(json)
	return
}

