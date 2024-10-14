CREATE OR REPLACE FUNCTION get_diner_endorsements(diner_uuids UUID[])
    RETURNS TABLE (endorsement TEXT) AS $$
BEGIN
    RETURN QUERY
        SELECT DISTINCT jsonb_array_elements_text(preferences)
        FROM diners
        WHERE id = ANY(diner_uuids);
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