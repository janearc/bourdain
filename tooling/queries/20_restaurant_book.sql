CREATE OR REPLACE FUNCTION public.restaurant_book(
    restaurant_uuid uuid,
    diner_uuids uuid[],
    req_start_time timestamp without time zone,
    req_end_time timestamp without time zone
) RETURNS uuid
    LANGUAGE plpgsql
AS $$
DECLARE
    reservation_uuid uuid;
    total_seating_capacity integer;
    party_size int;
    available_table_size int;
    available_tables RECORD;
    selected_tables uuid[];
    total_selected_capacity int := 0;
BEGIN
    -- Calculate the total seating capacity of the restaurant
    SELECT
        (cast(capacity->>'two-top' as integer) * 2) +
        (cast(capacity->>'four-top' as integer) * 4) +
        (cast(capacity->>'six-top' as integer) * 6)
    INTO total_seating_capacity
    FROM public.restaurants
    WHERE id = restaurant_uuid;

    -- Calculate the party size
    party_size := array_length(diner_uuids, 1);

    -- Ensure the party size doesn't exceed the seating capacity
    IF party_size > total_seating_capacity THEN
        RAISE EXCEPTION 'Party size exceeds the seating capacity of the restaurant.';
    END IF;

    -- Find available tables for the restaurant within the requested time frame
    FOR available_tables IN
        SELECT t.id, t.table_size
        FROM public.tops t
        WHERE t.restaurant_id = restaurant_uuid
          AND t.occupied = false
          AND NOT EXISTS (
            SELECT 1
            FROM public.reservations res
            WHERE res.restaurant_id = t.restaurant_id
              AND (res.start_time, res.end_time) OVERLAPS (req_start_time, req_end_time)
        )
        LOOP
            -- Add the available table to the selected tables array
            selected_tables := array_append(selected_tables, available_tables.id);
            total_selected_capacity := total_selected_capacity + available_tables.table_size;

            -- If the selected tables' capacity is enough to seat the party, stop
            IF total_selected_capacity >= party_size THEN
                EXIT;
            END IF;
        END LOOP;

    -- If we don't have enough tables, raise an exception
    IF total_selected_capacity < party_size THEN
        RAISE EXCEPTION 'Not enough available tables to seat the party.';
    END IF;

    -- Insert the new reservation
    INSERT INTO public.reservations (restaurant_id, start_time, end_time, num_diners)
    VALUES (restaurant_uuid, req_start_time, req_end_time, party_size)
    RETURNING id INTO reservation_uuid;

    -- Insert each diner into the reservation_diners table
    INSERT INTO public.reservation_diners (reservation_id, diner_id)
    SELECT reservation_uuid, unnest(diner_uuids);

    -- Mark the selected tables as occupied
    UPDATE public.tops
    SET occupied = true, reservation_id = reservation_uuid
    WHERE id = ANY(selected_tables);

    -- Return the reservation UUID
    RETURN reservation_uuid;
END;
$$;