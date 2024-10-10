CREATE OR REPLACE FUNCTION get_diner_endorsements(diner_uuids UUID[])
    RETURNS TABLE (endorsement TEXT) AS $$
BEGIN
    RETURN QUERY
        SELECT DISTINCT jsonb_array_elements_text(preferences)
        FROM diners
        WHERE id = ANY(diner_uuids);
END;
$$ LANGUAGE plpgsql;