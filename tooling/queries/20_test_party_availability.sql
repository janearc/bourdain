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
$$ LANGUAGE plpgsql