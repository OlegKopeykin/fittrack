// Package seed загружает глобальный каталог упражнений (из vault) в БД.
package seed

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/OlegKopeykin/fittrack/internal/db/gen"
	"github.com/OlegKopeykin/fittrack/internal/exercises"
)

//go:embed catalog.json
var catalogJSON []byte

type catalog struct {
	MuscleGroups []struct {
		Slug      string `json:"slug"`
		NameRu    string `json:"name_ru"`
		WeeklyMEV int64  `json:"weekly_mev"`
		WeeklyMAV int64  `json:"weekly_mav"`
	} `json:"muscle_groups"`
	Exercises []struct {
		Name        string   `json:"name"`
		Aliases     []string `json:"aliases"`
		MuscleGroup string   `json:"muscle_group"`
		Kind        string   `json:"kind"`
		PerArm      bool     `json:"per_arm"`
		Equipment   string   `json:"equipment"`
	} `json:"exercises"`
}

// LoadCatalog идемпотентно заливает глобальные группы мышц и упражнения:
// создаёт только отсутствующее (по slug / по имени глобального упражнения).
func LoadCatalog(ctx context.Context, conn *sql.DB) error {
	var cat catalog
	if err := json.Unmarshal(catalogJSON, &cat); err != nil {
		return fmt.Errorf("seed: разбор catalog.json: %w", err)
	}
	q := gen.New(conn)
	now := time.Now().UTC().Format(time.RFC3339)

	groupID := map[string]int64{}
	for i, g := range cat.MuscleGroups {
		existing, err := q.GetMuscleGroupBySlug(ctx, g.Slug)
		if err == nil {
			groupID[g.Slug] = existing.ID
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		created, err := q.InsertMuscleGroup(ctx, gen.InsertMuscleGroupParams{
			Slug:      g.Slug,
			NameRu:    g.NameRu,
			WeeklyMev: sql.NullInt64{Int64: g.WeeklyMEV, Valid: true},
			WeeklyMav: sql.NullInt64{Int64: g.WeeklyMAV, Valid: true},
			SortOrder: int64(i),
		})
		if err != nil {
			return fmt.Errorf("seed: группа %q: %w", g.Slug, err)
		}
		groupID[g.Slug] = created.ID
	}

	for _, ex := range cat.Exercises {
		if existing, err := q.GetGlobalExerciseByName(ctx, ex.Name); err == nil {
			// Уже есть — синхронизируем оборудование (бэкфилл на существующих).
			if existing.Equipment != ex.Equipment {
				if err := q.SetGlobalExerciseEquipment(ctx, gen.SetGlobalExerciseEquipmentParams{
					Equipment: ex.Equipment, Name: ex.Name,
				}); err != nil {
					return fmt.Errorf("seed: equipment %q: %w", ex.Name, err)
				}
			}
			continue
		} else if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		gid, ok := groupID[ex.MuscleGroup]
		if !ok {
			return fmt.Errorf("seed: упражнение %q ссылается на неизвестную группу %q", ex.Name, ex.MuscleGroup)
		}
		created, err := q.CreateExercise(ctx, gen.CreateExerciseParams{
			OwnerID:        sql.NullInt64{}, // глобальное
			Name:           ex.Name,
			MuscleGroupID:  gid,
			Kind:           ex.Kind,
			PerArm:         boolToInt(ex.PerArm),
			Equipment:      ex.Equipment,
			TechniqueNotes: "",
			CreatedAt:      now,
		})
		if err != nil {
			return fmt.Errorf("seed: упражнение %q: %w", ex.Name, err)
		}
		for _, alias := range ex.Aliases {
			if _, err := q.AddAlias(ctx, gen.AddAliasParams{
				ExerciseID: created.ID,
				Alias:      alias,
				AliasNorm:  exercises.Normalize(alias),
			}); err != nil {
				return fmt.Errorf("seed: алиас %q упражнения %q: %w", alias, ex.Name, err)
			}
		}
	}
	return nil
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
