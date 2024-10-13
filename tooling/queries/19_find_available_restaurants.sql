CREATE OR REPLACE FUNCTION find_available_restaurants(
    diner_endorsements jsonb
)
    RETURNS TABLE(restaurant_name text) AS $$
BEGIN
    RETURN QUERY
        SELECT r.name::text  -- Explicitly casting the name to text
        FROM restaurants r
        WHERE
            -- Check if restaurant endorsements include the diner preferences
            r.endorsements @> diner_endorsements;
END;
$$ LANGUAGE plpgsql;