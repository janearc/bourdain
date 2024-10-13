CREATE FUNCTION public.test_restaurant_book() RETURNS void
    LANGUAGE plpgsql
AS $$
DECLARE
    test_reservation_uuid uuid;
    available_restaurant_uuid uuid;
BEGIN
    -- Begin an exception-handling block to simulate a transaction without committing
    BEGIN
        -- Fetch an available restaurant UUID by joining on the restaurant name
        SELECT r.id INTO available_restaurant_uuid
        FROM public.restaurants r
                 JOIN (
            SELECT restaurant_name
            FROM public.check_restaurant_availability(
                    2, -- Party size
                    (SELECT ARRAY(SELECT diner_id FROM public.generate_party(2))),
                    '2024-10-14 18:00:00',
                    '2024-10-14 20:00:00'
                 )
        ) AS available_restaurant
                      ON r.name = available_restaurant.restaurant_name
        LIMIT 1;

        -- Test the restaurant_book function with the fetched restaurant UUID
        test_reservation_uuid := public.restaurant_book(
                available_restaurant_uuid, -- Use the available restaurant UUID
                (SELECT ARRAY(SELECT diner_id FROM public.generate_party(2))), -- Generate a party of 2 diners
                '2024-10-14 18:00:00',    -- Start time
                '2024-10-14 20:00:00'     -- End time
                                 );

        -- Optionally: Select the results to verify them
        RAISE NOTICE 'Test Reservation UUID: %', test_reservation_uuid;
        PERFORM * FROM reservations WHERE id = test_reservation_uuid;
        PERFORM * FROM reservation_diners WHERE reservation_id = test_reservation_uuid;

        -- Explicitly raise an exception to simulate a rollback
        RAISE EXCEPTION 'Test completed - simulating rollback';
    EXCEPTION WHEN OTHERS THEN
        -- Handle the rollback by catching the exception and ensuring no data is committed
        RAISE NOTICE 'Rolling back test transaction...';
    -- No ROLLBACK needed, as exceptions will automatically abort the transaction
    END;
END;
$$;