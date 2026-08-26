package regression

import("context";"database/sql";"fmt";"path/filepath";"strata-weave/internal/model";"strata-weave/internal/service";"strata-weave/internal/store";"sync";"testing")
func setupBug004(t *testing.T)(*service.Service,*sql.DB){t.Helper();db,e:=store.Open(filepath.Join(t.TempDir(),"field.db"));if e!=nil{t.Fatal(e)};t.Cleanup(func(){_=db.Close()});app:=service.New(db);if _,e=app.CreateTrench(context.Background(),model.Trench{ID:"t1",Code:"T-1",Site:"Ridge"});e!=nil{t.Fatal(e)};if _,e=app.CreateUnit(context.Background(),model.Unit{ID:"u1",TrenchID:"t1",Code:"U-1",Phase:1});e!=nil{t.Fatal(e)};return app,db}
func TestBug004_AlertLedgerConcurrentCreation(t *testing.T){app,_:=setupBug004(t);var wg sync.WaitGroup;for n:=0;n<24;n++{wg.Add(1);go func(n int){defer wg.Done();_,_=app.CreateAlert(context.Background(),model.Alert{ID:fmt.Sprintf("a-%d",n),UnitID:"u1",Severity:"high",Message:"water ingress"})}(n)};wg.Wait()}
