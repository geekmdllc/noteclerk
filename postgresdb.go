package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/geekmdio/ehrprotorepo/goproto"
	"github.com/pkg/errors"
	"strings"
)

type DbPostgres struct {
	db *sql.DB
}

// Initialize() initializes the connection to database. Ensure that the ./config/config.<environment>.json
// file has been created and properly configured with server and database values. Of note, the '<environment>'
// can be set to any value, so long as the NOTECLERK_ENVIRONMENT environmental variable's value matches.
// RETURNS: *sql.db, error
func (d *DbPostgres) Initialize(config *Config) (*sql.DB, error) {
	connStr := fmt.Sprintf("user=%v password=%v host=%v dbname=%v sslmode=%v port=%v",
		config.DbUsername, config.DbPassword, config.DbIp, config.DbName, config.DbSslMode, config.DbPort)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, errors.Wrapf(ErrPostgresDbInitFailedToOpenConn, "%v", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		return nil, errors.Wrapf(ErrPostgresDbInitFailedToPingDb, "%v", err)
	}

	d.db = db
	schemaErr := d.createSchema()
	if schemaErr != nil {
		return nil, errors.Wrapf(ErrPostgresDbInitFailedToCreateSchema, "%v", schemaErr)
	}

	return d.db, nil
}

func (d *DbPostgres) AddNote(n *ehrpb.Note) (id int64, err error) {

	row := d.db.QueryRow(addNoteQuery, n.DateCreated.GetSeconds(), n.DateCreated.GetNanos(),
		n.GetNoteGuid(), n.GetVisitGuid(), n.GetAuthorGuid(), n.GetPatientGuid(), n.GetType(),
		n.GetStatus())

	if scanErr := row.Scan(n.Id); scanErr != nil && scanErr != sql.ErrNoRows {
		return 0, errors.Wrapf(ErrPostgresDbAddNoteFailedToGetNewId, "%v", scanErr)
	}

	for _, v := range n.GetFragments() {
		_, _, err := d.AddNoteFragment(v)
		if err != nil && err != sql.ErrNoRows {
			return 0, errors.Wrapf(ErrPostgresDbAddNoteFailedToAddNoteFragments, "%v", err)
		}
	}

	for _, v := range n.GetTags() {
		_, err := d.AddNoteTag(n.GetNoteGuid(), v)
		if err != nil && err != sql.ErrNoRows{
			return 0, errors.Wrapf(ErrPostgresDbAddNoteFailedToAddNoteTagToDb, "%v", err)
		}
	}

	return n.Id,nil
}

func (d *DbPostgres) UpdateNote(n *ehrpb.Note) error {
	log.Fatal("Not implemented.")
	return nil
}

func (d *DbPostgres) DeleteNote(id int64) error {
	log.Fatal("Not implemented.")
	return nil
}

func (d *DbPostgres) AllNotes() ([]*ehrpb.Note, error) {
	rows, err := d.db.Query("SELECT * FROM note;")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	var notes []*ehrpb.Note
	for rows.Next() {
		tmpNote := NewNote()
		err := rows.Scan(&tmpNote.Id, &tmpNote.DateCreated.Seconds, &tmpNote.DateCreated.Nanos,
			&tmpNote.NoteGuid, &tmpNote.VisitGuid, &tmpNote.AuthorGuid,
			&tmpNote.PatientGuid, &tmpNote.Type, &tmpNote.Status)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		notes = append(notes, tmpNote)

	}
	return notes, nil
}

func (d *DbPostgres) AddNoteTag(noteGuid string, tag string) (id int64, err error) {
	row := d.db.QueryRow(addNoteTagQuery, noteGuid, tag)

	var newId int64
	if scanErr := row.Scan(newId); scanErr != nil && scanErr != sql.ErrNoRows {
		return 0, errors.Wrapf(ErrPostgresDbAddNoteTagFailedToGetNewId, "%v", scanErr)
	}

	return newId, nil
}

func (d *DbPostgres) GetNoteById(id int64) (*ehrpb.Note, error) {
	log.Fatal("Not implemented.")
	return nil, nil
}

func (d *DbPostgres) FindNote(filter NoteFindFilter) ([]*ehrpb.Note, error) {
	log.Fatal("Not implemented.")
	return nil, nil
}

func (d *DbPostgres) AllNoteFragments() ([]*ehrpb.NoteFragment, error) {
	log.Fatal("Not implemented.")
	return nil, nil
}

func (d *DbPostgres) AddNoteFragment(nf *ehrpb.NoteFragment)  (id int64, guid string, err error) {
	row := d.db.QueryRow(addNoteFragmentQuery, nf.DateCreated.Seconds, nf.DateCreated.Nanos,
		nf.GetNoteFragmentGuid(), nf.GetNoteGuid(), nf.GetIcd_10Code(), nf.GetIcd_10Long(),
		nf.GetDescription(), nf.GetStatus(), nf.GetPriority(), nf.GetTopic(), nf.GetContent())
	scanErr := row.Scan(nf.Id)
	if scanErr != nil && scanErr != sql.ErrNoRows {
		return 0, "", errors.Wrapf(ErrPostgresDbAddNoteFragmentFailedToGetNewId, "%v", scanErr)
	}

	for _, v := range nf.GetTags() {
		_, err := d.AddNoteTag(nf.GetNoteFragmentGuid(), v)
		if err != nil && err != sql.ErrNoRows{
			return 0, nf.NoteFragmentGuid, errors.Wrapf(ErrPostgresDbAddNoteFragmentFailedToAddNoteFragmentTagToDb, "%v", err)
		}
	}

	return nf.GetId(), nf.GetNoteFragmentGuid(),nil
}

func (d *DbPostgres) UpdateNoteFragment(n *ehrpb.NoteFragment)  error {
	log.Fatal("Not implemented.")
	return nil
}

func (d *DbPostgres) DeleteNoteFragment(id int64)  error {
	log.Fatal("Not implemented.")
	return nil
}

func (d *DbPostgres) GetNoteFragmentsById(id int64) (*ehrpb.NoteFragment, error) {
	log.Fatal("Not implemented.")
	return nil, nil
}

func (d *DbPostgres) FindNoteFragments(filter NoteFragmentFindFilter) ([]*ehrpb.NoteFragment, error) {
	log.Fatal("Not implemented.")
	return nil, nil
}


func (d *DbPostgres) AddNoteFragmentTag(noteGuid string, tag string) (id int64, err error) {
	row := d.db.QueryRow(addNoteFragmentTagQuery, noteGuid, tag)

	var newId int64
	if scanErr := row.Scan(newId); scanErr != nil && scanErr != sql.ErrNoRows {
		return 0, errors.Wrapf(ErrPostgresDbAddNoteTagFailedToGetNewId, "%v", scanErr)
	}

	return newId, nil
}

// https://www.calhoun.io/updating-and-deleting-postgresql-records-using-gos-sql-package/
func (d *DbPostgres) createSchema() error {
	err := d.createTable(createNoteTable)
	if notNilNotTableExists(err) {
		return errors.Wrapf(ErrPostgresDbCreateSchemaFails, "Target table: note. Error: %v", err)
	}
	if err == ErrPostgresDbInitTableAlreadyExistsErr {
		log.Warn("Table 'note' already exists.")
	}
	err = d.createTable(createNoteTagTable)
	if notNilNotTableExists(err) {
		return errors.Wrapf(ErrPostgresDbCreateSchemaFails, "Target table: note_tag. Error: %v", err)
	}
	if err == ErrPostgresDbInitTableAlreadyExistsErr {
		log.Warn("Table 'note_tag' already exists.")
	}
	err = d.createTable(createNoteFragmentTable)
	if notNilNotTableExists(err) {
		return errors.Wrapf(ErrPostgresDbCreateSchemaFails, "Target table: note_fragment. Error: %v", err)
	}
	if err == ErrPostgresDbInitTableAlreadyExistsErr {
		log.Warn("Table 'note_fragment' already exists.")
	}
	err = d.createTable(createNoteFragmentTagTable)
	if notNilNotTableExists(err) {
		return errors.Wrapf(ErrPostgresDbCreateSchemaFails, "Target table: note_fragment_tag. Error: %v", err)
	}
	if err == ErrPostgresDbInitTableAlreadyExistsErr {
		log.Warn("Table 'note_fragment_tag' already exists.")
	}

	//TODO: Remove this.
	tmpNote := NewNote()
	tmpNote.Tags = append(tmpNote.Tags, "note1Tag1", "note1Tag2")

	tmpFrag := NewNoteFragment()
	tmpFrag.NoteGuid = tmpNote.GetNoteGuid()
	tmpFrag.Tags = append(tmpFrag.Tags, "frag1Tag1", "frag1Tag2")

	tmpFrag2 := NewNoteFragment()
	tmpFrag2.NoteGuid = tmpNote.GetNoteGuid()
	tmpFrag2.Tags = append(tmpFrag.Tags, "frag2Tag1", "frag2Tag2")

	tmpNote.Fragments = append(tmpNote.Fragments, tmpFrag, tmpFrag2)

	d.AddNote(tmpNote)

	notes, notesErr := d.AllNotes()
	if notesErr != nil {
		log.Println(notesErr)
	}
	fmt.Println(notes)
	//End remove

	return nil
}

func (d *DbPostgres) createTable(query string) error {
	_, err := d.db.Exec(query)

	tableExistsError := strings.Contains(fmt.Sprintf("%v", err), "already exists")
	if tableExistsError {
		return ErrPostgresDbInitTableAlreadyExistsErr
	}
	if err != nil {
		return errors.Wrapf(ErrPostgresDbCreateTableFails, "%v", err)
	}
	return nil
}

func notNilNotTableExists(err error) bool {
	return err != nil && err != ErrPostgresDbInitTableAlreadyExistsErr
}
