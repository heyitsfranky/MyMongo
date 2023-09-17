package myMongo

import (
	"context"
	"testing"
)

type TestObject struct {
	UUID  string  `json:"uuid"`
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

func Test_Init(t *testing.T) {
	configPath := "template_MyMongo.config.json"
	err := Init(configPath)
	defer client.Disconnect(context.Background())
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
	if client == nil {
		t.Error("Expected DBClient to be non-nil after initialization, but it's nil")
	}
}

func Test_CreateFilterQuery(t *testing.T) {
	query := CreateFilterQuery("firstkey", "firstvalue", "nextkey", "nextvalue")
	expectedQuery := `{"firstkey":"firstvalue","nextkey":"nextvalue"}`
	if query != expectedQuery {
		t.Errorf("Expected query: %s, but got: %s", expectedQuery, query)
	}
}

func Test_All_CRUD_Operations(t *testing.T) {
	configPath := "template_MyMongo.config.json"
	err := Init(configPath)
	if err != nil {
		t.Fatalf("Failed to initialize MongoDB client: %v", err)
	}
	defer client.Disconnect(context.Background())
	testData := []TestObject{
		{UUID: "1", Name: "Object", Value: 42.0},
		{UUID: "2", Name: "Object", Value: 37.5},
		{UUID: "3", Name: "Object", Value: 12.8},
	}
	dbName := "myMongoTestDB"
	collectionName := "myMongoTestCollection"
	// Step 1: Create multiple objects
	for _, data := range testData {
		dataMap := map[string]interface{}{
			"uuid":  data.UUID,
			"name":  data.Name,
			"value": data.Value,
		}
		err := PerformDatabaseAction(dbName, collectionName, Create, dataMap)
		if err != nil {
			t.Fatalf("Failed to create object: %v", err)
		}
	}
	// Step 2: Search for one object of type TestObject
	obj, err := GetObject[*TestObject](CreateFilterQuery("uuid", "1"), dbName, collectionName)
	if err != nil || obj == nil {
		t.Fatalf("Failed to retrieve one object: %v", err)
	}
	// Step 3: Update all created objects
	for _, data := range testData {
		dataMap := map[string]interface{}{
			"uuid":  data.UUID,
			"name":  data.Name,
			"value": data.Value * 2, // Update the value
		}
		err := PerformDatabaseAction(dbName, collectionName, Update, dataMap)
		if err != nil {
			t.Fatalf("Failed to update object: %v", err)
		}
	}
	// Step 4: Search for multiple objects of type TestObject
	objects, err := GetMultipleObjects[*TestObject](CreateFilterQuery("name", "Object"), dbName, collectionName)
	if err != nil || objects == nil {
		t.Fatalf("Failed to retrieve multiple objects: %v", err)
	}
	// Ensure the expected number of objects is retrieved
	if len(objects) != len(testData) {
		t.Fatalf("Expected %d objects, but got %d", len(testData), len(objects))
	}
	// Step 5: Delete all created (and updated) objects
	for _, data := range testData {
		dataMap := map[string]interface{}{
			"uuid": data.UUID,
		}
		err := PerformDatabaseAction(dbName, collectionName, Delete, dataMap)
		if err != nil {
			t.Fatalf("Failed to delete object: %v", err)
		}
	}
}
