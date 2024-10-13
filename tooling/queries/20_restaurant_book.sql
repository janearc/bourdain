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
BEGIN
    -- Ensure the party size matches the number of diner UUIDs
    IF array_length(diner_uuids, 1) IS DISTINCT FROM (SELECT num_diners FROM restaurants WHERE id = restaurant_uuid) THEN
        RAISE EXCEPTION 'Party size does not match the number of diner UUIDs provided.';
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