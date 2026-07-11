BEGIN;

ALTER TABLE data_quality_reports
    ADD COLUMN reconciliation_task_id uuid;

ALTER TABLE data_quality_reports
    ADD CONSTRAINT data_quality_reports_reconciliation_task_fk
        FOREIGN KEY (reconciliation_task_id)
        REFERENCES derived_reconciliation_tasks(id)
        ON DELETE RESTRICT;

CREATE UNIQUE INDEX data_quality_reports_reconciliation_task_unique
    ON data_quality_reports (reconciliation_task_id)
    WHERE reconciliation_task_id IS NOT NULL;

ALTER TABLE flight_trajectories
    ADD COLUMN reconciliation_task_id uuid;

ALTER TABLE flight_trajectories
    ADD CONSTRAINT flight_trajectories_reconciliation_task_fk
        FOREIGN KEY (reconciliation_task_id)
        REFERENCES derived_reconciliation_tasks(id)
        ON DELETE RESTRICT;

CREATE UNIQUE INDEX flight_trajectories_reconciliation_task_unique
    ON flight_trajectories (reconciliation_task_id)
    WHERE reconciliation_task_id IS NOT NULL;

COMMIT;
