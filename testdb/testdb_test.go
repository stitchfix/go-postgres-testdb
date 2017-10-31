package testdb

import (
	"fmt"
	"github.com/phayes/freeport"
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
		fmt.Printf("Error creating temp dir %q: %s\n", tempDir, err)
		os.Exit(1)
	}

	tempDir = dir
}

func tearDown() {
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		os.Remove(tempDir)
	}

}

func TestStringInSlice(t *testing.T) {
	trueList := []string{"foo", "bar", "baz"}
	falseList := []string{"wip", "zoz", "woo"}

	assert.True(t, StringInSlice("bar", trueList), "Expected string is in list")
	assert.False(t, StringInSlice("bar", falseList), "Unexpected string is in list")
}

func TestPostgresInstalled(t *testing.T) {
	missing, ok := PostgresInstalled()

	if !ok {
		fmt.Printf("Missing Postgres executables: %s\n", missing)
		t.Fail()
	}
}

func TestInitDbDir(t *testing.T) {
	dbdir := fmt.Sprintf("%s/%s", tempDir, "db1")
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		err = os.Mkdir(dbdir, 0755)
		if err != nil {
			fmt.Printf("Error creating db dir %q: %s", dbdir, err)
			t.Fail()
		}
	}
	err := InitDbDir(dbdir)
	if err != nil {
		fmt.Printf("Error initializing db dir %s : %s\n", dbdir, err)
		t.Fail()
	}

	os.Remove(dbdir)
}

func TestPostgresRunning(t *testing.T) {
	ok, err := PostgresRunning(5432)
	if err != nil {
		fmt.Printf("Error checking for postgres process: %s\n", err)
		t.Fail()
	}

	if ok {
		fmt.Println("Postgres is running on standard port 5432.")
	} else {
		fmt.Println("Postgres is not running on standard port 5432.")
	}
}

func TestStartStopPostgres(t *testing.T) {
	dbName := "testing"
	dbDir := fmt.Sprintf("%s/%s", tempDir, dbName)

	port, err := freeport.GetFreePort()
	if err != nil {
		fmt.Printf("Error finding a free port to run postgres upon: %s\n", err)
		t.Fail()
	}

	running, err := PostgresRunning(port)

	if err != nil {
		fmt.Printf("Error checking to see if Postgres is running: %s\n", err)
		t.Fail()
	}

	if !running {

		err = InitDbDir(dbDir)
		if err != nil {
			fmt.Printf("Error initialiing db dir %q: %s\n", tempDir, err)
			t.Fail()
		}

		fmt.Println("Starting Postgres.")

		pid, err := StartPostgres(dbDir, port)

		if err != nil {
			fmt.Printf("Error starting postgres: %s\n", err)
			t.Fail()
		}

		fmt.Println("Success!")

		fmt.Printf("Postgres running with pid %d on port %d\n", pid, port)

		//give postgres a couple seconds to come up before we check it and try to create databases
		time.Sleep(5 * time.Second)

		running, err = PostgresRunning(port)

		if err != nil {
			fmt.Printf("Error checking to see if Postgres is running: %s\n", err)
			t.Fail()
		}

		assert.True(t, running, fmt.Sprintf("Postges is not demonstrably running on port %d.", port))

		err = CreateTestDb(dbName, port)

		if err != nil {
			fmt.Printf("Error creating testdb %q: %s\n", dbName, err)
			t.Fail()
		}

		exists, err := DbExists(dbName, port)

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

		running, err = PostgresRunning(port)

		if err != nil {
			fmt.Printf("Error checking to see if Postgres is running: %s\n", err)
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

	pid, port, err := StartTestDB(dbDir, dbName)
	if err != nil {
		fmt.Printf("Failed to start test db %q: %s\n", dbName, err)
	}

	running, err := PostgresRunning(port)

	if err != nil {
		fmt.Printf("Error checking to see if Postgres is running: %s\n", err)
		t.Fail()
	}

	if !running {
		fmt.Println("Starting Postgres.")

		fmt.Println("Success!")

		fmt.Printf("Postgres running with pid %d on port %d\n", pid, port)

		err = StopPostgres(pid)
		if err != nil {
			fmt.Printf("Failed to stop postgres process %d\n", pid)
			t.Fail()
		}

		fmt.Println("done.")

	} else {
		fmt.Println("Test cannot run since postgres is already running.")
	}

	os.Remove(dbDir)

}
