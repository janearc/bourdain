CREATE OR REPLACE FUNCTION get_endorsements_for_diners(
    diner_uuids uuid[]
) RETURNS jsonb AS $$
DECLARE
    endorsement_list jsonb;
BEGIN
    -- Fetch and aggregate endorsements for the diners
    SELECT jsonb_agg(DISTINCT endorsement) INTO endorsement_list
    FROM (
             SELECT jsonb_array_elements_text(preferences) AS endorsement
             FROM diners
             WHERE id = ANY(diner_uuids)
         ) AS diner_endorsements;

    RETURN endorsement_list;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION calculate_party_size(diner_uuids uuid[])
    RETURNS int AS $$
BEGIN
    RETURN array_length(diner_uuids, 1);  -- Get the number of UUIDs in the array
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION attempt_match(
    party_size int, current_endorsements jsonb, req_start_time timestamp, req_end_time timestamp
) RETURNS TABLE(restaurant_name text, matched_endorsements jsonb, message text) AS $$
BEGIN
    RETURN QUERY
        SELECT r.name::text, r.endorsements, 'Full match found'::text
        FROM restaurants r
        WHERE r.endorsements @> current_endorsements
          AND r.opening_time <= req_start_time::time
          AND r.closing_time >= req_end_time::time
          AND (cast(r.capacity->>'two-top' as integer) * 2) +
              (cast(r.capacity->>'four-top' as integer) * 4) +
              (cast(r.capacity->>'six-top' as integer) * 6) >= party_size
          AND NOT EXISTS (
            SELECT 1
            FROM reservations res
            WHERE res.restaurant_id = r.id
              AND (res.start_time, res.end_time) OVERLAPS (req_start_time, req_end_time)
        );
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION find_available_restaurants(
    party_size int, current_endorsements jsonb, req_start_time timestamp, req_end_time timestamp
) RETURNS TABLE(restaurant_name text, matched_endorsements jsonb, message text) AS $$
BEGIN
    RETURN QUERY
        SELECT r.name::text, r.endorsements, 'Full match found'::text
        FROM restaurants r
        WHERE
            (cast(r.capacity->>'two-top' as integer) * 2) +
            (cast(r.capacity->>'four-top' as integer) * 4) +
            (cast(r.capacity->>'six-top' as integer) * 6) >= party_size
          AND r.endorsements @> current_endorsements
          AND r.opening_time <= req_start_time::time
          AND r.closing_time >= req_end_time::time
          AND NOT EXISTS (
            SELECT 1
            FROM reservations res
            WHERE res.restaurant_id = r.id
              AND (res.start_time, res.end_time) OVERLAPS (req_start_time, req_end_time)
        );
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION predict_match_difficulty(diner_uuids uuid[])
    RETURNS TABLE(endorsement_count int, restaurant_count int) AS $$
BEGIN
    RETURN QUERY
        SELECT jsonb_array_length(d.preferences) AS endorsement_count,
               COUNT(DISTINCT r.id) AS restaurant_count
        FROM diners d
                 JOIN restaurants r ON r.endorsements @> d.preferences
        WHERE d.id = ANY(diner_uuids)
        GROUP BY endorsement_count
        ORDER BY endorsement_count DESC;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION check_restaurant_availability(
    diner_uuids uuid[], req_start_time timestamp, req_end_time timestamp
) RETURNS TABLE(restaurant_id uuid, restaurant_name text, matched_endorsements jsonb, message text) AS $$
DECLARE
    current_endorsements jsonb;
    party_size int;
BEGIN
    -- Step 1: Calculate the party size
    party_size := array_length(diner_uuids, 1);

    -- Step 2: Get the endorsements of the diners
    current_endorsements := get_endorsements_for_diners(diner_uuids);

    -- Step 3: Check if any restaurants match the endorsements
    IF NOT EXISTS (
        SELECT 1
        FROM restaurants r
        WHERE r.endorsements @> current_endorsements
    ) THEN
        -- Raise an exception if no restaurants match the endorsements
        RAISE EXCEPTION 'No restaurants match the given endorsements';
    END IF;

    -- Step 4: Proceed with normal availability check if matches are found
    RETURN QUERY
        SELECT r.id::uuid, r.name::text, r.endorsements, 'Match found'::text
        FROM restaurants r
        WHERE r.endorsements @> current_endorsements
          AND r.opening_time <= req_start_time::time
          AND r.closing_time >= req_end_time::time
          AND (cast(r.capacity->>'two-top' as integer) * 2) +
              (cast(r.capacity->>'four-top' as integer) * 4) +
              (cast(r.capacity->>'six-top' as integer) * 6) >= party_size
          AND NOT EXISTS (
            SELECT 1
            FROM reservations res
            WHERE res.restaurant_id = r.id
              AND (res.start_time, res.end_time) OVERLAPS (req_start_time, req_end_time)
        );
END;
$$ LANGUAGE plpgsql;