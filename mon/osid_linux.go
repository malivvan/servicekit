//go:build linux
// +build linux

package mon

func getOSID() (string, error) {
	id, err := readFile("/var/lib/dbus/machine-id")
	if err != nil {
		// fallback
		id, err = readFile("/etc/machine-id")
	}
	if err != nil {
		return "", err
	}
	return trim(string(id)), nil
}
