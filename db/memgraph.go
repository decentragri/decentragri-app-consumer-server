package memgraph

import (
	"context"
	"log"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"decentragri-app-cx-server/utils"
)

var driver neo4j.DriverWithContext

// InitMemGraph initializes a connection to the Memgraph database.
// It loads configuration from environment variables and returns a Neo4j driver.
func InitMemGraph() neo4j.DriverWithContext {

	uri := utils.GetEnv("MEMGRAPH_URI")
	username := utils.GetEnv("MEMGRAPH_USERNAME")
	password := utils.GetEnv("MEMGRAPH_PASSWORD")

	auth := neo4j.BasicAuth(username, password, "")
	d, err := neo4j.NewDriverWithContext(uri, auth, func(config *neo4j.Config) {
		config.Log = nil // You can set a logger here if needed
	})
	if err != nil {
		panic(err)
	}

	driver = d

	log.Println("Memgraph Initialized!")

	if err := driver.VerifyConnectivity(context.Background()); err != nil {
		log.Fatalf("Failed to connect to Memgraph: %s", err)
	}

	return driver
}

// GetDriver returns the initialized Neo4j driver.
func GetDriver() neo4j.DriverWithContext {
	return driver
}

// CloseDriver closes the Neo4j driver if it's open.
func CloseDriver() {
	if driver != nil {
		defer driver.Close(context.Background())
	}
}

// ExecuteRead is a utility to run a Cypher read query and return all records.
func ExecuteRead(query string, params map[string]interface{}) ([]*neo4j.Record, error) {
	ctx := context.Background()
	session := GetDriver().NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	recordsAny, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}
		records, err := result.Collect(ctx)
		if err != nil {
			return nil, err
		}
		return records, nil
	})
	if err != nil {
		return nil, err
	}
	return recordsAny.([]*neo4j.Record), nil
}

// ExecuteWrite is a utility to run a Cypher write query and return the summary or error.
func ExecuteWrite(query string, params map[string]interface{}) (neo4j.ResultSummary, error) {
	ctx := context.Background()
	session := GetDriver().NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	summaryAny, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}
		summary, err := result.Consume(ctx)
		if err != nil {
			return nil, err
		}
		return summary, nil
	})
	if err != nil {
		return nil, err
	}
	return summaryAny.(neo4j.ResultSummary), nil
}
