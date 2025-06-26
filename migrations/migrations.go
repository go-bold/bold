package migrations

import (
	"database/sql"
	"fmt"
	"strings"
)

var MySQL = &mysqlProvider{}
var PostgreSQL = &postgresqlProvider{}

type Column struct {
	Name      string
	Type      string
	Length    *int
	Nullable  bool
	Default   interface{}
	Primary   bool
	Unique    bool
	Comment   string
	After     string
}

type Blueprint interface {
	ID() ColumnBuilder
	String(name string, length int) ColumnBuilder
	Text(name string) ColumnBuilder
	Integer(name string) ColumnBuilder
	BigInteger(name string) ColumnBuilder
	Float(name string) ColumnBuilder
	Double(name string) ColumnBuilder
	Decimal(name string, precision, scale int) ColumnBuilder
	Boolean(name string) ColumnBuilder
	Date(name string) ColumnBuilder
	DateTime(name string) ColumnBuilder
	Timestamp(name string) ColumnBuilder
	JSON(name string) ColumnBuilder
	Binary(name string) ColumnBuilder
	UUID(name string) ColumnBuilder
	Timestamps()
	Index(columns ...string)
	UniqueIndex(columns ...string)
	Primary(columns ...string)
	FullTextIndex(columns ...string)
	Foreign(column string) ForeignKeyBuilder
	AddColumn(name, columnType string) ColumnBuilder
}

type MySQLBlueprint interface {
	Blueprint
	Enum(name string, values []string) ColumnBuilder
	Set(name string, values []string) ColumnBuilder
	Point(name string) ColumnBuilder
	Geometry(name string) ColumnBuilder
}

type PostgreSQLBlueprint interface {
	Blueprint
	Serial(name string) ColumnBuilder
	BigSerial(name string) ColumnBuilder
	JSONB(name string) ColumnBuilder
	Array(name string, baseType string) ColumnBuilder
	Inet(name string) ColumnBuilder
	CIDR(name string) ColumnBuilder
	MacAddr(name string) ColumnBuilder
	TsVector(name string) ColumnBuilder
	XML(name string) ColumnBuilder
	Money(name string) ColumnBuilder
	HStore(name string) ColumnBuilder
}

type ColumnBuilder interface {
	Nullable() ColumnBuilder
	NotNullable() ColumnBuilder
	Default(value interface{}) ColumnBuilder
	Unique() ColumnBuilder
	Primary() ColumnBuilder
	Comment(text string) ColumnBuilder
	After(column string) ColumnBuilder
	Index() ColumnBuilder
}

type ForeignKeyBuilder interface {
	References(column string) ForeignKeyBuilder
	On(table string) ForeignKeyBuilder
	OnDelete(action string) ForeignKeyBuilder
	OnUpdate(action string) ForeignKeyBuilder
}

type blueprint struct {
	tableName string
	columns   []Column
	indexes   []string
	foreigns  []string
	db        *sql.DB
}

func newBlueprint(tableName string, db *sql.DB) *blueprint {
	return &blueprint{
		tableName: tableName,
		columns:   []Column{},
		indexes:   []string{},
		foreigns:  []string{},
		db:        db,
	}
}

func (b *blueprint) AddColumn(name, columnType string) ColumnBuilder {
	column := Column{
		Name: name,
		Type: columnType,
	}
	b.columns = append(b.columns, column)
	return &columnBuilder{
		column:    &b.columns[len(b.columns)-1],
		blueprint: b,
	}
}

func (b *blueprint) ID() ColumnBuilder {
	return b.AddColumn("id", "BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY")
}

func (b *blueprint) String(name string, length int) ColumnBuilder {
	return b.AddColumn(name, fmt.Sprintf("VARCHAR(%d)", length))
}

func (b *blueprint) Text(name string) ColumnBuilder {
	return b.AddColumn(name, "TEXT")
}

func (b *blueprint) Integer(name string) ColumnBuilder {
	return b.AddColumn(name, "INT")
}

func (b *blueprint) BigInteger(name string) ColumnBuilder {
	return b.AddColumn(name, "BIGINT")
}

func (b *blueprint) Float(name string) ColumnBuilder {
	return b.AddColumn(name, "FLOAT")
}

func (b *blueprint) Double(name string) ColumnBuilder {
	return b.AddColumn(name, "DOUBLE")
}

func (b *blueprint) Decimal(name string, precision, scale int) ColumnBuilder {
	return b.AddColumn(name, fmt.Sprintf("DECIMAL(%d,%d)", precision, scale))
}

func (b *blueprint) Boolean(name string) ColumnBuilder {
	return b.AddColumn(name, "BOOLEAN")
}

func (b *blueprint) Date(name string) ColumnBuilder {
	return b.AddColumn(name, "DATE")
}

func (b *blueprint) DateTime(name string) ColumnBuilder {
	return b.AddColumn(name, "DATETIME")
}

func (b *blueprint) Timestamp(name string) ColumnBuilder {
	return b.AddColumn(name, "TIMESTAMP")
}

func (b *blueprint) JSON(name string) ColumnBuilder {
	return b.AddColumn(name, "JSON")
}

func (b *blueprint) Binary(name string) ColumnBuilder {
	return b.AddColumn(name, "BINARY")
}

func (b *blueprint) UUID(name string) ColumnBuilder {
	return b.AddColumn(name, "CHAR(36)")
}

func (b *blueprint) Timestamps() {
	b.AddColumn("created_at", "TIMESTAMP DEFAULT CURRENT_TIMESTAMP")
	b.AddColumn("updated_at", "TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP")
}

func (b *blueprint) Index(columns ...string) {
	indexName := strings.Join(columns, "_") + "_index"
	b.indexes = append(b.indexes, fmt.Sprintf("INDEX %s (%s)", indexName, strings.Join(columns, ", ")))
}

func (b *blueprint) UniqueIndex(columns ...string) {
	indexName := strings.Join(columns, "_") + "_unique"
	b.indexes = append(b.indexes, fmt.Sprintf("UNIQUE INDEX %s (%s)", indexName, strings.Join(columns, ", ")))
}

func (b *blueprint) Primary(columns ...string) {
	b.indexes = append(b.indexes, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(columns, ", ")))
}

func (b *blueprint) FullTextIndex(columns ...string) {
	indexName := strings.Join(columns, "_") + "_fulltext"
	b.indexes = append(b.indexes, fmt.Sprintf("FULLTEXT INDEX %s (%s)", indexName, strings.Join(columns, ", ")))
}

func (b *blueprint) Foreign(column string) ForeignKeyBuilder {
	return &foreignKeyBuilder{
		localColumn: column,
		blueprint:   b,
	}
}

type columnBuilder struct {
	column    *Column
	blueprint *blueprint
}

func (c *columnBuilder) Nullable() ColumnBuilder {
	c.column.Nullable = true
	return c
}

func (c *columnBuilder) NotNullable() ColumnBuilder {
	c.column.Nullable = false
	return c
}

func (c *columnBuilder) Default(value interface{}) ColumnBuilder {
	c.column.Default = value
	return c
}

func (c *columnBuilder) Unique() ColumnBuilder {
	c.column.Unique = true
	return c
}

func (c *columnBuilder) Primary() ColumnBuilder {
	c.column.Primary = true
	return c
}

func (c *columnBuilder) Comment(text string) ColumnBuilder {
	c.column.Comment = text
	return c
}

func (c *columnBuilder) After(column string) ColumnBuilder {
	c.column.After = column
	return c
}

func (c *columnBuilder) Index() ColumnBuilder {
	c.blueprint.Index(c.column.Name)
	return c
}

type foreignKeyBuilder struct {
	localColumn    string
	foreignTable   string
	foreignColumn  string
	onDelete       string
	onUpdate       string
	blueprint      *blueprint
}

func (f *foreignKeyBuilder) References(column string) ForeignKeyBuilder {
	f.foreignColumn = column
	return f
}


func (f *foreignKeyBuilder) OnDelete(action string) ForeignKeyBuilder {
	f.onDelete = action
	f.build()
	return f
}

func (f *foreignKeyBuilder) OnUpdate(action string) ForeignKeyBuilder {
	f.onUpdate = action
	f.build()
	return f
}

func (f *foreignKeyBuilder) On(table string) ForeignKeyBuilder {
	f.foreignTable = table
	f.build()
	return f
}

func (f *foreignKeyBuilder) build() {
	if f.foreignTable == "" || f.foreignColumn == "" {
		return
	}
	
	var parts []string
	parts = append(parts, fmt.Sprintf("FOREIGN KEY (%s)", f.localColumn))
	parts = append(parts, fmt.Sprintf("REFERENCES %s (%s)", f.foreignTable, f.foreignColumn))
	
	if f.onDelete != "" {
		parts = append(parts, fmt.Sprintf("ON DELETE %s", f.onDelete))
	}
	
	if f.onUpdate != "" {
		parts = append(parts, fmt.Sprintf("ON UPDATE %s", f.onUpdate))
	}
	
	foreignSQL := strings.Join(parts, " ")
	f.blueprint.foreigns = append(f.blueprint.foreigns, foreignSQL)
}