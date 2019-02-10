-- add delete cascade to tables so deleting a domain also deletes all related data
-- table: tasks
ALTER TABLE tasks DROP CONSTRAINT "tasks_domain_id_fkey";
ALTER TABLE tasks ADD CONSTRAINT "tasks_domain_id_fkey_del_cascade" FOREIGN KEY (domain_id) REFERENCES domains ON DELETE CASCADE;
-- table: minion_domain
ALTER TABLE minion_domain DROP CONSTRAINT "minion_domain_domain_id_fkey";
ALTER TABLE minion_domain ADD CONSTRAINT "minion_domain_domain_id_fkey_del_cascade" FOREIGN KEY (domain_id) REFERENCES domains ON DELETE CASCADE;
-- If tasks are dropped, it should also drop assignments
ALTER TABLE task_assignments DROP CONSTRAINT "task_assignments_task_id_fkey";
ALTER TABLE task_assignments ADD CONSTRAINT "task_assignments_task_id_fkey_del_cascade" FOREIGN KEY (task_id) REFERENCES tasks ON DELETE CASCADE;
INSERT INTO version (point) VALUES (2);