-- Adding a table so multiple people can share a domain
CREATE TABLE minion_domain (minion_id INT NOT NULL REFERENCES minions (id), domain_id INT NOT NULL REFERENCES domains (id) );
INSERT INTO version (point) VALUES (1);