package jobshop

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// LoadDataset reads a JSS instance in standard Taillard/OR-Library format.
//
// Format (standard):
//   Line 1: <jobs> <machines>
//   Lines 2..jobs+1: pairs of (machine, duration) for each operation in the job
//
// Example (3 jobs, 3 machines):
//   3 3
//   0 3  1 2  2 2
//   0 2  2 1  1 4
//   1 4  2 3  0 1
//
// First job: machine 0 for 3 time units, then machine 1 for 2, then machine 2 for 2.
func LoadDataset(path string) (*Dataset, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	// Skip comment lines (starting with #) and blank lines.
	var firstLine string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		firstLine = line
		break
	}

	// Parse dimensions.
	parts := strings.Fields(firstLine)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid header: expected '<jobs> <machines>', got: %s", firstLine)
	}
	numJobs, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid job count: %s", parts[0])
	}
	numMachines, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid machine count: %s", parts[1])
	}

	ds := &Dataset{
		Jobs:     numJobs,
		Machines: numMachines,
		JobList:  make([]Job, numJobs),
	}

	// Parse job lines.
	for j := 0; j < numJobs; j++ {
		if !scanner.Scan() {
			return nil, fmt.Errorf("unexpected end of file at job %d", j)
		}
		line := strings.TrimSpace(scanner.Text())
		// Skip comments within job data.
		for line == "" || strings.HasPrefix(line, "#") {
			if !scanner.Scan() {
				return nil, fmt.Errorf("unexpected end of file at job %d", j)
			}
			line = strings.TrimSpace(scanner.Text())
		}

		fields := strings.Fields(line)
		if len(fields) < numMachines*2 {
			return nil, fmt.Errorf("job %d: expected %d values (machine,duration pairs), got %d", j, numMachines*2, len(fields))
		}

		job := Job{ID: j, Operations: make([]Operation, numMachines)}
		for op := 0; op < numMachines; op++ {
			machine, err := strconv.Atoi(fields[op*2])
			if err != nil {
				return nil, fmt.Errorf("job %d op %d: invalid machine: %s", j, op, fields[op*2])
			}
			duration, err := strconv.Atoi(fields[op*2+1])
			if err != nil {
				return nil, fmt.Errorf("job %d op %d: invalid duration: %s", j, op, fields[op*2+1])
			}

			operation := Operation{
				JobID:   j,
				OpIndex: op,
				Machine: machine,
				Duration: duration,
			}
			job.Operations[op] = operation
			ds.AllOps = append(ds.AllOps, operation)
		}
		ds.JobList[j] = job
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading: %w", err)
	}

	// Derive name from filename.
	ds.Name = path

	return ds, nil
}
