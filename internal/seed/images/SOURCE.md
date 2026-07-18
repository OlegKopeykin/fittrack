# Иллюстрации упражнений

Источник: **free-exercise-db** — https://github.com/yuhonas/free-exercise-db

Лицензия: **Public Domain** (The Unlicense). Атрибуция не требуется.

Файлы названы по id упражнения в исходной базе (`<Id>.jpg`), сопоставление
с нашим каталогом — в `catalog.json` (поле `image`). Встраиваются в бинарь
через `//go:embed` и ставятся глобальным упражнениям при старте (`LoadCatalog`),
если картинки ещё нет.
