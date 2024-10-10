
CREATE OR REPLACE FUNCTION generate_party(party_size INT)
    RETURNS TABLE(diner_id UUID) AS $$
BEGIN
    RETURN QUERY
        SELECT id
        FROM diners
        ORDER BY random()
        LIMIT party_size;
END;
$$ LANGUAGE plpgsql;