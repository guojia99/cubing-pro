package formula

type FormulaCatalogImport struct {
	ID       int  `gorm:"primaryKey;autoIncrement:false"`
	Complete bool `gorm:"not null"`
}

type FormulaClass struct {
	ID       string       `gorm:"size:32;primaryKey"`
	Cube     string       `gorm:"size:32;not null;uniqueIndex:idx_formula_class_name"`
	Name     string       `gorm:"size:191;not null;uniqueIndex:idx_formula_class_name"`
	Position int          `gorm:"not null"`
	Sets     []FormulaSet `gorm:"foreignKey:ClassID"`
}

type FormulaSet struct {
	ID       string         `gorm:"size:32;primaryKey"`
	ClassID  string         `gorm:"size:32;not null;index"`
	Name     string         `gorm:"type:text;not null"`
	Position int            `gorm:"not null"`
	Groups   []FormulaGroup `gorm:"foreignKey:SetID"`
}

type FormulaGroup struct {
	ID       string        `gorm:"size:32;primaryKey"`
	SetID    string        `gorm:"size:32;not null;index"`
	Name     string        `gorm:"type:text;not null"`
	Position int           `gorm:"not null"`
	Cases    []FormulaCase `gorm:"foreignKey:GroupID"`
}

type FormulaCase struct {
	ID        string            `gorm:"size:32;primaryKey"`
	GroupID   string            `gorm:"size:32;not null;index"`
	Name      string            `gorm:"type:text;not null"`
	Position  int               `gorm:"not null"`
	Scrambles []string          `gorm:"serializer:json;type:text"`
	Solutions []FormulaSolution `gorm:"foreignKey:CaseID"`
}

type FormulaSolution struct {
	ID       string `gorm:"size:32;primaryKey"`
	CaseID   string `gorm:"size:32;not null;index:idx_formula_solution_position"`
	Position int    `gorm:"not null;index:idx_formula_solution_position"`
	Moves    string `gorm:"type:text;not null"`
}
