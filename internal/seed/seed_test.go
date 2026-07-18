package seed_test

import (
	"testing"

	"github.com/OlegKopeykin/fittrack/internal/db/gen"
	"github.com/OlegKopeykin/fittrack/internal/seed"
	"github.com/OlegKopeykin/fittrack/internal/testutil"
)

func TestLoadCatalog(t *testing.T) {
	conn := testutil.NewTestDB(t)
	q := gen.New(conn)
	ctx := t.Context()

	if err := seed.LoadCatalog(ctx, conn); err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}

	groups, err := q.ListMuscleGroups(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(groups) != 16 {
		t.Errorf("групп мышц %d, want 16", len(groups))
	}

	// Глобальные упражнения (owner_id IS NULL) видит любой пользователь.
	list, err := q.ListExercisesForUser(ctx, testutil.NullInt(0))
	if err != nil {
		t.Fatal(err)
	}
	if len(list) < 50 {
		t.Errorf("упражнений %d, want >= 50", len(list))
	}

	// Ключевые упражнения на месте.
	if _, err := q.GetGlobalExerciseByName(ctx, "Присед в Смите"); err != nil {
		t.Errorf("нет упражнения «Присед в Смите»: %v", err)
	}
}

func TestLoadCatalogIsIdempotent(t *testing.T) {
	conn := testutil.NewTestDB(t)
	q := gen.New(conn)
	ctx := t.Context()

	for i := 0; i < 2; i++ {
		if err := seed.LoadCatalog(ctx, conn); err != nil {
			t.Fatalf("LoadCatalog #%d: %v", i+1, err)
		}
	}

	groups, _ := q.ListMuscleGroups(ctx)
	if len(groups) != 16 {
		t.Errorf("после повторного сида групп %d, want 16 (дублей быть не должно)", len(groups))
	}
	list, _ := q.ListExercisesForUser(ctx, testutil.NullInt(0))
	if len(list) != 53 {
		t.Errorf("после повторного сида упражнений %d, want 53", len(list))
	}
	byName := map[string]string{}
	for _, e := range list {
		byName[e.Name] = e.Equipment
	}
	if byName["Жим ногами"] != "machine" {
		t.Errorf("оборудование «Жим ногами» = %q, want machine", byName["Жим ногами"])
	}
	if byName["Молотки с гантелями"] != "dumbbell" {
		t.Errorf("оборудование «Молотки с гантелями» = %q, want dumbbell", byName["Молотки с гантелями"])
	}
}

func TestCatalogAliasesSearchable(t *testing.T) {
	conn := testutil.NewTestDB(t)
	q := gen.New(conn)
	ctx := t.Context()
	if err := seed.LoadCatalog(ctx, conn); err != nil {
		t.Fatal(err)
	}

	// «РДЛ» — алиас «Румынская тяга с гантелями».
	ids, err := q.SearchAliasExerciseIDs(ctx, "%рдл%")
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 {
		t.Fatalf("поиск по алиасу «рдл» вернул %d, want 1", len(ids))
	}
	ex, _ := q.GetExercise(ctx, ids[0])
	if ex.Name != "Румынская тяга с гантелями" {
		t.Errorf("алиас «рдл» → %q, want «Румынская тяга с гантелями»", ex.Name)
	}
}
