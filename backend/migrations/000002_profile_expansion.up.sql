-- Profile expansion: country fields
ALTER TABLE profiles ADD COLUMN country VARCHAR(100);
ALTER TABLE profiles ADD COLUMN country_code VARCHAR(2);
