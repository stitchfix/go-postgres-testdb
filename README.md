# go-postgres-test
[![Circle CI](https://circleci.com/gh/stitchfix/go-postgres-testdb.svg?style=shield)](https://circleci.com/gh/stitchfix/go-postgres-testdb)

[![Go Report Card](https://goreportcard.com/badge/github.com/stitchfix/go-postgres-testdb)](https://goreportcard.com/report/github.com/stitchfix/go-postgres-testdb)

A helper library for managing ephemeral test databases in Postgres.

It won't do anything if Postgres is not installed, but if the binaries are available, this library will allow you to create an ephemeral database, use it for the life of a test, and then clean up afterwards.

It was created cos there's currently no 'in memory' postgres clone that can be used for testing.

It's intended to be used on a laptop, or within an ephemeral container, not anyplace where a real production postgres instance is being used.


In a test, set up the following:

        var tempDir string
        var dbPid int
        var dbDir string
        var dbName string

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

            dbName = "fargle"
            dbDir = fmt.Sprintf("%s/%s", tempDir, dbName)

            dbPid, dbPort, err = StartTestDB(dbDir, dbName)
            if err != nil {
                fmt.Printf("Failed to start test db %q: %s", dbName, err)
            }
            
            running, err := PostgresRunning(dbPort)
            if err != nil {
                fmt.Printf("Error Checking to see if postgres is running: %s", err)
                os.Exit(1
            }
            if running {
                fmt.Printf("Postgres is running with pid %d on port %d", dbPid, dbPort)
            }
        }

        func tearDown() {
            err := StopPostgres(dbPid)
            if err != nil {
                fmt.Printf("Failed to stop postgres process %d", dbPid)
            } else {
                fmt.Println("database shut down.")
            }

            if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
                os.Remove(tempDir)
            }

        }
