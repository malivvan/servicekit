// Package servicekit implements a service wrapper providing a cli for service management.
package servicekit

import (
	"github.com/kardianos/service"
	"os"
	"bufio"
	"runtime"
	"path/filepath"
	"strings"
	"io"
	"os/user"
	"errors"
	"strconv"
	"fmt"
)

type Config struct {
	Name        string
	DisplayName string
	Description string
	Version     string
}

type Interface interface {
	Start(workdir Workdir) error
	Stop() error
}

type wrapper struct {
	workdir Workdir
	program Interface
}

func (w *wrapper) Start(s service.Service) error {
	return w.program.Start(w.workdir)
}

func (w *wrapper) Stop(s service.Service) error {
	return w.program.Stop()
}

func Wrap(config Config, program Interface) {
	serviceConfig, err := config.evaluate()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	serviceWrapper := &wrapper{program: program}
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "version":
			fmt.Fprintln(os.Stdout, config.Name, "v"+config.Version)
			os.Exit(0)

		case "run":

			// ensure working directory is define and exists
			if len(os.Args) < 3 {
				fmt.Fprintln(os.Stderr, "error: missing working directory")
				os.Exit(1)
			}
			err := os.MkdirAll(os.Args[2], 0700)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error creating workdir:", err)
				os.Exit(1)
			}

			// create and run service
			serviceWrapper.workdir = Workdir(os.Args[2])
			serviceConfig.Arguments = []string{"run", os.Args[2]}
			s, err := service.New(serviceWrapper, serviceConfig)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error creating service:", err)
				os.Exit(1)
			}
			err = s.Run()
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}
			os.Exit(0)

		case "install":

			// read missing parameters and validate
			username := ""
			workdir := ""
			defaultWorkdir := getDefaultWorkdir(config.Name)
			if runtime.GOOS == "linux" {
				username, err = userInput("Username", 2, "")
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				workdir, err = userInput("Working Directory", 3, defaultWorkdir)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			} else {
				workdir, err = userInput("Working Directory", 2, defaultWorkdir)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			}
			if !filepath.IsAbs(workdir) {
				fmt.Fprintln(os.Stderr, "error: working directory must be absolute path")
				os.Exit(1)
			}
			if username != "" {
				if _, err := user.Lookup(username); err != nil {
					fmt.Fprintln(os.Stderr, "error: user", username, "does not exist")
					os.Exit(1)
				}
			}

			// define install path
			installPath := filepath.Join(workdir, config.Name)
			if runtime.GOOS == "windows" {
				installPath += ".exe"
			}

			// define executable name and fix case where windows cmd call omits .exe suffix
			executableName := os.Args[0]
			if runtime.GOOS == "windows" && !strings.HasSuffix(executableName, ".exe") {
				executableName += ".exe"
			}

			// install service first - installation will fail if user has
			// insufficient permissions without altering the filesystem
			serviceConfig.Executable = installPath
			serviceConfig.Arguments = []string{"run", workdir}
			serviceConfig.UserName = username
			s, err := service.New(serviceWrapper, serviceConfig)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error installing service:", err)
				os.Exit(1)
			}
			err = s.Install()
			if err != nil {
				fmt.Fprintln(os.Stderr, "error installing service:", err)
				os.Exit(1)
			}

			// ensure working directory initialized
			err = os.MkdirAll(workdir, 0700)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error creating working directory:", err)
				s.Uninstall()
				os.Exit(1)
			}

			// ensure empty install path
			isEmpty, err := ensureEmptyInstallPath(installPath)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				s.Uninstall()
				os.Exit(1)
			}
			if !isEmpty {
				fmt.Fprintln(os.Stdout, "Service installation aborted!")
				s.Uninstall()
				os.Exit(0)
			}

			// copy currently running executable
			err = copyFile(installPath, executableName)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error installing executable:", err)
				s.Uninstall()
				os.Exit(1)
			}

			// give folder and executable ownership to specified user
			if username != "" {
				err = chown(username, installPath)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					s.Uninstall()
					os.Exit(1)
				}
				err = chown(username, workdir)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					s.Uninstall()
					os.Exit(1)
				}
			}

			fmt.Fprintln(os.Stdout, "Service", config.Name, "installed!")
			os.Exit(0)
		case "uninstall":
			s, err := service.New(serviceWrapper, serviceConfig)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error creating service:", err)
				os.Exit(1)
			}

			// ensure stopped
			s.Stop()

			err = s.Uninstall()
			if err != nil {
				fmt.Fprintln(os.Stderr, "error uninstalling service:", err)
				os.Exit(1)
			}

			fmt.Fprintln(os.Stdout, "Service", config.Name, "uninstalled!")
			os.Exit(0)
		case "start", "stop", "restart":
			s, err := service.New(serviceWrapper, serviceConfig)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error creating service:", err)
				os.Exit(1)
			}
			err = service.Control(s, os.Args[1])
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				os.Exit(1)
			}

			fmt.Fprintln(os.Stdout, "Service", config.Name, os.Args[1]+"ed!")
			os.Exit(0)
		}
	}

	usage(config)
	os.Exit(0)
}

func usage(config Config) {
	fmt.Fprintln(os.Stdout, "SERVICE", config.Name, "v"+config.Version)
	fmt.Fprintln(os.Stdout)
	if config.Description != "" {
		cursor := 0
		for _, word := range strings.Fields(config.Description) {
			if cursor+len(word) > 60 {
				cursor = 0
				fmt.Fprintln(os.Stdout)
			}
			fmt.Fprint(os.Stdout, word)
			fmt.Fprint(os.Stdout, " ")
			cursor += len(word) + 1
		}
		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout)
	}
	fmt.Fprintln(os.Stdout, "Available Commands:")
	fmt.Fprintln(os.Stdout, "  run [workdir]")
	if runtime.GOOS == "linux" {
		fmt.Fprintln(os.Stdout, "  install (user) (workdir)")
	} else {
		fmt.Fprintln(os.Stdout, "  install (workdir)")
	}
	fmt.Fprintln(os.Stdout, "  uninstall")
	fmt.Fprintln(os.Stdout, "  start")
	fmt.Fprintln(os.Stdout, "  stop")
	fmt.Fprintln(os.Stdout, "  restart")
	fmt.Fprintln(os.Stdout, "  version")
}

func (config *Config) evaluate() (*service.Config, error) {
	if config.Name == "" {
		return nil, errors.New("service name is empty")
	}
	if strings.Contains(config.Name, " ") {
		return nil, errors.New("service name cannot contain whitespaces")
	}
	if config.Version == "" {
		return nil, errors.New("service version is empty")
	}
	if config.DisplayName == "" {
		config.DisplayName = config.Name
	}
	return &service.Config{
		Name:        config.Name,
		DisplayName: config.DisplayName,
		Description: config.Description,
	}, nil
}

func userInput(prompt string, argIndex int, defaultValue string) (string, error) {
	var value string
	if len(os.Args) > argIndex {
		value = os.Args[argIndex]
		fmt.Fprintln(os.Stdout, prompt+":", value)
	} else if defaultValue != "" {
		value = readline(prompt + " [" + defaultValue + "]: ")
		if value == "" {
			value = defaultValue
		}
	} else {
		value = readline(prompt + ": ")
		if value == "" {
			return "", errors.New("error: invalid user input")
		}
	}
	return value, nil
}

func readline(prompt string) string {
	fmt.Fprint(os.Stdout, prompt)
	value, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	value = strings.TrimSuffix(value, "\n")
	return strings.TrimSuffix(value, "\r")
}

func ensureEmptyInstallPath(path string) (bool, error) {
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			return false, errors.New("error: directory exist at " + path)
		}
		if readline("Overwrite file at "+path+"? (y/n) ") != "y" {
			return false, nil
		}
		err = os.Remove(path)
		if err != nil {
			return false, errors.New("error clearing install path: " + err.Error())
		}
	}
	return true, nil
}

func copyFile(dst, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return nil
}

func chown(name string, path string) error {
	u, err := user.Lookup(name)
	if err != nil {
		return errors.New("error: user '" + name + "' does not exist")
	}
	uid, err1 := strconv.Atoi(u.Uid)
	gid, err2 := strconv.Atoi(u.Gid)
	if err1 != nil || err2 != nil || os.Chown(path, uid, gid) != nil {
		return errors.New("error: taking ownership of '" + path + "' to '" + name + "' failed")
	}
	return nil
}

var defaultWorkdir = map[string]string{
	"linux":   "/opt/",
	"windows": "C:\\",
}

func getDefaultWorkdir(name string) string {
	if defaultWorkdir, exist := defaultWorkdir[runtime.GOOS]; exist {
		return filepath.Join(defaultWorkdir, name)
	}
	return ""
}
