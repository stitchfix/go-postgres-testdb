# go-postgres-test
[![Circle CI](https://circleci.com/gh/stitchfix/go-postgres-testdb.svg?style=shield)](https://circleci.com/gh/stitchfix/go-postgres-testdb)

[![Go Report Card](https://goreportcard.com/badge/github.com/stitchfix/go-postgres-testdb)](https://goreportcard.com/report/github.com/stitchfix/go-postgres-testdb)

A helper library for managing ephemeral test databases in Postgres


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

            running, err := PostgresRunning()
            if err != nil {
                fmt.Printf("Error checking to see if Postgres is running: %s", err)
                os.Exit(1)
            }

            if !running {
                dbPid, err = StartTestDB(dbDir, dbName)
                if err != nil {
                    fmt.Printf("Failed to start test db %q: %s", dbName, err)
                }
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
