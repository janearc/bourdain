CREATE TABLE public.reservations (
                                     id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
                                     restaurant_id uuid NOT NULL,
                                     start_time timestamp without time zone NOT NULL,
                                     end_time timestamp without time zone NOT NULL,
                                     num_diners integer NOT NULL,
                                     PRIMARY KEY (id),
                                     FOREIGN KEY (restaurant_id) REFERENCES public.restaurants(id) ON DELETE CASCADE
);