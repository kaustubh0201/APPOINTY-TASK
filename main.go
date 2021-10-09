package main

import (
  "fmt"
  "log"
  "net/http"
  "encoding/json"
  "context"

  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/bson/primitive"
  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/mongo/options"
)

var db *mongo.Database
var userCollection *mongo.Collection
var postCollection *mongo.Collection
var userPostCollection *mongo.Collection

func ConnectDB() *mongo.Database {

  clientOptions := options.Client().ApplyURI("mongodb://127.0.0.1:27017/?directConnection=true&serverSelectionTimeoutMS=2000")

  client, err := mongo.Connect(context.TODO(), clientOptions)

  if err != nil {
    log.Fatal(err)
  }

  fmt.Println("Connected to MongoDB!")

  database := client.Database("golang_task")

  return database
}

func init() {
  db = ConnectDB()
  userCollection = db.Collection("user")
  postCollection = db.Collection("post")
  userPostCollection = db.Collection("userPost")
}

type User struct{
  _Id primitive.ObjectID
	Name string
	Email string 
	Password string
}

type Post struct{
	_Id primitive.ObjectID
	Caption string
	ImageURL string 
	PostedTimestamp string
}

type UserPost struct{
	UserID primitive.ObjectID
	PostID []primitive.ObjectID
}

func parseID(parse string) (string){
  last_index := 0

  for i := 0; i < len(parse); i++ {
    character := parse[i:i+1]
    if(character == "/"){
      last_index = i
    }
  }

  result := parse[last_index+1:]

  return result
}

func getUserID(w http.ResponseWriter, r *http.Request) {
if r.Method == "GET" {
    w.Header().Set("Content-Type", "application/json")

    var u User

    path := r.URL.Path
    userid := parseID(path)

    realid, _ := primitive.ObjectIDFromHex(userid)

    filter := bson.M{"_id": realid}

    err := userCollection.FindOne(context.TODO(), filter).Decode(&u)

    if err != nil {
      w.Write([]byte(`{"message": "Problem Faced"}`))
    }

    json.NewEncoder(w).Encode(u)
  }
}

func createUser(w http.ResponseWriter, r *http.Request) {
  if r.Method == "POST" {
    var u User

    _ = json.NewDecoder(r.Body).Decode(&u)

    result, _ := userCollection.InsertOne(context.TODO(), u)

    var up UserPost

    EArr := make([]primitive.ObjectID, 0)

    up.UserID = (result.InsertedID).(primitive.ObjectID)
    up.PostID = EArr

    userPostCollection.InsertOne(context.TODO(), up)

    json.NewEncoder(w).Encode(result)
  }
}

type postField struct {
  UserId string
  Caption string
  ImageURL string
  PostedTimestamp string
}

func createPost(w http.ResponseWriter, r *http.Request) {
  if r.Method == "POST" {
    var pF postField

    _ = json.NewDecoder(r.Body).Decode(&pF)

    var p Post 

    p.Caption = pF.Caption
    p.ImageURL = pF.ImageURL
    p.PostedTimestamp = pF.PostedTimestamp

    result, err := postCollection.InsertOne(context.TODO(), p)
    
    if err != nil {
      w.Write([]byte(`{"message": "Failed to Create User"}`))
    } else {
      get_id, _ := primitive.ObjectIDFromHex(pF.UserId)
      query := bson.M {"_id": get_id}
      update := bson.M {"$push" : bson.M{"postids": result.InsertedID.(primitive.ObjectID)}}

      userPostCollection.FindOneAndUpdate(context.TODO(), query, update)
    }
  }
}

func getPostsID(w http.ResponseWriter, r *http.Request) {
  if r.Method == "GET" {

    w.Header().Set("Content-Type", "application/json")

    var p Post

    path := r.URL.Path
    postid := parseID(path)

    realid, _ := primitive.ObjectIDFromHex(postid)

    filter := bson.M{"_id": realid}

    err := postCollection.FindOne(context.TODO(), filter).Decode(&p)

    if err != nil {
      w.Write([]byte(`{"message": "Problem Faced"}`))
    }

    json.NewEncoder(w).Encode(p)
  }
}

func getPostUserID(w http.ResponseWriter, r *http.Request) {
  if r.Method == "GET" {

    w.Header().Set("Content-Type", "application/json")

    var up UserPost

    path := r.URL.Path
    userid := parseID(path)

    realid, _ := primitive.ObjectIDFromHex(userid)

    filter := bson.M{"_id": realid}

    err := userPostCollection.FindOne(context.TODO(), filter).Decode(&up)

    if err != nil {
      w.Write([]byte(`{"message": "Problem Faced"}`))
    }

    json.NewEncoder(w).Encode(up)
  }
}

func main() {
  http.HandleFunc("/users/", getUserID)
  http.HandleFunc("/posts/", getPostsID)
  http.HandleFunc("/users", createUser)
  http.HandleFunc("/posts", createPost)
  http.HandleFunc("/posts/users/", getPostUserID)

  log.Fatal(http.ListenAndServe(":8000", nil))
}
