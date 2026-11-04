package scheduler

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Opt in with SCHEDULER_TEST_DSN pointing at LOCAL disposable PostgreSQL.
// Only a freshly-created, uniquely named schema is changed/dropped. No .env is read.
func schedulerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("SCHEDULER_TEST_DSN")
	if dsn == "" {
		t.Skip("set SCHEDULER_TEST_DSN to a local disposable PostgreSQL to run integration tests")
	}
	cfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal("invalid scheduler test connection configuration")
	}
	if cfg.Host != "localhost" && cfg.Host != "127.0.0.1" && cfg.Host != "::1" {
		t.Fatal("scheduler integration tests require localhost")
	}
	root := stdlib.OpenDB(*cfg)
	t.Cleanup(func() { root.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	schema := fmt.Sprintf("scheduler_test_%d", time.Now().UnixNano())
	if _, err := root.ExecContext(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatal("cannot create isolated scheduler test schema")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := root.ExecContext(ctx, "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Error("cannot remove isolated scheduler test schema")
		}
	})
	cfg.RuntimeParams["search_path"] = schema
	pool := stdlib.OpenDB(*cfg)
	pool.SetMaxOpenConns(8)
	t.Cleanup(func() { pool.Close() })
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: pool}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal("cannot open isolated scheduler database")
	}
	execSQL(t, db, `
 CREATE TABLE events (id bigserial PRIMARY KEY, title text NOT NULL DEFAULT '', start_time timestamptz,
 end_time timestamptz, status text DEFAULT 'upcoming', created_at timestamptz, updated_at timestamptz, deleted_at timestamptz);
 CREATE TABLE event_registrations (id bigserial PRIMARY KEY, event_id bigint NOT NULL REFERENCES events(id),
 user_id bigint NOT NULL, reminder_sent boolean NOT NULL DEFAULT false, created_at timestamptz, updated_at timestamptz, deleted_at timestamptz);
 CREATE TABLE notifications (id bigserial PRIMARY KEY, user_id bigint NOT NULL, notification_type text NOT NULL,
 title text NOT NULL, body text NOT NULL, reference_type text NOT NULL, reference_id bigint NOT NULL,
 redirect_url text NOT NULL, is_read boolean NOT NULL DEFAULT false, created_at timestamptz, updated_at timestamptz, deleted_at timestamptz);`)
	return db
}

func execSQL(t *testing.T, db *gorm.DB, query string, args ...interface{}) {
	t.Helper()
	if err := db.Exec(query, args...).Error; err != nil {
		t.Fatalf("fixture SQL failed: %v", err)
	}
}

func seedReminders(t *testing.T, db *gorm.DB, now time.Time, count int) {
	t.Helper()
	execSQL(t, db, "INSERT INTO events (id, title, start_time) VALUES (1, 'Alumni', ?)", now.Add(24*time.Hour))
	execSQL(t, db, "INSERT INTO event_registrations (event_id, user_id) SELECT 1, generate_series(1, ?)", count)
}

func assertReminderCounts(t *testing.T, db *gorm.DB, want int64) {
	t.Helper()
	var sent, notifications int64
	if err := db.Table("event_registrations").Where("reminder_sent").Count(&sent).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Table("notifications").Count(&notifications).Error; err != nil {
		t.Fatal(err)
	}
	if sent != want || notifications != want {
		t.Fatalf("sent=%d notifications=%d want=%d", sent, notifications, want)
	}
}

func TestPostgresRemindersConcurrentBoundedAndIdempotent(t *testing.T) {
	db := schedulerTestDB(t)
	now := time.Now()
	seedReminders(t, db, now, 205)
	n, err := persistReminderBatch(context.Background(), db, now)
	if err != nil || n != reminderBatchSize {
		t.Fatalf("first batch=%d err=%v", n, err)
	}
	assertReminderCounts(t, db, 100)
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := RunEventRemindersOnceContext(context.Background(), db, nil); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	assertReminderCounts(t, db, 205)
	if err := RunEventRemindersOnceContext(context.Background(), db, nil); err != nil {
		t.Fatal(err)
	}
	assertReminderCounts(t, db, 205)
	var duplicates int64
	if err := db.Raw("SELECT count(*) FROM (SELECT user_id, reference_id FROM notifications GROUP BY user_id, reference_id HAVING count(*) > 1) d").Scan(&duplicates).Error; err != nil {
		t.Fatal(err)
	}
	if duplicates != 0 {
		t.Fatal("duplicate reminders")
	}
}

func TestPostgresRemindersSkipLocked(t *testing.T) {
	db := schedulerTestDB(t)
	now := time.Now()
	seedReminders(t, db, now, 2)
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	defer tx.Rollback()
	execSQL(t, tx, "SELECT id FROM event_registrations WHERE id=1 FOR UPDATE")
	n, err := persistReminderBatch(context.Background(), db, now)
	if err != nil || n != 1 {
		t.Fatalf("unlocked batch=%d err=%v", n, err)
	}
	assertReminderCounts(t, db, 1)
	if err := tx.Rollback().Error; err != nil {
		t.Fatal(err)
	}
	n, err = persistReminderBatch(context.Background(), db, now)
	if err != nil || n != 1 {
		t.Fatalf("retry batch=%d err=%v", n, err)
	}
	assertReminderCounts(t, db, 2)
}

func TestPostgresReminderRollback(t *testing.T) {
	for _, tc := range []struct{ name, constraint string }{
		{"insert failure", "ALTER TABLE notifications ADD CONSTRAINT injected_failure CHECK (user_id < 0)"},
		{"flag failure after insert", "ALTER TABLE event_registrations ADD CONSTRAINT injected_failure CHECK (NOT reminder_sent)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := schedulerTestDB(t)
			now := time.Now()
			seedReminders(t, db, now, 2)
			execSQL(t, db, tc.constraint)
			n, err := persistReminderBatch(context.Background(), db, now)
			if err == nil || n != 0 {
				t.Fatalf("expected rollback, count=%d err=%v", n, err)
			}
			assertReminderCounts(t, db, 0)
		})
	}
}

func TestPostgresReminderEligibility(t *testing.T) {
	db := schedulerTestDB(t)
	now := time.Now().UTC().Truncate(time.Second)
	for i, tc := range []struct {
		offset                                  time.Duration
		eventDeleted, registrationDeleted, sent bool
	}{
		{23 * time.Hour, false, false, false}, {25 * time.Hour, false, false, false},
		{23*time.Hour - time.Second, false, false, false}, {24 * time.Hour, true, false, false},
		{24 * time.Hour, false, true, false}, {24 * time.Hour, false, false, true},
	} {
		execSQL(t, db, "INSERT INTO events(id, start_time, deleted_at) VALUES (?, ?, CASE WHEN ? THEN now() ELSE NULL END)", i+1, now.Add(tc.offset), tc.eventDeleted)
		execSQL(t, db, "INSERT INTO event_registrations(event_id,user_id,reminder_sent,deleted_at) VALUES (?, ?, ?, CASE WHEN ? THEN now() ELSE NULL END)", i+1, i+1, tc.sent, tc.registrationDeleted)
	}
	n, err := persistReminderBatch(context.Background(), db, now)
	if err != nil || n != 1 {
		t.Fatalf("eligible=%d err=%v", n, err)
	}
	var recipient uint
	if err := db.Table("notifications").Select("user_id").Scan(&recipient).Error; err != nil {
		t.Fatal(err)
	}
	if recipient != 1 {
		t.Fatalf("recipient=%d", recipient)
	}
}

func TestPostgresReminderCancellationRollsBack(t *testing.T) {
	db := schedulerTestDB(t)
	now := time.Now()
	seedReminders(t, db, now, 1)
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	defer tx.Rollback()
	execSQL(t, tx, "LOCK TABLE notifications IN ACCESS EXCLUSIVE MODE")
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	started := time.Now()
	if _, err := persistReminderBatch(ctx, db, now); err == nil {
		t.Fatal("expected cancellation")
	}
	if time.Since(started) > 2*time.Second {
		t.Fatal("cancellation did not interrupt DB wait")
	}
	if err := tx.Rollback().Error; err != nil {
		t.Fatal(err)
	}
	assertReminderCounts(t, db, 0)
	if _, err := persistReminderBatch(context.Background(), db, now); err != nil {
		t.Fatal(err)
	}
	assertReminderCounts(t, db, 1)
}

func TestPostgresStatusTransitions(t *testing.T) {
	db := schedulerTestDB(t)
	now := time.Now()
	cases := []struct {
		name, from, want string
		start, end       *time.Time
		deleted          bool
	}{
		{name: "future", from: "upcoming", want: "upcoming", start: timePtr(now.Add(time.Hour))},
		{name: "started", from: "upcoming", want: "ongoing", start: timePtr(now.Add(-time.Hour))},
		{name: "ended", from: "upcoming", want: "completed", start: timePtr(now.Add(-time.Hour)), end: timePtr(now.Add(-time.Minute))},
		{name: "24h fallback", from: "ongoing", want: "completed", start: timePtr(now.Add(-25 * time.Hour))},
		{name: "explicit future end", from: "ongoing", want: "ongoing", start: timePtr(now.Add(-25 * time.Hour)), end: timePtr(now.Add(time.Hour))},
		{name: "cancelled", from: "cancelled", want: "cancelled", start: timePtr(now.Add(-48 * time.Hour))},
		{name: "completed", from: "completed", want: "completed", start: timePtr(now.Add(-48 * time.Hour))},
		{name: "no start", from: "upcoming", want: "upcoming"},
		{name: "deleted", from: "upcoming", want: "upcoming", start: timePtr(now.Add(-48 * time.Hour)), deleted: true},
	}
	for i, tc := range cases {
		execSQL(t, db, "INSERT INTO events(id,status,start_time,end_time,deleted_at) VALUES(?,?,?,?,CASE WHEN ? THEN now() ELSE NULL END)", i+1, tc.from, tc.start, tc.end, tc.deleted)
	}
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := RunEventStatusSyncOnceContext(context.Background(), db); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	for i, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got string
			if err := db.Table("events").Select("status").Where("id=?", i+1).Scan(&got).Error; err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Fatalf("status=%s want=%s", got, tc.want)
			}
		})
	}
	var before, after []struct {
		ID        uint
		Status    string
		UpdatedAt *time.Time
	}
	if err := db.Table("events").Order("id").Find(&before).Error; err != nil {
		t.Fatal(err)
	}
	if err := RunEventStatusSyncOnceContext(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	if err := db.Table("events").Order("id").Find(&after).Error; err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatal("repeat run changed completed transitions")
	}
}

func timePtr(t time.Time) *time.Time { return &t }
