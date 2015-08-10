# divan
Mini CouchDB implementation backed by PostgreSQL.

## Get Started

Only run with the -b flag the first time if you don't have divan table setup yet. Divan table is used for configuration. Start with this command:

```
DATABASE_URL=postgres://username:password@dbhost.com:5432/database go run main.go -b
```

In another terminal, run this commands to test (replace any ids and revisions with the uuids that get generated for you):

```
# Create a database
curl -XPUT http://localhost:8080/test

# Post a document
curl -XPOST --data '{}' http://localhost:8080/test
{"id":"45c21dac-d6db-45c5-b82e-7d2f794d9568","ok":true,"rev":"3bc8d889-57db-4180-b879-57f500e16e86"}

# Put the document
curl -XPUT --data '{"_rev": "3bc8d889-57db-4180-b879-57f500e16e86", "test": "test"}' http://localhost:8080/test/45c21dac-d6db-45c5-b82e-7d2f794d9568
{"id":"45c21dac-d6db-45c5-b82e-7d2f794d9568","ok":true,"rev":"a574228d-9b9d-4b1b-9ad6-3cdf49389220"}

# Get the document
curl http://localhost:8080/test/45c21dac-d6db-45c5-b82e-7d2f794d9568
{"_id":"45c21dac-d6db-45c5-b82e-7d2f794d9568","_rev":"a574228d-9b9d-4b1b-9ad6-3cdf49389220","test":"test"}
```

## HTTP Document API

On bootup, the output should give you a route map of the implemented HTTP methods.