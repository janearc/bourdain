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
			CREATE OR REPLACE FUNCTION find_available_restaurants(party_size int, diner_endorsements jsonb)
			RETURNS TABLE(restaurant_name text) AS $$
			BEGIN
			  RETURN QUERY
				SELECT r.name
				FROM restaurants r
				WHERE 
				  (cast(r.capacity->>'two-top' as integer) * 2) +
				  (cast(r.capacity->>'four-top' as integer) * 4) +
				  (cast(r.capacity->>'six-top' as integer) * 6) >= party_size
				AND r.endorsements @> diner_endorsements;
			END;
			$$ LANGUAGE plpgsql;`)

	if arErr != nil {
		logrus.Fatalf("Error creating available_restaurants function: %v", endErr)
	}

	// create the distinct endorsements function
	_, bpErr := db.Exec(`
		CREATE OR REPLACE FUNCTION generate_party(party_size INT)
		RETURNS TABLE(diner_id UUID) AS $$
		BEGIN
			RETURN QUERY
			SELECT id
			FROM diners
			ORDER BY random()
			LIMIT party_size;
		END;
		$$ LANGUAGE plpgsql;`)

	if bpErr != nil {
		logrus.Fatalf("Error creating generate_party function: %v", endErr)
	}

	logrus.Info("Function generate_party created successfully")

	// this is a test function to verify from psql that the logic works in
	// the db, regardless of what golang thinks of that

	_, tpErr := db.Exec(`
		CREATE OR REPLACE FUNCTION test_party_availability(party_size int)
		RETURNS TABLE(restaurant_name text) AS $$
		BEGIN
		  RETURN QUERY
			SELECT * 
			FROM find_available_restaurants(
			  party_size,
			  (SELECT jsonb_agg(endorsement) FROM get_diner_endorsements(
				  (SELECT ARRAY(SELECT * FROM generate_party(party_size)))
			  ))
			);
		END;
		$$ LANGUAGE plpgsql;`)

	if tpErr != nil {
		logrus.Fatalf("Error creating test_party_availability function: %v", tpErr)
	}

	logrus.Info("Function test_party_availability created successfully")

	return
}
