package vrptw

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// LoadDataset reads a Solomon VRPTW instance file.
// Solomon format:
//   Line 1: Name
//   Lines 2-4: blank/headers
//   Line 5: NUMBER CAPACITY
//   Lines 6-8: blank/headers
//   Line 9+: CUST_NO  XCOORD  YCOORD  DEMAND  READY_TIME  DUE_DATE  SERVICE_TIME
//   Customer 0 is the depot.
func LoadDataset(path string) (*Dataset, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	return ParseSolomon(lines)
}

// ParseSolomon parses Solomon format lines into a Dataset.
func ParseSolomon(lines []string) (*Dataset, error) {
	if len(lines) < 10 {
		return nil, fmt.Errorf("insufficient lines in Solomon file: got %d", len(lines))
	}

	ds := &Dataset{}
	ds.Name = strings.TrimSpace(lines[0])

	// Find the VEHICLE section: line with NUMBER and CAPACITY values.
	vehicleLine := findDataLine(lines, 4, 6)
	if vehicleLine == "" {
		return nil, fmt.Errorf("could not find vehicle data line")
	}
	vFields := strings.Fields(vehicleLine)
	if len(vFields) >= 2 {
		ds.Vehicles, _ = strconv.Atoi(vFields[0])
		ds.Capacity, _ = strconv.Atoi(vFields[1])
	}

	// Find customer data: starts after the CUSTOMER header section.
	// Look for the first line with numeric data after line 7.
	dataStart := -1
	for i := 7; i < len(lines); i++ {
		fields := strings.Fields(strings.TrimSpace(lines[i]))
		if len(fields) >= 7 {
			if _, err := strconv.Atoi(fields[0]); err == nil {
				dataStart = i
				break
			}
		}
	}
	if dataStart < 0 {
		return nil, fmt.Errorf("could not find customer data section")
	}

	// Parse customer lines. Customer 0 is the depot.
	for i := dataStart; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 7 {
			continue
		}

		id, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		x, _ := strconv.ParseFloat(fields[1], 64)
		y, _ := strconv.ParseFloat(fields[2], 64)
		demand, _ := strconv.Atoi(fields[3])
		readyTime, _ := strconv.Atoi(fields[4])
		dueDate, _ := strconv.Atoi(fields[5])
		service, _ := strconv.Atoi(fields[6])

		if id == 0 {
			// Depot.
			ds.Depot = Depot{
				ID:        id,
				X:         x,
				Y:         y,
				ReadyTime: readyTime,
				DueDate:   dueDate,
				Service:   service,
			}
		} else {
			ds.Customers = append(ds.Customers, Customer{
				ID:        id,
				X:         x,
				Y:         y,
				Demand:    demand,
				ReadyTime: readyTime,
				DueDate:   dueDate,
				Service:   service,
			})
		}
	}

	ds.Dimension = 1 + len(ds.Customers)

	if ds.Capacity == 0 {
		return nil, fmt.Errorf("no capacity specified")
	}
	if len(ds.Customers) == 0 {
		return nil, fmt.Errorf("no customers found")
	}

	return ds, nil
}

// findDataLine searches for the first non-blank line with numeric content
// in the given range [start, end).
func findDataLine(lines []string, start, end int) string {
	if end > len(lines) {
		end = len(lines)
	}
	for i := start; i < end; i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 1 {
			if _, err := strconv.Atoi(fields[0]); err == nil {
				return line
			}
		}
	}
	return ""
}
