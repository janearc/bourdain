CREATE OR REPLACE FUNCTION check_restaurant_availability(
    diner_uuids uuid[], req_start_time timestamp, req_end_time timestamp
) RETURNS TABLE(restaurant_name text, matched_endorsements jsonb, message text) AS $$
DECLARE
    endorsement_list jsonb;
    current_endorsements jsonb;
    match_found boolean := false;
    party_size int;
BEGIN
    -- Step 1: Calculate the party size dynamically
    party_size := array_length(diner_uuids, 1);  -- Get the number of UUIDs in the array

    -- Step 2: Get the endorsements of the diners
    SELECT jsonb_agg(endorsement) INTO endorsement_list
    FROM get_diner_endorsements(diner_uuids);

    -- Step 3: Initialize the current set of endorsements
    current_endorsements := endorsement_list;

    -- Start a loop to check for a full match and possibly reduce endorsements
    LOOP
        -- Attempt to find a restaurant with the current set of endorsements
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

        -- If no match found, reduce endorsements
        IF NOT FOUND THEN
            -- If no more endorsements left to drop, exit the loop
            IF jsonb_array_length(current_endorsements) = 0 THEN
                RAISE NOTICE 'No full match found, trying partial matches';
                EXIT;
            END IF;

            -- Drop the first endorsement in the array
            current_endorsements := jsonb_set(
                    current_endorsements, ARRAY['0'], 'null'::jsonb, true
                                    );

            -- Remove all 'null' values from the array
            current_endorsements := jsonb_strip_nulls(current_endorsements);
        ELSE
            match_found := true;
            EXIT;
        END IF;
    END LOOP;

    -- If no match was found, return partial matches
    IF NOT match_found THEN
        RETURN QUERY
            SELECT r.name::text, r.endorsements, 'Partial match found, not all preferences met'::text
            FROM restaurants r
            WHERE
                r.endorsements @> current_endorsements
              AND r.opening_time <= req_start_time::time
              AND r.closing_time >= req_end_time::time
              AND (
                      (cast(r.capacity->>'two-top' as integer) * 2) +
                      (cast(r.capacity->>'four-top' as integer) * 4) +
                      (cast(r.capacity->>'six-top' as integer) * 6)
                      ) >= party_size
              AND NOT EXISTS (
                SELECT 1
                FROM reservations res
                WHERE res.restaurant_id = r.id
                  AND (res.start_time, res.end_time) OVERLAPS (req_start_time, req_end_time)
            );
    END IF;
END;
$$ LANGUAGE plpgsql;