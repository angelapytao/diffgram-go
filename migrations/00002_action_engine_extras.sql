-- +goose Up
ALTER TABLE `action_run`
    ADD COLUMN `runner_name` VARCHAR(100) NULL,
    ADD COLUMN `config_data` JSON NULL,
    ADD COLUMN `event_payload` JSON NULL,
    ADD COLUMN `error_message` TEXT NULL,
    ADD COLUMN `action_template_id` INT NULL;

ALTER TABLE `action_template`
    ADD COLUMN `runner_name` VARCHAR(100) NULL,
    ADD COLUMN `event_type` VARCHAR(100) NULL,
    ADD COLUMN `config_data` JSON NULL,
    ADD COLUMN `project_id` INT NULL;

CREATE INDEX `idx_action_run_status` ON `action_run` (`status`);
CREATE INDEX `idx_action_template_event_type` ON `action_template` (`event_type`);

-- +goose Down
DROP INDEX `idx_action_template_event_type` ON `action_template`;
DROP INDEX `idx_action_run_status` ON `action_run`;

ALTER TABLE `action_template`
    DROP COLUMN `project_id`,
    DROP COLUMN `config_data`,
    DROP COLUMN `event_type`,
    DROP COLUMN `runner_name`;

ALTER TABLE `action_run`
    DROP COLUMN `action_template_id`,
    DROP COLUMN `error_message`,
    DROP COLUMN `event_payload`,
    DROP COLUMN `config_data`,
    DROP COLUMN `runner_name`;
