-- initial table to keep track of where we are
CREATE TABLE version (point INT, last_update TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
INSERT INTO version (point) VALUES (0);