-- +goose Up
-- Смена набора иллюстраций (фото → схематичный лайн-арт): чистим картинки
-- глобальных упражнений, чтобы LoadCatalog при старте залил новые. Картинки
-- пользовательских упражнений (owner_id IS NOT NULL) не трогаем.
DELETE FROM exercise_images
WHERE exercise_id IN (SELECT id FROM exercises WHERE owner_id IS NULL);

-- +goose Down
-- Необратимо: удалённые blob-картинки не восстановить (перезальются сидом).
SELECT 1;
