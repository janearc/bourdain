create TABLE public.tops (
                             id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
                             restaurant_id uuid NOT NULL,
                             table_size integer NOT NULL,
                             occupied boolean DEFAULT false NOT NULL,
                             reservation_id uuid,
                             PRIMARY KEY (id),
                             FOREIGN KEY (restaurant_id) REFERENCES public.restaurants(id) ON DELETE CASCADE,
                             FOREIGN KEY (reservation_id) REFERENCES public.reservations(id) ON DELETE SET NULL
);

    DO $$
    DECLARE
        restaurant RECORD;
        table_id UUID;
        num_two_top int;
        num_four_top int;
        num_six_top int;
        i int;
    BEGIN
        -- Loop through each restaurant
        FOR restaurant IN
            SELECT id, capacity
            FROM restaurants
            LOOP
                -- Get the number of two-tops, four-tops, and six-tops for the current restaurant
                num_two_top := (restaurant.capacity->>'two-top')::int;
                num_four_top := (restaurant.capacity->>'four-top')::int;
                num_six_top := (restaurant.capacity->>'six-top')::int;

                -- Insert two-top tables
                IF num_two_top > 0 THEN
                    FOR i IN 1..num_two_top LOOP
                            INSERT INTO tops (restaurant_id, table_size, occupied)
                            VALUES (restaurant.id, 2, false);
                        END LOOP;
                END IF;

                -- Insert four-top tables
                IF num_four_top > 0 THEN
                    FOR i IN 1..num_four_top LOOP
                            INSERT INTO tops (restaurant_id, table_size, occupied)
                            VALUES (restaurant.id, 4, false);
                        END LOOP;
                END IF;

                -- Insert six-top tables
                IF num_six_top > 0 THEN
                    FOR i IN 1..num_six_top LOOP
                            INSERT INTO tops (restaurant_id, table_size, occupied)
                            VALUES (restaurant.id, 6, false);
                        END LOOP;
                END IF;
            END LOOP;
    END $$;

CREATE OR REPLACE FUNCTION get_available_tops(
    restaurant_uuid uuid, req_start_time timestamp, req_end_time timestamp
) RETURNS TABLE(table_id uuid, table_size int) AS $$
BEGIN
    RETURN QUERY
        SELECT t.id, t.table_size
        FROM tops t
        WHERE t.restaurant_id = restaurant_uuid
          AND NOT EXISTS (
            SELECT 1
            FROM reservations res
            WHERE res.id = t.reservation_id  -- Check if the table is reserved
              AND (res.start_time, res.end_time) OVERLAPS (req_start_time, req_end_time)  -- Time overlap check
        );
END;
$$ LANGUAGE plpgsql;