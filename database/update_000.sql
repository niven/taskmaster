-- initial table to keep track of where we are
CREATE TABLE version (point INT, last_update TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP);
INSERT INTO version (point) VALUES (0);
-- initial tables
CREATE TABLE minions (id SERIAL PRIMARY KEY, email VARCHAR(255) NOT NULL UNIQUE, name VARCHAR(255) NOT NULL );
CREATE TABLE domains (id SERIAL PRIMARY KEY, owner INTEGER, name VARCHAR(255) NOT NULL, last_reset_date DATE NOT NULL DEFAULT CURRENT_DATE, CONSTRAINT domains_owner_ref_minions_id_fkey FOREIGN KEY (owner) REFERENCES minions(id) );
CREATE TABLE tasks (id SERIAL PRIMARY KEY, domain_id INTEGER, name VARCHAR(255) NOT NULL, weekly BOOLEAN DEFAULT false, description TEXT, count INTEGER NOT NULL DEFAULT 1, CONSTRAINT tasks_domain_id_ref_domains_id_fkey FOREIGN KEY (domain_id) REFERENCES domains(id));
CREATE TYPE enum_status AS ENUM ('pending', 'done_and_stashed', 'done_and_available');
CREATE TABLE task_assignments (id SERIAL PRIMARY KEY, task_id INTEGER, minion_id INTEGER, assigned_on DATE NOT NULL, status enum_status default 'pending', CONSTRAINT task_assignment_task_id_ref_tasks_id_fkey FOREIGN KEY (task_id) REFERENCES tasks(id), CONSTRAINT task_assignment_minion_id_ref_minions_id_fkey FOREIGN KEY (minion_id) REFERENCES minions(id));
-- System data
INSERT INTO minions (id, email, name) VALUES(0, 'unused', 'System');
INSERT INTO domains (id, owner, name) VALUES(0 , 0, 'System');