CREATE OR REPLACE FUNCTION find_available_restaurants(party_size int, diner_endorsements jsonb)
    RETURNS TABLE(restaurant_name text) AS $$
BEGIN
    RETURN QUERY
        SELECT r.name::text  -- Explicitly casting the name to text
        FROM restaurants r
        WHERE
            (cast(r.capacity->>'two-top' as integer) * 2) +
            (cast(r.capacity->>'four-top' as integer) * 4) +
            (cast(r.capacity->>'six-top' as integer) * 6) >= party_size
          AND r.endorsements @> diner_endorsements;
END;
$$ LANGUAGE plpgsql