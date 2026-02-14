// pkg/database/schema.go
package database

import "os"

func Table(name string) string {
	schema := os.Getenv("DB_SCHEMA")
	if schema == "" {
		schema = "myschema"
	}
	return schema + "." + name
}
