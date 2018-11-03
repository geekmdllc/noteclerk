package main

import (
	"context"
	"github.com/geekmdio/ehrprotorepo/v1/generated/goproto"
	"github.com/geekmdio/noted"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"
	"testing"
)

func TestDbPostgres_InitializeWithEmptyConfig_ThrowsError(t *testing.T) {
	config := &Config{}
	db := &DbPostgres{}
	if err := db.Initialize(config); err == nil {
		t.Fatalf("Initializing with a blank config should return an error, but returned %v", err)
	}
}

func TestNoteClerkServer_CreateNote(t *testing.T) {
	s := &Server{}
	s.Initialize(&Config{}, mockDb)
	c := context.Background()
	cnr := &ehrpb.CreateNoteRequest{
		Note: &ehrpb.Note{
			Id:          0,                      // Generated by NewNote (via primary key auto-increment)
			DateCreated: &timestamp.Timestamp{}, // Generated by NewNote
			NoteGuid:    "",                     // Generated by NewNote
			VisitGuid:   uuid.New().String(),
			AuthorGuid:  uuid.New().String(),
			PatientGuid: uuid.New().String(),
			Type:        ehrpb.NoteType_CONTINUED_CARE_DOCUMENTATION,
		},
	}
	res, err := s.CreateNote(c, cnr)

	if err != nil {
		t.Fatalf("%v", err)
	}

	if res.Status.HttpCode != ehrpb.StatusCodes_OK {
		t.Fatalf("Status could should be OK")
	}

	if res.Note.GetDateCreated().Seconds == 0 {
		t.Fatalf("The timestamp was not created.")
	}

	_, noteGuidParseErr := uuid.Parse(res.Note.GetNoteGuid())
	if noteGuidParseErr != nil {
		t.Fatalf("The note did not have a new GUID created for it, and is likely invalid.")
	}

	_, visitGuidParseErr := uuid.Parse(res.Note.GetVisitGuid())
	if visitGuidParseErr != nil {
		t.Fatalf("The note did not have a new GUID created for it, and is likely invalid.")
	}

	_, authorGuidParseErr := uuid.Parse(res.Note.GetAuthorGuid())
	if authorGuidParseErr != nil {
		t.Fatalf("The note did not have a new GUID created for it, and is likely invalid.")
	}

	_, patientGuidParseErr := uuid.Parse(res.Note.GetPatientGuid())
	if patientGuidParseErr != nil {
		t.Fatalf("The note did not have a new GUID created for it, and is likely invalid.")
	}

	if res.Note.Type != cnr.Note.Type {
		t.Fatalf("Note type was not transferred to resulting note.")
	}
}

func TestNoteClerkServer_CreateNote_WithNoteThatAlreadyHasId_ReturnsError(t *testing.T) {
	s := &Server{}
	s.Initialize(&Config{}, mockDb)
	c := context.Background()
	cnr := &ehrpb.CreateNoteRequest{
		Note: &ehrpb.Note{
			Id:          1, // This should throw an error
			DateCreated: &timestamp.Timestamp{},
			NoteGuid:    "",
			VisitGuid:   uuid.New().String(),
			AuthorGuid:  uuid.New().String(),
			PatientGuid: uuid.New().String(),
			Type:        ehrpb.NoteType_CONTINUED_CARE_DOCUMENTATION,
		},
	}
	_, err := s.CreateNote(c, cnr)

	if err == nil {
		t.Fatalf("Should return an error when a note with a similar Id already exists, but error was %v", err)
	}
}

func TestNoteClerkServer_CreateNote_WithFragmentsRetainsFragments(t *testing.T) {
	s := &Server{}
	s.Initialize(&Config{}, mockDb)

	expectedFragId := int64(44)
	noteFrag := noted.NewNoteFragment()
	noteFrag.Id = expectedFragId
	cnr := &ehrpb.CreateNoteRequest{Note: noted.NewNote()}

	cnr.Note.Fragments = append(cnr.Note.Fragments, noteFrag)
	res, err := s.CreateNote(context.Background(), cnr)
	if err != nil {
		t.Fatalf("Error creating a new note, err %v", err)
	}

	if len(res.Note.Fragments) <= 0 {
		t.Fatalf("Note fragments were note appended.")
	}

	firstFrag := res.Note.Fragments[0]
	if firstFrag.Id != expectedFragId {
		t.Fatalf("Was expecting an id of %v and got %v", expectedFragId, firstFrag.Id)
	}
}

func TestNoteClerkServer_CreateNote_WithTagsRetainsTags(t *testing.T) {
	s := &Server{}
	s.Initialize(&Config{}, mockDb)
	c := context.Background()
	expectedTag := "mytag"
	cnr := &ehrpb.CreateNoteRequest{
		Note: &ehrpb.Note{
			Id: 0,
		},
	}
	cnr.Note.Tags = append(cnr.Note.Tags, expectedTag)

	res, err := s.CreateNote(c, cnr)
	if err != nil {
		t.Fatalf("Error creating a new note, err %v", err)
	}

	if len(res.Note.Tags) <= 0 {
		t.Fatalf("Should be one tag present")
	}

	firstTag := res.Note.Tags[0]
	if firstTag != expectedTag {
		t.Fatalf("Expected tag %v, but got %v", expectedTag, firstTag)
	}
}

func TestNoteClerkServer_CreateNote_WithNonZeroIdIsRejected(t *testing.T) {
	s := &Server{}
	s.Initialize(&Config{}, mockDb)
	cnr := &ehrpb.CreateNoteRequest{
		Note: noted.NewNote(),
	}
	cnr.Note.Id = 1
	res, err := s.CreateNote(context.Background(), cnr)
	if err == nil {
		t.Fatalf("Note should be rejected for non-zero id.")
	}

	if res != nil {
		t.Fatalf("The response should be nil because note was rejected")
	}

}

func TestNoteClerkServer_DeleteNote(t *testing.T) {
	s := &Server{}
	s.Initialize(&Config{}, mockDb)

	notes, _ := s.db.AllNotes()

	guidToDelete := notes[0].GetNoteGuid()
	delReq := &ehrpb.DeleteNoteRequest{
		Guid: guidToDelete,
	}

	res, err := s.DeleteNote(context.Background(), delReq)
	if err != nil {
		t.Fatalf("Failed to perform delete request.")
	}

	if res.Status.HttpCode != ehrpb.StatusCodes_OK {
		t.Fatalf("Status response should be OK")
	}

	allNotes, _ := s.db.AllNotes()
	idPresent := false
	for _, n := range allNotes {
		if n.GetNoteGuid() == guidToDelete {
			idPresent = true
		}
	}
	if idPresent {
		t.Fatalf("Note is still present in the database, confimed by id.")
	}
}

func TestNoteClerkServer_DeleteNote_WhichDoestExistReturnsError(t *testing.T) {
	s := &Server{}
	s.Initialize(&Config{}, mockDb)

	idToDelete := int64(-1)
	delReq := &ehrpb.DeleteNoteRequest{
		Id: idToDelete,
	}

	res, err := s.DeleteNote(context.Background(), delReq)
	if err == nil {
		t.Fatalf("Should not be able to delete a note with a negative id, which doesn't exist")
	}

	status := res.Status.HttpCode
	if status != ehrpb.StatusCodes_NOT_MODIFIED {
		t.Fatalf("status returned %v, but should be %v", status, ehrpb.StatusCodes_NOT_MODIFIED)
	}
}

func TestNoteClerkServer_RetrieveNote(t *testing.T) {
	s := &Server{}
	s.Initialize(&Config{}, mockDb)

	notes, _ := s.db.AllNotes()
	expectedGuid := notes[0].NoteGuid

	retReq := &ehrpb.RetrieveNoteRequest{
		Guid: expectedGuid,
	}

	res, err := s.RetrieveNote(context.Background(), retReq)
	if err != nil {
		t.Fatalf("Failed to perform retrieval request. Err: %v", err)
	}

	if res.Status.HttpCode != ehrpb.StatusCodes_OK {
		t.Fatalf("Status response should be OK")
	}

	if res.Note == nil {
		t.Fatalf("No note was retrieved")
	}

	if res.Note.GetNoteGuid() != expectedGuid {
		t.Fatalf("The note Id was %v, but should have been %v", res.Note.Id, expectedGuid)
	}
}

func TestNoteClerkServer_RetrieveNote_ByIdThatDoesntExist_ReturnsError(t *testing.T) {
	s := &Server{}
	s.Initialize(&Config{}, mockDb)

	expectedId := int64(-1)

	retReq := &ehrpb.RetrieveNoteRequest{
		Id: expectedId,
	}

	res, err := s.RetrieveNote(context.Background(), retReq)
	if err == nil {
		t.Fatalf("Should not be able to find note with negative Id.")
	}

	if res.Status.HttpCode != ehrpb.StatusCodes_NOT_FOUND {
		t.Fatalf("Status response should be NOT FOUND")
	}
}

func TestNoteClerkServer_FindNote(t *testing.T) {
	s := &Server{}
	s.Initialize(&Config{}, mockDb)

	found, err := s.db.AllNotes()
	firstNote := found[0]

	findReq := &ehrpb.SearchNotesRequest{
		VisitGuid: firstNote.GetVisitGuid(),
	}

	res, err := s.SearchNotes(context.Background(), findReq)
	if err != nil {
		t.Fatalf("Failed to find note.")
	}

	if res.Status.HttpCode != ehrpb.StatusCodes_OK {
		t.Fatalf("Should result with status OK.")
	}

	noteFound := false
	for _, n := range res.Notes {
		if n.GetVisitGuid() == firstNote.VisitGuid {
			noteFound = true
			break
		}
	}
	if !noteFound {
		t.Fatalf("Failed to find a note associted with visit GUID %v", firstNote.VisitGuid)
	}

}

func TestNoteClerkServer_FindNote_WithNonExistentGuid_ReturnsError(t *testing.T) {
	s := &Server{}
	s.Initialize(&Config{}, mockDb)

	findReq := &ehrpb.SearchNotesRequest{
		VisitGuid: uuid.New().String(),
	}

	res, err := s.SearchNotes(context.Background(), findReq)
	if err == nil {
		t.Fatalf("A note with this newly generated GUID should not be found in the database.")
	}

	if res.Status.HttpCode != ehrpb.StatusCodes_NOT_FOUND {
		t.Fatalf("Should return NOT FOUND")
	}
}

func TestNoteClerkServer_UpdateNote(t *testing.T) {
	mockDb := mockDb
	err := mockDb.Initialize(nil)
	if err != nil {
		t.Fatalf("Failed to initialize mock database.")
	}

	firstNote := mockDb.db[0]

	retReq := &ehrpb.RetrieveNoteRequest{
		Guid: firstNote.GetNoteGuid(),
	}

	s := &Server{}
	s.db = mockDb
	res, _ := s.RetrieveNote(context.Background(), retReq)

	noteToUpdate := res.Note
	noteToUpdate.Tags = append(noteToUpdate.Tags, "appendedTag")

	updateReq := &ehrpb.UpdateNoteRequest{
		Id:   0,
		Note: noteToUpdate,
	}
	updateRes, updateErr := s.UpdateNote(context.Background(), updateReq)
	if updateErr != nil {
		t.Fatalf("Failed to update note")
	}

	if updateRes.Status.HttpCode != ehrpb.StatusCodes_OK {
		t.Fatalf("Status should return OK.")
	}

}

func TestNoteClerkServer_UpdateNote_NoteDoesNotExistReturnsError(t *testing.T) {
	s := &Server{}
	s.Initialize(&Config{}, mockDb)

	note := noted.NewNote()
	note.Id = -1
	updateReq := &ehrpb.UpdateNoteRequest{
		Id:   note.Id,
		Note: note,
	}
	updateRes, updateErr := s.UpdateNote(context.Background(), updateReq)
	if updateErr == nil {
		t.Fatalf("should not be able to updated note with negative Id, which doesn't exist")
	}

	if updateRes.Status.HttpCode != ehrpb.StatusCodes_NOT_FOUND {
		t.Fatalf("Status should return NOT FOUND.")
	}

}

func TestNoteClerkServer_UpdateNote_NoteIdDoesntMatchUpdateId(t *testing.T) {
	s := &Server{}
	s.Initialize(&Config{}, mockDb)

	note := noted.NewNote()
	note.Id = 0
	updateReq := &ehrpb.UpdateNoteRequest{
		Id:   1,
		Note: note,
	}
	updateRes, updateErr := s.UpdateNote(context.Background(), updateReq)
	if updateErr == nil {
		t.Fatalf("should not be able to updated note with negative Id, which doesn't exist")
	}

	if updateRes.Status.HttpCode != ehrpb.StatusCodes_CONFLICT {
		t.Fatalf("Status should return CONFLICT, but returned %v.", updateRes.Status.HttpCode)
	}

}

func TestNoteClerkServer_Initialize_WithNilConfig_ReturnsError(t *testing.T) {
	s := &Server{}
	err := s.Initialize(nil, mockDb)
	if err == nil {
		t.Fatalf("Initialize should throw error with nil config file.")
	}
}

func TestNoteClerkServer_Initialize_WithNilDb_ReturnsError(t *testing.T) {
	s := &Server{}
	err := s.Initialize(&Config{}, nil)
	if err == nil {
		t.Fatalf("Initialize should throw error with nil database.")
	}
}