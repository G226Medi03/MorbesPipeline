package store

import (
	"context"
	"encoding/json"
	"github.com/G226Medi03/MorbesPipeline/sourceutil"
	"path/filepath"
	"testing"

	"github.com/G226Medi03/MedicalEntity/ename/eactor"
	testdb "github.com/G226Medi03/MorbesPipeline/internal/testdb"
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	"github.com/G226Medi03/SourceEntity"
)

func TestDeviceActionLoadsIntoGraph(t *testing.T) {
	ctx := context.Background()
	db, err := testdb.Open(filepath.Join(t.TempDir(), "device-graph.db"))
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close
	st := New(ctx, db)
	if err := st.InitSchema(); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}
	device, err := st.SaveDevice(eactor.Device{Name: "도뇨관"}, []source.SourceRef{{SourceID: sourceutil.SourceID(1)}})
	if err != nil {
		t.Fatalf("SaveDevice: %v", err)
	}
	action, err := st.SaveAction(entity.NewAction(
		"cause_status",
		json.RawMessage(`{"status":{"kind":"inflammation","actor":"방광"}}`),
		[]source.SourceRef{{SourceID: sourceutil.SourceID(1)}},
	))
	if err != nil {
		t.Fatalf("SaveAction: %v", err)
	}
	if err := st.LinkDeviceAction(entity.DeviceID(device.ID), action.Id(), sourceutil.SourceID(2)); err != nil {
		t.Fatalf("LinkDeviceAction: %v", err)
	}
	container, err := st.LoadGraph()
	if err != nil {
		t.Fatalf("LoadGraph: %v", err)
	}
	actor, ok := container.Actors[device.ID]
	if !ok {
		t.Fatalf("device actor not loaded: %#v", container.Actors)
	}
	if _, ok := actor.(eactor.Device); !ok {
		t.Fatalf("device actor = %#v", actor)
	}
	ids := container.ActionActors[action.Id()]
	if len(ids) != 1 || ids[0] != device.ID {
		t.Fatalf("action actors = %#v", ids)
	}
}
