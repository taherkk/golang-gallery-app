-- +goose Up
-- +goose StatementBegin
ALTER TABLE sessions ADD CONSTRAINT fk_users_sessions_id FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE;

-- +goose StatementEnd
-- +goose Down
ALTER TABLE sessions
DROP CONSTRAINT fk_users_sessions_id;

-- +goose StatementBegin
-- +goose StatementEnd