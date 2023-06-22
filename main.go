package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/jackc/pgx/v4"
)

type ChildProfile struct {
	ID         string  `json:"id"`
	GivenName  string  `json:"givenName"`
	MiddleName *string `json:"middleName:omitempty"`
	FamilyName string  `json:"familyName"`
	BirthDate  string  `json:"birthDate"`
}

var db *pgx.Conn

// const (
// 	dbHost     = "database-1.cluster-c4vnvehic8i0.us-east-1.rds.amazonaws.com"
// 	dbPort     = "5432"
// 	dbName     = "postgres"
// 	dbUser     = "postgres"
// 	dbPassword = "password"
// )

func initDB() error {
	config, err := pgx.ParseConfig("postgres://postgres:postgrespassword@localhost:5432/postgres")

	// config, err := pgx.ParseConfig(fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPassword, dbHost, dbPort, dbName))

	if err != nil {
		return err
	}
	conn, err := pgx.ConnectConfig(context.Background(), config)
	if err != nil {
		return err
	}
	fmt.Println("Connect Succefully...")
	db = conn
	return nil
}

func getChildProfile(params graphql.ResolveParams) (interface{}, error) {
	id, ok := params.Args["id"].(string)
	if !ok {
		return nil, errors.New("invalid ID")
	}

	var childProfile ChildProfile
	var middleName *string // Use a pointer to string to handle NULL value

	err := db.QueryRow(context.Background(), "SELECT id, given_name, middle_name, family_name, birthdate FROM child.person WHERE id = $1", id).Scan(&childProfile.ID, &childProfile.GivenName, &middleName, &childProfile.FamilyName, &childProfile.BirthDate)

	if err != nil {
		return nil, err
	}

	childProfile.MiddleName = middleName // Assign the value to the struct field

	return childProfile, nil
}

func createChildProfile(params graphql.ResolveParams) (interface{}, error) {
	givenName, _ := params.Args["givenName"].(string)
	middleName, _ := params.Args["middleName"].(string)
	familyName, _ := params.Args["familyName"].(string)
	birthDate, _ := params.Args["birthDate"].(string)

	// Generate a new ID for the user
	id := uuid.New().String()

	_, err := db.Exec(context.Background(), "INSERT INTO child.person (id, given_name, middle_name, family_name, birthdate) VALUES ($1, $2, $3, $4, $5)",
		id, givenName, middleName, familyName, birthDate)
	if err != nil {
		return nil, err
	}

	childProfile := ChildProfile{
		ID:         id,
		GivenName:  givenName,
		MiddleName: &middleName,
		FamilyName: familyName,
		BirthDate:  birthDate,
	}

	return childProfile, nil
}

func main() {
	err := initDB()
	if err != nil {
		log.Fatal("Failed to connect to the database:", err)
	}

	// Define the childProfile object type
	childprofileObject := graphql.NewObject(graphql.ObjectConfig{
		Name: "ChildProfile",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"givenName": &graphql.Field{
				Type: graphql.String,
			},
			"middleName": &graphql.Field{
				Type: graphql.String,
			},
			"familyName": &graphql.Field{
				Type: graphql.String,
			},
			"birthDate": &graphql.Field{
				Type: graphql.String,
			},
		},
	})

	// Define the GraphQL fields and types
	fields := graphql.Fields{
		"getChildProfile": &graphql.Field{
			Type:        childprofileObject,
			Description: "Get a user by ID",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: getChildProfile,
		},
		"createChildProfile": &graphql.Field{
			Type:        childprofileObject,
			Description: "Create a new user",
			Args: graphql.FieldConfigArgument{
				"givenName": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"middleName": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"familyName": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
				"birthDate": &graphql.ArgumentConfig{
					Type: graphql.NewNonNull(graphql.String),
				},
			},
			Resolve: createChildProfile,
		},
	}

	// Define the root query object
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}

	// Define the root mutation object
	rootMutation := graphql.ObjectConfig{
		Name:   "RootMutation",
		Fields: fields,
		// Fields: graphql.Fields{
		// 	"createChildProfile": &graphql.Field{
		// 		Type:        childprofileObject,
		// 		Description: "Create a new child profile",
		// 		Args: graphql.FieldConfigArgument{
		// 			"givenName": &graphql.ArgumentConfig{
		// 				Type: graphql.NewNonNull(graphql.String),
		// 			},
		// 			"middleName": &graphql.ArgumentConfig{
		// 				Type: graphql.String,
		// 			},
		// 			"familyName": &graphql.ArgumentConfig{
		// 				Type: graphql.NewNonNull(graphql.String),
		// 			},
		// 			"birthDate": &graphql.ArgumentConfig{
		// 				Type: graphql.NewNonNull(graphql.String),
		// 			},
		// 		},
		// 		Resolve: createChildProfile,
		// 	},
		// },
	}

	// Define the schema
	schemaConfig := graphql.SchemaConfig{
		Query:    graphql.NewObject(rootQuery),
		Mutation: graphql.NewObject(rootMutation),
	}

	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatal("Failed to create GraphQL schema:", err)
	}

	// Create a new HTTP handler for GraphQL requests
	graphqlHandler := handler.New(&handler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	http.Handle("/graphql", graphqlHandler)

	log.Println("Server started on http://localhost:8082/graphql")
	log.Fatal(http.ListenAndServe(":8082", nil))

}
