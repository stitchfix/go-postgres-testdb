package testdb

import (
	"fmt"
	"github.com/phayes/freeport"
	"github.com/pkg/errors"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// PostgresInstalled Tests to see whether Postgres binaries are installed and available to the command shell
func PostgresInstalled() (missing []string, ok bool) {
	executables := []string{
		"postgres",
		"initdb",
		"createdb",
		"dropdb",
		"psql",
	}

	missing = make([]string, 0)

	for _, x := range executables {
		_, err := exec.LookPath(x)
		if err != nil {
			missing = append(missing, x)
		}
	}

	if len(missing) > 0 {
		return missing, false
	}

	return missing, true

}

// PostgresRunning Detect if postgres process is currently running
func PostgresRunning(port int) (running bool, err error) {
	if runtime.GOOS == "darwin" {
		// lsof -n | grep PGSQL | awk '{ print $8 }' | cut -d '.' -f4
		cmd := exec.Command("bash", "-c", "lsof -n | grep PGSQL | awk '{ print $8 }' | cut -d '.' -f4")

		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error running command: %s\n", err)
			return running, err
		}

		lines := strings.Split(string(output), "\n")

		if StringInSlice(strconv.Itoa(port), lines) {
			running = true
			return running, err
		}

	} else if runtime.GOOS == "linux" {
		// sudo netstat -pntl | grep postgres | grep 0.0.0.0 | awk '{print $4}' | cut -d ':' -f 2
		cmd := exec.Command("bash", "-c", "netstat -pntl | grep postgres | grep 0.0.0.0 | awk '{print $4}' | cut -d ':' -f2")

		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error running command: %s\n", err)
			return running, err
		}

		lines := strings.Split(string(output), "\n")

		if StringInSlice(strconv.Itoa(port), lines) {
			running = true
			return running, err
		}

	} else {
		err = errors.New(fmt.Sprintf("Unsuppported OS: %s", runtime.GOOS))
		return running, err
	}

	return running, err
}

// InitDbDir Takes <dir> as an argument and runs initdb on that dir.
// Returns success or failure and errors if present
func InitDbDir(dir string) (err error) {
	path, err := exec.LookPath("initdb")
	if err != nil {
		return err
	}

	cmd := exec.Command(path, dir)

	err = cmd.Run()

	if err != nil {
		return err
	}

	return err
}

// StartPostgres Starts Postgres.  Keep track of whether we started it ourselves, and remember to stop it
func StartPostgres(dbDir string, port int) (pid int, err error) {
	path, err := exec.LookPath("postgres")
	if err != nil {
		return pid, err
	}

	cmd := exec.Command(path, "-D", dbDir, "-p", strconv.Itoa(port))

	err = cmd.Start()
	if err != nil {
		return pid, err
	}

	pid = cmd.Process.Pid

	return pid, err
}

// StopPostgres Stops the Postgres processes with pid of <pid>
func StopPostgres(pid int) (err error) {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	err = p.Kill()
	if err != nil {
		fmt.Printf("Kill Error: %s", err)
		return err
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return err

	}

	// we don't really care about the result, we just want to call wait so that things clean up

	proc.Wait()

	return err
}

// CreateTestDb Creates a database of the name given
func CreateTestDb(dbName string, port int) (err error) {
	path, err := exec.LookPath("createdb")
	if err != nil {
		fmt.Printf("Command createdb doesn't exist")
		return err
	}

	cmd := exec.Command(path, "-p", strconv.Itoa(port), dbName)

	output, err := cmd.Output()

	if err != nil {
		fmt.Printf("Db Creation Output: %q\n", output)

		return err
	}

	return err
}

// CreateTestDbUser  Creates a user in the test db.
func CreateTestDbUser(userName string, port int) (err error) {
	path, err := exec.LookPath("createuser")
	if err != nil {
		fmt.Printf("Command createuser doesn't exist")
		return err
	}

	cmd := exec.Command(path, "-p", strconv.Itoa(port), userName)

	output, err := cmd.Output()

	if err != nil {
		fmt.Printf("User Creation Output: %q\n", output)

		return err
	}

	return err

}

// DbExists Checks whether a database of the name given exists
func DbExists(dbName string, port int) (exists bool, err error) {
	path, err := exec.LookPath("psql")
	if err != nil {
		return exists, err
	}
	cmd := exec.Command(path, "-ltq", dbName, "-p", strconv.Itoa(port))

	outputBytes, err := cmd.Output()

	if err != nil {

		return exists, err
	}

	dbNames := make([]string, 0)

	lines := strings.Split(string(outputBytes), "\n")

	for _, line := range lines {
		parts := strings.Split(line, "|")

		name := strings.TrimSpace(parts[0])

		if name != "" {
			dbNames = append(dbNames, name)
		}
	}

	exists = StringInSlice(dbName, dbNames)

	return exists, err
}

// StartTestDB Convenience function that performs checks and starts db, creates the db if it doesn't exist and returns the pid and any errors
func StartTestDB(dbDir string, dbName string) (pid int, port int, err error) {
	port, err = freeport.GetFreePort()
	if err != nil {
		return pid, port, err
	}

	running, err := PostgresRunning(port)

	if err != nil {
		return pid, port, err
	}

	if !running {
		err = InitDbDir(dbDir)
		if err != nil {
			return pid, port, err
		}

		pid, err = StartPostgres(dbDir, port)

		if err != nil {
			return pid, port, err
		}

		//give postgres a couple seconds to come up before we check it and try to create databases
		time.Sleep(5 * time.Second)

		ok, err := PostgresRunning(port)

		if err != nil {
			return pid, port, err
		}

		if !ok {
			err = errors.New("Postgres failed to start.")

			return pid, port, err
		}

		err = CreateTestDb(dbName, port)

		if err != nil {
			return pid, port, err
		}

		ok, err = DbExists(dbName, port)

		if err != nil {
			return pid, port, err
		}

		err = CreateTestDbUser(dbName, port)
		if err != nil {
			return pid, port, err
		}

		if !ok {
			err = errors.New("Testdb failed to create.")
			return pid, port, err
		}

		return pid, port, err

	}

	err = errors.New("Postgres is already running.")
	return pid, port, err
}

// StringInSlice Checks to see if string given is in the slice given.
func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
