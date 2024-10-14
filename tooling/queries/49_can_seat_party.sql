CREATE OR REPLACE FUNCTION can_seat_party_at_time(
    restaurant_id uuid,
    party_size int,
    req_start_time timestamp,
    req_end_time timestamp
) RETURNS boolean AS $$
DECLARE
    available_seats int := 0;
    table_record RECORD;
BEGIN
    -- Step 1: Check available tables for the restaurant that are not occupied at the requested time
    FOR table_record IN
        SELECT t.id, t.table_size
        FROM tops t
        WHERE t.restaurant_id = can_seat_party_at_time.restaurant_id  -- Correct reference to the input restaurant_id
          AND NOT EXISTS (
            -- Check if the table is already reserved during the requested time
            SELECT 1 FROM reservations r
            WHERE r.restaurant_id = t.restaurant_id  -- Ensure reservations match the same restaurant
              AND r.id = t.reservation_id
              AND (r.start_time, r.end_time) OVERLAPS (req_start_time, req_end_time)
        )
        LOOP
            -- Accumulate the available seating capacity
            available_seats := available_seats + table_record.table_size;

            -- If we've found enough seating, return true
            IF available_seats >= party_size THEN
                RETURN true;
            END IF;
        END LOOP;

    -- If we didn't find enough seating, return false
    RETURN false;
END;
$$ LANGUAGE plpgsql;