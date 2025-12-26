CREATE TABLE numbers_diapason (
    code SMALLINT, 
    from_n INT NOT NULL CHECK (from_n <= to_n), 
    to_n INT NOT NULL CHECK (to_n >= from_n), 
    capacity INT, 
    operator TEXT, 
    region TEXT, 
    territory TEXT, 
    inn BIGINT
);