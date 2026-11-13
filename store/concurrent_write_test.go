package store

import (
	"context"
	"encoding/json"
	"github.com/G226Medi03/MorbesPipeline/sourceutil"
	"path/filepath"
	"sync"
	"testing"

	"github.com/G226Medi03/MedicalEntity/ename/eactor"
	testdb "github.com/G226Medi03/MorbesPipeline/internal/testdb"
	entity "github.com/G226Medi03/MorbesPipeline/legacy"
	source "github.com/G226Medi03/SourceEntity"
)

func TestConcurrentEntityUpserts(t *testing.T) {
	ctx := context.Background()
	db, err := testdb.Open(filepath.Join(t.TempDir(), "concurrent.db"))
	if err != nil {
		t.Fatalf("OpenDB: %v", err)
	}
	defer db.Close() //nolint:errcheck // deferred close
	st := New(ctx, db)
	if err := st.InitSchema(); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}

	const workers = 16
	var wg sync.WaitGroup
	errs := make(chan error, workers)
	actorIDs := make(chan entity.ActorID, workers)
	deviceIDs := make(chan entity.DeviceID, workers)
	statusIDs := make(chan entity.StatusID, workers)
	actionIDs := make(chan entity.ActionID, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			actor, err := st.SaveOrgan(eactor.UnknownActor{Name: "동시 actor"}, []source.SourceRef{{SourceID: sourceutil.SourceID(1)}})
			if err != nil {
				errs <- err
				return
			}
			actorIDs <- actor.ID
			device, err := st.SaveDevice(eactor.Device{Name: "동시 device"}, []source.SourceRef{{SourceID: sourceutil.SourceID(1)}})
			if err != nil {
				errs <- err
				return
			}
			deviceIDs <- entity.DeviceID(device.ID)
			status, err := st.SaveStatus(entity.NewStatus(
				"have_disease",
				json.RawMessage(`{"disease":"동시 질환"}`),
				[]source.SourceRef{{SourceID: sourceutil.SourceID(1)}},
			))
			if err != nil {
				errs <- err
				return
			}
			statusIDs <- status.Id()
			action, err := st.SaveAction(entity.NewAction(
				"cause_status",
				json.RawMessage(`{"status":{"kind":"have_disease","disease":"동시 질환"}}`),
				[]source.SourceRef{{SourceID: sourceutil.SourceID(1)}},
			))
			if err != nil {
				errs <- err
				return
			}
			actionIDs <- action.Id()
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent save: %v", err)
	}
	assert_single_actor_id(t, actorIDs)
	assert_single_device_id(t, deviceIDs)
	assert_single_status_id(t, statusIDs)
	assert_single_action_id(t, actionIDs)
}

func assert_single_actor_id(t *testing.T, values chan entity.ActorID) {
	t.Helper()
	close(values)
	var first entity.ActorID
	for value := range values {
		if first == 0 {
			first = value
			continue
		}
		if value != first {
			t.Fatalf("actor ids = %d, %d", first, value)
		}
	}
}

func assert_single_device_id(t *testing.T, values chan entity.DeviceID) {
	t.Helper()
	close(values)
	var first entity.DeviceID
	for value := range values {
		if first == 0 {
			first = value
			continue
		}
		if value != first {
			t.Fatalf("device ids = %d, %d", first, value)
		}
	}
}

func assert_single_status_id(t *testing.T, values chan entity.StatusID) {
	t.Helper()
	close(values)
	var first entity.StatusID
	for value := range values {
		if first == 0 {
			first = value
			continue
		}
		if value != first {
			t.Fatalf("status ids = %d, %d", first, value)
		}
	}
}

func assert_single_action_id(t *testing.T, values chan entity.ActionID) {
	t.Helper()
	close(values)
	var first entity.ActionID
	for value := range values {
		if first == 0 {
			first = value
			continue
		}
		if value != first {
			t.Fatalf("action ids = %d, %d", first, value)
		}
	}
}
