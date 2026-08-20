DROP TABLE IF EXISTS inpatient_bookings, inpatient_beds, inpatient_rooms CASCADE;
DELETE FROM permissions WHERE code IN ('inpatient:read','inpatient:write');
