package main

import (
	"bufio"
	context2 "context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"log"
	"os"
	"os/user"
	"strconv"
	"strings"
)

// convertString converts char string[] from eBPF to string in golang.
func convertString(int8Value []int8) string {
	stringValue := ""
	for _, val := range int8Value {
		stringValue += string(val)
	}
	return stringValue
}

// isNewline checks if a string is just an empty line with enter pressed
func isEmptyLine(str string) bool {
	sum := 0
	for _, val := range []byte(str) {
		sum = sum + int(val)
	}
	return sum == 0
}

// getProcName retrieves the name of the process by given PID.
func getProcName(pid int) string {
	// Convert PID to string
	pidStr := strconv.Itoa(pid)

	// Open /proc/<pid>/status file
	statusFile, err := os.Open("/proc/" + pidStr + "/status")
	if err != nil {
		return "UNKNOWN"
	}
	defer statusFile.Close()

	// Read file line by line
	scanner := bufio.NewScanner(statusFile)
	for scanner.Scan() {
		line := scanner.Text()

		// Find line that starts with "Name:"
		if len(line) > 5 && line[0:5] == "Name:" {
			// Extract process name from line
			name := line[6:]
			return name
		}
	}

	// Process name not found
	return "UNKNOWN"
}

// getUsername retrieves the string username by given UID.
func getUsername(uid int) string {
	u, err := user.LookupId(strconv.Itoa(int(uid)))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return "UNKNOWN"
	}
	return u.Username
}

// getProcInfo checks if a process has cgroup enabled inside.
// This function will check if the process had any system.slice and has "docker" keyword inside it.
// If the cgroup had "docker" keyword inside it, it will return the Docker ID with the tag.
func getProcInfo(pid int) string {
	// Convert PID to string
	pidStr := strconv.Itoa(pid)

	// Open /proc/<pid>/cgroup file
	statusFile, err := os.Open("/proc/" + pidStr + "/cgroup")
	if err != nil {
		return ""
	}
	defer statusFile.Close()

	// Read file line by line
	scanner := bufio.NewScanner(statusFile)
	for scanner.Scan() {
		line := scanner.Text()

		// Containers contain namespaces based upon .scope name.
		parts := strings.Split(line, "/")
		containerID := parts[len(parts)-1]
		if strings.Contains(containerID, "docker") {
			containerID = strings.Replace(containerID, ".scope", "", 1)
			return containerID
		}
	}

	// Process was native process.
	return ""
}

func retrieveK8sPod(containerID string) {

}

// handleContext handles the given context.
func handleContext(ctx context) {
	procName := getProcName(int(ctx.pid))
	username := getUsername(int(ctx.uid))
	containerName := getProcInfo(int(ctx.pid))

	// Print out the current event on log.
	if verbose {
		log.Printf("PID: %s(%d) by %s(%d) PPID: %s(%d): %s", procName, ctx.pid, username, ctx.uid, ctx.parentProcName, ctx.ppid, ctx.line)
		if containerName != "" {
			log.Printf(" - container: %s\n", containerName)
		} else {
			log.Printf(" - non-container\n")
		}
	}

	// Send data into API server.
	err := sendEvent(ctx, master+"/event_send")
	if err != nil {
		log.Printf("Could not send data to API server %v\n", err)
	}
}

// retrieveInode retrieves inode value from a filename.
func retrieveInode(filename string) int {
	info, err := os.Stat(filename)
	if err != nil {
		return 0
	}

	// This is stupid, but info.Sys does not have an interface that seems to be compatibable
	// This shall be converted into Stats_t struct for better performance.
	strStat := fmt.Sprintf("%v", info.Sys())
	strSlices := strings.Split(strStat, " ")
	inode := strSlices[1]

	num, err := strconv.Atoi(inode)
	if err != nil {
		return 0
	}

	return num
}

// intContains checks if a slice of strings contains a particular string.
func intContains(slice []int, target int) bool {
	for _, s := range slice {
		if target == s {
			return true
		}
	}
	return false
}

// getHostname retrieves the host name of this current machine.
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "UNKNOWN"
	}

	return hostname
}

// retrieveDockerBinBash retrieves all docker /bin/bash files in each container's overlayFS merged directory.
func retrieveDockerBinBash() []string {
	var ret []string
	ret = []string{}

	// Create a new Docker client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	// Get all containers.
	containers, err := cli.ContainerList(context2.Background(), types.ContainerListOptions{All: true})
	if err != nil {
		panic(err)
	}

	// For all containers, perform inspect and retrieve the merged overlayfs dir
	for _, container := range containers {
		containerJSON, err := cli.ContainerInspect(context2.Background(), container.ID)
		if err != nil {
			panic(err)
		}

		// Access the GraphDriver.Data.MergedDir field
		mergedDir := containerJSON.GraphDriver.Data["MergedDir"]
		bashFile := fmt.Sprintf("%s/bin/bash", mergedDir)
		_, err = os.Stat(mergedDir)
		if os.IsNotExist(err) {
			continue
		} else {
			ret = append(ret, bashFile)
		}
	}

	return ret
}
