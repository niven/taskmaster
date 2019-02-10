-- Adding a table so multiple people can share a domain
CREATE TABLE minion_domain (minion_id INT, domain_id INT, CONSTRAINT minion_domain_minion_id_ref_minion_id_fkey FOREIGN KEY (minion_id) REFERENCES minions(id), CONSTRAINT minion_domain_domain_id_ref_domains_id_fkey FOREIGN KEY (domain_id) REFERENCES domains(id));
INSERT INTO version (point) VALUES (1);