package regression

import("context";"database/sql";"path/filepath";"strata-weave/internal/model";"strata-weave/internal/service";"strata-weave/internal/store";"sync";"testing")
func setupBug002(t *testing.T)(*service.Service,*sql.DB){t.Helper();db,e:=store.Open(filepath.Join(t.TempDir(),"field.db"));if e!=nil{t.Fatal(e)};t.Cleanup(func(){_=db.Close()});app:=service.New(db);if _,e=app.CreateTrench(context.Background(),model.Trench{ID:"t1",Code:"T-1",Site:"Ridge"});e!=nil{t.Fatal(e)};if _,e=app.CreateUnit(context.Background(),model.Unit{ID:"u1",TrenchID:"t1",Code:"U-1",Phase:1});e!=nil{t.Fatal(e)};return app,db}
func TestBug002_ReviewQueueConcurrentSubmission(t *testing.T){app,_:=setupBug002(t);r,e:=app.CreateRecord(context.Background(),model.FieldRecord{ID:"r1",UnitID:"u1",Author:"a",Notes:"layer note"});if e!=nil{t.Fatal(e)};var wg sync.WaitGroup;for n:=0;n<20;n++{wg.Add(1);go func(){defer wg.Done();_=app.SubmitRecord(context.Background(),r.ID)}()};wg.Wait()}
