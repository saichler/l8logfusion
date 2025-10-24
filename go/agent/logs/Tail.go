package logs

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"time"

	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8utils/go/utils/queues"
)

// TailFile tails a file and prints each new line.
// It handles file truncation by detecting when the file size becomes smaller.
// The location parameter specifies where to start tailing:
//   - location == 0: start from the beginning of the file
//   - location == -1: start from the end of the file (only new lines)
//   - location > 0: start from the specific byte offset (useful for resuming after crash)
func TailFile(filename string, nic ifs.IVNic, location int64) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get initial file size before seeking
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}
	initialSize := fileInfo.Size()

	// Seek to the appropriate position based on location parameter
	var lastSize int64
	if location == -1 {
		// Start from the end of the file (only new lines)
		_, err = file.Seek(0, io.SeekEnd)
		if err != nil {
			return fmt.Errorf("failed to seek to end: %w", err)
		}
		lastSize = initialSize
	} else if location == 0 {
		// Start from the beginning of the file
		_, err = file.Seek(0, io.SeekStart)
		if err != nil {
			return fmt.Errorf("failed to seek to start: %w", err)
		}
		lastSize = 0
	} else {
		// Start from a specific byte offset (resume after crash)
		_, err = file.Seek(location, io.SeekStart)
		if err != nil {
			return fmt.Errorf("failed to seek to position %d: %w", location, err)
		}
		lastSize = location
	}

	reader := bufio.NewReader(file)

	// Cooldown buffer configuration
	const cooldownDuration = 100 * time.Millisecond
	const maxBufferAge = 3 * time.Second
	const maxBufferLines = 1000 // Prevent OOM from massive log bursts

	lineBuffer := queues.NewQueue("LineBuffer", 30000)
	lastLineTime := time.Now() // Initialize to current time instead of zero
	var lastFlushTime atomic.Int64
	lastFlushTime.Store(time.Now().UnixNano())

	flushBuffer := func() {
		lines := lineBuffer.Clear()
		if len(lines) > 0 {
			localBuff := make([]string, len(lines))
			for i, line := range lines {
				localBuff[i] = line.(string)
			}
			SendLogs(filename, nic, localBuff...)
			lastFlushTime.Store(time.Now().UnixNano())
		}
	}

	for {
		// Check current file size
		fileInfo, err := file.Stat()
		if err != nil {
			return fmt.Errorf("failed to stat file: %w", err)
		}
		currentSize := fileInfo.Size()

		// Detect truncation: if current size is less than last known size
		if currentSize < lastSize {
			// Flush buffer before handling truncation
			go flushBuffer()

			// File was truncated, reopen and seek to beginning
			file.Close()
			file, err = os.Open(filename)
			if err != nil {
				return fmt.Errorf("failed to reopen file after truncation: %w", err)
			}

			reader = bufio.NewReader(file)
			lastSize = 0
		}

		// Read new lines
		hasNewLines := false
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					// No more data available
					break
				}
				return fmt.Errorf("error reading file: %w", err)
			}

			// Add line to buffer
			lineBuffer.Add(line)
			lastLineTime = time.Now()
			hasNewLines = true

			// Flush if buffer reaches max size to prevent OOM
			if lineBuffer.Size() >= maxBufferLines {
				go flushBuffer()
			}
		}

		// Update last known size
		fileInfo, err = file.Stat()
		if err != nil {
			return fmt.Errorf("failed to stat file: %w", err)
		}
		lastSize = fileInfo.Size()

		// Check flush conditions
		timeSinceLastLine := time.Since(lastLineTime)
		lastFlushTimeValue := time.Unix(0, lastFlushTime.Load())
		timeSinceLastFlush := time.Since(lastFlushTimeValue)

		if lineBuffer.Size() > 0 {
			// Flush if cooldown period passed without new lines
			// OR if max buffer age exceeded
			if (!hasNewLines && timeSinceLastLine >= cooldownDuration) ||
				timeSinceLastFlush >= maxBufferAge {
				go flushBuffer()
			}
		}

		// Brief sleep before next check
		time.Sleep(10 * time.Millisecond)
	}
}
