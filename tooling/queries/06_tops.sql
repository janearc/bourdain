CREATE TABLE public.tops (
                             id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
                             restaurant_id uuid NOT NULL,
                             table_size integer NOT NULL,
                             occupied boolean DEFAULT false NOT NULL,
                             reservation_id uuid,
                             PRIMARY KEY (id),
                             FOREIGN KEY (restaurant_id) REFERENCES public.restaurants(id) ON DELETE CASCADE,
                             FOREIGN KEY (reservation_id) REFERENCES public.reservations(id) ON DELETE SET NULL
);

CREATE OR REPLACE FUNCTION get_available_tops(
    restaurant_uuid uuid, req_start_time timestamp, req_end_time timestamp
) RETURNS TABLE(table_id uuid, table_size int) AS $$
BEGIN
    RETURN QUERY
        SELECT t.id, t.table_size
        FROM tops t
        WHERE t.restaurant_id = restaurant_uuid  -- Fully qualifying `restaurant_id`
          AND NOT EXISTS (
            SELECT 1
            FROM reservations res
            WHERE res.id = t.reservation_id  -- Check if the table is reserved
              AND (res.start_time, res.end_time) OVERLAPS (req_start_time, req_end_time)  -- Time overlap check
        );
END;
$$ LANGUAGE plpgsql;