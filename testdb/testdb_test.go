package testdb

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

var tempDir string

func TestMain(m *testing.M) {
	setUp()

	code := m.Run()

	tearDown()

	os.Exit(code)
}

func setUp() {
	dir, err := ioutil.TempDir("", "testdb")
	if err != nil {
		fmt.Printf("Error creating temp dir %q: %s", tempDir, err)
		os.Exit(1)
	}

	tempDir = dir
}

func tearDown() {
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		os.Remove(tempDir)
	}

}

func TestPostgresInstalled(t *testing.T) {
	missing, ok := PostgresInstalled()

	if !ok {
		fmt.Printf("Missing Postgres executables:", missing)
	}
}

func TestPostgresRunning(t *testing.T) {
	ok, err := PostgresRunning()
	if err != nil {
		fmt.Printf("Error checking for postgres process: %s")
		t.Fail()
	}

	if ok {
		fmt.Println("Postgres IS running.")
	} else {
		fmt.Println("Postgres IS NOT running.")
	}
}

func TestInitDbDir(t *testing.T) {
	dbdir := fmt.Sprintf("%s/%s", tempDir, "db1")
	err := InitDbDir(dbdir)
	if err != nil {
		fmt.Printf("Error initialiing db dir %q: %s", tempDir, err)
		t.Fail()
	} else {
		fmt.Println("Db dir successfully initialized.")
	}

	os.Remove(dbdir)

}

func TestStringInSlice(t *testing.T) {
	trueList := []string{"foo", "bar", "baz"}
	falseList := []string{"wip", "zoz", "woo"}

	assert.True(t, StringInSlice("bar", trueList), "Expected string is in list")
	assert.False(t, StringInSlice("bar", falseList), "Unexpected string is in list")
}

func TestStartStopPostgres(t *testing.T) {
	dbName := "testing"
	dbDir := fmt.Sprintf("%s/%s", tempDir, dbName)

	running, err := PostgresRunning()

	if err != nil {
		fmt.Printf("Error checking to see if Postgres is running: %s", err)
		t.Fail()
	}

	if !running {

		err = InitDbDir(dbDir)
		if err != nil {
			fmt.Printf("Error initialiing db dir %q: %s", tempDir, err)
			t.Fail()
		}

		fmt.Println("Starting Postgres.")

		pid, err := StartPostgres(dbDir)

		if err != nil {
			fmt.Printf("Error starting postgres: %s", err)
			t.Fail()
		}

		fmt.Println("Success!")

		fmt.Printf("Postgres running with pid %d\n", pid)

		//give postgres a couple seconds to come up before we check it and try to create databases
		time.Sleep(5 * time.Second)

		running, err = PostgresRunning()

		if err != nil {
			fmt.Printf("Error checking to see if Postgres is running: %s", err)
			t.Fail()
		}

		assert.True(t, running, "Postges is not demonstrably running.")

		err = CreateTestDb(dbName)

		if err != nil {
			fmt.Printf("Error creating testdb %q: %s", dbName, err)
			t.Fail()
		}

		exists, err := DbExists(dbName)

		if err != nil {
			fmt.Printf("Error testing to see if db %q exists: %s\n", dbName, err)
			t.Fail()
		}

		assert.True(t, exists, "Test db does not exist")

		fmt.Println("Stopping postgres.")
		err = StopPostgres(pid)

		if err != nil {
			fmt.Printf("Failed to stop postgres process %d\n", pid)
			t.Fail()
		}

		time.Sleep(5 * time.Second)

		running, err = PostgresRunning()

		if err != nil {
			fmt.Printf("Error checking to see if Postgres is running: %s", err)
			t.Fail()
		}

		if !running {
			fmt.Println("Postgres stopped.")
		} else {
			fmt.Println("Postgres failed to stop.")
			t.Fail()
		}

		os.Remove(dbDir)

	} else {
		fmt.Println("Test cannot run since postgres is already running.")
	}

}

func TestStartTestDB(t *testing.T) {
	fmt.Println("Running TestStartTestDb")
	dbName := "fargle"
	dbDir := fmt.Sprintf("%s/%s", tempDir, dbName)

	running, err := PostgresRunning()

	if err != nil {
		fmt.Printf("Error checking to see if Postgres is running: %s", err)
		t.Fail()
	}

	if !running {
		fmt.Println("Starting Postgres.")
		pid, err := StartTestDB(dbDir, dbName)
		if err != nil {
			fmt.Printf("Failed to start test db %q: %s", dbName, err)
		}

		fmt.Println("Success!")

		fmt.Printf("Postgres running with pid %d\n", pid)

		err = StopPostgres(pid)
		if err != nil {
			fmt.Printf("Failed to stop postgres process %d", pid)
			t.Fail()
		}

		fmt.Println("done.")

	} else {
		fmt.Println("Test cannot run since postgres is already running.")
	}

	os.Remove(dbDir)

}
