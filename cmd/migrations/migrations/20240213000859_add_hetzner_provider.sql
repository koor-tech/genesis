-- +goose Up
-- +goose StatementBegin
insert into "providers" ("id", "name")
values ('80be226b-8355-4dea-b41a-6e17ea37559a', 'hetzner');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM
    "providers"
WHERE id = '80be226b-8355-4dea-b41a-6e17ea37559a'
-- +goose StatementEnd
