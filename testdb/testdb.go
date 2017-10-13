package testdb

import (
	"errors"
	"fmt"
	"github.com/mitchellh/go-ps"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

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

// Detect if postgres process is running
func PostgresRunning() (running bool, err error) {
	processes, err := ps.Processes()
	if err != nil {
		return running, err
	}

	for _, p := range processes {
		if p.Executable() == "postgres" {

			pcall := syscall.Kill(p.Pid(), syscall.Signal(0))

			running = pcall == nil
		}
	}

	return running, err
}

// Takes <dir> as an argument and runs initdb on that dir.
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

// start postgres.  Keep track of whether we started it ourselves, and remember to stop it
func StartPostgres(dbDir string) (pid int, err error) {
	path, err := exec.LookPath("postgres")
	if err != nil {
		return pid, err
	}

	cmd := exec.Command(path, "-D", dbDir)

	err = cmd.Start()
	if err != nil {
		return pid, err
	}

	pid = cmd.Process.Pid

	return pid, err
}

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

func CreateTestDb(dbName string) (err error) {
	path, err := exec.LookPath("createdb")
	if err != nil {
		fmt.Printf("Command createdb doesn't exist")
		return err
	}

	cmd := exec.Command(path, dbName)

	output, err := cmd.Output()

	if err != nil {
		fmt.Printf("Db Creation Output: %q\n", output)

		return err
	}

	return err
}

func DbExists(dbName string) (exists bool, err error) {
	path, err := exec.LookPath("psql")
	if err != nil {
		return exists, err
	}
	cmd := exec.Command(path, "-ltq", dbName)

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

func StartTestDB(dbDir string, dbName string) (pid int, err error) {
	running, err := PostgresRunning()

	if err != nil {
		return pid, err
	}

	if !running {
		err = InitDbDir(dbDir)
		if err != nil {
			return pid, err
		}

		pid, err = StartPostgres(dbDir)

		if err != nil {
			return pid, err
		}

		//give postgres a couple seconds to come up before we check it and try to create databases
		time.Sleep(5 * time.Second)

		ok, err := PostgresRunning()

		if err != nil {
			return pid, err
		}

		if !ok {
			err = errors.New("Postgres failed to start.")

			return pid, err
		}

		err = CreateTestDb(dbName)

		if err != nil {
			return pid, err
		}

		ok, err = DbExists(dbName)

		if err != nil {
			return pid, err
		}

		if !ok {
			err = errors.New("Testdb failed to create.")
			return pid, err
		}

		return pid, err

	} else {
		err = errors.New("Postgres is already running.")
		return pid, err
	}
}

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
