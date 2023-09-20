package myMongo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"time"

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

var config *Config

type Config struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     string `json:"port"`
}

func Init(configPath string) error {
	err := readConfig(configPath)
	if err != nil {
		return err
	}
	clientOptions := options.Client().ApplyURI("mongodb://" + config.Host + ":" + config.Port).SetAuth(options.Credential{
		Username: config.Username,
		Password: config.Password,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}
	return nil
}

func readConfig(configFilePath string) error {
	// Remove the packageDir logic, as we're now using the provided configFilePath.
	configFile, err := os.Open(configFilePath)
	if err != nil {
		return err
	}
	defer configFile.Close()
	var tempConfig map[string]interface{}
	decoder := json.NewDecoder(configFile)
	if err := decoder.Decode(&tempConfig); err != nil {
		return err
	}
	// Actual checking step
	config = &Config{}
	configValue := reflect.ValueOf(config).Elem()
	for i := 0; i < configValue.NumField(); i++ {
		fieldName := configValue.Type().Field(i).Tag.Get("json")
		if value, exists := tempConfig[fieldName]; exists {
			configFieldValue := configValue.Field(i)
			configFieldType := configFieldValue.Type()
			if configFieldType.Kind() == reflect.String {
				configFieldValue.SetString(value.(string))
			} else {
				return fmt.Errorf("unsupported field type for '%s'", fieldName)
			}
		} else {
			return fmt.Errorf("missing key '%s' in the config file", fieldName)
		}
	}
	return nil
}

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
