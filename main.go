package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	age "github.com/bearbin/go-age"
	// _ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

var db *sql.DB 

const (
	host = "ec2-34-253-119-24.eu-west-1.compute.amazonaws.com"
	port = "5432"
	user = "kchiduwjmrugzz"
	password = "f2c55181c53484e8c6593e811bdf9b675f6f206c2f322a8f8f46e0d2fd553a9b"
	dbname = "dcgn7ehlnlj5sq"
)

//PERSON STRUCTURE
type Person struct {
	Id int64 `json:"id"`
	FirstName string `json:"firstName"`
	LastName string `json:"lastName"`
	Age int `json:"age"`
	Dob string `json:"dob"`
	Image string `json:"image"`
	Address string `json:"address"`
} 

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin","http://localhost:4200")
    (*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Access-Control-Allow-Headers, Origin,Accept, X-Requested-With, Content-Type, Access-Control-Request-Method, Access-Control-Request-Headers")
	/* 
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
 */    
}
//CALCULATE AGE BY DOB
func getDOB(dob string)  time.Time  {
	split := strings.Split(dob, "-")
	year,_ := strconv.Atoi(split[0])
	month,_ := strconv.Atoi(split[1])
	day,_ := strconv.Atoi(split[1])
	
   age := time.Date(year, time.Month(month),day, 0, 0, 0, 0, time.UTC)
    
   return age 
}



//GET LIST OF PERSONS
func getPersons(w http.ResponseWriter, r *http.Request){
	// setupResponse(&w, r)
	var persons []Person

	
	w.Header().Set("Content-Type","application/json")
	rows, err := db.Query("select id, firstName, lastName, dob, image, address from persons")
	if err != nil {
		log.Fatal(err)
	}
	// rows.Close()

	for rows.Next() {
		var person Person
		err := rows.Scan(&person.Id, &person.FirstName,&person.LastName, &person.Dob,&person.Image, &person.Address)
		if err != nil {
			log.Fatal(err)
		}

		// fmt.Println(person.Dob)
		// json.NewEncoder(w).Encode(rows)
		calculatedAge :=getDOB(person.Dob)
//  fmt.Println(calculatedAge)
		 person.Age = age.Age(calculatedAge) 
		persons = append(persons,person)



		// age "github.com/bearbin/go-age"
	}

	json.NewEncoder(w).Encode(persons)
}

//GET SINGLE PERSON
func getPerson(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type","application/json")
	params := mux.Vars(r)
	var person Person
    // Query for a value based on a single row.
     err := db.QueryRow("SELECT * from persons where id = $1",params["id"]).Scan(&person.Id, &person.FirstName,&person.LastName,&person.Address,&person.Dob,&person.Image)

	 if err != nil {
		// fmt.Println("in if",err)
	
		w.WriteHeader(404) // Return 500 Internal Server Error.
		return
	}
	calculatedAge :=getDOB(person.Dob)
	person.Age = age.Age(calculatedAge) 
    json.NewEncoder(w).Encode(person)
	
}



//CREATE NEW PERSON
func createPerson(w http.ResponseWriter, r *http.Request){
	// setupResponse(&w, r)
	
	w.Header().Set("Content-Type","application/json")

	var person Person

	_ = json.NewDecoder(r.Body).Decode(&person)


/* 	calculatedAge :=getDOB(person.Dob)
	fmt.Println("_____________","_________________",calculatedAge)
person.Age = age.Age(calculatedAge) */
insertSqlStatement := "INSERT INTO persons (firstName,lastName,dob,image,address) VALUES ($1, $2, $3, $4, $5) "

	_, err := db.Exec(insertSqlStatement,person.FirstName,person.LastName,person.Dob,person.Image,person.Address)
if err != nil {
	log.Fatal(err)
}
// lastId, err := res.LastInsertId()
if err != nil {
	log.Fatal(err)
}
/* rowCnt, err := res.RowsAffected()
if err != nil {
	log.Fatal(err)
} */
// person.Id=lastId
// log.Printf("ID = %d, affected = %d\n", lastId, rowCnt)

json.NewEncoder(w).Encode(&person) 

}

//UPDATE SINGLE PERSON
func updatePerson(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	var person Person
	json.NewDecoder(r.Body).Decode(&person)
	insert, err := db.Query(
		"UPDATE persons SET firstName = '" + person.FirstName + "', lastName = '" + person.LastName + "', dob = '" + person.Dob + "', image = '" + person.Image + "' , address = '" + person.Address + "' WHERE id = '" + params["id"] + "'")

	if err != nil {
		fmt.Println(err.Error())
	}
	defer insert.Close()

	json.NewEncoder(w).Encode("Person Updated Successfully") 
}

//DELETE SINGLE PERSON
func deletePerson(w http.ResponseWriter, r *http.Request)  {
	w.Header().Set("Content-Type","application/json")
	params := mux.Vars(r)
	result, err := db.Exec("Delete from persons where id=$1", params["id"])
	if err != nil {
		// return 0
	} 
		
	rowCnt, err := result.RowsAffected()
if err != nil {
	log.Fatal(err)
} 
fmt.Println(rowCnt)

json.NewEncoder(w).Encode("Person Deleted Successfully") 
	

}

func CheckError(err error){
	if err != nil {
		panic(err)
	}
}

func main(){
	// port := os.Getenv("PORT")
	// port1 := "8000"
	psqlconn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user,password,host,port,dbname)
	// psqlconn := fmt.Sprintf("host = %s port = %d user= %s password= %s dbname = %s sslmode=disable",host,port,user,password,dbname)

	tempDB,err := sql.Open("postgres",psqlconn)
	CheckError(err)


	// tempDB, err := sql.Open("mysql", "root:S@lman005@tcp(127.0.0.1:3306)/persondb")
	db = tempDB

	CheckError(err)

	// defer db.Close()

	fmt.Println("Successfull Connected")



	//INIT ROUTER
	r := mux.NewRouter()


	//MOCK DATA
	// persons = append(persons,Person{id:"0",firstName: "Salman",address: "house # 1"})

	//ROUTE HANDLERS / ENDPOINTS
	r.HandleFunc("/api/persons",getPersons).Methods("GET")
	r.HandleFunc("/api/persons/{id}",getPerson).Methods("GET")
	r.HandleFunc("/api/persons",createPerson).Methods("POST")
	r.HandleFunc("/api/persons/{id}",updatePerson).Methods("PUT")
	r.HandleFunc("/api/persons/{id}",deletePerson).Methods("DELETE")



	http.Handle("/", r)
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
	})
	handler := c.Handler(r)
	port1 := os.Getenv("PORT")
	if port1 == ""{
		port1 = "8000"
	}

	//RUN THE SERVER	
	log.Fatal(http.ListenAndServe(":" +port1, handler))
}