-- Index on endorsements
CREATE INDEX idx_restaurants_endorsements ON restaurants USING gin(endorsements);

-- Index on capacity (jsonb field for two-top, four-top, six-top)
CREATE INDEX idx_restaurants_capacity ON restaurants((cast(capacity->>'two-top' as integer)),
                                                     (cast(capacity->>'four-top' as integer)),
                                                     (cast(capacity->>'six-top' as integer)));

-- Index on opening and closing times
CREATE INDEX idx_restaurants_times ON restaurants(opening_time, closing_time);

-- Create a materialized view to aggregate restaurant endorsements
CREATE MATERIALIZED VIEW restaurant_endorsements AS
SELECT r.id AS restaurant_id,
       jsonb_agg(DISTINCT endorsement) AS combined_endorsements
FROM restaurants r
         JOIN reservations res ON r.id = res.restaurant_id
         JOIN reservation_diners rd ON res.id = rd.reservation_id
         JOIN diners d ON d.id = rd.diner_id,
     LATERAL (SELECT jsonb_array_elements_text(d.preferences) AS endorsement) AS diner_endorsements
GROUP BY r.id;

-- Create an index to speed up querying the materialized view
CREATE INDEX idx_restaurant_endorsements ON restaurant_endorsements USING gin(combined_endorsements);