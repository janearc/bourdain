CREATE FUNCTION public.restaurant_book(
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
BEGIN
    -- Calculate the total seating capacity of the restaurant
    SELECT
        (cast(capacity->>'two-top' as integer) * 2) +
        (cast(capacity->>'four-top' as integer) * 4) +
        (cast(capacity->>'six-top' as integer) * 6)
    INTO total_seating_capacity
    FROM public.restaurants
    WHERE id = restaurant_uuid;

    -- Ensure the party size matches the available seating capacity
    IF array_length(diner_uuids, 1) > total_seating_capacity THEN
        RAISE EXCEPTION 'Party size exceeds the seating capacity of the restaurant.';
    END IF;

    -- Insert the new reservation
    INSERT INTO public.reservations (restaurant_id, start_time, end_time, num_diners)
    VALUES (restaurant_uuid, req_start_time, req_end_time, array_length(diner_uuids, 1))
    RETURNING id INTO reservation_uuid;

    -- Insert each diner into the reservation_diners table
    INSERT INTO public.reservation_diners (reservation_id, diner_id)
    SELECT reservation_uuid, unnest(diner_uuids);

    -- Return the reservation UUID
    RETURN reservation_uuid;
END;
$$;