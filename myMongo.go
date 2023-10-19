package myMongo

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/heyitsfranky/MyConfig/src/myConfig"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type ActionET int

const (
	Create ActionET = iota
	Delete
	Update
)

var Data *InitData

type InitData struct {
	Username string `yaml:"mongo-username"`
	Password string `yaml:"mongo-password"`
	Host     string `yaml:"mongo-host"`
	Port     int    `yaml:"mongo-port"`
}

func Init(configPath string) error {
	if Data == nil {
		err := myConfig.Init(configPath, &Data)
		if err != nil {
			return err
		}
	}
	clientOptions := options.Client().ApplyURI("mongodb://" + Data.Host + ":" + strconv.Itoa(Data.Port)).SetAuth(options.Credential{
		Username: Data.Username,
		Password: Data.Password,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}
	return nil
}

// Creates simple filter queries
// i.e. "key1", "value1", "key2", "value2", etc.
// e.g. "id", 5, "name", "John", "age", 42
func CreateFilterQuery(values ...interface{}) string {
	query := make(map[string]interface{})
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			panic("Invalid key type")
		}
		query[key] = values[i+1]
	}
	jsonBytes, err := json.Marshal(query)
	if err != nil {
		panic(err)
	}
	return string(jsonBytes)
}

func get(filter string, dbName string, collectionName string) (*mongo.Cursor, error) {
	collection := client.Database(dbName).Collection(collectionName)
	var filterBson bson.M
	if filter != "" {
		err := bson.UnmarshalExtJSON([]byte(filter), true, &filterBson)
		if err != nil {
			return nil, err
		}
	}
	cur, err := collection.Find(context.Background(), filterBson)
	if err != nil {
		return nil, err
	}
	return cur, nil
}

func PerformDatabaseAction(dbName string, collectionName string, givenAction ActionET, givenData map[string]interface{}) error {
	collection := client.Database(dbName).Collection(collectionName)
	var err error
	uuid, ok := givenData["uuid"].(string)
	if !ok {
		return errors.New("missing or invalid 'uuid' key in received data")
	}
	switch givenAction {
	case ActionET(Create):
		_, err = collection.InsertOne(context.Background(), givenData)
	case ActionET(Delete):
		filter := bson.M{"uuid": uuid}
		_, err = collection.DeleteOne(context.Background(), filter)
	case ActionET(Update):
		update := bson.M{"$set": givenData}
		_, err = collection.UpdateOne(context.Background(), bson.M{"uuid": uuid}, update)
	default:
		return errors.New("unsupported action")
	}
	if err != nil {
		return err
	}
	return nil
}

func GetObject[T any](filter string, dbName string, collectionName string) (T, error) {
	var elem T
	cur, err := get(filter, dbName, collectionName)
	if err != nil {
		return elem, err
	}
	defer cur.Close(context.Background())
	// Check if there is at least one result
	if cur.Next(context.Background()) {
		err = cur.Decode(&elem)
		if err != nil {
			return elem, err
		}
		return elem, nil
	}
	if err = cur.Err(); err != nil {
		return elem, err
	}
	//no object found
	return elem, nil
}

func GetMultipleObjects[T any](filter string, dbName string, collectionName string) ([]T, error) {
	var results []T
	cur, err := get(filter, dbName, collectionName)
	if err != nil {
		return results, err
	}
	defer cur.Close(context.Background())
	for cur.Next(context.Background()) {
		var elem T
		err = cur.Decode(&elem)
		if err != nil {
			return results, err
		}
		results = append(results, elem)
	}
	if err = cur.Err(); err != nil {
		return results, err
	}
	//no objects found
	return results, nil
}

// Creates a complex BSON filter query
// i.e. bson.M{"key": bson.M{"operation": value}}
// e.g. bson.M{"age": bson.M{"$ne": 42}}
func CreateAdvancedFilterQuery(key string, operation string, value interface{}) string {
	filter := bson.M{key: bson.M{operation: value}}
	jsonBytes, err := bson.MarshalExtJSON(filter, true, false)
	if err != nil {
		panic(err)
	}
	return string(jsonBytes)
}

// Takes any BSON and creates a filter query
// Allows for more flexibility than CreateAdvancedFilterQuery, however bson.M must then be included in the package that is calling this function.
func CreateBSONFilterQuery(filter bson.M) string {
	jsonBytes, err := bson.MarshalExtJSON(filter, true, false)
	if err != nil {
		panic(err)
	}
	return string(jsonBytes)
}
