# golang-graphql-postgresql
An implementation of GraphQL with PostgreSQL in GoLang.

To start the server run
go run main.go

To test the GraphQL endpoints - 
URL http://localhost:8082/graphql
Open postman and and change request method to POST and Choose GraphQL

1. To Test CreateChildProfile mutation

mutation {
    createChildProfile(
            givenName:"Tim"
            familyName:"William"
            middleName: "J"
            birthDate:"2008-11-20")
            {
                 id
                givenName
                middleName
                 familyName
                birthDate
                }
}

2. To Test GetChildProfile query

query{
  getChildProfile(id: "9bde890a-6a53-4d05-9e0a-44f82006699a") {
    id
    givenName
    middleName
    familyName
    birthDate
  }
}


SQL
CREATE SCHEMA child;


CREATE TABLE child.person(
  id uuid NOT NULL,
  given_name text NOT NULL,
  middle_name text,
  family_name text NOT NULL,
  birthdate text NOT NULL
)
