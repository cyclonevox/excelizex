package testdata_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/cyclonevox/excelizex/v2/e2e/fixture"
)

var rewriteFixtures = flag.Bool("rewrite", false, "regenerate committed .xlsx fixtures in this directory")

func testdataDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve testdata directory")
	}

	return filepath.Dir(file)
}

func writeFixture(t *testing.T, dir, name string, buf *bytes.Buffer) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRewriteFixtures(t *testing.T) {
	if os.Getenv("EXCELIZEX_REWRITE_FIXTURES") != "1" && !*rewriteFixtures {
		t.Skip("skipped; run with -rewrite or EXCELIZEX_REWRITE_FIXTURES=1 to regenerate fixtures")
	}

	dir := testdataDir(t)
	joined := fixture.ParseJoined("2024-09-01")

	fixtures := map[string]*bytes.Buffer{
		"students_notice_ok.xlsx": fixture.BuildDirtyNoticeImport(t, [][]string{
			{"张三", "110101199001011234", "18", "A"},
			{"李四", "110101199002021234", "20", "B"},
		}),
		"students_notice_empty_name.xlsx": fixture.BuildDirtyNoticeImport(t, [][]string{
			{"张三", "110101199001011234", "18", "A"},
			{"", "110101199002021234", "20", "B"},
		}),
		"students_notice_partial_fail.xlsx": fixture.BuildDirtyNoticeImport(t, [][]string{
			{"张三", "110101199001011234", "18", "A"},
			{"", "110101199002021234", "19", "A"},
			{"王五", "110101199003031234", "bad", "A"},
			{"赵六", "110101199004041234", "22", "Z"},
		}),
		"students_notice_bad_type.xlsx": fixture.BuildDirtyNoticeImport(t, [][]string{
			{"张三", "", "not-int", "A"},
		}),
		"students_notice_failfast.xlsx": fixture.BuildDirtyNoticeImport(t, [][]string{
			{"张三", "", "18", "A"},
			{"李四", "", "bad-age", "A"},
			{"王五", "", "20", "A"},
		}),
		"students_notice_service_callback.xlsx": fixture.BuildDirtyNoticeImport(t, [][]string{
			{"张三", "", "18", "A"},
			{"李四", "", "20", "B"},
			{"重复考生", "", "21", "A"},
			{"王五", "", "22", "B"},
		}),
		"students_notice_fixed.xlsx": fixture.BuildDirtyNoticeImport(t, [][]string{
			{"张三", "110101199001011234", "18", "A"},
			{"钱七", "110101199005051234", "21", "B"},
		}),
		"students_reordered.xlsx":      fixture.BuildReorderedHeadersFile(t),
		"students_missing_column.xlsx": fixture.BuildMissingColumnFile(t),
		"students_empty_rows.xlsx":     fixture.BuildDirtyEmptyRowsFile(t),
		"students_notice_legacy.xlsx":  fixture.BuildLegacyNoticeStudents(t),
		"students_header_legacy.xlsx":  fixture.BuildLegacyHeaderStudents(t),
		"inline_address.xlsx":          fixture.BuildInlineAddressFile(t),
		"time_bool.xlsx": fixture.WriteTimeBoolSheet(t,
			fixture.TimeBoolRow{Name: "张三", Active: true, Joined: joined},
			fixture.TimeBoolRow{Name: "李四", Active: false, Joined: joined},
		),
		"scores_header.xlsx": fixture.WriteHeaderDataScores(t,
			fixture.ScoreRow{Name: "李四", Score: 95},
			fixture.ScoreRow{Name: "王五", Score: 88},
		),
	}

	for name, buf := range fixtures {
		writeFixture(t, dir, name, buf)
	}
}
