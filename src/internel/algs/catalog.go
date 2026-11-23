package algs

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/guojia99/cubing-pro/src/internel/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// The source key is used once, at import. Display text and positions can then
// change in MySQL without changing links, selections or discussion targets.
func catalogID(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		fmt.Fprintf(h, "%d:%s", len(part), part)
	}
	return hex.EncodeToString(h.Sum(nil)[:8])
}

type formulaCatalogImport struct {
	ID       int  `gorm:"primaryKey;autoIncrement:false"`
	Complete bool `gorm:"not null"`
}

type formulaClass struct {
	ID       string       `gorm:"size:32;primaryKey"`
	Cube     string       `gorm:"size:32;not null;uniqueIndex:idx_formula_class_name"`
	Name     string       `gorm:"size:191;not null;uniqueIndex:idx_formula_class_name"`
	Position int          `gorm:"not null"`
	Sets     []formulaSet `gorm:"foreignKey:ClassID"`
}

type formulaSet struct {
	ID       string         `gorm:"size:32;primaryKey"`
	ClassID  string         `gorm:"size:32;not null;index"`
	Name     string         `gorm:"type:text;not null"`
	Position int            `gorm:"not null"`
	Groups   []formulaGroup `gorm:"foreignKey:SetID"`
}

type formulaGroup struct {
	ID       string        `gorm:"size:32;primaryKey"`
	SetID    string        `gorm:"size:32;not null;index"`
	Name     string        `gorm:"type:text;not null"`
	Position int           `gorm:"not null"`
	Cases    []formulaCase `gorm:"foreignKey:GroupID"`
}

type formulaCase struct {
	ID        string            `gorm:"size:32;primaryKey"`
	GroupID   string            `gorm:"size:32;not null;index"`
	Name      string            `gorm:"type:text;not null"`
	Position  int               `gorm:"not null"`
	Scrambles []string          `gorm:"serializer:json;type:text"`
	Solutions []formulaSolution `gorm:"foreignKey:CaseID"`
}

type formulaSolution struct {
	ID       string `gorm:"size:32;primaryKey"`
	CaseID   string `gorm:"size:32;not null;index:idx_formula_solution_position"`
	Position int    `gorm:"not null;index:idx_formula_solution_position"`
	Moves    string `gorm:"type:text;not null"`
}

func MigrateCatalog(db *gorm.DB) error {
	return InitCatalog(db, "")
}

func InitCatalog(db *gorm.DB, trainerPath string) error {
	return database.WithMigrationLock(db, "cubing_pro_formula_init", func(conn *gorm.DB) error {
		return initCatalog(conn, trainerPath)
	})
}

func initCatalog(db *gorm.DB, trainerPath string) error {
	schema := db
	if db.Dialector.Name() == "mysql" {
		schema = db.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci")
	}
	if err := schema.AutoMigrate(&formulaCatalogImport{}, &formulaClass{}, &formulaSet{}, &formulaGroup{}, &formulaCase{}, &formulaSolution{}); err != nil {
		return err
	}
	if trainerPath == "" {
		return nil
	}
	return db.Transaction(func(tx *gorm.DB) error {
		importState := formulaCatalogImport{ID: 1}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&importState).Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&importState, 1).Error; err != nil {
			return err
		}
		if importState.Complete {
			return nil
		}
		catalog, err := loadStatic(trainerPath)
		if err != nil {
			return err
		}
		for _, cube := range CubeKeyList {
			for pos, class := range catalog[cube].ClassList {
				if err := importClass(tx, cube, pos, class); err != nil {
					return fmt.Errorf("import %s/%s: %w", cube, class.Name, err)
				}
			}
		}
		return tx.Model(&importState).Update("complete", true).Error
	})
}

func importClass(tx *gorm.DB, cube string, position int, class AlgorithmClass) error {
	row := formulaClass{ID: catalogID(cube, class.Name), Cube: cube, Name: class.Name, Position: position}
	if err := tx.Create(&row).Error; err != nil {
		return err
	}
	var sets []formulaSet
	var groups []formulaGroup
	var cases []formulaCase
	var solutions []formulaSolution
	for si, set := range class.Sets {
		setID := catalogID(row.ID, set.Name)
		sets = append(sets, formulaSet{ID: setID, ClassID: row.ID, Name: set.Name, Position: si})
		for gi := range set.AlgorithmGroups {
			group := &set.AlgorithmGroups[gi]
			group.GroupID = catalogID(setID, group.Name)
			groups = append(groups, formulaGroup{ID: group.GroupID, SetID: setID, Name: group.Name, Position: gi})
			for ci := range group.Algorithms {
				c, caseSolutions := importCase(&group.Algorithms[ci], group.GroupID, ci)
				cases = append(cases, c)
				solutions = append(solutions, caseSolutions...)
			}
		}
	}
	if err := insertRows(tx, sets); err != nil {
		return err
	}
	if err := insertRows(tx, groups); err != nil {
		return err
	}
	if err := insertRows(tx, cases); err != nil {
		return err
	}
	if err := insertRows(tx, solutions); err != nil {
		return err
	}
	loaded, err := ReadClass(tx, cube, class.Name)
	if err != nil {
		return err
	}
	expected, err := json.Marshal(class)
	if err != nil {
		return err
	}
	actual, err := json.Marshal(loaded)
	if err != nil {
		return err
	}
	if !bytes.Equal(expected, actual) {
		return fmt.Errorf("database round trip changed the catalog")
	}
	return nil
}

func importCase(alg *Algorithm, groupID string, position int) (formulaCase, []formulaSolution) {
	alg.CaseID = catalogID(groupID, alg.sourceKey)
	alg.SolutionIDs = make([]*string, len(alg.Algs))
	occurrences := make(map[string]int)
	solutions := make([]formulaSolution, len(alg.Algs))
	for i, moves := range alg.Algs {
		var id string
		switch {
		case moves == "":
			// Empty positions have an internal row ID but no public solution ID.
			id = catalogID(alg.CaseID, "", fmt.Sprint(i))
		case occurrences[moves] == 0:
			id = catalogID(alg.CaseID, moves)
		default:
			id = catalogID(alg.CaseID, moves, fmt.Sprint(occurrences[moves]))
		}
		solutions[i] = formulaSolution{ID: id, CaseID: alg.CaseID, Position: i, Moves: moves}
		if moves != "" {
			alg.SolutionIDs[i] = &solutions[i].ID
			occurrences[moves]++
		}
	}
	c := formulaCase{
		ID: alg.CaseID, GroupID: groupID, Name: alg.Name, Position: position,
		Scrambles: alg.Scrambles,
	}
	return c, solutions
}

func insertRows[T any](db *gorm.DB, rows []T) error {
	if len(rows) == 0 {
		return nil
	}
	return db.Omit(clause.Associations).CreateInBatches(rows, 100).Error
}

func ReadClass(db *gorm.DB, cube, class string) (*AlgorithmClass, error) {
	var row formulaClass
	inOrder := func(q *gorm.DB) *gorm.DB { return q.Order("position") }
	err := db.Preload("Sets", inOrder).Preload("Sets.Groups", inOrder).Preload("Sets.Groups.Cases", inOrder).
		Preload("Sets.Groups.Cases.Solutions", inOrder).
		Where(&formulaClass{Cube: cube, Name: class}).First(&row).Error
	if err != nil {
		return nil, err
	}
	if row.Cube != cube || row.Name != class {
		return nil, gorm.ErrRecordNotFound
	}
	result := &AlgorithmClass{Name: row.Name, Sets: []AlgorithmSet{}, SetKeys: []string{}}
	for _, set := range row.Sets {
		s := AlgorithmSet{Name: set.Name, AlgorithmGroups: []AlgorithmGroup{}, GroupsKeys: []string{}}
		for _, group := range set.Groups {
			g := AlgorithmGroup{GroupID: group.ID, Name: group.Name, Algorithms: []Algorithm{}}
			for _, c := range group.Cases {
				g.Algorithms = append(g.Algorithms, c.algorithm())
			}
			s.AlgorithmGroups = append(s.AlgorithmGroups, g)
			s.GroupsKeys = append(s.GroupsKeys, g.Name)
		}
		result.Sets = append(result.Sets, s)
		result.SetKeys = append(result.SetKeys, s.Name)
	}
	return result, nil
}

func (c *formulaCase) algorithm() Algorithm {
	a := Algorithm{
		CaseID: c.ID, Name: c.Name, Scrambles: c.Scrambles,
		Algs: make([]string, len(c.Solutions)), SolutionIDs: make([]*string, len(c.Solutions)),
	}
	for i := range c.Solutions {
		sol := &c.Solutions[i]
		a.Algs[i] = sol.Moves
		if sol.Moves != "" {
			a.SolutionIDs[i] = &sol.ID
		}
	}
	return a
}
