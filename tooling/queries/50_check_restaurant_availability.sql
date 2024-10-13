CREATE OR REPLACE FUNCTION check_restaurant_availability(
    diner_uuids uuid[],
    req_start_time timestamp,
    req_end_time timestamp
)
    RETURNS TABLE(restaurant_name text) AS $$
BEGIN
    RETURN QUERY
        SELECT r.name::text
        FROM restaurants r
        WHERE
          -- Calculate party size from the number of UUIDs in the array
            (cast(r.capacity->>'two-top' as integer) * 2) +
            (cast(r.capacity->>'four-top' as integer) * 4) +
            (cast(r.capacity->>'six-top' as integer) * 6) >= array_length(diner_uuids, 1)
          AND
          -- Check if restaurant endorsements include the diner preferences
            r.endorsements @> (
                SELECT jsonb_agg(endorsement)
                FROM get_diner_endorsements(diner_uuids)
            )
          AND
          -- Check if the restaurant is open during the requested time
            r.opening_time <= req_start_time::time
          AND
            r.closing_time >= req_end_time::time
          AND
          -- Ensure there are no conflicting reservations
            NOT EXISTS (
                SELECT 1
                FROM reservations res
                WHERE res.restaurant_id = r.id
                  AND (
                    (res.start_time, res.end_time) OVERLAPS (req_start_time, req_end_time)
                    )
            );
END;
$$ LANGUAGE plpgsql;