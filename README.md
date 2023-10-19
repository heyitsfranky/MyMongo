# MyMongo

**MyMongo** is a Go package designed for simplified interaction with MongoDB databases. It offers easy-to-use functions for CRUD operations and object retrieval from MongoDB collections.

## Features

- Initialize a MongoDB client with authentication using a configuration file.
- Perform Create, Read, Update, and Delete (CRUD) operations on MongoDB collections.
- Easily retrieve single or multiple objects from a collection based on filter criteria.
- Flexible and customizable configuration via JSON files.
- Simple and intuitive API for working with MongoDB in your Go applications.

## Installation

To use **MyMongo** in your Go project, you can simply run:

```bash
go get github.com/heyitsfranky/MyMongo@latest
```

## Usage

Here's a basic example of how to use **MyMongo** to initialize a MongoDB client and perform CRUD operations:
```go
package main

import (
    "fmt"
    "github.com/heyitsfranky/MyMongo"
)

func main() {
    // Initialize the MongoDB client with a configuration file path.
    if err := myMongo.Init("path/to/your/config.yaml"); err != nil {
        fmt.Printf(err)
        return
    }

    your_db_name := "foo"
    your_col_name := "bar"
    
    // Perform CRUD operations
    // Example: Create a new document
    data := map[string]interface{}{
        "uuid":      "12345",
        "firstName": "John",
        "lastName":  "Doe",
    }
    if err := myMongo.PerformDatabaseAction(your_db_name, your_col_name, myMongo.Create, data); err != nil {
        fmt.Printf(err)
    }

    // Example: Retrieve an object by filter (multiple objects also possible)
    filter := CreateFilterQuery("uuid", "12345", "firstName", "John")
    obj, err := myMongo.GetObject[interface{}](filter, your_db_name, your_col_name)
    if err != nil {
        fmt.Printf(err)
    } else {
        fmt.Printf("Retrieved object: %+v\n", obj)
    }

    // Example: Update a document
    updateData := map[string]interface{}{
        "uuid": "12345",
        "lastName": "Smith",
    }
    if err := myMongo.PerformDatabaseAction(your_db_name, your_col_name, myMongo.Update, updateData); err != nil {
        fmt.Printf(err)
    }

    // Example: Delete a document
    if err := myMongo.PerformDatabaseAction(your_db_name, your_col_name, myMongo.Delete, data); err != nil {
        fmt.Printf(err)
    }
}
```
**Note**: ensure that each set of data includes a unique UUID identified by the "uuid" key. This UUID serves as the unique identifier for data updates, and deletions.

## Configuration

To configure your MongoDB connection, create a YAML configuration file with the following structure:

```yaml
mongo-username: your_username,
mongo-password: your_password,
mongo-host: localhost,
mongo-port: 27017
```

An up-to-date template can *always* be found under 'template_MyMongo.cfg.yaml'.

## License

This package is distributed under the MIT License.
Feel free to contribute or report issues on GitHub.

Happy coding!