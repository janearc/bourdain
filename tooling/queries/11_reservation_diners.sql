CREATE TABLE public.reservation_diners (
                                           reservation_id uuid NOT NULL,
                                           diner_id uuid NOT NULL,
                                           PRIMARY KEY (reservation_id, diner_id),
                                           FOREIGN KEY (reservation_id) REFERENCES public.reservations(id) ON DELETE CASCADE,
                                           FOREIGN KEY (diner_id) REFERENCES public.diners(id) ON DELETE CASCADE
);