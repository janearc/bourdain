package main

import (
	"database/sql"
	"github.com/sirupsen/logrus"
)

func createPlSqlFunctions(db *sql.DB) {
	remoteDB, rderr := getCurrentDatabase(db)

	if rderr != nil {
		logrus.Fatalf("Error getting current database: %v", rderr)
	} else {
		logrus.Infof("[createtables] Current database: %s", remoteDB)
	}

	// create the distinct endorsements function
	_, endErr := db.Exec(`
		CREATE OR REPLACE FUNCTION get_diner_endorsements(diner_uuids UUID[])
		RETURNS TABLE (endorsement TEXT) AS $$
		BEGIN
			RETURN QUERY
			SELECT DISTINCT jsonb_array_elements_text(preferences) 
			FROM diners
			WHERE id = ANY(diner_uuids);
		END;
		$$ LANGUAGE plpgsql;`)

	if endErr != nil {
		logrus.Fatalf("Error creating get_diner_endorsements function: %v", endErr)
	}

	logrus.Info("Function get_diner_endorsements created successfully")

	// create the distinct endorsements function
	_, arErr := db.Exec(`
		CREATE OR REPLACE FUNCTION find_available_restaurants(party_size int, endorsements jsonb)
		RETURNS TABLE(name text) AS $$
		BEGIN
		  RETURN QUERY
			SELECT name
			FROM restaurants
			WHERE 
			  (cast(capacity->>'two-top' as integer) * 2) +
			  (cast(capacity->>'four-top' as integer) * 4) +
			  (cast(capacity->>'six-top' as integer) * 6) >= party_size
			AND endorsements @> endorsements;
		END;
		$$ LANGUAGE plpgsql;`)

	if arErr != nil {
		logrus.Fatalf("Error creating available_restaurants function: %v", endErr)
	}

	logrus.Info("Function available_restaurants created successfully")

	return
}
